package main

import (
	"os"
	"path/filepath"
	"testing"

	"go.rgst.io/stencil/internal/modules/resolver"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

// TestCanUpgradeModules tests that the upgrade command can upgrade
// modules in a project.
func TestCanUpgradeModules(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewUpgradeCommand(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewUpgradeCommand() to not return nil")

	// Create a project that consumes stencil-golang at a known low
	// version. If it upgrades, we consider success. Given this is a
	// high-level test, we rely on other unit tests to ensure that the
	// underlying resolver works as expected.
	mf := &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
	}

	lock := &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				// https://github.com/rgst-io/stencil-golang/releases/tag/v0.3.1
				Tag:    "v0.3.1",
				Commit: "6f031a70bea1bb06fe57db48abcea52a287eae7f",
			},
		}},
	}

	// Write the manifest and lockfile to disk
	mfBytes, err := yaml.Marshal(mf)
	assert.NilError(t, err, "failed to marshal manifest")

	lockBytes, err := yaml.Marshal(lock)
	assert.NilError(t, err, "failed to marshal lockfile")

	err = os.WriteFile(filepath.Join(tmpDir, "stencil.yaml"), mfBytes, 0o644)
	assert.NilError(t, err, "failed to write manifest")

	os.WriteFile(filepath.Join(tmpDir, stencil.LockfileName), lockBytes, 0o644)
	assert.NilError(t, err, "failed to write lockfile")

	// Run the upgrade
	err = testRunCommand(t, cmd, tmpDir)
	if err != nil {
		// Right now it errors due to no go.mod, so allow that error to
		// occur. It doesn't indicate the upgrade failure.
		assert.ErrorContains(t, err, "failed to run post run command")
	}

	// Read the lockfile back in and ensure that the version has been
	// updated.
	lf, err := stencil.LoadLockfile(tmpDir)
	assert.NilError(t, err, "expected LoadLockfile() to not error")
	assert.Equal(t, len(lf.Modules), 1, "expected exactly one module in lockfile")
	assert.Check(t, lf.Modules[0].Version.Tag != "v0.3.1", "expected module to be upgraded")
}
