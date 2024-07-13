package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

// prepareTestRun sets up the environment for running a stencil command.
//
// dir is the directory to change into before running the command. If
// dir is an empty string, a temporary directory will be created.
func prepareTestRun(t *testing.T, dir string) {
	// Change into the repo root.
	// GOMOD: <module path>/go.mod
	b, err := exec.Command("go", "env", "GOMOD").Output()
	assert.NilError(t, err)
	repoRoot := strings.TrimSuffix(strings.TrimSpace(string(b)), "/go.mod")
	chdir(t, repoRoot)

	// Use a temporary directory for the test if one is not provided.
	if dir == "" {
		dir = t.TempDir()
	}
	chdir(t, dir)
}

// testRunApp runs the provided cli.App with the provided arguments.
func testRunApp(t *testing.T, dir string, app *cli.App, args ...string) error {
	prepareTestRun(t, dir)

	return app.Run(append([]string{"test"}, args...))
}

// testRunCommand runs a command with the provided arguments. It does
// not support global flags.
func testRunCommand(t *testing.T, cmd *cli.Command, dir string, args ...string) error {
	prepareTestRun(t, dir)

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
	log := slogext.NewTestLogger(t)
	cmd := NewCreateModuleCommand(log)
	assert.Assert(t, cmd != nil)
	assert.NilError(t, testRunCommand(t, cmd, "", "test-module"))

	// Ensure it created the expected files.
	_, err := os.Stat("stencil.yaml")
	assert.NilError(t, err)
}

func TestCreateModuleFailsWhenFilesExist(t *testing.T) {
	log := slogext.NewTestLogger(t)
	cmd := NewCreateModuleCommand(log)
	assert.Assert(t, cmd != nil)

	tmpDir := t.TempDir()

	// Create a file to trigger the error.
	f, err := os.Create(filepath.Join(tmpDir, "test-file"))
	assert.NilError(t, err)
	assert.NilError(t, f.Close())

	err = testRunCommand(t, cmd, tmpDir, "test-module")
	assert.ErrorContains(t, err, "directory is not empty, found test-file")
}
