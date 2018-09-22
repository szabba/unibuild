// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild_test

import (
	"testing"

	"github.com/szabba/assert"

	"github.com/szabba/unibuild"
)

func TestLibraryIsSortedBeforeAnApplicationUsingIt(t *testing.T) {
	// given
	libID := unibuild.RequirementIdentity{Name: "lib"}

	var lib unibuild.Project = &Project{
		Info_: unibuild.ProjectInfo{Name: "lib"},
		Builds_: []unibuild.RequirementVersion{
			{ID: libID},
		},
	}

	var app unibuild.Project = &Project{
		Info_: unibuild.ProjectInfo{Name: "app"},
		Uses_: []unibuild.Requirement{
			Requirement{ID_: libID},
		},
	}

	suite := unibuild.NewProjectSuite(lib, app)

	// when
	ordSuite, err := suite.ResolveOrder()
	order := ordSuite.Order()

	// then
	assert.That(err == nil, t.Errorf, "unexpected error reported: %s", err)
	assert.That(len(order) == 2, t.Fatalf, "got %d projects in order, want %d", len(order), 2)
	assert.That(order[0] == lib, t.Errorf, "got 0-th project %#v, want %#v", order[0].Info(), lib.Info())
	assert.That(order[1] == app, t.Errorf, "got 1-st project %#v, want %#v", order[1].Info(), app.Info())
}

func TestDirectCycleIsDetected(t *testing.T) {
	// given
	idA := unibuild.RequirementIdentity{Name: "a"}
	idB := unibuild.RequirementIdentity{Name: "b"}

	var prjA unibuild.Project = &Project{
		Info_:   unibuild.ProjectInfo{Name: "a"},
		Uses_:   []unibuild.Requirement{Requirement{ID_: idB}},
		Builds_: []unibuild.RequirementVersion{{ID: idA}},
	}
	var prjB unibuild.Project = &Project{
		Info_:   unibuild.ProjectInfo{Name: "b"},
		Uses_:   []unibuild.Requirement{Requirement{ID_: idA}},
		Builds_: []unibuild.RequirementVersion{{ID: idB}},
	}

	suite := unibuild.NewProjectSuite(prjA, prjB)

	// when
	_, err := suite.ResolveOrder()

	// then
	assert.That(err != nil, t.Errorf, "got no error when one is expected")
}

func TestIndirectCycleIsDetected(t *testing.T) {
	// given
	idA := unibuild.RequirementIdentity{Name: "a"}
	idB := unibuild.RequirementIdentity{Name: "b"}
	idC := unibuild.RequirementIdentity{Name: "c"}

	var prjA unibuild.Project = &Project{
		Info_:   unibuild.ProjectInfo{Name: "a"},
		Uses_:   []unibuild.Requirement{Requirement{ID_: idC}},
		Builds_: []unibuild.RequirementVersion{{ID: idA}},
	}
	var prjB unibuild.Project = &Project{
		Info_:   unibuild.ProjectInfo{Name: "b"},
		Uses_:   []unibuild.Requirement{Requirement{ID_: idA}},
		Builds_: []unibuild.RequirementVersion{{ID: idB}},
	}
	var prjC unibuild.Project = &Project{
		Info_:   unibuild.ProjectInfo{Name: "c"},
		Uses_:   []unibuild.Requirement{Requirement{ID_: idB}},
		Builds_: []unibuild.RequirementVersion{{ID: idC}},
	}

	suite := unibuild.NewProjectSuite(prjA, prjB, prjC)

	// when
	_, err := suite.ResolveOrder()

	// then
	assert.That(err != nil, t.Errorf, "got no error when one is expected")
}
