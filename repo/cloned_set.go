// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"context"
	"os"
	"path/filepath"

	"github.com/samsarahq/go/oops"
)

type ClonedSet struct {
	dir   string
	repos map[string]Local
}

func NewClonedSet(dir string) *ClonedSet {
	return &ClonedSet{
		dir:   dir,
		repos: map[string]Local{},
	}
}

func CloneAll(ctx context.Context, set *Set, dir string) (*ClonedSet, error) {
	wrap := func(err error) error {
		return oops.Wrapf(err, "problem cloning repository set")
	}

	clones := NewClonedSet(dir)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return clones, wrap(err)
	}

	err = set.EachTry(func(r Remote) error {
		return clones.Clone(ctx, r)
	})
	return clones, wrap(err)
}

func (set *ClonedSet) Clone(ctx context.Context, r Remote) error {
	wrap := func(err error) error {
		return oops.Wrapf(err, "problem cloning remote repository %#v inside directory %s", r, set.dir)
	}

	if prev, present := set.repos[r.Name]; present && r != prev.Remote {
		return wrap(ErrDuplicateName)
	}

	l, err := r.Clone(ctx, set.dir)
	if err != nil {
		return wrap(err)
	}

	set.repos[l.Name] = l
	return nil
}

func (set *ClonedSet) Clear() error {
	wrap := func(err error) error {
		return oops.Wrapf(err, "problem clearing cloned repository set at %s", set.dir)
	}

	dir, err := os.Open(set.dir)
	if err != nil {
		return wrap(err)
	}

	children, err := dir.Readdirnames(-1)
	if err != nil {
		return wrap(err)
	}

	for _, c := range children {
		err := os.RemoveAll(filepath.Join(set.dir, c))
		if err != nil {
			return wrap(err)
		}
	}
	return nil
}

func (set *ClonedSet) Size() int { return len(set.repos) }

func (set *ClonedSet) Each(f func(Local)) {
	for _, l := range set.repos {
		f(l)
	}
}

func (set *ClonedSet) EachTry(f func(Local) error) error {
	for _, l := range set.repos {
		err := f(l)
		if err != nil {
			return err
		}
	}
	return nil
}
