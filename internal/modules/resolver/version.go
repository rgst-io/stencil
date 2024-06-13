// Copyright (C) 2024 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resolver

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Version represents a version found in a Git repository. Versions are
// only discovered if a tag or branch points to a commit (individual
// commits will never be automatically discovered unless they are
// manually passed in).
type Version struct {
	// Commit is the underlying commit hash for this version.
	Commit string `yaml:"commit,omitempty"`

	// Tag is the underlying tag for this version, if set.
	Tag string `yaml:"tag,omitempty"`
	sv  *semver.Version

	// Virtual is version that was injected either through the local
	// file-system or through a replacement. The value of this is set
	// depending on the context of how this was set. A virtual version is
	// never returned from a resolver.
	Virtual string `yaml:"virtual,omitempty"`

	// Branch is the underlying branch for this version, if set.
	Branch string `yaml:"branch,omitempty"`
}

// Equal returns true if the two versions are equal.
func (v *Version) Equal(other *Version) bool {
	// If either is nil, they must both be nil.
	if v == nil || other == nil {
		return v == other
	}

	// Otherwise, check all fields.
	return v.Commit == other.Commit && v.Tag == other.Tag && v.Branch == other.Branch
}

// String is a user-friendly representation of the version that can be
// used in error messages.
func (v *Version) String() string {
	switch {
	case v.Virtual != "":
		return fmt.Sprintf("virtual (source: %s)", v.Virtual)
	case v.Tag != "":
		return fmt.Sprintf("tag %s (%s)", v.Tag, v.Commit)
	case v.Branch != "":
		return fmt.Sprintf("branch %s (%s)", v.Branch, v.Commit)
	default:
		return v.Commit
	}
}

// GitRef returns a Git reference that can be used to check out the
// version.
func (v *Version) GitRef() string {
	switch {
	// TODO(jaredallard): This will require native ext handling.
	case v.Virtual != "":
		return "NOT_A_VALID_GIT_VERSION"
	case v.Tag != "":
		return "refs/tags/" + v.Tag
	case v.Branch != "":
		return "refs/heads/" + v.Branch
	default:
		return v.Commit
	}
}
