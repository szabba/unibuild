// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package maven

import "errors"

var (
	errNoGroupID    = errors.New("no groupId")
	errNoArtifactID = errors.New("no artifactId")
	errNoVersion    = errors.New("no version")
)

// An EffectivePom contains the interesting parts of the mvn help:effective-pom output for a multi-module project.
type EffectivePom struct {
	Projects []EffectiveModule `xml:"project"`
}

// An EffectiveModule contains the interesting parts of the mvn help:effective-pom output for a single-module project.
type EffectiveModule struct {
	Header
	Dependencies []Identity `xml:"dependencies>dependency"`
}

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

// Validate reports issues with the header value.
func (head Header) Validate() error {

	if head.Version == "" && head.Parent.Version == "" {
		return errNoVersion
	}
	if head.ArtifactID == "" && head.Parent.ArtifactID == "" {
		return errNoArtifactID
	}
	if head.GroupID == "" && head.Parent.GroupID == "" {
		return errNoGroupID
	}
	return nil
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
