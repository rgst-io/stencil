// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains code for the create command

package main

import (
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/slogext"
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
