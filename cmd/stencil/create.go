// Copyright (C) 2024 stencil contributors
// Copyright (C) 2022-2023 Outreach Corporation
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

// Description: This file contains code for the create command

package main

import (
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// NewCreateCommand returns a new urfave/cli.Command for the
// create command
func NewCreateCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "create",
		Usage:       "create a new stencil project or module",
		Description: "Commands to create template repositories, or stencil powered repositories",
		Subcommands: []*cli.Command{
			NewCreateModuleCommand(log),
		},
	}
}
