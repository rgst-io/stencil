// Copyright (C) 2026 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package stdouttest contains helpers for capturing stdout in tests.
package stdouttest

import (
	"os"
	"sync"
	"testing"

	"gotest.tools/v3/assert"
)

// stdouterrLock is used to lock modification of both [os.Stdout] and
// [os.Stderr].
var stdouterrLock = sync.Mutex{}

// RunOptions are options for [Run].
type RunOptions struct {
	// Stdout denotes if stdout should be captured.
	Stdout bool

	// Stderr denotes if stderr should be captured.
	Stderr bool
}

// Run runs the following code with stdout and stderr captured by
// default, or as configured with [RunOptions].
//
// Note: Run does NOT support being ran in parallel tests and will fail
// the test if detected.
func Run(t *testing.T, fn func(), opts ...*RunOptions) []byte {
	t.Helper()
	if len(opts) > 1 {
		t.Fatal("Run: RunOptions can only be provided at most once")
	}

	var opt *RunOptions
	if len(opts) == 0 {
		opt = &RunOptions{
			Stdout: true,
			Stderr: true,
		}
	} else {
		opt = opts[0]
	}

	if !stdouterrLock.TryLock() {
		t.Fatal("failed to obtain stdout/err lock for capturing. " +
			"Parallel tests are not supported.")
	}
	defer stdouterrLock.Unlock()

	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "stdouttest-*")
	assert.NilError(t, err,
		"failed to create temp file for stdout/err capturing",
	)

	if opt.Stdout {
		origStdout := os.Stdout
		t.Cleanup(func() {
			os.Stdout = origStdout
		})

		os.Stdout = tmpFile
	}

	if opt.Stderr {
		origStderr := os.Stderr
		t.Cleanup(func() {
			os.Stderr = origStderr
		})

		os.Stderr = tmpFile
	}

	fn()

	assert.NilError(t, tmpFile.Close(),
		"failed to close temp file used for stdout/err capturing",
	)

	b, err := os.ReadFile(tmpFile.Name())
	assert.NilError(t, err,
		"failed to read temp file used for stdout/err capturing",
	)
	return b
}
