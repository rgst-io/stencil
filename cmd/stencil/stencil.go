// Copyright (C) 2026 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Package main implements the stencil CLI. This is the entrypoint for
// the CLI.
package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"go.rgst.io/jaredallard/slogext/v2"
	"go.rgst.io/stencil/v2/internal/cmd/stencil"
	"go.rgst.io/stencil/v2/internal/version"
	"go.rgst.io/stencil/v2/pkg/configuration"
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

		return stencil.NewCommand(log, manifest, &stencil.NewCommandOpts{
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
			NewUseCommand(),
		},
	}
}
