// Copyright (C) 2024 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"path"
	"testing"

	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

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
			name:                "TestPruneNoChange",
			initStencilYaml:     "name: stencil\nmodules:\n  - name: test\n",
			initStencilLock:     "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
			makeTestFile:        true,
			expectedStencilLock: "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
		},
		{
			name:                "TestPruneMissingFile",
			initStencilYaml:     "name: stencil\nmodules:\n  - name: test\n",
			initStencilLock:     "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
			expectedStencilLock: "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n",
		},
		{
			name:                "TestPruneMissingFileNotInPassedList",
			initStencilYaml:     "name: stencil\nmodules:\n  - name: test\n",
			initStencilLock:     "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
			pruneArgs:           []string{"--file", "somethingelse"},
			expectedStencilLock: "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
		},
		{
			name:                "TestPruneMissingFileInPassedList",
			initStencilYaml:     "name: stencil\nmodules:\n  - name: test\n",
			initStencilLock:     "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n",
			pruneArgs:           []string{"--file", "testfile"},
			expectedStencilLock: "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n",
		},
		{
			name:                "TestPruneMissingModuleNotInPassedList",
			initStencilYaml:     "name: stencil\n",
			initStencilLock:     "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n",
			expectedStencilLock: "version: v1.6.2\nmodules: []\nfiles: []\n",
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
