package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
	"gotest.tools/v3/assert"
)

// testRunCommand runs a command with the provided arguments. It does
// not support global flags.
func testRunCommand(t *testing.T, cmd *cli.Command, args ...string) error {
	// Change into the repo root.
	// GOMOD: <module path>/go.mod
	b, err := exec.Command("go", "env", "GOMOD").Output()
	assert.NilError(t, err)
	repoRoot := strings.TrimSuffix(strings.TrimSpace(string(b)), "/go.mod")
	chdir(t, repoRoot)

	// Build stencil in case it's required for this test.
	bCmd := exec.Command("mise", "run", "build")
	bCmd.Stderr = os.Stderr
	bCmd.Stdout = os.Stdout
	assert.NilError(t, bCmd.Run(), "failed to build stencil")

	// Temporarily change os.Args[0] to point to stencil.
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{filepath.Join(repoRoot, "bin", "stencil")}

	// Use a temporary directory for the test.
	chdir(t, t.TempDir())

	app := cli.NewApp()
	app.Commands = []*cli.Command{cmd}
	return app.Run(append([]string{"test", cmd.Name}, args...))
}

// chdir changes the current working directory to the provided directory
// and sets up a cleanup function to change it back to the original
// directory when the test is done. If the cleanup function fails, the
// test will panic.
func chdir(t *testing.T, dir string) {
	origDir, err := os.Getwd()
	assert.NilError(t, err)
	assert.NilError(t, os.Chdir(dir))
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			// Failed, not safe to run other tests.
			panic(err)
		}
	})
}

func TestCanCreateModule(t *testing.T) {
	cmd := NewCreateModule()
	assert.Assert(t, cmd != nil)
	assert.NilError(t, testRunCommand(t, cmd, "test-module"))

	// Ensure it created the expected files.
	_, err := os.Stat("stencil.yaml")
	assert.NilError(t, err)
}
