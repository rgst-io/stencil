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

// Package main contains the implementation for the stencil CLI.
package main

import (
	"context"
	"os"

	"go.rgst.io/stencil/pkg/slogext"
)

// entrypoint is the main entrypoint for the stencil CLI. It is
// separated from main to allow for defers to run before exiting on
// error, which main handles.
func entrypoint(log slogext.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewStencil(log)
	return app.RunContext(ctx, os.Args)
}

// main calls the entrypoint, logs errors, and exits with a non-zero
// status code if an error occurs. Logic should be in entrypoint.
func main() {
	log := slogext.New()

	if err := entrypoint(log); err != nil {
		//nolint:gocritic // Why: We're OK not canceling context in this case.
		log.WithError(err).Error("failed to run")
		os.Exit(1)
	}
}
