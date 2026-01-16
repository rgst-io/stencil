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
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"go.rgst.io/stencil/v2/internal/cmd/stencil"
	"go.rgst.io/stencil/v2/internal/version"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// Set the version printer to do nothing but print the version.
//
//nolint:gochecknoinits // Why: This is acceptable.
func init() {
	cli.VersionPrinter = func(_ *cli.Command) {
		fmt.Println(version.Version.String())
	}
}

var description = "" +
	"Stencil is a smart templating engine that helps you create and manage templates\n" +
	"for your projects. Using go templates, you can create pluggable modules with\n" +
	"templates or, native extensions in any language.\n\n" +
	"Checkout our documentation at https://stencil.rgst.io for more information."

// NewStencilAction returns a new cli.ActionFunc for the plain stencil
// command.
func NewStencilAction(log slogext.Logger) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		log.Infof("stencil %s", c.Root().Version)

		// We don't accept arguments, a user is likely trying to run a
		// subcommand here anyways (e.g., typo).
		if c.NArg() > 0 {
			return fmt.Errorf("unexpected arguments: %v", c.Args().Slice())
		}

		if c.Bool("debug") {
			log.SetLevel(slogext.DebugLevel)
			log.Debug("Debug logging enabled")
		}

		manifest, err := configuration.LoadDefaultManifest()
		if err != nil {
			return fmt.Errorf("failed to parse stencil.yaml: %w", err)
		}

		return stencil.NewCommand(log, manifest, &stencil.NewCommandOpt{
			DryRun:      c.Bool("dry-run"),
			Adopt:       c.Bool("adopt"),
			SkipPostRun: c.Bool("skip-post-run"),
			FailIgnored: c.Bool("fail-ignored"),
		}).Run(ctx)
	}
}

// NewStencil returns a new CLI application for stencil.
func NewStencil(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Version:     version.Version.GitVersion,
		Name:        "stencil",
		Usage:       "a smart templating engine for project development",
		Description: description,
		Action:      NewStencilAction(log),
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
			&cli.BoolFlag{
				Name:  "adopt",
				Usage: "Uses heuristics to detect code that should go into blocks to assist with first-time adoption of templates",
			},
			&cli.BoolFlag{
				Name:  "skip-post-run",
				Usage: "Skips running post-run commands",
			},
			&cli.BoolFlag{
				Name:  "fail-ignored",
				Usage: "Fails if there are ignored files via .stencilignore",
			},
		},
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			NewDescribeCommand(),
			NewCreateCommand(log),
			NewUpgradeCommand(log),
			NewLockfileCommand(log),
			NewModuleCommand(log),
		},
	}
}
