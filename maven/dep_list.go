// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/samsarahq/go/oops"
)

const (
	mavenOutputPrefix  = "[INFO]"
	mavenDepListHeader = "The following files have been resolved:"

	groupIDIx    = 0
	artifactIDIx = 1
	versionIx    = 3
	depChunks    = 5
)

var (
	errNoDepListHeader   = errors.New("could not find a dependency list header")
	errInvalidDepListing = errors.New("dependency list line invalid")
	errDepListOver       = errors.New("no more dependencies in list")

	newline = regexp.MustCompile("\n|\r\n")
)

// ListDeps finds the dependencies of all the modules inside a multi-module project directory.
// The results reflect how maven will resolve the dependencies.
func ListDeps(ctx context.Context, path string, mods []Identity) ([]Identity, error) {
	all := map[Identity]bool{}
	for _, m := range mods {
		deps, err := ListModuleDeps(ctx, path, m)
		if err != nil {
			return nil, err
		}
		for _, d := range deps {
			all[d] = true
		}
	}

	out := make([]Identity, 0, len(all))
	for d := range all {
		out = append(out, d)
	}
	return out, nil
}

// ListModuleDeps finds the dependencies of a module inside a multi-module project directory.
// The results reflect how maven will resolve the dependencies.
func ListModuleDeps(ctx context.Context, dir string, mod Identity) ([]Identity, error) {
	modRef := mod.GroupID + ":" + mod.ArtifactID
	cmd := exec.CommandContext(ctx, "mvn", "dependency:list", "-pl", modRef)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, oops.Wrapf(err, "problem listing maven dependencies in project at %s, module %s", dir, modRef)
	}
	lines := newline.Split(string(out), -1)
	return newDepListParser(lines).parse()
}

type depListParser struct {
	lines []string
	deps  []Identity
}

func newDepListParser(lines []string) *depListParser {
	return &depListParser{
		lines: lines,
		deps:  make([]Identity, 0, len(lines)),
	}
}

func (p *depListParser) parse() ([]Identity, error) {
	err := p.skipPastHeader()
	if err != nil {
		return nil, err
	}
	return p.parseAll()
}

func (p *depListParser) skipPastHeader() error {
	for len(p.lines) > 1 {
		l := p.lines[0]
		p.lines = p.lines[1:]
		if strings.Contains(l, mavenDepListHeader) {
			return nil
		}
	}
	return errNoDepListHeader
}

func (p *depListParser) parseAll() ([]Identity, error) {
	deps := make([]Identity, 0, len(p.lines))
	for len(p.lines) > 1 {
		err := p.parseOne()
		if err == errDepListOver {
			return p.deps, nil
		}
		if err != nil {
			return nil, err
		}
	}
	return deps, nil
}

func (p *depListParser) parseOne() error {
	line := p.lines[0]
	line = strings.TrimPrefix(line, mavenOutputPrefix)
	line = strings.TrimSpace(line)

	if line == "" || line == "none" {
		return errDepListOver
	}

	chunks := strings.Split(line, ":")

	if len(chunks) != depChunks {
		return oops.Errorf("dependency format invalid: %s", line)
	}

	id := Identity{
		GroupID:    chunks[groupIDIx],
		ArtifactID: chunks[artifactIDIx],
		Version:    chunks[versionIx],
	}
	p.lines, p.deps = p.lines[1:], append(p.deps, id)
	return nil
}
