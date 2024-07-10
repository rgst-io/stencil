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
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
)

// NewLockfilePruneCommand returns a new urfave/cli.Command for the
// lockfile prune command.
func NewLockfilePruneCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "prune",
		Description: "Prunes any non-existent files from the lockfile (will recreate any file.Once files on next run)",
		Args:        true,
		ArgsUsage:   "[filename] [filename 2] [...] - only prunes passed filenames from the lockfile",
		Action: func(c *cli.Context) error {
			l, err := stencil.LoadLockfile("")
			if err != nil {
				return errors.Wrap(err, "failed to load lockfile")
			}

			prunedList := l.Prune(c.Args().Slice())
			if len(prunedList) == 0 {
				log.Info("No changes made to lockfile")
				return nil
			}

			for _, lf := range prunedList {
				log.Infof("Pruned missing file %s from lockfile", lf)
			}

			log.Info("Writing out modified lockfile")
			return l.Write()
		},
	}
}
