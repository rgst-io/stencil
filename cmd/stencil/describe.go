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
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/pkg/stencil"
)

// NewDescribeCommand returns a new urfave/cli.Command for the describe
// command
func NewDescribeCommand() *cli.Command {
	return &cli.Command{
		Name:        "describe",
		Description: "Print information about a known file rendered by a template",
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return errors.New("expected exactly one argument, path to file")
			}

			return describeFile(c.Args().First())
		},
	}
}

// cleanPath ensures that a path is always relative to the current working directory
// with no .., . or other path elements.
func cleanPath(path string) (string, error) {
	// make absolute so we can handle .. and other weird path things
	// defaults to nothing if already absolute
	path, err := filepath.Abs(path)
	if err != nil {
		return "", errors.Wrap(err, "failed to get absolute path")
	}

	// convert absolute -> relative
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current working directory")
	}
	path = "." + strings.TrimPrefix(path, cwd)
	return filepath.Clean(path), nil
}

// describeFile prints information about a file rendered by a template
func describeFile(filePath string) error {
	l, err := stencil.LoadLockfile("")
	if err != nil {
		return errors.Wrap(err, "failed to load lockfile")
	}

	// check if the file exists on disk before we try to find
	// it in the lockfile
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file %q does not exist", filePath)
	}

	relativeFilePath, err := cleanPath(filePath)
	if err != nil {
		return errors.Wrap(err, "failed to clean path for searching lockfile")
	}

	for _, f := range l.Files {
		if f.Name == relativeFilePath {
			fmt.Printf("%s was created by module https://%s (template: %s)\n", f.Name, f.Module, f.Template)
			return nil
		}
	}

	return fmt.Errorf("file %q isn't created by stencil", filePath)
}
