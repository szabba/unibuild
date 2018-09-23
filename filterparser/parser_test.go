// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package filterparser_test

//go:generate mockgen -source parser.go -destination mock_builder_test.go -package filterparser_test FiltersBuilder

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/szabba/assert"

	"github.com/szabba/unibuild/filterparser"
)

func TestOneProject(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	gomock.InOrder(
		builder.EXPECT().Include("A"),
		builder.EXPECT().Build())

	// when
	_, err := filterparser.Parse(builder, "A")

	// then
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
}

func TestTwoProjects(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	gomock.InOrder(
		builder.EXPECT().Include("A"),
		builder.EXPECT().Include("B"),
		builder.EXPECT().Build())

	// when
	_, err := filterparser.Parse(builder, "A", "B")

	// then
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
}

func TestDepsModifierMustNotBeFirst(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	// when
	_, err := filterparser.Parse(builder, "+deps")

	// then
	assert.That(err != nil, t.Errorf, "got no error, while one was expected")
}

func TestDependentModifierMustNotBeFirst(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	// when
	_, err := filterparser.Parse(builder, "+dependent")

	// then
	assert.That(err != nil, t.Errorf, "got no error, while one was expected")
}

func TestExcludeModifierMustNotBeFirst(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	// when
	_, err := filterparser.Parse(builder, "+exclude")

	// then
	assert.That(err != nil, t.Errorf, "got no error, while one was expected")
}

func TestProjectWithDeps(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	gomock.InOrder(
		builder.EXPECT().Include("A"),
		builder.EXPECT().WithDeps("A"),
		builder.EXPECT().Build())

	// when
	_, err := filterparser.Parse(builder, "A", "+deps")

	// then
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
}

func TestProjectWithDependent(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	gomock.InOrder(
		builder.EXPECT().Include("A"),
		builder.EXPECT().WithDependents("A"),
		builder.EXPECT().Build())

	// when
	_, err := filterparser.Parse(builder, "A", "+dependent")

	// then
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
}

func TestProjectExcluded(t *testing.T) {
	// given
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builder := NewMockFiltersBuilder(ctrl)

	gomock.InOrder(
		builder.EXPECT().Include("A"),
		builder.EXPECT().Exclude("A"),
		builder.EXPECT().Build())

	// when
	_, err := filterparser.Parse(builder, "A", "+exclude")

	// then
	assert.That(err == nil, t.Errorf, "unexpected error: %s", err)
}
