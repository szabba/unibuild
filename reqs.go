// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import "errors"

var (
	ErrWrongVersion  = errors.New("wrong version")
	ErrCannotSatisfy = errors.New("cannot satisfy")
)

type RequirementIdentity struct {
	// TODO: To handle multi-ecosystem builds.
	// Ecosystem string
	Name string
}

type Requirement interface {
	ID() RequirementIdentity
	// TODO: Handle versioning
	// Accepts(v Version) bool
}

type RequirementVersion struct {
	ID RequirementIdentity
	// TODO: Handle versioning
	// Version Version
}

func Satisfies(reqver RequirementVersion, req Requirement) bool {
	return reqver.ID == req.ID()
}
