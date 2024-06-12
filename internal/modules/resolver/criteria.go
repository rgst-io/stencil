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
	"regexp"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
)

// Criteria represents a set of criteria that a version must satisfy to
// be able to be selected.
type Criteria struct {
	// Below are fields for internal use only. Specifically used for
	// constraint parsing and checking.
	c          *semver.Constraints
	prerelease string

	once sync.Once

	// Constraint is a semantic versioning constraint that the version
	// must satisfy.
	//
	// Example: ">=1.0.0 <2.0.0"
	Constraint string

	// Branch is the branch that the version must point to. This
	// constraint will only be satisfied if the branch currently points to
	// the commit being considered.
	//
	// If a branch is provided, it will always be used over other
	// versions. For this reason, top-level modules should only ever use
	// branches.
	Branch string
}

// Parse parses the criteria's constraint into a semver constraint. If
// the constraint is already parsed, this is a no-op.
func (c *Criteria) Parse() error {
	var err error
	c.once.Do(func() {
		if c.Constraint == "" {
			// No constraint, no need to parse.
			return
		}

		if strings.Contains(c.Constraint, "||") || strings.Contains(c.Constraint, "&&") {
			// We don't support complex constraints.
			err = fmt.Errorf("complex constraints are not supported")
			return
		}

		// Create a "version" from the constraint
		// TODO: make a variable for this regexp
		cv := regexp.MustCompile(`^[^v\d]+`).ReplaceAllString(c.Constraint, "")

		// Attempt to parse the constraint as a version for detecting
		// per-release versions.
		vc, err := semver.NewVersion(cv)
		if err == nil {
			c.prerelease = strings.Split(vc.Prerelease(), ".")[0]
		}

		c.c, err = semver.NewConstraint(c.Constraint)
		if err != nil {
			return
		}
	})

	return err
}

// Check returns true if the version satisfies the criteria. If a
// prerelease is included then the provided criteria will be mutated to
// support pre-releases as well as ensure that the prerelease string
// matches the provided version. If a branch is provided, then the
// criteria will always be satisfied unless the criteria is looking for
// a specific branch, in which case it will be satisfied only if the
// branches match.
func (c *Criteria) Check(v *Version, prerelease, branch string) bool {
	if c.Branch != "" && v.Branch == c.Branch {
		return true
	}

	// Looking for a specific branch, but we're not asking for a branch,
	// so return success because we cannot compare these versions.
	if branch != "" && c.Branch == "" {
		return true
	}

	if c.c != nil && v.sv != nil {
		if c.prerelease != "" && c.prerelease != prerelease {
			// The provided criteria has a pre-release version, but the
			// version we're checking against does not match. This means
			// that we should not consider this version.
			return false
		}

		// If we're eligible for pre-releases but our constraint doesn't
		// allow for them, then we need to change our constraint to allow
		// for pre-releases.
		if prerelease != "" && c.prerelease == "" {
			// We need to add the pre-release to the constraint.
			c.Constraint = fmt.Sprintf("%s-%s", c.Constraint, prerelease)

			// TODO: Better error handling and location for this logic since
			// doing this on every call is pretty awful and inefficient.
			var err error
			c.c, err = semver.NewConstraint(c.Constraint)
			if err != nil {
				// This should never happen since we've already parsed
				// the constraint once.
				panic(fmt.Sprintf("failed to parse constraint: %v", err))
			}
			c.prerelease = prerelease
		}

		return c.c.Check(v.sv)
	}

	// Otherwise, doesn't match.
	return false
}

// Equals returns true if the criteria is equal to the other criteria.
func (c *Criteria) Equals(other *Criteria) bool {
	if c.Constraint != other.Constraint {
		return false
	}

	if c.Branch != other.Branch {
		return false
	}

	return true
}

// String returns a user-friendly representation of the criteria.
func (c *Criteria) String() string {
	if c.Branch != "" {
		return fmt.Sprintf("branch %s", c.Branch)
	}

	return c.Constraint
}
