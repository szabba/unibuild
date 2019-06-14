// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/szabba/unibuild/output"

	"github.com/szabba/unibuild/prefixio"
)

type Remote struct {
	Name string
	URL  string
}

func (r Remote) Clone(ctx context.Context, dir string) (Local, error) {
	cmd := r.Command(ctx, "git", "clone", r.URL)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		return Local{}, err
	}
	return Local{
		Remote: r,
		Path:   filepath.Join(dir, r.Name),
	}, nil
}

func (r Remote) Command(ctx context.Context, cmdName string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	r.wrapOutput(ctx, cmd)
	return cmd
}

func (r Remote) wrapOutput(ctx context.Context, cmd *exec.Cmd) {
	// FIXME? Possible abuse of context.
	prefix := r.Name + " | "
	stdout := output.FromContext(ctx)
	stderr := output.ErrFromContext(ctx)
	cmd.Stdout = prefixio.NewWriter(stdout, prefix)
	cmd.Stderr = prefixio.NewWriter(stderr, prefix)
}
