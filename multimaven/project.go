// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package multimaven

import (
	"context"
	"io"
	"log"
	"os/exec"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/pom"
	"github.com/szabba/uninbuild/repo"
)

type Project struct {
	name     string
	version  string
	dir      string
	deps     []unibuild.ProjectInfo
	pomHeads []pom.Header
}

var _ unibuild.Project = Project{}

func NewProject(ctx context.Context, clone repo.Local) (Project, error) {
	pomHeads, err := pom.Scan(ctx, clone.Path)
	if err != nil {
		return Project{}, oops.Wrapf(err, "failed scanning for maven modules in %s", clone.Path)
	}

	prj := Project{
		name:     clone.Name,
		version:  pomHeads[0].EffectiveVersion(),
		dir:      clone.Path,
		pomHeads: pomHeads,
	}
	return prj, nil
}

func (prj Project) Info() unibuild.ProjectInfo {
	return unibuild.ProjectInfo{
		Name:    prj.name,
		Version: prj.version,
	}
}

func (prj Project) Deps() []unibuild.ProjectInfo { return prj.deps }

func (prj Project) POMHeaders() []pom.Header { return prj.pomHeads }

func (prj Project) FindDeps(prjs []Project) (Project, error) {
	// TODO
	log.SetFlags(log.Flags() | log.Lshortfile)
	log.Panic("Project.FindDeps not implemented yet")
	return Project{}, nil
}

func (prj Project) Build(ctx context.Context, arts map[unibuild.ProjectInfo][]unibuild.Artifact, logTo io.Writer) ([]unibuild.Artifact, error) {
	cmd := exec.CommandContext(ctx, "mvn", "clean", "deploy")
	cmd.Dir = prj.dir
	cmd.Stdout = logTo
	cmd.Stderr = logTo
	err := cmd.Run()
	return nil, err
}
