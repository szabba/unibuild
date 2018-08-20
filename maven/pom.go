// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import (
	"encoding/xml"
	"errors"
	"io"
	"os"

	"github.com/samsarahq/go/oops"
)

var (
	errNoGroupID    = errors.New("no groupId")
	errNoArtifactID = errors.New("no artifactId")
	errNoVersion    = errors.New("no version")
)

// A Header corresponds to the parts of a POM that determine the identity of a maven module.
type Header struct {
	Parent Identity `xml:"parent"`
	Identity
}

// An Identity of a maven module.
type Identity struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// ParseHeaderFromPath parses the Header from a POM with the given path.
// Parsing will fail if the POM is missing information required to compute an Identity.
// The path should locate the POM file, not the directory it resides in.
func ParseHeaderFromPath(path string) (Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return Header{}, oops.Wrapf(err, "cannot open file to parse version from")
	}
	defer f.Close()
	return ParseHeader(f)
}

// ParseHeader parses the Header of a POM from the given io.Reader.
// Parsing will fail if the POM is missing information required to compute an Identity.
func ParseHeader(r io.Reader) (Header, error) {
	head := Header{}
	err := xml.NewDecoder(r).Decode(&head)
	if err != nil {
		return Header{}, err
	}

	if head.Version == "" && head.Parent.Version == "" {
		return Header{}, errNoVersion
	}
	if head.ArtifactID == "" && head.Parent.ArtifactID == "" {
		return Header{}, errNoArtifactID
	}
	if head.GroupID == "" && head.Parent.GroupID == "" {
		return Header{}, errNoGroupID
	}

	return head, nil
}

// EffectiveIdentity calculates the effective identity of a module based on it's Header.
func (head Header) EffectiveIdentity() Identity {
	return Identity{
		GroupID:    head.EffectiveGroupID(),
		ArtifactID: head.EffectiveArtifactID(),
		Version:    head.EffectiveVersion(),
	}
}

// EffectiveGroupID is the groupId of a POM.
// If a groupId was not specified explicitly, the parent one is used.
func (head Header) EffectiveGroupID() string {
	if head.GroupID == "" {
		return head.Parent.GroupID
	}
	return head.GroupID
}

// EffectiveArtifactID is the artifactId of a POM.
// If an artifactId was not specified explicitly, the parent one is used.
func (head Header) EffectiveArtifactID() string {
	if head.ArtifactID == "" {
		return head.Parent.ArtifactID
	}
	return head.ArtifactID
}

// EffectiveVersion is the version of a POM.
// If a version was not specified explicitly, the parent one is used.
func (head Header) EffectiveVersion() string {
	if head.Version == "" {
		return head.Parent.Version
	}
	return head.Version
}
