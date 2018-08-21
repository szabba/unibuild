// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/samsarahq/go/oops"
	"github.com/xanzy/go-gitlab"

	"github.com/szabba/uninbuild"
	"github.com/szabba/uninbuild/multimaven"
	"github.com/szabba/uninbuild/repo"
)

func main() {
	flags := new(Flags)
	flags.Parse()

	if flags.logUTC {
		log.SetFlags(log.Flags() | log.LUTC)
	}

	repos, err := getRepos(flags.authToken, flags.group)
	if err != nil {
		log.Fatalf("problem getting repos: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if flags.timeout > 0 {
		ctx, _ = context.WithTimeout(ctx, flags.timeout)
	}

	clones, err := repo.CloneAll(ctx, repos, ".")
	if err != nil {
		clones.Clear()
		log.Fatalf("problem cloning repos: %s", err)
	}
	defer clones.Clear()

	space := multimaven.NewWorkspace("", clones)
	runBuild(ctx, space)
}

type Flags struct {
	authToken string
	group     string
	timeout   time.Duration
	logUTC    bool
}

func (fs *Flags) Parse() {
	flag.StringVar(&fs.authToken, "auth-token", "", "gitlab API authentication token (required)")
	flag.StringVar(&fs.group, "group", "", "gitlab group to clone repositories from (required)")
	flag.DurationVar(&fs.timeout, "timeout", time.Duration(0), "the timeout for the build (ignored if <= 0)")
	flag.BoolVar(&fs.logUTC, "log-utc", false, "when present, the time in logs is in UTC (local otherwise)")

	flag.Parse()

	noAuthToken := fs.authToken == ""
	noGroup := fs.group == ""

	if noAuthToken {
		fmt.Println("an authentication token needs to be specified")
	}
	if noGroup {
		fmt.Println("a gitlab group needs to be specified")
	}

	if noAuthToken || noGroup {
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func getRepos(authToken, name string) (*repo.Set, error) {
	cli := gitlab.NewClient(nil, authToken)
	group, _, err := cli.Groups.GetGroup(name)
	if err != nil {
		return nil, oops.Wrapf(err, "cannot retrieve gitlab group %s", name)
	}

	repos := repo.NewSet()
	for _, prj := range group.Projects {
		err := repos.Add(repo.Remote{
			Name: prj.Name,
			URL:  prj.SSHURLToRepo,
		})
		if err != nil {
			return nil, oops.Wrapf(err, "cannot build repository set for group %s", name)
		}
	}
	return repos, nil
}

func runBuild(ctx context.Context, ws unibuild.Workspace) {
	start := time.Now()
	_, err := unibuild.Build(ctx, ws)
	log.Printf("build took %s", time.Now().Sub(start))
	if err != nil {
		log.Fatalf("build failed: %s", err)
	}
	log.Print("build ok")
}
