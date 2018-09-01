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
	projects     []Project
	depGraph     graph.Directed
	order, cycle []Project
	cycleErr     error
}

func NewWorkspace(projects []Project) *Workspace {
	w := &Workspace{
		projects: append([]Project{}, projects...),
	}
	w.resolveOrder()
	return w
}

func (w *Workspace) resolveOrder() {
	providers, err := w.buildProviderMap()
	if err != nil {
		w.cycleErr = oops.Wrapf(err, "problem building providers map")
		return
	}

	w.depGraph = w.buildDepGraph(providers)
	order, cycle := w.depGraph.Topological()
	if len(cycle) > 0 {
		w.cycle = w.orderProjects(cycle)
		w.cycleErr = NewDependencyCycleError(w.cycle)
		return
	}
	w.order = w.orderProjects(order)
}

func (w *Workspace) BuildOrder() ([]Project, error) {
	return append([]Project{}, w.order...), w.cycleErr
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

func (w *Workspace) buildDepGraph(providerIxs map[RequirementIdentity]int) graph.Directed {
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
