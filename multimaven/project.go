// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"
	"io"
	"os/exec"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/maven"
	"github.com/szabba/uninbuild/repo"
)

type Project struct {
	name     string
	version  string
	dir      string
	deps     []unibuild.ProjectInfo
	pomHeads []maven.Header
}

var _ unibuild.Project = Project{}

func NewProject(ctx context.Context, clone repo.Local) (Project, error) {
	pomHeads, err := maven.Scan(ctx, clone.Path)
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

func (prj Project) MavenModuleHeaders() []maven.Header { return prj.pomHeads }

func (prj Project) WithDependnecies(ctx context.Context, providers map[maven.Identity]Project) (Project, error) {
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

func (prj Project) FindMavenDeps(ctx context.Context) ([]maven.Identity, error) {
	mods := make([]maven.Identity, len(prj.pomHeads))
	for i, h := range prj.pomHeads {
		mods[i] = h.EffectiveIdentity()
	}
	return maven.ListDeps(ctx, prj.dir, mods)
}
