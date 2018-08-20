// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"context"
	"io"
)

type Project interface {
	Info() ProjectInfo
	Deps() []ProjectInfo
	Build(ctx context.Context, arts map[ProjectInfo][]Artifact, logTo io.Writer) ([]Artifact, error)
}

type ProjectInfo struct {
	Name    string
	Version string
}