package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
	"sigs.k8s.io/yaml"
)

// mustYamlMarshal marshals the provided data as YAML and returns it as
// a string. If it fails, this function panics.
func mustYamlMarshal(d any) string {
	b, err := yaml.Marshal(d)
	if err != nil {
		panic(fmt.Errorf("failed to marshal data as yaml: %w", err))
	}

	return string(b)
}

// TestLockfilePrune is a test matrix runner for combos against lockfile prune
func TestLockfilePrune(t *testing.T) {
	log := slogext.NewTestLogger(t)

	tests := []struct {
		name                string
		initStencilYaml     string
		initStencilLock     string
		pruneArgs           []string
		makeTestFile        bool
		expectedStencilLock string
	}{
		{
			name: "TestPruneNoChange",
			initStencilYaml: mustYamlMarshal(map[string]any{
				"name":    "stencil",
				"modules": []map[string]any{{"name": "test"}},
			}),
			initStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
			makeTestFile: true,
			expectedStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
		},
		{
			name: "TestPruneMissingFile",
			initStencilYaml: mustYamlMarshal(map[string]any{
				"name":    "stencil",
				"modules": []map[string]any{{"name": "test"}},
			}),
			initStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
			expectedStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []any{},
			}),
		},
		{
			name: "TestPruneMissingFileNotInPassedList",
			initStencilYaml: mustYamlMarshal(map[string]any{
				"name":    "stencil",
				"modules": []map[string]any{{"name": "test"}},
			}),
			initStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
			pruneArgs: []string{"--file", "somethingelse"},
			expectedStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
		},
		{
			name: "TestPruneMissingFileInPassedList",
			initStencilYaml: mustYamlMarshal(map[string]any{
				"name":    "stencil",
				"modules": []map[string]any{{"name": "test"}},
			}),
			initStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []map[string]any{
					{"name": "testfile"},
				},
			}),
			pruneArgs: []string{"--file", "testfile"},
			expectedStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []any{},
			}),
		},
		{
			name: "TestPruneMissingModuleNotInPassedList",
			initStencilYaml: mustYamlMarshal(map[string]any{
				"name": "stencil",
			}),
			initStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []map[string]any{
					{"name": "test", "url": "", "version": nil},
				},
				"files": []any{},
			}),
			expectedStencilLock: mustYamlMarshal(map[string]any{
				"version": "v1.6.2",
				"modules": []any{},
				"files":   []any{},
			}),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			td := t.TempDir()

			err := os.WriteFile(path.Join(td, "stencil.yaml"), []byte(tc.initStencilYaml), 0o666)
			assert.NilError(t, err)

			err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(tc.initStencilLock), 0o666)
			assert.NilError(t, err)

			if tc.makeTestFile {
				err = os.WriteFile(path.Join(td, "testfile"), []byte("shrug"), 0o666)
				assert.NilError(t, err)
			}

			cmd := NewLockfilePruneCommand(log)
			err = testRunCommand(t, cmd, td, tc.pruneArgs...)
			assert.NilError(t, err)

			conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
			assert.NilError(t, err)

			assert.Equal(t, tc.expectedStencilLock, string(conts))
		})
	}
}
