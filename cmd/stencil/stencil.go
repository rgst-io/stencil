// Copyright (C) 2023 stencil contributors
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

	"os"

	"github.com/urfave/cli/v2"

	"github.com/pkg/errors"
	"go.rgst.io/stencil/internal/cmd/stencil"
	"go.rgst.io/stencil/internal/slogext"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
)

// main is the entrypoint for the stencil CLI.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slogext.New()

	app := cli.App{
		Version:     version.Version,
		Name:        "stencil",
		Description: "a smart templating engine for project development",
		Action: func(c *cli.Context) error {
			log.Infof("stencil %s", c.App.Version)

			if c.Bool("debug") {
				log.SetLevel(slogext.DebugLevel)
				log.Debug("Debug logging enabled")
			}

			manifest, err := configuration.NewDefaultManifest()
			if err != nil {
				return errors.Wrap(err, "failed to parse stencil.yaml")
			}

			cmd := stencil.NewCommand(log, manifest, c.Bool("dry-run"),
				c.Bool("frozen-lockfile"), c.Bool("allow-major-version-upgrades"))
			return errors.Wrap(cmd.Run(ctx), "run codegen")
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"dryrun"},
				Usage:   "Don't write files to disk",
			},
			&cli.BoolFlag{
				Name:  "frozen-lockfile",
				Usage: "Use versions from the lockfile instead of the latest",
			},
			&cli.BoolFlag{
				Name:  "allow-major-version-upgrades",
				Usage: "Allow major version upgrades without confirmation",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enables debug logging for version resolution, template renderer, and other useful information",
				Aliases: []string{"d"},
			},
		},
		Commands: []*cli.Command{
			NewDescribeCmd(),
			NewCreateCommand(),
			NewDocsCommand(),
			NewConfigureCommand(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		//nolint:gocritic // Why: We're OK not canceling context in this case.
		log.With(err).Error("failed to run")
	}
}
