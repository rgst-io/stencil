package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
)

// TestCanUseModule ensures that we can run 'stencil use <modulePath'
// and have it mutate stencil.yaml.
func TestCanUseModule(t *testing.T) {
	testDir := t.TempDir()

	assert.NilError(t,
		os.WriteFile(
			filepath.Join(testDir, "stencil.yaml"),
			[]byte(fmt.Sprintf("name: %s", t.Name())), 0o600,
		),
		"expected stencil.yaml creation to not fail",
	)

	cmd := NewUseCommand()
	assert.Assert(t, cmd != nil)
	assert.NilError(t, testRunCommand(t, cmd, testDir, "github.com/rgst-io/test-module"))

	// Ensure it created the expected files.
	b, err := os.ReadFile(filepath.Join(testDir, "stencil.yaml"))
	assert.NilError(t, err, "expected reading stencil.yaml to not fail")

	assert.Equal(t, string(b),
		fmt.Sprintf("name: %s\nmodules:\n  - name: github.com/rgst-io/test-module\n", t.Name()),
	)
}

// TestCanUseVersionedModuled ensures that we can run 'stencil use
// <modulePath>@v2' and have it set the version.
func TestCanUseVersionedModuled(t *testing.T) {
	testDir := t.TempDir()

	assert.NilError(t,
		os.WriteFile(
			filepath.Join(testDir, "stencil.yaml"),
			[]byte(fmt.Sprintf("name: %s", t.Name())), 0o600,
		),
		"expected stencil.yaml creation to not fail",
	)

	cmd := NewUseCommand()
	assert.Assert(t, cmd != nil)
	assert.NilError(t, testRunCommand(t, cmd, testDir, "github.com/rgst-io/test-module@v2"))

	// Ensure it created the expected files.
	b, err := os.ReadFile(filepath.Join(testDir, "stencil.yaml"))
	assert.NilError(t, err, "expected reading stencil.yaml to not fail")

	assert.Equal(t, string(b),
		fmt.Sprintf("name: %s\nmodules:\n  - name: github.com/rgst-io/test-module\n    version: v2\n", t.Name()),
	)
}

// TestCanUseReplacement ensures that we can run 'stencil use
// ../path/to/module' and have it configure a replacement.
func TestCanUseReplacement(t *testing.T) {
	testDir := t.TempDir()
	moduleDir := t.TempDir()

	assert.NilError(t,
		os.WriteFile(
			filepath.Join(testDir, "stencil.yaml"),
			[]byte(fmt.Sprintf("name: %s", t.Name())), 0o600,
		),
		"expected stencil.yaml creation to not fail",
	)

	assert.NilError(t,
		os.WriteFile(
			filepath.Join(moduleDir, "manifest.yaml"),
			[]byte(fmt.Sprintf("name: %s", t.Name()+"Module")), 0o600,
		),
		"expected manifest.yaml creation to not fail",
	)

	cmd := NewUseCommand()
	assert.Assert(t, cmd != nil)
	assert.NilError(t, testRunCommand(t, cmd, testDir, moduleDir))

	// Ensure it created the expected files.
	b, err := os.ReadFile(filepath.Join(testDir, "stencil.yaml"))
	assert.NilError(t, err, "expected reading stencil.yaml to not fail")

	assert.Equal(t, string(b),
		fmt.Sprintf("name: %s\nreplacements:\n  %s: %s\n", t.Name(), t.Name()+"Module", moduleDir),
	)
}
