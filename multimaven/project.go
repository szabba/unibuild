// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"
	"io"
	"os/exec"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/unibuild"
	"github.com/szabba/unibuild/maven"
	"github.com/szabba/unibuild/repo"
)

type Project struct {
	name    string
	version string
	dir     string
	uses    []unibuild.Requirement
	builds  []unibuild.RequirementVersion
}

var _ unibuild.Project = Project{}

func NewProject(ctx context.Context, clone repo.Local) (Project, error) {
	pomHeads, err := maven.Scan(ctx, clone.Path)
	if err != nil {
		return Project{}, oops.Wrapf(err, "problem scanning for maven modules in %s", clone.Path)
	}

	uses, err := findUses(ctx, clone, pomHeads)
	if err != nil {
		return Project{}, oops.Wrapf(err, "problem resolving maven dependencies in %s", clone.Path)
	}

	prj := Project{
		name:    clone.Name,
		version: pomHeads[0].EffectiveVersion(),
		dir:     clone.Path,
		uses:    uses,
		builds:  headsToBuilds(pomHeads),
	}

	return prj, nil
}

func findUses(ctx context.Context, clone repo.Local, heads []maven.Header) ([]unibuild.Requirement, error) {
	mods := make([]maven.Identity, len(heads))
	for i, h := range heads {
		mods[i] = h.EffectiveIdentity()
	}

	depIDs, err := maven.ListDeps(ctx, clone.Path, mods)
	if err != nil {
		return nil, oops.Wrapf(err, "problem listing maven deps in %s", clone.Path)
	}

	reqs := make([]unibuild.Requirement, len(depIDs))
	for i, id := range depIDs {
		reqs[i] = NewRequirement(id)
	}

	return reqs, nil
}

func headsToBuilds(heads []maven.Header) []unibuild.RequirementVersion {
	vreqs := make([]unibuild.RequirementVersion, len(heads))
	for i, h := range heads {
		vreqs[i] = unibuild.RequirementVersion{
			ID: unibuild.RequirementIdentity{
				Name: h.EffectiveGroupID() + ":" + h.EffectiveArtifactID(),
			},
			// TODO: Handle versioning
			// Version: h.EffectiveVersion(),
		}
	}
	return vreqs
}

func (prj Project) Info() unibuild.ProjectInfo {
	return unibuild.ProjectInfo{
		Name:    prj.name,
		Version: prj.version,
	}
}

func (prj Project) Uses() []unibuild.Requirement { return prj.uses }

func (prj Project) Builds() []unibuild.RequirementVersion { return prj.builds }

func (prj Project) Build(ctx context.Context, logTo io.Writer) error {
	cmd := exec.CommandContext(ctx, "mvn", "clean", "deploy")
	cmd.Dir = prj.dir
	cmd.Stdout = logTo
	cmd.Stderr = logTo
	return cmd.Run()
}
