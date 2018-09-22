// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package unibuild_test

import (
	"context"
	"github.com/szabba/unibuild"
	"io"
)

type Project struct {
	Info_   unibuild.ProjectInfo
	Builds_ []unibuild.RequirementVersion
	Uses_   []unibuild.Requirement
	Err     error
}

var _ unibuild.Project = Project{}

func (prj Project) Info() unibuild.ProjectInfo                 { return prj.Info_ }
func (prj Project) Builds() []unibuild.RequirementVersion      { return prj.Builds_ }
func (prj Project) Uses() []unibuild.Requirement               { return prj.Uses_ }
func (prj Project) Build(_ context.Context, _ io.Writer) error { return prj.Err }
