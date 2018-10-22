// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/samsarahq/go/oops"
)

var newlinePattern = regexp.MustCompile("\n|\r\n")

// Scan parses the Headers of the POM in dir and all of it's submodules recursively.
// If no error occurrs, the returned slice is non-empty.
func Scan(ctx context.Context, dir string) ([]Header, error) {
	paths, err := scanPaths(ctx, dir)
	if err != nil {
		return nil, oops.Wrapf(err, "cannot scan for POMs in %s", dir)
	}

	headers := make([]Header, 0, len(paths))
	for _, path := range paths {
		hdr, err := ParseHeaderFromPath(path)
		if err != nil {
			return nil, oops.Wrapf(err, "cannot parse header from file %s", path)
		}
		headers = append(headers, hdr)
	}

	return headers, nil
}

func scanPaths(ctx context.Context, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "mvn", "-q", "--also-make", "exec:exec", "-Dexec.executable="+_WorkDirCmd)
	cmd.Dir = path
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, oops.Wrapf(err, "cannot list POMs in directory %s", path)
	}

	lines := newlinePattern.Split(string(out), -1)
	poms := make([]string, 0, len(lines))
	for _, l := range lines {
		if l == "" {
			continue
		}
		absPom := filepath.Join(l, "pom.xml")
		poms = append(poms, absPom)
	}

	if len(poms) < 1 {
		return poms, oops.Errorf("no POMs found in directory %s", path)
	}
	return poms, nil
}
