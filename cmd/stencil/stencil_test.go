package main

import (
	"os"
	"path/filepath"
	"testing"

	"go.rgst.io/stencil/v2/internal/yaml"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

func TestFailsWhenUnknownArgumentsArePassed(t *testing.T) {
	cmd := NewStencil(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewStencil() to not return nil")

	err := testRunCommand(t, cmd, "", "im-not-a-command")
	assert.ErrorContains(t, err, "unexpected arguments: [im-not-a-command]")
}

func TestStencilKeepsIgnoredFiles(t *testing.T) {
	cmd := NewStencil(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewStencil() to not return nil")

	tmpDir := t.TempDir()

	assert.NilError(t, os.WriteFile(filepath.Join(tmpDir, ".stencilignore"), []byte("go.mod"), 0o644),
		"expected write to .stencilignore to not fail")

	assert.NilError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("hello, world!"), 0o644),
		"expected write to go.mod to not fail")

	manifest := &configuration.Manifest{
		Name: "test-123",
		Arguments: map[string]any{
			"org": "fake-org",
		},
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
	}
	manifestBytes, err := yaml.Marshal(manifest)
	assert.NilError(t, err, "expected manifest marshal to not fail")
	assert.NilError(t, os.WriteFile(filepath.Join(tmpDir, "stencil.yaml"), manifestBytes, 0o755),
		"expected write to stencil.yaml to not fail")

	assert.NilError(t, testRunCommand(t, cmd, tmpDir, "--skip-post-run"), "expected stencil run to not fail")

	goModContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
	assert.NilError(t, err, "expected read of go.mod to not fail")
	assert.Equal(t, string(goModContent), "hello, world!", "expected go.mod to remain unchanged")
}
