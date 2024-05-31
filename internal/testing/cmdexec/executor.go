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
	"os/exec"
	"sync"
)

// Contains package globals to control which executor is used by the
// package as well as locks to ensure this package is thread-safe.
var (
	// executor is the function used to create new commands. By default,
	// this is set to [stdExecutor], but can be replaced with a mock
	// executor using [UseMockExecutor].
	executor executorFn = stdExecutor

	// Locks to control the accessing of the executor variable. We don't
	// use a [sync.RWMutex] here because we want to be able to lock the
	// read and write operations separately.
	executorRLock = new(sync.Mutex)
	executorWLock = new(sync.Mutex)
)

// stdExecutor is the default executor used by exectest. It's a simple
// wrapper around [exec.CommandContext] to return the Cmd interface.
func stdExecutor(ctx context.Context, name string, arg ...string) Cmd {
	return exec.CommandContext(ctx, name, arg...)
}

// executorFn is a function that returns a new Cmd based on the given
// arguments.
type executorFn func(context.Context, string, ...string) Cmd
