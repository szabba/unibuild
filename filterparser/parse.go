// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filterparser

import (
	"errors"
	"unicode"
	"unicode/utf8"

	"github.com/samsarahq/go/oops"

	"github.com/szabba/unibuild"
)

var ErrInvalidFilter = errors.New("invalid filter")

func Parse(args ...string) ([]unibuild.Filter, error) {
	fs := make([]unibuild.Filter, 0, 2*len(args))
	err := error(nil)

	for i, arg := range args {
		fs, err = parseOne(arg, fs)
		if err != nil {
			return nil, oops.Wrapf(err, "problem parsing %d-th filter %q", i, arg)
		}
	}

	return fs, nil
}

func parseOne(arg string, fs []unibuild.Filter) ([]unibuild.Filter, error) {
	r, w := utf8.DecodeRuneInString(arg)
	rest := arg[w:]

	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return append(fs, unibuild.WithDeps(arg)), nil
	} else if r == '+' {
		return append(fs, unibuild.WithDeps(rest), unibuild.WithDependents(rest)), nil
	} else if r == '.' {
		return append(fs, unibuild.Exactly(rest)), nil
	} else if r == '/' {
		return append(fs, unibuild.Exclude(rest)), nil
	} else {
		return nil, ErrInvalidFilter
	}
}
