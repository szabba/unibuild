// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"

	"github.com/samsarahq/go/oops"

	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/maven"
	"github.com/szabba/uninbuild/repo"
)

type Workspace struct {
	dir      string
	clones   *repo.ClonedSet
	projects []unibuild.Project
}

func NewWorkspace(ctx context.Context, clones *repo.ClonedSet) (*Workspace, error) {
	ws := &Workspace{
		dir:    clones.Dir(),
		clones: clones,
	}
	prjs, err := ws.findProjects(ctx)
	if err != nil {
		return nil, err
	}
	ws.projects = prjs
	return ws, nil
}

func (ws Workspace) Projects() []unibuild.Project { return ws.projects }

func (ws Workspace) findProjects(ctx context.Context) ([]unibuild.Project, error) {
	depless, err := ws.identifyProjects(ctx)
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

func (ws Workspace) identifyProjects(ctx context.Context) ([]Project, error) {
	prjs := make([]Project, 0, ws.clones.Size())
	err := ws.clones.EachTry(func(cln repo.Local) error {
		p, err := NewProject(ctx, cln)
		if err != nil {
			return oops.Wrapf(err, "problem identifying project of repo %s", cln.Name)
		}
		prjs = append(prjs, p)
		return nil
	})

	if err != nil {
		return nil, err
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

func (ws Workspace) identifyProviders(prjs []Project) (map[maven.Identity]Project, error) {
	providers := map[maven.Identity]Project{}
	for _, p := range prjs {
		err := ws.addProvided(p, providers)
		if err != nil {
			return nil, err
		}
	}
	return providers, nil
}

func (ws Workspace) addProvided(prj Project, providers map[maven.Identity]Project) error {
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
