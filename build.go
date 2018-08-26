// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"context"
	"fmt"
	"os"

	"github.com/samsarahq/go/oops"
	"github.com/soniakeys/graph"
)

func BuildProjects(ctx context.Context, projects []Project) error {
	orded, err := orderProjects(projects)
	if err != nil {
		return oops.Wrapf(err, "cannot resolve build order")
	}

	for _, prj := range orded {

		err := prj.Build(ctx, os.Stdout)
		if err != nil {
			return oops.Wrapf(err, "cannot build project %v", prj.Info())
		}
	}

	return nil
}

func orderProjects(prjs []Project) ([]Project, error) {
	dg := depGraph(prjs)

	ord, cycle := dg.Topological()
	if ord == nil {
		prjCycle := sequenceProjects(prjs, cycle)
		err := fmt.Errorf("dependency cycle: %v", prjCycle)
		return nil, err
	}

	return sequenceProjects(prjs, ord), nil
}

func depGraph(prjs []Project) graph.Directed {
	ixs := map[ProjectInfo]int{}
	for i, prj := range prjs {
		ixs[prj.Info()] = i
	}

	adjList := make(graph.AdjacencyList, len(prjs))
	for i, prjs := range prjs {
		deps := prjs.Deps()
		edgeEnds := make([]graph.NI, len(deps))

		for j, dep := range deps {
			edgeEnds[j] = graph.NI(ixs[dep])
		}
		adjList[i] = edgeEnds
	}

	return graph.Directed{adjList}
}

func sequenceProjects(prjs []Project, ord []graph.NI) []Project {
	out := make([]Project, len(ord))
	for i := range ord {
		out[i] = prjs[ord[i]]
	}
	return out
}
