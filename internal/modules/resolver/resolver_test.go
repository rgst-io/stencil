//go:build !test_no_internet

package resolver_test

import (
	"context"
	"testing"

	"go.rgst.io/stencil/internal/modules/resolver"
	"gotest.tools/v3/assert"
)

// TestDoesntSupportComplexConstraints ensures that the resolver does
// not support complex constraints. Specifically, it does not support
// constraints that include boolean operators.
func TestDoesntSupportComplexConstraints(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	_, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil",
		&resolver.Criteria{
			Constraint: ">=1.0.0 && <1.23.1",
		},
	)
	assert.ErrorContains(t, err, "failed to parse criteria: complex constraints are not supported")
}

func TestResolverErrorsIfNotCriteria(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	_, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil")
	assert.ErrorContains(t, err, "no criteria provided")
}

// TestReturnsTheLatestVersions ensures that the resolver returns the
// latest version.
func TestReturnsTheLatestVersions(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	v, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil",
		&resolver.Criteria{
			Constraint: ">0.0.0",
		},
		&resolver.Criteria{
			Constraint: "<0.2.0",
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, v.Tag, "v0.1.0")
}

// TestCanResolvePrereleases ensures that the resolver successfully
// considers pre-releases when asked through the criteria.
func TestCanResolvePrereleases(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	// TODO(jaredallard): When we have in-memory testing for this, or a
	// repo that has pre-releases, we should test this on a different
	// repository. For now, the test isn't fragile, but still sucks to use
	// something we don't control.
	v, err := r.Resolve(ctx, "https://github.com/getoutreach/stencil-golang",
		&resolver.Criteria{
			Constraint: ">=1.23.0",
		},
		&resolver.Criteria{
			Constraint: "=1.23.1-rc.1",
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, v.Tag, "v1.23.1-rc.1")
}

// TestDoesNotConsiderPrereleasesWhenNotAsked ensures that the resolver
// does not consider pre-releases when not asked through the criteria.
func TestDoesNotConsiderPrereleasesWhenNotAsked(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	// TODO(jaredallard): When we have in-memory testing for this, or a
	// repo that has pre-releases, we should test this on a different
	// repository. For now, the test isn't fragile, but still sucks to use
	// something we don't control.
	v, err := r.Resolve(ctx, "https://github.com/getoutreach/stencil-golang",
		&resolver.Criteria{
			Constraint: ">=1.23.0",
		},
		&resolver.Criteria{
			Constraint: "<1.23.1",
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, v.Tag, "v1.23.0")
}

// TestSupportsComparingPrereleaseVersions ensures that the resolver
// supports comparing -rc.1, -rc.2, etc. versions.
func TestSupportsComparingPrereleaseVersions(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	// TODO(jaredallard): When we have in-memory testing for this, or a
	// repo that has pre-releases, we should test this on a different
	// repository. For now, the test isn't fragile, but still sucks to use
	// something we don't control.
	v, err := r.Resolve(ctx, "https://github.com/getoutreach/stencil-golang",
		&resolver.Criteria{
			Constraint: ">=1.23.1-rc.0",
		},
		&resolver.Criteria{
			Constraint: "=1.23.1-rc.1",
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, v.Tag, "v1.23.1-rc.1")
}

// TestDoesntSupportMultiplePrereleases ensures that the resolver does
// not support multiple pre-releases tracks, because we cannot compare
// those versions.
func TestDoesntSupportMultiplePrereleases(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	_, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil",
		&resolver.Criteria{
			Constraint: "=1.23.1-rc.1",
		},
		&resolver.Criteria{
			Constraint: "=1.23.1-alpha.1",
		},
	)
	assert.ErrorContains(t, err, "unable to satisfy multiple pre-release constraints (rc, alpha)")
}

// TestUsesBranchOverConstraints ensures that the resolver ranks
// branches higher than constraints.
func TestUsesBranchOverConstraints(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	v, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil",
		&resolver.Criteria{
			Branch: "main",
		},
		&resolver.Criteria{
			Constraint: ">=1.0.0",
		},
	)
	assert.NilError(t, err)
	assert.Equal(t, v.Branch, "main")
}

// TestCannotMixBranches ensures that the resolver does not support
// mixing branches.
func TestCannotMixBranches(t *testing.T) {
	ctx := context.Background()

	r := new(resolver.Resolver)

	_, err := r.Resolve(ctx, "https://github.com/rgst-io/stencil",
		&resolver.Criteria{
			Branch: "main",
		},
		&resolver.Criteria{
			Branch: "master",
		},
	)
	assert.ErrorContains(t, err, "unable to satisfy multiple branch constraints (main, master)")
}
