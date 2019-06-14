// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os"
	"os/exec"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/unibuild/prefixio"
)

type Local struct {
	Remote
	Path string
}

func (l Local) Reset(ctx context.Context) error {
	err := l.Run(ctx, "git", "reset", "--hard")
	return oops.Wrapf(err, "in repository at %s, failed to reset", l.Path)
}

func (l Local) Fetch(ctx context.Context) error {
	err := l.Run(ctx, "git", "fetch", "--force", "--prune", "--tags")
	return oops.Wrapf(err, "in repository at %s, failed to fetch", l.Path)

}

func (l Local) Checkout(ctx context.Context, ref string) error {
	err := l.Run(ctx, "git", "checkout", "-B", ref, "origin/"+ref)
	return oops.Wrapf(err, "in repository at %s, failed to checkout %s", l.Path, ref)
}

func (l Local) CheckoutFirst(ctx context.Context, ref string, otherRefs ...string) error {
	allRefs := append([]string{ref}, otherRefs...)
	for _, ref := range allRefs {
		err := l.Checkout(ctx, ref)
		if err == nil {
			return nil
		}
	}
	return oops.Errorf("in repository at %s, none of the refs %q could be checked out", l.Path, allRefs)
}

func (l Local) CurrentHash(ctx context.Context) (string, error) {
	cmd := l.Command(ctx, "git", "show", "--format", "format:%H", "-s")
	out, err := cmd.Output()
	if err != nil {
		return "", oops.Wrapf(err, "cannot get current commit hash of repo at %s", l.Path)
	}
	return string(out), nil
}

func (l Local) Run(ctx context.Context, cmdName string, args ...string) error {
	cmd := l.Command(ctx, cmdName, args...)
	out, err := cmd.CombinedOutput()
	prefixio.NewWriter(os.Stdout, l.Name+" | ").Write(out)
	return err
}

func (l Local) Command(ctx context.Context, cmdName string, args ...string) *exec.Cmd {
	cmd := l.Remote.Command(ctx, cmdName, args...)
	cmd.Dir = l.Path
	return cmd
}
