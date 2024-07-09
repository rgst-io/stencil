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
	"go.rgst.io/stencil/pkg/stencil"
)

// NewLockfilePruneCommand returns a new urfave/cli.Command for the
// lockfile prune command.
func NewLockfilePruneCommand() *cli.Command {
	return &cli.Command{
		Name:        "prune",
		Description: "Prunes any non-existent files from the lockfile (will recreate any file.Once files on next run)",
		Action: func(_ *cli.Context) error {
			l, err := stencil.LoadLockfile("")
			if err != nil {
				return errors.Wrap(err, "failed to load lockfile")
			}

			prunedList := l.Prune()
			if len(prunedList) == 0 {
				fmt.Printf("No changes made to lockfile\n")
				return nil
			}

			for _, lf := range prunedList {
				fmt.Printf("Pruned missing file %s from lockfile\n", lf)
			}

			fmt.Printf("Writing out modified lockfile\n")
			return l.Write()
		},
	}
}
