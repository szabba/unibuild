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
		ctx, cancel = context.WithTimeout(ctx, flags.timeout)
		defer cancel()
	}

	start := time.Now()

	clones, err := repo.SyncAll(ctx, repos, ".")
	if err != nil {
		log.Fatalf("problem syncing repos: %s", err)
	}

	err = clones.EachTry(func(l repo.Local) error {
		return l.CheckoutFirst(ctx, flags.branches.topic, flags.branches.default_)
	})
	if err != nil {
		log.Fatalf("problem checking out appropriate branches: %s", err)
	}

	space, err := multimaven.NewWorkspace(ctx, clones)
	if err != nil {
		log.Fatalf("problem building workspace: %s", err)
	}

	prjs := space.Projects()

	ws := unibuild.NewWorkspace(prjs)
	order, err := ws.FindBuildOrder()
	if err != nil {
		log.Fatalf("problem finding build order: %s", err)
	}

	err = runBuild(ctx, order)
	log.Printf("build took %s", time.Now().Sub(start))

	if err != nil {
		log.Fatalf("build failed: %s", err)
	}
	log.Printf("build ok")
}

type Flags struct {
	logUTC    bool
	timeout   time.Duration
	authToken string
	group     string
	branches  struct {
		topic    string
		default_ string
	}
}

func (fs *Flags) Parse() {
	flag.BoolVar(&fs.logUTC, "log-utc", false, "when present, the time in logs is in UTC (local otherwise)")
	flag.DurationVar(&fs.timeout, "timeout", time.Duration(0), "the timeout for the build (ignored if <= 0)")
	flag.StringVar(&fs.authToken, "auth-token", "", "gitlab API authentication token (required)")
	flag.StringVar(&fs.group, "group", "", "gitlab group to clone repositories from (required)")
	flag.StringVar(&fs.branches.topic, "topic-branch", "", "topic branch to checkout, if available (ignored when empty)")
	flag.StringVar(&fs.branches.default_, "default-branch", "master", "the branch to default to when the topic branch is not used")

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

func runBuild(ctx context.Context, prjs []unibuild.Project) error {
	for _, p := range prjs {
		err := p.Build(ctx, os.Stdout)
		if err != nil {
			return oops.Wrapf(err, "problem building project %s", p.Info().Name)
		}
		log.Printf("succesfully built %s", p.Info().Name)
	}
	return nil
}
