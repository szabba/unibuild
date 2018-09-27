// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"flag"
	"strings"
)

var ErrNoBranches = errors.New("no branches were specified")

type CommaList struct {
	list []string
}

var _ flag.Value = new(CommaList)

func (cl *CommaList) Set(s string) error {
	cl.list = strings.Split(s, ",")
	if len(cl.list) == 0 {
		return ErrNoBranches
	}
	return nil
}

func (cl *CommaList) String() string {
	return strings.Join(cl.list, ",")
}
