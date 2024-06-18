// Copyright (C) 2024 stencil contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package main implements the stencil CLI. This is the entrypoint for
// the CLI.
package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go.rgst.io/stencil/internal/cmd/stencil"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
)

// Set the version printer to do nothing but print the version.
//
//nolint:gochecknoinits
func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Version)
	}
}

// NewStencil returns a new CLI application for stencil.
func NewStencil(log slogext.Logger) *cli.App {
	return &cli.App{
		Version:     version.Version.String(),
		Name:        "stencil",
		Description: "a smart templating engine for project development",
		Action: func(c *cli.Context) error {
			log.Infof("stencil %s", c.App.Version)

			// We don't accept arguments, a user is likely trying to run a
			// subcommand here anyways (e.g., typo).
			if c.NArg() > 0 {
				return fmt.Errorf("unexpected arguments: %v", c.Args().Slice())
			}

			if c.Bool("debug") {
				log.SetLevel(slogext.DebugLevel)
				log.Debug("Debug logging enabled")
			}

			manifest, err := configuration.NewDefaultManifest()
			if err != nil {
				return fmt.Errorf("failed to parse stencil.yaml: %w", err)
			}

			return stencil.NewCommand(log, manifest, c.Bool("dry-run")).Run(c.Context)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"dryrun"},
				Usage:   "Don't write files to disk",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enables debug logging for version resolution, template renderer, and other useful information",
				Aliases: []string{"d"},
			},
		},
		Commands: []*cli.Command{
			NewDescribeCommand(),
			NewCreateCommand(),
			NewUpgradeCommand(log),
		},
	}
}
