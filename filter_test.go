// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild_test

import (
	"testing"

	"github.com/szabba/assert"

	"github.com/szabba/unibuild"
)

func TestFilters(t *testing.T) {
	idA := unibuild.RequirementIdentity{Name: "a"}
	idB := unibuild.RequirementIdentity{Name: "b"}
	idC := unibuild.RequirementIdentity{Name: "c"}
	idD := unibuild.RequirementIdentity{Name: "d"}

	var (
		prjA unibuild.Project = Project{
			Info_:   unibuild.ProjectInfo{Name: "a"},
			Builds_: []unibuild.RequirementVersion{{ID: idA}},
		}
		prjB unibuild.Project = Project{
			Info_:   unibuild.ProjectInfo{Name: "b"},
			Uses_:   []unibuild.Requirement{Requirement{ID_: idA}},
			Builds_: []unibuild.RequirementVersion{{ID: idB}},
		}
		prjC unibuild.Project = Project{
			Info_:   unibuild.ProjectInfo{Name: "c"},
			Uses_:   []unibuild.Requirement{Requirement{ID_: idB}},
			Builds_: []unibuild.RequirementVersion{{ID: idC}},
		}
		prjD unibuild.Project = Project{
			Info_:   unibuild.ProjectInfo{Name: "d"},
			Uses_:   []unibuild.Requirement{Requirement{ID_: idA}},
			Builds_: []unibuild.RequirementVersion{{ID: idD}},
		}
	)

	ordSuite, err := unibuild.NewProjectSuite(prjA, prjB, prjC, prjD).ResolveOrder()
	assert.That(err == nil, t.Fatalf, "unexpected error: %s", err)

	t.Run("ExactlyB", func(t *testing.T) {
		// given
		filter := unibuild.Exactly("b")

		// when
		filterSuite := ordSuite.Filter(filter)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjB)
	})

	t.Run("BWithDeps", func(t *testing.T) {
		// given
		filter := unibuild.WithDeps("b")

		// when
		filterSuite := ordSuite.Filter(filter)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjA, prjB)
	})

	t.Run("CWithDeps", func(t *testing.T) {
		// given
		filter := unibuild.WithDeps("c")

		// when
		filterSuite := ordSuite.Filter(filter)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjA, prjB, prjC)
	})

	t.Run("BWithDependents", func(t *testing.T) {
		// given
		filter := unibuild.WithDependents("b")

		// when
		filterSuite := ordSuite.Filter(filter)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjB, prjC)
	})

	t.Run("BDepsAndDependents", func(t *testing.T) {
		// given
		withDeps := unibuild.WithDeps("b")
		withDependents := unibuild.WithDependents("b")
		exclude := unibuild.Exclude("b")

		// when
		filterSuite := ordSuite.Filter(withDeps, withDependents, exclude)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjA, prjC)
	})

	t.Run("ExcludeCancelsInclude", func(t *testing.T) {
		// given
		include := unibuild.Exactly("b")
		exclude := unibuild.Exclude("b")

		// when
		filterSuite := ordSuite.Filter(include, exclude)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order)
	})

	t.Run("IncludeCancelsExclude", func(t *testing.T) {
		// given
		include := unibuild.Exactly("b")
		exclude := unibuild.Exclude("b")

		// when
		filterSuite := ordSuite.Filter(include, exclude, include)

		// then
		order := filterSuite.Order()
		assertOrder(t.Errorf, order, prjB)
	})
}

func assertOrder(onErr assert.ErrorFunc, got []unibuild.Project, want ...unibuild.Project) {
	if len(got) != len(want) {
		onErr("got order of %d projects, want %d", len(got), len(want))
		return
	}
	for i := range got {
		pGot, pWant := got[i], want[i]
		assert.That(pGot.Info() == pWant.Info(), onErr, "got %d-th project %#v, want %#v", i, pGot.Info(), pWant.Info())
	}
}
