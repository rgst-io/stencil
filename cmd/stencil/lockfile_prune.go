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

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
)

// NewLockfilePruneCommand returns a new urfave/cli.Command for the
// lockfile prune command.
func NewLockfilePruneCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "prune",
		Description: "Prunes any non-existent files from the lockfile (will recreate any file.Once files on next run)",

		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:  "file",
				Usage: "If any file options are passed, prune only checks the passed filenames for pruning",
			},
			&cli.StringSliceFlag{
				Name:  "module",
				Usage: "If any module options are passed, prune only checks the passed modulenames for pruning",
			},
		},
		Action: func(c *cli.Context) error {
			l, err := stencil.LoadLockfile("")
			if err != nil {
				return errors.Wrap(err, "failed to load lockfile")
			}

			manifest, err := configuration.NewDefaultManifest()
			if err != nil {
				return fmt.Errorf("failed to parse stencil.yaml: %w", err)
			}

			prunedFiles := l.PruneFiles(c.StringSlice("file"))

			for _, lf := range prunedFiles {
				log.Infof("Pruned missing file %s from lockfile", lf)
			}

			newModules := []string{}
			for _, m := range manifest.Modules {
				newModules = append(newModules, m.Name)
			}

			prunedModules := l.PruneModules(newModules, c.StringSlice("module"))

			for _, lf := range prunedModules {
				log.Infof("Pruned missing module %s from lockfile", lf)
			}

			if len(prunedFiles) == 0 && len(prunedModules) == 0 {
				log.Info("No changes made to lockfile")
				return nil
			}

			log.Info("Writing out modified lockfile")
			return l.Write()
		},
	}
}
