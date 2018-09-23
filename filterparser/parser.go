// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filterparser

import (
	"errors"
	"fmt"

	"github.com/samsarahq/go/oops"

	"github.com/szabba/unibuild"
)

const (
	DepsToken      = "+deps"
	DependentToken = "+dependent"
	ExcludeToken   = "+exclude"
)

var ErrInvalidFilter = errors.New("invalid filter")

type FiltersBuilder interface {
	Include(project string)
	WithDeps(project string)
	WithDependents(project string)
	Exclude(project string)
	Build() []unibuild.Filter
}

type parserState func(builder FiltersBuilder, arg string) (parserState, error)

func Parse(builder FiltersBuilder, tokens ...string) ([]unibuild.Filter, error) {
	state, err := start, error(nil)
	for i, tok := range tokens {
		state, err = state(builder, tok)
		if err != nil {
			return nil, oops.Wrapf(err, "problem at filter token %d", i)
		}
	}

	return builder.Build(), nil
}

func start(builder FiltersBuilder, tok string) (parserState, error) {
	if isModifierToken(tok) {
		return nil, fmt.Errorf("modifier token %q must come after a project name", tok)
	}
	builder.Include(tok)
	return afterProject(tok), nil
}

func afterProject(name string) parserState {
	return func(builder FiltersBuilder, tok string) (parserState, error) {

		switch tok {
		case DepsToken:
			builder.WithDeps(name)
		case DependentToken:
			builder.WithDependents(name)
		case ExcludeToken:
			builder.Exclude(name)
		default:
			builder.Include(tok)
		}

		if isModifierToken(tok) {
			return afterProject(name), nil
		}
		return start, nil
	}
}

func isModifierToken(tok string) bool {
	return tok == DepsToken || tok == DependentToken || tok == ExcludeToken
}
