// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"github.com/szabba/unibuild"
)

type Requirement struct {
	id unibuild.RequirementIdentity
}

func NewRequirement(id Identity) Requirement {
	return Requirement{
		id: unibuild.RequirementIdentity{
			Name: id.GroupID + ":" + id.ArtifactID,
		},
	}
}

var _ unibuild.Requirement = Requirement{}

func (req Requirement) ID() unibuild.RequirementIdentity { return req.id }
