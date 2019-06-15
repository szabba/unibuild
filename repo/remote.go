// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/szabba/unibuild/prefixio"
)

type Remote struct {
	Name string
	URL  string
}

func (r Remote) Clone(ctx context.Context, dir string) (Local, error) {
	cmd := r.Command(ctx, "git", "clone", r.URL)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	prefixio.NewWriter(os.Stdout, r.Name+" | ").Write(out)
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
	return cmd
}
