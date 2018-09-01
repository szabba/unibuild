// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package binhash

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/samsarahq/go/oops"
)

type Sha256 = [sha256.Size]byte

func OwnHash() (Sha256, error) {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		return Sha256{}, err
	}
	return Hash(path)
}

func Hash(path string) (Sha256, error) {
	wrap := func(err error) error { return oops.Wrapf(err, "cannot hash file %s", path) }

	f, err := os.Open(path)
	if err != nil {
		return Sha256{}, wrap(err)
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return Sha256{}, wrap(err)
	}

	out := Sha256{}
	h.Sum(out[:0])
	return out, nil
}
