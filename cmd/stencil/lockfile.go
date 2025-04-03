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

// Description: This file contains code for the lockfile command

package main

import (
	"github.com/urfave/cli/v3"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// NewLockfileCommand returns a new urfave/cli.Command for the
// lockfile command set
func NewLockfileCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:  "lockfile",
		Usage: "modify/examine the current lockfile (stencil.lock)",
		Commands: []*cli.Command{
			NewLockfilePruneCommand(log),
		},
	}
}
