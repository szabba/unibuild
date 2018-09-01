// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import "fmt"

type DependencyCycleError struct {
	cycle []Project
}

func NewDependencyCycleError(cycle []Project) error {
	return &DependencyCycleError{cycle}
}

func (err *DependencyCycleError) Error() string {
	return fmt.Sprintf("dependency cycle detected: %#v", err.cycle)
}

func (err *DependencyCycleError) DependencyCycle() []Project {
	return append([]Project{}, err.cycle...)
}
