package nativeext_test

import (
	"context"
	"os"
	"testing"

	"github.com/jaredallard/cmdexec"
	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/v2/internal/modules/nativeext"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

// TestCanImportNativeExtension ensures that we can resolve a repo's
// version, download the release, extract it, register it with the
// extension host, and then finally execute a template function provided
// by the extension.
func TestCanImportNativeExtension(t *testing.T) {
	if os.Getenv("CI") == "true" {
		originalGITHUB := os.Getenv("GITHUB_TOKEN")
		originalGH := os.Getenv("GH_TOKEN")

		assert.NilError(t, os.Unsetenv("GITHUB_TOKEN"))
		assert.NilError(t, os.Unsetenv("GH_TOKEN"))

		// restore later
		t.Cleanup(func() {
			assert.NilError(t, os.Setenv("GITHUB_TOKEN", originalGITHUB))
			assert.NilError(t, os.Setenv("GH_TOKEN", originalGH))
		})
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ext, err := nativeext.NewHost(slogext.NewTestLogger(t))
	assert.NilError(t, err, "expected NewHost to not fail")
	defer ext.Close()

	ver, err := resolver.NewResolver().Resolve(ctx, "https://github.com/rgst-io/stencil-golang", &resolver.Criteria{
		Constraint: "=2.0.1",
	})
	assert.NilError(t, err, "failed to resolve version")

	// Set an empty token so that we don't try to use it for fetching the
	// license data. Github Actions tokens don't have access to read
	// licenses.
	t.Setenv("GITHUB_TOKEN", "")
	cmdexec.UseMockExecutor(t, cmdexec.NewMockExecutor(&cmdexec.MockCommand{
		Name:   "gh",
		Args:   []string{"auth", "token"},
		Stdout: []byte("xyz"),
	}))

	err = ext.RegisterExtension(ctx, "https://github.com/rgst-io/stencil-golang", "github.com/rgst-io/stencil-golang", ver)
	assert.NilError(t, err, "failed to register extension")

	caller, err := ext.GetExtensionCaller(ctx)
	assert.NilError(t, err, "failed to get extension caller")

	resp, err := caller.Call("github.com/rgst-io/stencil-golang.GetLicense", "GPL-3.0")
	assert.NilError(t, err, "failed to call extension")
	assert.Assert(t, resp != "")
}
