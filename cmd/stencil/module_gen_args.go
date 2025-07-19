// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/urfave/cli/v3"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"

	_ "embed"
)

//go:embed embed/arguments.md.tpl
var argumentTpl []byte

// NewModuleGenArgsDocsCommand returns a [cli.Command] for the module
// gen-args command.
func NewModuleGenArgsDocsCommand(_ slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "gen-args-docs",
		Aliases:     []string{"generate-args-docs"},
		Usage:       "Generate documentation for the current stencil module",
		Description: "Generate documentation for the current stencil module.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "output-file",
				Aliases:     []string{"o"},
				DefaultText: "File to output to, set to '-' for stdout",
				Value:       "docs/arguments.md",
			},
		},
		Action: func(_ context.Context, c *cli.Command) error {
			mf, err := configuration.LoadDefaultTemplateRepositoryManifest()
			if err != nil {
				return err
			}

			tpl, err := template.New("arguments.md.tpl").Funcs(sprig.FuncMap()).Parse(string(argumentTpl))
			if err != nil {
				return fmt.Errorf("failed to parse embedded arguments template: %w", err)
			}

			var out io.Writer

			outFile := filepath.Clean(c.String("output-file"))
			if outFile == "-" {
				out = os.Stdout
			} else {
				outDir := filepath.Dir(outFile)
				if outDir != outFile {
					if err := os.MkdirAll(outDir, 0o750); err != nil {
						return fmt.Errorf(
							"failed to create parent directory of output file %s, %s: %w", outFile, outDir, err,
						)
					}
				}

				f, err := os.Create(outFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer f.Close()

				out = f
			}

			return tpl.Execute(out, mf)
		},
	}
}
