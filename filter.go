// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"github.com/soniakeys/graph"
)

// type alias Filter = DirectedGraph Project -> Dict Project bool -> Dict Project Bool

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

func WithDeps(prjName string) Filter { return nil }

func WithDependents(prjName string) Filter { return nil }

func Exclude(prjName string) Filter { return nil }
