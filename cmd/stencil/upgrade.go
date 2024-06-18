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

package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/internal/cmd/stencil"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
)

// NewUpgradeCommand returns a new urfave/cli.Command for the upgrade
// command.
func NewUpgradeCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "upgrade",
		Usage:       "upgrade stencil modules",
		Description: "Runs stencil with newer modules and updates stencil.lock to use them",
		UsageText:   "upgrade",
		Action: func(c *cli.Context) error {
			log.Infof("stencil %s", c.App.Version)

			if c.Bool("debug") {
				log.SetLevel(slogext.DebugLevel)
				log.Debug("Debug logging enabled")
			}

			manifest, err := configuration.NewDefaultManifest()
			if err != nil {
				return fmt.Errorf("failed to parse stencil.yaml: %w", err)
			}

			return stencil.NewCommand(log, manifest, c.Bool("dry-run")).Upgrade(c.Context)
		},
	}
}
