// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"github.com/soniakeys/graph"
)

type Filter interface {
	Filter([]Project, graph.Directed, []bool)
}

func Exactly(prjName string) Filter { return exactly(prjName) }

type exactly string

func (ex exactly) Filter(ps []Project, _ graph.Directed, include []bool) {
	for i, p := range ps {
		if p.Info().Name == string(ex) {
			include[i] = true
		}
	}
}

func WithDependents(prjName string) Filter { return withDependents(prjName) }

type withDependents string

func (wd withDependents) Filter(ps []Project, deps graph.Directed, include []bool) {
	for i, p := range ps {
		if p.Info().Name == string(wd) {
			wd.markDeps(deps, include, i)
		}
	}
}

func (wd withDependents) markDeps(deps graph.Directed, include []bool, i int) {
	queue, nextQueue := []graph.NI{graph.NI(i)}, []graph.NI{}
	for len(queue) > 0 {
		for _, ni := range queue {
			include[ni] = true
			nextQueue = append(nextQueue, deps.AdjacencyList[ni]...)
		}
		queue, nextQueue = nextQueue, queue[:0]
	}

}

func WithDeps(prjName string) Filter { return withDeps(prjName) }

type withDeps string

func (wd withDeps) Filter(ps []Project, deps graph.Directed, include []bool) {
	invDeps, _ := deps.Transpose()
	withDependents(string(wd)).Filter(ps, invDeps, include)
}

func Exclude(prjName string) Filter { return exclude(prjName) }

type exclude string

func (ex exclude) Filter(ps []Project, _ graph.Directed, include []bool) {
	for i, p := range ps {
		if p.Info().Name == string(ex) {
			include[i] = false
		}
	}
}
