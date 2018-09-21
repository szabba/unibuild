// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"log"

	"github.com/samsarahq/go/oops"
	"github.com/soniakeys/graph"
)

type ProjecSuite struct {
	projects []Project
	depGraph graph.Directed
	order    []Project
	cycleErr error
}

func NewProjecSuite(projects []Project) *ProjecSuite {
	ps := &ProjecSuite{
		projects: append([]Project{}, projects...),
	}
	ps.resolveOrder()
	return ps
}

func (ps *ProjecSuite) resolveOrder() {
	providers, err := ps.buildProviderMap()
	if err != nil {
		ps.cycleErr = oops.Wrapf(err, "problem building providers map")
		return
	}

	ps.depGraph = ps.buildDepGraph(providers)
	order, cycle := ps.depGraph.Topological()
	if len(cycle) > 0 {
		pjsCycle := ps.orderProjects(cycle)
		ps.cycleErr = NewDependencyCycleError(pjsCycle)
		return
	}
	ps.order = ps.orderProjects(order)
}

func (ps *ProjecSuite) BuildOrder() ([]Project, error) {
	return append([]Project{}, ps.order...), ps.cycleErr
}

func (ps *ProjecSuite) buildProviderMap() (map[RequirementIdentity]int, error) {
	providers := map[RequirementIdentity]int{}
	for i, p := range ps.projects {
		for _, b := range p.Builds() {

			prev, present := providers[b.ID]
			if present && prev != i {
				return nil, oops.Errorf(
					"both %s and %s build %s",
					ps.projects[i].Info().Name,
					ps.projects[prev].Info().Name,
					b.ID)
			}
			providers[b.ID] = i
		}
	}
	return providers, nil
}

func (ps *ProjecSuite) buildDepGraph(providerIxs map[RequirementIdentity]int) graph.Directed {
	adjList := make(graph.AdjacencyList, len(ps.projects))
	for i, p := range ps.projects {
		adjList[i] = ps.edgeEnds(p, providerIxs)
	}
	return graph.Directed{adjList}
}

func (ps *ProjecSuite) edgeEnds(p Project, providerIxs map[RequirementIdentity]int) []graph.NI {
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

func (ps *ProjecSuite) orderProjects(order []graph.NI) []Project {
	pjs := make([]Project, len(order))
	for i, pIX := range order {
		pjs[i] = ps.projects[pIX]
	}
	return pjs
}
