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
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gopkg.in/yaml.v3"
)

// encodeToFile encodes the provided data [d] to the provided file path,
// it is done by streaming to the created file.
func encodeToFile(d any, outputFilePath string) error {
	f, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	if err := enc.Encode(d); err != nil {
		return err
	}

	return enc.Close()
}

// generateTemplateRepository generates a template repository manifest
// (manifest.yaml) based on the provided input.
func generateTemplateRepository(name string, hasNativeExt bool) *configuration.TemplateRepositoryManifest {
	tr := &configuration.TemplateRepositoryManifest{
		Name: name,
	}

	tr.Type = configuration.TemplateRepositoryTypes{}
	if hasNativeExt {
		tr.Type = append(tr.Type, configuration.TemplateRepositoryTypeTemplates, configuration.TemplateRepositoryTypeExt)
	}

	return tr
}

// generateStencilYaml generates a stencil.yaml manifest based on the
// provided input.
func generateStencilYaml(name string, hasNativeExt bool) *configuration.Manifest {
	mf := &configuration.Manifest{
		Name: path.Base(name),
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-module",
		}},
		Arguments: map[string]any{
			"org": strings.Split(strings.TrimPrefix(name, "github.com/"), "/")[0],
		},
	}

	if hasNativeExt {
		mf.Arguments["commands"] = []string{
			"plugin",
		}
	} else {
		mf.Arguments["library"] = true
	}

	return mf
}

// NewCreateModuleCommand returns a new urfave/cli.Command for the
// create module command.
func NewCreateModuleCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "module",
		Description: "Creates a module with the provided name in the current directory",
		ArgsUsage:   "create module <name>",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "native-extension",
				Usage: "Generate a module with a native extension. ",
			},
		},
		Action: func(c *cli.Context) error {
			var stencilManifestName = "stencil.yaml"

			// ensure we have a name
			if c.NArg() != 1 {
				return errors.New("must provide a name for the module")
			}

			moduleName := c.Args().Get(0)
			hasNativeExt := c.Bool("native-extension")

			// stencil-golang requires Github right now, so it doesn't make
			// sense to generate broken templates on some other VCS provider.
			// Note that you can still have template modules on, say, Gitlab,
			// but we just can't generate them (yet!).
			if !strings.HasPrefix(moduleName, "github.com/") {
				return fmt.Errorf("currently, only github based modules are supported")
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

			// create stencil.yaml
			if err := encodeToFile(generateStencilYaml(moduleName, hasNativeExt), stencilManifestName); err != nil {
				return fmt.Errorf("failed to serialize %s: %w", stencilManifestName, err)
			}

			if err := encodeToFile(generateTemplateRepository(moduleName, hasNativeExt), "manifest.yaml"); err != nil {
				return fmt.Errorf("failed to serialize generated manifest.yaml: %w", err)
			}

			// Run the standard stencil command.
			if err := NewStencilAction(log)(cli.NewContext(c.App, flag.NewFlagSet("", flag.ExitOnError), c)); err != nil {
				return err
			}

			fmt.Println()
			log.Info("Created module successfully", "module", moduleName)
			log.Info("- Ensure that 'stencil.yaml' is configured to your liking (e.g., license)")
			log.Info("- For configuration options provided by stencil-golang, see the docs:")
			log.Info("  https://github.com/rgst-io/stencil-golang")
			return nil
		},
	}
}
