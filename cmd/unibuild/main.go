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
	gitlab "github.com/xanzy/go-gitlab"

	"github.com/szabba/unibuild"
	"github.com/szabba/unibuild/binhash"
	"github.com/szabba/unibuild/filterparser"
	"github.com/szabba/unibuild/multimaven"
	"github.com/szabba/unibuild/repo"
)

const (
	DefaultBaseURL = "https://gitlab.com/"
)

func main() {
	hash, err := binhash.OwnHash()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("running binary hash: %x", hash)

	flags := new(Flags)
	flags.Parse()

	if flags.logUTC {
		log.SetFlags(log.Flags() | log.LUTC)
	}

	repos, err := getRepos(flags.baseURL, flags.authToken, flags.group)
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
	err = runBuild(ctx, repos, flags)
	log.Printf("build took %s", time.Now().Sub(start))

	if err != nil {
		log.Fatalf("build failed: %s", err)
	}
	log.Printf("build ok")
}

type Flags struct {
	baseURL   string
	logUTC    bool
	timeout   time.Duration
	branches  CommaList
	authToken string
	group     string
	filters   []unibuild.Filter
}

func (fs *Flags) Parse() {
	flag.BoolVar(&fs.logUTC, "log-utc", false, "when present, the time in logs is in UTC (local otherwise)")
	flag.DurationVar(&fs.timeout, "timeout", time.Duration(0), "the timeout for the build (ignored if <= 0)")
	flag.StringVar(&fs.baseURL, "base-url", DefaultBaseURL, "gitlab API base URL (must end with /)")
	flag.StringVar(&fs.authToken, "auth-token", "", "gitlab API authentication token (required)")
	flag.StringVar(&fs.group, "group", "", "gitlab group to clone repositories from (required)")
	flag.Var(&fs.branches, "branches", "comma-separated list of branches to try checking out")
	fs.branches.Set("master")

	flag.Parse()

	noAuthToken := fs.authToken == ""
	noGroup := fs.group == ""

	if noAuthToken {
		fs.fail("an authentication token needs to be specified")
	}
	if noGroup {
		fs.fail("a gitlab group needs to be specified")
	}

	err := fs.parseFilters()
	if err != nil {
		fs.fail(err.Error())
	}
}

func (fs *Flags) parseFilters() error {
	builder := filterparser.NewBuilder()
	filters, err := filterparser.Parse(builder, flag.Args()...)
	if err != nil {
		return err
	}
	fs.filters = filters
	return nil
}

func (fs *Flags) fail(message string) {
	fmt.Println(message)
	fmt.Println()
	flag.Usage()
	os.Exit(1)
}

func runBuild(ctx context.Context, repos *repo.Set, flags *Flags) error {
	clones, err := repo.SyncAll(ctx, repos, ".")
	if err != nil {
		return oops.Wrapf(err, "problem syncing repos")
	}

	err = clones.EachTry(func(l repo.Local) error {
		return l.CheckoutFirst(ctx, flags.branches.list[0], flags.branches.list[1:]...)
	})
	if err != nil {
		return oops.Wrapf(err, "problem checking out appropriate branches")
	}

	prjs, err := analyzeProjects(ctx, clones)
	if err != nil {
		return oops.Wrapf(err, "problem analyzing projects")
	}

	ps := unibuild.NewProjectSuite(prjs...)
	ordSuite, err := ps.ResolveOrder()
	if err != nil {
		return oops.Wrapf(err, "problem finding build order")
	}

	filterSuite := ordSuite.Filter(flags.filters...)

	for _, p := range filterSuite.Order() {
		err := p.Build(ctx, os.Stdout)
		if err != nil {
			return oops.Wrapf(err, "problem building project %s", p.Info().Name)
		}
	}

	return nil
}

func getRepos(baseURL, authToken, name string) (*repo.Set, error) {
	cli := gitlab.NewClient(nil, authToken)
	err := cli.SetBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
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

func analyzeProjects(ctx context.Context, clones *repo.ClonedSet) ([]unibuild.Project, error) {
	prjs := make([]unibuild.Project, 0, clones.Size())
	err := clones.EachTry(func(cln repo.Local) error {
		p, err := multimaven.NewProject(ctx, cln)
		if err != nil {
			log.Printf("problem analyzing project in repo at %s: %s", cln.Path, err)
			return nil
		}
		prjs = append(prjs, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return prjs, nil
}
