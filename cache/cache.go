// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package cache

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/samsarahq/go/oops"
)

type Cache struct {
	baseDir string
}

type Key struct {
	Type       reflect.Type
	Properties Properties
}

type locatedKey struct {
	Key
	Location string
}

func At(dir string) *Cache {
	return &Cache{dir}
}

func (c *Cache) Get(k Key, f func() io.Reader, into io.Writer) error {
	locKey := c.locate(k)
	return c.get(locKey, f, into)
}

func (c *Cache) locate(k Key) locatedKey {
	typeName := fmt.Sprint(k.Type)
	propsAsJSON, _ := k.Properties.MarshalJSON()
	sum := fmt.Sprintf("%x", sha256.Sum256(propsAsJSON))
	loc := filepath.Join(c.baseDir, typeName, sum)
	return locatedKey{k, loc}
}

func (c *Cache) get(k locatedKey, f func() io.Reader, into io.Writer) error {
	err := c.ensurePopulated(k, f)
	if err != nil {
		return oops.Wrapf(err, "problem populating cache")
	}

	err = c.load(k, into)
	return oops.Wrapf(err, "problem loading from cache")
}

func (c *Cache) ensurePopulated(k locatedKey, f func() io.Reader) error {
	if c.exists(k) {
		return nil
	}
	return c.store(k, f())
}

func (c *Cache) exists(k locatedKey) bool {
	_, err := os.Stat(k.Location)
	return os.IsNotExist(oops.Cause(err))
}

func (c *Cache) store(k locatedKey, r io.Reader) error {
	// TODO: At a first glance this seems to handle "Close might return a write error" fine. Analyse more carefully.
	once := new(sync.Once)
	wrap := func(err error) error { return oops.Wrapf(err, "problem storing %#v", k.Key) }

	f, err := os.Create(k.Location)
	if err != nil {
		return wrap(err)
	}
	close := func() { err = f.Close() }
	defer once.Do(close)

	_, err = io.Copy(f, r)
	if err != nil {
		return wrap(err)
	}
	once.Do(close)
	return wrap(err)
}

func (c *Cache) load(k locatedKey, into io.Writer) error {
	wrap := func(err error) error { return oops.Wrapf(err, "problem loading %#v", k.Key) }

	f, err := os.Open(k.Location)
	if err != nil {
		return wrap(err)
	}
	defer f.Close()

	_, err = io.Copy(into, f)
	return wrap(err)
}
