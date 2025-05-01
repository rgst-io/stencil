//go:build !test_no_internet

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"go.rgst.io/stencil/v2/pkg/stencil"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

// selectUpgradeCanary returns a file from the provided directory that
// can be used as a canary for re-running stencil detection.
func selectUpgradeCanary(t *testing.T, dir string) string {
	dirs, err := os.ReadDir(dir)
	assert.NilError(t, err, "expected ReadDir() to not fail")

	var file string
	for _, de := range dirs {
		if de.IsDir() {
			continue
		}

		// Don't use any files owned by us.
		if de.Name() == stencil.LockfileName || de.Name() == "stencil.yaml" {
			continue
		}

		file = filepath.Join(dir, de.Name())
	}

	t.Logf("Using %s as upgrade canary", file)
	return file
}

// writeML writes the provided manifest and lock to their known
// locations in the provided directory. If any of the provided manifest
// or lockfile arguments are nil, they are not written to disk.
func writeML(t *testing.T, mf *configuration.Manifest, lock *stencil.Lockfile, dir string) {
	if mf != nil {
		// Write the manifest and lockfile to disk
		mfBytes, err := yaml.Marshal(mf)
		assert.NilError(t, err, "failed to marshal manifest")

		err = os.WriteFile(filepath.Join(dir, "stencil.yaml"), mfBytes, 0o644)
		assert.NilError(t, err, "failed to write manifest")
	}

	if lock != nil {
		lockBytes, err := yaml.Marshal(lock)
		assert.NilError(t, err, "failed to marshal lockfile")

		os.WriteFile(filepath.Join(dir, stencil.LockfileName), lockBytes, 0o644)
		assert.NilError(t, err, "failed to write lockfile")
	}
}

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
	writeML(t, &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Arguments: map[string]any{
			"org": "rgst-io",
		},
	}, &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				// https://github.com/rgst-io/stencil-golang/releases/tag/v0.3.1
				Tag:    "v0.3.1",
				Commit: "6f031a70bea1bb06fe57db48abcea52a287eae7f",
			},
		}},
	}, tmpDir)

	// Run the upgrade
	if err := testRunCommand(t, cmd, tmpDir); err != nil {
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

// TestUpgradeIncludesNewModules tests that the upgrade command can install
// new modules in a project.
func TestUpgradeIncludesNewModules(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewUpgradeCommand(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewUpgradeCommand() to not return nil")

	writeML(t, &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Arguments: map[string]any{
			"org": "rgst-io",
		},
	}, &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{},
	}, tmpDir)

	// Run the upgrade
	assert.NilError(t, testRunCommand(t, cmd, tmpDir), "expected upgrade to not error")

	// Read the lockfile back in and ensure that the module was added.
	lf, err := stencil.LoadLockfile(tmpDir)
	assert.NilError(t, err, "expected LoadLockfile() to not error")
	assert.Equal(t, len(lf.Modules), 1, "expected exactly one module in lockfile")
	assert.Check(t, lf.Modules[0].Version.Tag != "", "expected module to be latest version")
}

// TestUpgradeReRunsStencil tests an upgrade command re-runs stencil
// even when there is no new version.
func TestUpgradeReRunsStencil(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewUpgradeCommand(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewUpgradeCommand() to not return nil")

	writeML(t, &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Arguments: map[string]any{
			"org": "rgst-io",
		},
	}, &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{},
	}, tmpDir)

	// Run the upgrade
	assert.NilError(t, testRunCommand(t, cmd, tmpDir), "expected upgrade to not error")

	file := selectUpgradeCanary(t, tmpDir)
	assert.NilError(t, os.Remove(file))
	assert.NilError(t, testRunCommand(t, cmd, tmpDir), "expected upgrade to not error")

	_, err := os.Stat(file)
	assert.NilError(t, err, "expected file %s to exist", file)
}

// TestUpgradeDoesNotReRunStencilWhenTold tests an upgrade command does
// not re-run stencil when told to do so via a flag.
func TestUpgradeDoesNotReRunStencilWhenTold(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewUpgradeCommand(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewUpgradeCommand() to not return nil")

	writeML(t, &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Arguments: map[string]any{
			"org": "rgst-io",
		},
	}, &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{},
	}, tmpDir)

	// Run the upgrade
	assert.NilError(t, testRunCommand(t, cmd, tmpDir), "expected upgrade to not error")

	file := selectUpgradeCanary(t, tmpDir)
	assert.NilError(t, os.Remove(file))
	assert.NilError(t, testRunCommand(t, cmd, tmpDir, "--skip-render-no-changes"), "expected upgrade to not error")

	_, err := os.Stat(file)
	assert.ErrorIs(t, err, os.ErrNotExist, "expected file %s to not exist", file)
}

// TestCanRunWithoutLock ensures that stencil upgrade can be ran without
// a lock file, but do nothing.
func TestCanRunWithoutLock(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewUpgradeCommand(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewUpgradeCommand() to not return nil")

	// Create a project that consumes stencil-golang at a known low
	// version. If it upgrades, we consider success. Given this is a
	// high-level test, we rely on other unit tests to ensure that the
	// underlying resolver works as expected.
	writeML(t, &configuration.Manifest{
		Name: "testing",
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Arguments: map[string]any{
			"org": "rgst-io",
		},
	}, nil, tmpDir)

	assert.NilError(t, testRunCommand(t, cmd, tmpDir), "expected stencil to not fail")

	dirs, err := os.ReadDir(tmpDir)
	assert.NilError(t, err, "expected ReadDir() to not fail")
	assert.Equal(t, len(dirs), 1, "expected stencil to have not ran")
}
