// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/szabba/unibuild/prefixio"
)

type Remote struct {
	Name string
	URL  string
	out  io.Writer
	log  *log.Logger
}

func (r Remote) Out() io.Writer {
	if r.out == nil {
		r.out = prefixio.NewWriter(os.Stdout, r.Name+" | ")
	}
	return r.out
}

func (r Remote) Log() *log.Logger {
	if r.log == nil {
		r.log = log.New(r.Out(), "", 0)
	}
	return r.log
}

func (r Remote) Clone(ctx context.Context, dir string) (Local, error) {
	cmd := r.Command(ctx, "git", "clone", r.URL)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	r.Out().Write(out)
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
