// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"bytes"
	"context"
	"io"

	"github.com/szabba/unibuild/output"
	"github.com/szabba/unibuild/repo"
)

type cloneWrap struct {
	repo.Local
}

func (cw cloneWrap) Run(ctx context.Context, cmd string, args ...string) error {
	var buf bytes.Buffer

	stdOut := output.FromContext(ctx)
	ctx = output.WithOutput(ctx, &buf)
	ctx = output.WithErr(ctx, &buf)

	err := cw.Command(ctx, cmd, args...).Run()
	io.Copy(stdOut, &buf)
	return err
}
