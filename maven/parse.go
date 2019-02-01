// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/samsarahq/go/oops"
	"github.com/szabba/unibuild/prefixio"

	"github.com/szabba/unibuild/repo"
)

func ParseEffectivePomOfClone(ctx context.Context, cln repo.Local) (EffectivePom, error) {
	tmpfile, err := ioutil.TempFile("", "effective-pom-*.xml")
	if err != nil {
		return EffectivePom{}, err
	}
	defer os.Remove(tmpfile.Name())
	err = tmpfile.Close()
	if err != nil {
		return EffectivePom{}, err
	}

	// TODO: on Windows the path might contain spaces...
	err = writeEffectivePomTo(ctx, cln, tmpfile.Name())
	if err != nil {
		return EffectivePom{}, err
	}

	tmpfile, err = os.Open(tmpfile.Name())
	if err != nil {
		return EffectivePom{}, err
	}
	defer tmpfile.Close()

	return ParseEffectivePom(tmpfile)
}

func writeEffectivePomTo(ctx context.Context, cln repo.Local, dst string) error {
	cmd := exec.CommandContext(
		ctx,
		"mvn", "org.apache.maven.plugins:maven-help-plugin:3.1.0:effective-pom",
		"-Doutput="+dst)
	cmd.Dir = cln.Path
	cmd.Stdout = prefixio.NewWriter(os.Stdout, cln.Name+" | ")
	cmd.Stderr = prefixio.NewWriter(os.Stdout, cln.Name+" | ")
	return cmd.Run()
}

func ParseEffectivePom(r io.Reader) (EffectivePom, error) {
	buf := new(bytes.Buffer)
	multi, errMulti := parseMultiModuleProject(io.TeeReader(r, buf))
	single, errSingle := parseSingleModuleProject(buf)

	if errMulti == nil {
		return multi, nil
	}
	if errSingle == nil {
		return single, nil
	}
	return EffectivePom{}, oops.Errorf("cannot parse effective POM")
}

func parseMultiModuleProject(r io.Reader) (EffectivePom, error) {
	var pom EffectivePom
	dec := xml.NewDecoder(r)
	err := dec.Decode(&pom)
	if err != nil {
		return EffectivePom{}, err
	}
	if len(pom.Projects) == 0 {
		return EffectivePom{}, oops.Errorf("not a multimodule project")
	}
	return pom, nil
}

func parseSingleModuleProject(r io.Reader) (EffectivePom, error) {
	var mod EffectiveModule
	dec := xml.NewDecoder(r)
	err := dec.Decode(&mod)
	if err != nil {
		return EffectivePom{}, err
	}
	pom := EffectivePom{
		Projects: []EffectiveModule{mod},
	}
	return pom, nil
}
