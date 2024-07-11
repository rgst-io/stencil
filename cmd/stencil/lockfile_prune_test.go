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

	"github.com/stretchr/testify/assert"
	"go.rgst.io/stencil/pkg/slogext"
)

// TestPruneNoChange tests the lockfile prune behavior with an existing file and module reference
func TestPruneNoChange(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\nmodules:\n    - name: test\n"), 0o666)
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "testfile"), []byte("shrug"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)
	assert.Equal(t, lockStartConts, string(conts))
}

// TestPruneMissingFile tests the lockfile prune behavior with a missing file to prune with no passed file args
func TestPruneMissingFile(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\nmodules:\n  - name: test\n"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)

	lockEndConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n"
	assert.Equal(t, lockEndConts, string(conts))
}

// TestPruneMissingFileNotInPassedList tests the lockfile prune behavior with a missing file to prune but not passed into args
func TestPruneMissingFileNotInPassedList(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\nmodules:\n  - name: test\n"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune", "--file", "somethingelse"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)

	assert.Equal(t, lockStartConts, string(conts))
}

// TestPruneMissingFileInPassedList tests the lockfile prune behavior with a missing file to prune passed into arg list
func TestPruneMissingFileInPassedList(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\nmodules:\n  - name: test\n"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles:\n  - name: testfile\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune", "--file", "testfile"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)

	lockEndConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n"
	assert.Equal(t, lockEndConts, string(conts))
}

// TestPruneMissingModulNotInPassedList tests the lockfile prune behavior with a missing module to prune not passed into arg list
func TestPruneMissingModulNotInPassedList(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\n"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)

	lockEndConts := "version: v1.6.2\nmodules: []\nfiles: []\n"
	assert.Equal(t, lockEndConts, string(conts))
}

// TestPruneMissingModuleInPassedList tests the lockfile prune behavior with a missing module to prune passed into arg list
func TestPruneMissingModuleInPassedList(t *testing.T) {
	log := slogext.NewTestLogger(t)

	td, err := os.MkdirTemp("", "tr")
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(td, "stencil.yaml"), []byte("name: stencil\n"), 0o666)
	assert.NoError(t, err)

	lockStartConts := "version: v1.6.2\nmodules:\n    - name: test\n      url: \"\"\n      version: null\nfiles: []\n"
	err = os.WriteFile(path.Join(td, "stencil.lock"), []byte(lockStartConts), 0o666)
	assert.NoError(t, err)

	cli := NewStencil(log)
	os.Chdir(td)
	err = cli.Run([]string{"", "lockfile", "prune", "--module", "test"})
	assert.NoError(t, err)

	conts, err := os.ReadFile(path.Join(td, "stencil.lock"))
	assert.NoError(t, err)

	lockEndConts := "version: v1.6.2\nmodules: []\nfiles: []\n"
	assert.Equal(t, lockEndConts, string(conts))
}
