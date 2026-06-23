package nativeext_test

import (
	"context"
	"testing"

	"go.rgst.io/jaredallard/slogext/v2"
	"go.rgst.io/jaredallard/vcs/v2/resolver"
	"go.rgst.io/stencil/v2/internal/modules/nativeext"
	"gotest.tools/v3/assert"
)

// TestCanImportNativeExtension ensures that we can resolve a repo's
// version, download the release, extract it, register it with the
// extension host, and then finally execute a template function provided
// by the extension.
func TestCanImportNativeExtension(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ext, err := nativeext.NewHost(slogext.NewTestLogger(t))
	assert.NilError(t, err, "expected NewHost to not fail")
	defer ext.Close()

	ver, err := resolver.NewResolver().Resolve(ctx, "https://github.com/rgst-io/stencil-golang", &resolver.Criteria{
		Constraint: "=1.1.0",
	})
	assert.NilError(t, err, "failed to resolve version")

	err = ext.RegisterExtension(ctx, "https://github.com/rgst-io/stencil-golang", "github.com/rgst-io/stencil-golang", ver)
	assert.NilError(t, err, "failed to register extension")

	caller, err := ext.GetExtensionCaller(ctx)
	assert.NilError(t, err, "failed to get extension caller")

	resp, err := caller.Call("github.com/rgst-io/stencil-golang.GetLicense", "GPL-3.0")
	assert.NilError(t, err, "failed to call extension")
	assert.Assert(t, resp != "")
}
