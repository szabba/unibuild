// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os"
	"os/exec"

	"github.com/samsarahq/go/oops"
)

type Local struct {
	Remote
	Path string
}

func (l Local) Reset(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "reset", "--hard")
	cmd.Dir = l.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l Local) Fetch(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "fetch", "--force", "--prune", "--tags")
	cmd.Dir = l.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (l Local) Checkout(ctx context.Context, ref string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-B", ref, "origin/"+ref)
	cmd.Dir = l.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
	cmd := exec.CommandContext(ctx, "git", "show", "--format", "format:%H", "-s")
	cmd.Dir = l.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", oops.Wrapf(err, "cannot get current commit hash of repo at %s", l.Path)
	}
	return string(out), nil
}
