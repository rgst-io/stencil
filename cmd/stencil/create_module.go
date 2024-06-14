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
	"os"
	"os/exec"
	"path"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/configuration"
	"gopkg.in/yaml.v3"
)

// NewCreateModuleCommand returns a new urfave/cli.Command for the
// create module command.
func NewCreateModuleCommand() *cli.Command {
	return &cli.Command{
		Name:        "module",
		Description: "Creates a module with the provided name in the current directory",
		ArgsUsage:   "create module <name>",
		Action: func(c *cli.Context) error {
			var manifestFileName = "stencil.yaml"

			// ensure we have a name
			if c.NArg() != 1 {
				return errors.New("must provide a name for the module")
			}

			allowedFiles := map[string]struct{}{
				".git": {},
			}
			files, err := os.ReadDir(".")
			if err != nil {
				return err
			}

			// ensure we don't have any files in the current directory, except for
			// the allowed files
			for _, file := range files {
				if _, ok := allowedFiles[file.Name()]; !ok {
					return fmt.Errorf("directory is not empty, found %s", file.Name())
				}
			}

			tm := &configuration.Manifest{
				Name: path.Base(c.Args().Get(0)),
			}

			f, err := os.Create(manifestFileName)
			if err != nil {
				return err
			}
			defer f.Close()

			enc := yaml.NewEncoder(f)
			if err := enc.Encode(tm); err != nil {
				return err
			}
			if err := enc.Close(); err != nil {
				return err
			}

			//nolint:gosec // Why: intentional
			cmd := exec.CommandContext(c.Context, os.Args[0])
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			return errors.Wrap(cmd.Run(), "failed to run stencil")
		},
	}
}
