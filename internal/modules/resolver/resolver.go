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

// Package resolver implements a way to resolve versions using a set of
// criteria. Note that only semantic versioning is supported for tags.
package resolver

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"go.rgst.io/stencil/internal/git"
)

// Resolver is an instance of a version resolver that resolves versions
// based on the provided criteria. Version lists are fetched exactly
// once and are cached for the lifetime of the resolver.
type Resolver struct {
	// versions is a map of URIs to versions that have been fetched.
	versions map[string][]Version

	// versionsMu is a mutex that protects the versions map, allowing
	// for concurrent access.
	versionsMu sync.Mutex
}

// NewResolver creates a new resolver instance.
func NewResolver() *Resolver {
	return &Resolver{
		versions: make(map[string][]Version),
	}
}

// fetchVersionsIfNecessary fetches versions for the provided URI if not
// already fetched. If versions are already fetched, they are returned
// immediately.
func (r *Resolver) fetchVersionsIfNecessary(ctx context.Context, uri string) ([]Version, error) {
	// Prevent anything else from reading/writing while we're determining
	// if we need to fetch or write new versions. This ensures that we
	// never accidentally write to the same block twice, since only one
	// would ever be able to determine if it needs to fetch or not.
	r.versionsMu.Lock()
	defer r.versionsMu.Unlock()

	if r.versions == nil {
		r.versions = make(map[string][]Version)
	}

	// We have it already, noop.
	if versions, ok := r.versions[uri]; ok {
		return versions, nil
	}

	// Fetch versions for the URI.
	remoteStrs, err := git.ListRemote(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes for %s: %w", uri, err)
	}

	versions := make([]Version, 0)
	for _, remoteStr := range remoteStrs {
		if len(remoteStr) != 2 {
			continue
		}

		commit := remoteStr[0]
		ref := remoteStr[1]
		switch {
		case strings.HasPrefix(ref, "refs/tags/"):
			if strings.HasSuffix(ref, "^{}") {
				// Skip annotated tags.
				continue
			}

			tag := strings.TrimPrefix(ref, "refs/tags/")
			sv, err := semver.NewVersion(tag)
			if err != nil {
				// Skip tags that do not follow semantic versioning. We do not
				// support them.
				continue
			}

			versions = append(versions, Version{
				Commit: commit,
				Tag:    tag,
				sv:     sv,
			})
		case strings.HasPrefix(ref, "refs/heads/"):
			branch := strings.TrimPrefix(ref, "refs/heads/")
			versions = append(versions, Version{
				Commit: commit,
				Branch: branch,
			})
		default:
			continue
		}
	}

	// Write the versions to the cache.
	r.versions[uri] = versions

	return versions, nil
}

// Resolve returns the latest version matching the provided criteria.
// If multiple criteria are provided, the version must satisfy all of
// them. If no versions are found, an error is returned.
//
// TODO(jaredallard): Return resolution errors as a type that can be
// unwrapped for getting information about why it failed.
func (r *Resolver) Resolve(ctx context.Context, uri string, criteria ...*Criteria) (*Version, error) {
	if len(criteria) == 0 {
		return nil, fmt.Errorf("no criteria provided")
	}

	// Parse the criteria so we can call Check() later, but also to see if
	// we have any "wins once" criteria (prerelease track and branches).
	var prerelease string
	var branch string
	for _, criterion := range criteria {
		if criterion.Branch != "" {
			if branch != "" && branch != criterion.Branch {
				return nil, fmt.Errorf("unable to satisfy multiple branch constraints (%s, %s)", branch, criterion.Branch)
			}

			branch = criterion.Branch
		}

		if err := criterion.Parse(); err != nil {
			return nil, fmt.Errorf("failed to parse criteria: %w", err)
		}

		// See if pre-releases are included in any of the provided
		// constraints.
		if criterion.c != nil && criterion.prerelease != "" {
			if prerelease != "" && prerelease != criterion.prerelease {
				return nil, fmt.Errorf(
					"unable to satisfy multiple pre-release constraints (%s, %s)", prerelease, criterion.prerelease,
				)
			}

			prerelease = criterion.prerelease
		}
	}

	versions, err := r.fetchVersionsIfNecessary(ctx, uri)
	if err != nil {
		return nil, err
	}

	// Sort the versions by semantic versioning. Branches are always at
	// the end of the list because we only want to consider them if no
	// tags are available.
	sort.Slice(versions, func(i, j int) bool {
		// Tags are always at the beginning of the list and are sorted by
		// version.
		if versions[i].sv != nil && versions[j].sv != nil {
			return versions[i].sv.GreaterThan(versions[j].sv)
		}

		// Branches are always at the end of the list.
		if versions[i].sv != nil {
			return true
		}
		if versions[j].sv != nil {
			return false
		}

		// Both are branches, sort by branch name just for predictability.
		return versions[i].Branch < versions[j].Branch
	})

	// If we have pre-releases, then we need to make sure that none of the
	// criteria's are failing due to pre-releases _not_ being included.

	// Find the latest version that satisfies all criteria.
	var latest *Version
	for i := range versions {
		version := &versions[i]

		var satisfied bool
		for _, criterion := range criteria {
			satisfied = criterion.Check(version, prerelease, branch)
			if !satisfied {
				break
			}
		}
		if satisfied {
			// We found a version that satisfies all criteria, return it
			// because we already sorted the list and know it's the best
			// possible version.
			latest = version
			break
		}
	}
	if latest != nil {
		return latest, nil
	}

	return nil, fmt.Errorf("no versions found that satisfy criteria")
}
