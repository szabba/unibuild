// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild

import (
	"context"
	"io"
)

type Artifact interface {
	Info() ArtifactInfo
	Persist(ctx context.Context, logTo io.Writer) error
}

type ArtifactInfo struct {
	Project ProjectInfo
	Name    string
	Version string
	Props   map[string]string
}
