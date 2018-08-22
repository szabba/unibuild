// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/assert"
)

func SyncAll(ctx context.Context, set *Set, dir string) (*ClonedSet, error) {
	wrap := func(err error) error {
		return oops.Wrapf(err, "problem syncing repository set")
	}

	clones := NewClonedSet(dir)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return clones, wrap(err)
	}

	err = set.EachTry(func(r Remote) error { return Sync(ctx, clones, r) })

	return clones, wrap(err)
}

func Sync(ctx context.Context, clones *ClonedSet, r Remote) error {
	assert.That(r.Name != "", log.Panicf, "remote repository name must not be empty: %#v", r)

	wrap := func(err error) error {
		return oops.Wrapf(err, "problem syncing repository %s in directory %s", r.Name, clones.Dir())
	}

	dstPath := filepath.Join(clones.Dir(), r.Name)

	// We ignore the possibility that the location exists but is not a directory.
	_, err := os.Stat(dstPath)
	if os.IsNotExist(err) {
		return wrap(clones.Clone(ctx, r))
	}
	if err != nil {
		return wrap(err)
	}

	l := Local{Remote: r, Path: dstPath}
	err = clones.add(l)
	if err != nil {
		return wrap(err)
	}

	return wrap(l.Fetch(ctx))
}
