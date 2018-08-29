// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

type Remote struct {
	Name string
	URL  string
}

func (r Remote) Clone(ctx context.Context, dir string) (Local, error) {
	cmd := exec.CommandContext(ctx, "git", "clone", r.URL)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return Local{}, err
	}
	return Local{
		Remote: r,
		Path:   filepath.Join(dir, r.Name),
	}, nil
}
