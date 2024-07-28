package nativeext_test

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/internal/modules/nativeext"
	"go.rgst.io/stencil/pkg/slogext"
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

		// debugging
		cmd := exec.Command("gh", "auth", "status")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		assert.NilError(t, cmd.Run())

		// restore later
		t.Cleanup(func() {
			assert.NilError(t, os.Setenv("GITHUB_TOKEN", originalGITHUB))
			assert.NilError(t, os.Setenv("GH_TOKEN", originalGH))
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ext := nativeext.NewHost(slogext.NewTestLogger(t))
	defer ext.Close()

	ver, err := resolver.NewResolver().Resolve(ctx, "https://github.com/getoutreach/stencil-golang", &resolver.Criteria{
		Constraint: "=1.23.3",
	})
	assert.NilError(t, err, "failed to resolve version")

	err = ext.RegisterExtension(ctx, "https://github.com/getoutreach/stencil-golang", "github.com/getoutreach/stencil-golang", ver)
	assert.NilError(t, err, "failed to register extension")

	caller, err := ext.GetExtensionCaller(ctx)
	assert.NilError(t, err, "failed to get extension caller")

	resp, err := caller.Call("github.com/getoutreach/stencil-golang.ParseGoMod", "go.mod", "module test\n\ngo 1.19")
	assert.NilError(t, err, "failed to call extension")

	moduleMap := resp.(map[string]any)["Module"].(map[string]any)
	spew.Dump(moduleMap)
	assert.Equal(t, moduleMap["Syntax"].(map[string]any)["Token"].([]any)[1], "test", "failed to parse go.mod")
}
