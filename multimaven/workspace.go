// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"

	"github.com/szabba/uninbuild/pom"

	"github.com/samsarahq/go/oops"

	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/repo"
)

type Workspace struct {
	dir   string
	repos []repo.Remote
}

var _ unibuild.Workspace = Workspace{}

func NewWorkspace(dir string, repos []repo.Remote) *Workspace {
	return &Workspace{dir, repos}
}

func (ws Workspace) Projects(ctx context.Context) ([]unibuild.Project, error) {
	depless, err := ws.identifyProjects(ctx, ws.repos)
	if err != nil {
		return nil, oops.Wrapf(err, "problem identifying projects")
	}

	deped, err := ws.resolveDeps(ctx, depless)
	if err != nil {
		return nil, oops.Wrapf(err, "problem resolving cross-project dependencies")
	}

	prjs := make([]unibuild.Project, 0, len(deped))
	for _, p := range deped {
		prjs = append(prjs, p)
	}
	return prjs, nil
}

func (ws Workspace) identifyProjects(ctx context.Context, repos []repo.Remote) ([]Project, error) {
	prjs := make([]Project, 0, len(repos))
	for _, r := range repos {

		p, err := ws.repoToProject(ctx, r)
		if err != nil {
			return nil, oops.Wrapf(err, "")
		}

		prjs = append(prjs, p)
	}
	return prjs, nil
}

func (ws Workspace) resolveDeps(ctx context.Context, depless []Project) ([]Project, error) {
	providers, err := ws.identifyProviders(depless)
	if err != nil {
		return nil, err
	}

	prjs := make([]Project, 0, len(depless))
	for _, dp := range depless {

		p, err := dp.WithDependnecies(ctx, providers)
		if err != nil {
			return nil, err
		}
		prjs = append(prjs, p)
	}

	return prjs, nil
}

func (ws Workspace) identifyProviders(prjs []Project) (map[pom.Identity]Project, error) {
	providers := map[pom.Identity]Project{}
	for _, p := range prjs {
		err := ws.addProvided(p, providers)
		if err != nil {
			return nil, err
		}
	}
	return providers, nil
}

func (ws Workspace) addProvided(prj Project, providers map[pom.Identity]Project) error {
	for _, h := range prj.MavenModuleHeaders() {
		hID := h.EffectiveIdentity()
		if _, present := providers[hID]; present {
			otherPrj := providers[hID]
			return oops.Errorf(
				"projects %s and %s both provide %s:%s",
				prj.Info().Name, otherPrj.Info().Name, hID.GroupID, hID.ArtifactID)
		}
	}
	return nil
}

func (ws Workspace) repoToProject(ctx context.Context, r repo.Remote) (Project, error) {
	c, err := r.Clone(ctx, ws.dir)
	if err != nil {
		return Project{}, err
	}
	return NewProject(ctx, c)
}
