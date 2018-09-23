// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filterparser

import (
	"github.com/szabba/unibuild"
)

type builder struct {
	filters []unibuild.Filter
}

var _ FiltersBuilder = new(builder)

func NewBuilder() FiltersBuilder { return new(builder) }

func (b *builder) Build() []unibuild.Filter {
	filters := make([]unibuild.Filter, 0, len(b.filters))
	return append(filters, b.filters...)
}

func (b *builder) Include(project string) {
	b.append(unibuild.Exactly(project))
}

func (b *builder) WithDeps(project string) {
	b.append(unibuild.WithDeps(project))
}

func (b *builder) WithDependents(project string) {
	b.append(unibuild.WithDependents(project))
}

func (b *builder) Exclude(project string) {
	b.append(unibuild.Exclude(project))
}

func (b *builder) append(f unibuild.Filter) {
	b.filters = append(b.filters, f)
}
