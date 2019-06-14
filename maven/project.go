// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"context"
	"io"
	"regexp"

	"github.com/samsarahq/go/oops"

	"github.com/szabba/unibuild"
	"github.com/szabba/unibuild/repo"
)

// A Project that is built using maven.
type Project struct {
	name    string
	version string
	clone   repo.Local
	uses    []unibuild.Requirement
	builds  []unibuild.RequirementVersion
}

var _ unibuild.Project = Project{}

// NewProject attempts to create a maven project given a locally cloned repository.
func NewProject(ctx context.Context, clone repo.Local) (Project, error) {
	effPom, err := ParseEffectivePomOfClone(ctx, clone)
	if err != nil {
		return Project{}, oops.Wrapf(err, "problem scanning effective POM in %s", clone.Path)
	}

	if len(effPom.Projects) == 0 {
		return Project{}, oops.Errorf("effective POM invalid: has no projects")
	}

	builds := findBuilds(effPom)

	prj := Project{
		name:    clone.Name,
		version: effPom.Projects[0].EffectiveVersion(),
		clone:   clone,
		uses:    findUses(effPom, builds),
		builds:  builds,
	}

	return prj, nil
}

func findBuilds(effPom EffectivePom) []unibuild.RequirementVersion {
	builds := make([]unibuild.RequirementVersion, 0, len(effPom.Projects))
	for _, prj := range effPom.Projects {
		bld := unibuild.RequirementVersion{
			ID: unibuild.RequirementIdentity{
				Name: prj.EffectiveGroupID() + ":" + prj.EffectiveArtifactID(),
			},
			// TODO: Handle versioning
			// Version: h.EffectiveVersion(),
		}
		builds = append(builds, bld)
	}
	return builds
}

func findUses(effPom EffectivePom, builds []unibuild.RequirementVersion) []unibuild.Requirement {
	all := make(map[unibuild.Requirement]bool)
	for _, prj := range effPom.Projects {
		for _, dep := range prj.Dependencies {
			req := NewRequirement(dep)
			if isSatisifed(req, builds) {
				continue
			}
			all[req] = true
		}
	}

	out := make([]unibuild.Requirement, 0, len(all))
	for req := range all {
		out = append(out, req)
	}
	return out
}

func isSatisifed(req unibuild.Requirement, builds []unibuild.RequirementVersion) bool {
	satisfied := false
	for _, bld := range builds {
		satisfied = satisfied || unibuild.Satisfies(bld, req)
	}
	return satisfied
}

func (prj Project) Info() unibuild.ProjectInfo {
	return unibuild.ProjectInfo{
		Name:    prj.name,
		Version: prj.version,
	}
}

func (prj Project) Uses() []unibuild.Requirement { return prj.uses }

func (prj Project) Builds() []unibuild.RequirementVersion { return prj.builds }

var _Line = regexp.MustCompile(`[^\r\n]*\r?\n`)

func (prj Project) Build(ctx context.Context, logTo io.Writer) error {
	return cloneWrap{prj.clone}.Run(ctx, "mvn", "-U", "-B", "clean", "deploy")
}
