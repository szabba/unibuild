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

type Header struct {
	Parent Identity `xml:"parent"`
	Identity
}

type Identity struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

func ParseHeaderFromPath(path string) (Header, error) {
	f, err := os.Open(path)
	if err != nil {
		return Header{}, oops.Wrapf(err, "cannot open file to parse version from")
	}
	defer f.Close()
	return ParseHeader(f)
}

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

func (head Header) EffectiveIdentity() Identity {
	return Identity{
		GroupID:    head.EffectiveGroupID(),
		ArtifactID: head.EffectiveArtifactID(),
		Version:    head.EffectiveVersion(),
	}
}

func (head Header) EffectiveGroupID() string {
	if head.GroupID == "" {
		return head.Parent.GroupID
	}
	return head.GroupID
}

func (head Header) EffectiveArtifactID() string {
	if head.ArtifactID == "" {
		return head.Parent.ArtifactID
	}
	return head.ArtifactID
}

func (head Header) EffectiveVersion() string {
	if head.Version == "" {
		return head.Parent.Version
	}
	return head.Version
}
