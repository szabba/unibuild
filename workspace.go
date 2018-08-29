// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"log"

	"github.com/samsarahq/go/oops"
	"github.com/soniakeys/graph"
)

type Workspace struct {
	projects []Project
}

func NewWorkspace(projects []Project) *Workspace {
	return &Workspace{
		projects: append([]Project{}, projects...),
	}
}

func (w *Workspace) FindBuildOrder() ([]Project, error) {
	providers, err := w.buildProviderMap()
	if err != nil {
		return nil, oops.Wrapf(err, "problem building providers map")
	}

	g := w.depGraph(providers)
	order, cycle := g.Topological()

	if order == nil {
		return nil, oops.Errorf("build dependency cycle %s", w.orderProjects(cycle))
	}
	return w.orderProjects(order), nil
}

func (w *Workspace) buildProviderMap() (map[RequirementIdentity]int, error) {
	providers := map[RequirementIdentity]int{}
	for i, p := range w.projects {
		for _, b := range p.Builds() {

			prev, present := providers[b.ID]
			if present && prev != i {
				return nil, oops.Errorf(
					"both %s and %s build %s",
					w.projects[i].Info().Name,
					w.projects[prev].Info().Name,
					b.ID)
			}
			providers[b.ID] = i
		}
	}
	return providers, nil
}

func (w *Workspace) depGraph(providerIxs map[RequirementIdentity]int) graph.Directed {
	adjList := make(graph.AdjacencyList, len(w.projects))
	for i, p := range w.projects {
		adjList[i] = w.edgeEnds(p, providerIxs)
	}
	return graph.Directed{adjList}
}

func (w *Workspace) edgeEnds(p Project, providerIxs map[RequirementIdentity]int) []graph.NI {
	uses := p.Uses()
	ends := make([]graph.NI, 0, len(uses))
	for _, req := range uses {

		ix, present := providerIxs[req.ID()]
		if !present {
			log.Printf("no provider for %#v", req.ID())
			continue
		}
		ends = append(ends, graph.NI(ix))

	}
	return ends
}

func (w *Workspace) orderProjects(order []graph.NI) []Project {
	ps := make([]Project, len(order))
	for i, pIX := range order {
		ps[i] = w.projects[pIX]
	}
	return ps
}
