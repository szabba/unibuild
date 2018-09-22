// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"log"

	"github.com/samsarahq/go/oops"
	"github.com/soniakeys/graph"
)

type ProjectSuite struct {
	projects []Project
}

func NewProjectSuite(projects ...Project) *ProjectSuite {
	return &ProjectSuite{
		projects: append([]Project{}, projects...),
	}
}

func (ps *ProjectSuite) ResolveOrder() (OrderedProjectSuite, error) {
	ixOrder, depGraph, err := ps.resolveOrder()
	if err != nil {
		return OrderedProjectSuite{}, err
	}
	order := ps.orderProjects(ixOrder)
	ordSuite := OrderedProjectSuite{ps.projects, depGraph, ixOrder, order}
	return ordSuite, nil
}

func (ps *ProjectSuite) resolveOrder() ([]graph.NI, graph.Directed, error) {
	providers, err := ps.buildProviderMap()
	if err != nil {
		return nil, graph.Directed{}, oops.Wrapf(err, "problem building providers map")
	}

	depGraph := ps.buildDepGraph(providers)
	order, cycle := depGraph.Topological()
	if len(cycle) > 0 {
		pjsCycle := ps.orderProjects(cycle)
		return nil, depGraph, NewDependencyCycleError(pjsCycle)
	}
	return order, depGraph, nil
}

func (ps *ProjectSuite) buildProviderMap() (map[RequirementIdentity]int, error) {
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

func (ps *ProjectSuite) buildDepGraph(providerIxs map[RequirementIdentity]int) graph.Directed {
	adjList := make(graph.AdjacencyList, len(ps.projects))
	for i, p := range ps.projects {
		adjList[i] = ps.edgeEnds(p, providerIxs)
	}
	inverse := graph.Directed{adjList}
	depGraph, _ := inverse.Transpose()
	return depGraph
}

func (ps *ProjectSuite) edgeEnds(p Project, providerIxs map[RequirementIdentity]int) []graph.NI {
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

func (ps *ProjectSuite) orderProjects(order []graph.NI) []Project {
	pjs := make([]Project, len(order))
	for i, pIX := range order {
		pjs[i] = ps.projects[pIX]
	}
	return pjs
}

type OrderedProjectSuite struct {
	projects []Project
	depGraph graph.Directed
	ixOrder  []graph.NI
	order    []Project
}

func (ops OrderedProjectSuite) Order() []Project {
	return append([]Project{}, ops.order...)
}

func (ops OrderedProjectSuite) Filter(fs ...Filter) FilteredProjectSuite {
	include := make([]bool, len(ops.projects))
	for _, f := range fs {
		f.Filter(ops.projects, ops.depGraph, include)
	}
	order := make([]Project, 0, len(ops.projects))
	for _, i := range ops.ixOrder {
		if include[i] {
			nextProj := ops.projects[i]
			order = append(order, nextProj)
		}
	}
	return FilteredProjectSuite{order}
}

type FilteredProjectSuite struct {
	order []Project
}

func (fps FilteredProjectSuite) Order() []Project {
	return append([]Project{}, fps.order...)
}
