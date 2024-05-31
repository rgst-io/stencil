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

// Package cmdexec provides a way to execute commands using the exec
// package while supporting mocking for testing purposes. The default
// behaviour of the package is to simply wrap [exec.Command] and it's
// context accepting counterpart, [exec.CommandContext]. However, when
// running in tests, the package can be configured to use a mock
// executor that allows for controlling the output and behaviour of the
// commands executed for testing purposes.
package cmdexec

import (
	"context"
	"testing"
)

// Cmd is an interface to be used instead of [*exec.Cmd] for mocking
// purposes.
type Cmd interface {
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}

// Command returns a new Cmd that will call the given command with the
// given arguments. See [exec.Command] for more information.
func Command(name string, arg ...string) Cmd {
	return CommandContext(context.Background(), name, arg...)
}

// CommandContext returns a new Cmd that will call the given command with
// the given arguments and the given context. See [exec.CommandContext]
// for more information.
func CommandContext(ctx context.Context, name string, arg ...string) Cmd {
	executorRLock.Lock()
	defer executorRLock.Unlock()

	return executor(ctx, name, arg...)
}

// UseMockExecutor replaces the executor used by exectest with a mock
// executor that can be used to control the output of all commands
// created after this function is called. A cleanup function is added
// to the test to ensure that the original executor is restored after
// the test has finished.
//
// Note: This function can only ever be called once per test. If it's
// called again it will deadlock the test.
//
// Usage:
//
//	func TestSomething(t *testing.T) {
//	    mock := exectest.NewMockExecutor()
//	    mock.AddCommand(&exectest.MockCommand{
//	        Name:   "echo",
//	        Args:   []string{"hello", "world"},
//	        Stdout: []byte("hello world\n"),
//	    })
//
//	    exectest.UseMockExecutor(t, mock)
//
//	    // Your test code here.
//	}
func UseMockExecutor(t *testing.T, mock *MockExecutor) {
	// Prevent new mock executors from being used until this test has finished.
	executorWLock.Lock()

	// Lock the reader to prevent new commands from being created while we
	// swap out the executor.
	executorRLock.Lock()
	originalExecutor := executor
	executor = mock.executor
	executorRLock.Unlock()

	t.Cleanup(func() {
		// Lock the reader again to prevent new commands from being created
		// while we restore the original executor.
		executorRLock.Lock()

		// Unlock the reader and writer once we're done.
		defer executorRLock.Unlock()
		defer executorWLock.Unlock()

		// Restore the original executor.
		executor = originalExecutor
	})
}
