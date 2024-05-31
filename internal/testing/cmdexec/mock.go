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

package cmdexec

import (
	"context"
	"fmt"
)

// MockExecutor provides an executor that returns mock data.
type MockExecutor struct {
	cmds []*MockCommand
}

// MockCommand is a command that can be executed by the MockExecutor.
type MockCommand struct {
	Name   string
	Args   []string
	Stdout []byte
	Stderr []byte
	Err    error
}

func (c *MockCommand) Output() ([]byte, error) {
	return c.Stdout, c.Err
}

func (c *MockCommand) CombinedOutput() ([]byte, error) {
	return append(c.Stdout, c.Stderr...), c.Err
}

// NewMockExecutor returns a new MockExecutor with the given commands.
func NewMockExecutor(cmds ...*MockCommand) *MockExecutor {
	return &MockExecutor{cmds}
}

// AddCommand adds a command to the executor.
//
// Note: This is not thread-safe.
func (e *MockExecutor) AddCommand(cmd *MockCommand) {
	e.cmds = append(e.cmds, cmd)
}

// executor implements the [executorFn] type, returning a Cmd based on
// the provided arguments. If no commands are available based on the
// provided input, this function will panic.
func (e *MockExecutor) executor(_ context.Context, name string, arg ...string) Cmd {
	if len(e.cmds) == 0 {
		panic("no commands to execute")
	}

	// argsEqual checks if two slices of strings are equal.
	var argsEqual = func(a, b []string) bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}

		return true
	}

	// Check if we have a command that matches the input name and args.
	for i := range e.cmds {
		cmd := e.cmds[i]
		if cmd.Name == name && argsEqual(cmd.Args, arg) {
			return cmd
		}
	}

	// Did you forget to call [AddCommand]?
	panic(fmt.Errorf("no mocked output registered for %s %v", name, arg))
}
