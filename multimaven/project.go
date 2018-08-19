// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"strings"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/pom"
	"github.com/szabba/uninbuild/repo"
)

type Project struct {
	name     string
	version  string
	dir      string
	deps     []unibuild.ProjectInfo
	pomHeads []pom.Header
}

var _ unibuild.Project = Project{}

func NewProject(ctx context.Context, clone repo.Local) (Project, error) {
	pomHeads, err := pom.Scan(ctx, clone.Path)
	if err != nil {
		return Project{}, oops.Wrapf(err, "failed scanning for maven modules in %s", clone.Path)
	}

	prj := Project{
		name:     clone.Name,
		version:  pomHeads[0].EffectiveVersion(),
		dir:      clone.Path,
		pomHeads: pomHeads,
	}
	return prj, nil
}

func (prj Project) Info() unibuild.ProjectInfo {
	return unibuild.ProjectInfo{
		Name:    prj.name,
		Version: prj.version,
	}
}

func (prj Project) Deps() []unibuild.ProjectInfo { return prj.deps }

func (prj Project) MavenModuleHeaders() []pom.Header { return prj.pomHeads }

func (prj Project) WithDependnecies(ctx context.Context, providers map[pom.Identity]Project) (Project, error) {
	mvnDeps, err := prj.FindMavenDeps(ctx)
	if err != nil {
		return prj, oops.Wrapf(err, "problem resolving deps of %s", prj.name)
	}

	depSet := make(map[unibuild.ProjectInfo]bool, len(mvnDeps))
	for _, mvnDep := range mvnDeps {
		d, ok := providers[mvnDep]
		if !ok {
			continue
		}
		depSet[d.Info()] = true
	}

	out := prj
	out.deps = make([]unibuild.ProjectInfo, 0, len(depSet))
	for di := range depSet {
		out.deps = append(out.deps, di)
	}

	return out, nil
}

func (prj Project) Build(ctx context.Context, arts map[unibuild.ProjectInfo][]unibuild.Artifact, logTo io.Writer) ([]unibuild.Artifact, error) {
	cmd := exec.CommandContext(ctx, "mvn", "clean", "deploy")
	cmd.Dir = prj.dir
	cmd.Stdout = logTo
	cmd.Stderr = logTo
	err := cmd.Run()
	return nil, err
}

func (prj Project) FindMavenDeps(ctx context.Context) ([]pom.Identity, error) {
	allDeps := map[pom.Identity]bool{}
	for _, mod := range prj.MavenModuleHeaders() {

		ids, err := prj.listMavenDeps(ctx, mod)
		if err != nil {
			return nil, oops.Wrapf(err, "problem listing deps of project %s module %s", prj.name, mod.ArtifactID+":"+mod.GroupID)
		}
		for _, id := range ids {
			allDeps[id] = true
		}
	}

	ids := make([]pom.Identity, 0, len(allDeps))
	for id := range allDeps {
		ids = append(ids, id)
	}
	return ids, nil
}

func (prj Project) listMavenDeps(ctx context.Context, mod pom.Header) ([]pom.Identity, error) {
	modPrjName := mod.GroupID + ":" + mod.ArtifactID
	cmd := exec.CommandContext(ctx, "mvn", "dependency:list", "-pl", modPrjName)
	cmd.Dir = prj.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, oops.Wrapf(err, "cannot list dependencies project %s module %s", prj.Info().Name, modPrjName)
	}
	depIDs, err := prj.parseDepList(string(out))
	return depIDs, oops.Wrapf(err, "cannot parse dependency list of project %s module %s", prj.Info().Name, modPrjName)
}

const (
	_MavenOutputPrefix  = "[INFO]"
	_MavenDepListHeader = "The following files have been resolved:"

	_GroupIDIx    = 0
	_ArtifactIDIx = 1
	_VersionIx    = 3
	_DepChunks    = 5
)

var (
	errNoDepListHeader   = errors.New("could not find a dependency list header")
	errInvalidDepListing = errors.New("dependency list line invalid")
	errDepListOver       = errors.New("no more dependencies in list")

	newline = regexp.MustCompile("\n|\r\n")
)

func (prj Project) parseDepList(mvnOut string) ([]pom.Identity, error) {
	lines := newline.Split(mvnOut, -1)
	lines, err := prj.skipPastHeader(lines)
	if err != nil {
		return nil, err
	}
	return prj.extractIdentities(lines)
}

func (prj Project) skipPastHeader(lines []string) ([]string, error) {
	for len(lines) > 1 {
		l, rest := lines[0], lines[1:]
		if strings.Contains(l, _MavenDepListHeader) {
			return rest, nil
		}
		lines = rest
	}
	return lines, errNoDepListHeader
}

func (prj Project) extractIdentities(lines []string) ([]pom.Identity, error) {
	deps := []pom.Identity{}
	for len(lines) > 1 {
		l, rest := lines[0], lines[1:]
		d, err := prj.extractIdentity(l)
		if err == errDepListOver {
			break
		}
		if err != nil {
			return nil, err
		}
		deps = append(deps, d)
		lines = rest
	}
	return deps, nil
}

func (prj Project) extractIdentity(line string) (pom.Identity, error) {
	line = strings.TrimPrefix(line, _MavenOutputPrefix)
	line = strings.TrimSpace(line)

	if line == "" || line == "none" {
		return pom.Identity{}, errDepListOver
	}

	chunks := strings.Split(line, ":")

	if len(chunks) != _DepChunks {
		return pom.Identity{}, oops.Errorf("dependency format invalid: %s", line)
	}

	id := pom.Identity{
		GroupID:    chunks[_GroupIDIx],
		ArtifactID: chunks[_ArtifactIDIx],
		Version:    chunks[_VersionIx],
	}
	return id, nil
}
