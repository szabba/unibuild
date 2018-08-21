// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package repo

import (
	"errors"
)

var ErrDuplicateName = errors.New("duplicate repository name")

type Set struct {
	repos map[string]Remote
}

func NewSet() *Set {
	return &Set{
		repos: map[string]Remote{},
	}
}

func (set *Set) Add(ri Remote) error {
	if prev, present := set.repos[ri.Name]; present && prev != ri {
		return ErrDuplicateName
	}
	set.repos[ri.Name] = ri
	return nil
}

func (set *Set) Copy() *Set {
	cp := NewSet()
	for _, ri := range set.repos {
		cp.Add(ri)
	}
	return cp
}

func (set *Set) Size() int { return len(set.repos) }

func (set *Set) Each(f func(Remote)) {
	for _, r := range set.repos {
		f(r)
	}
}

func (set *Set) EachTry(f func(Remote) error) error {
	for _, r := range set.repos {
		err := f(r)
		if err != nil {
			return err
		}
	}
	return nil
}
