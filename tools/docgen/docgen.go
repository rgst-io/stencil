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

// Package main implements the docgen command. This command generates
// documentation for stencil.
package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"sort"
	"strings"

	// We're using embed
	_ "embed"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/princjef/gomarkdoc"
	"github.com/princjef/gomarkdoc/lang"
	"github.com/princjef/gomarkdoc/logger"
	"go.rgst.io/stencil/pkg/slogext"
)

//go:embed function.md.tpl
var functionsTemplate string

type file struct {
	Name     string
	Contents string
}

// generateMarkdown generates the markdown files for all functions.
func generateMarkdown(_ slogext.Logger) ([]file, error) {
	// Create a renderer to output data
	rOpts := []gomarkdoc.RendererOption{gomarkdoc.WithTemplateOverride("func", functionsTemplate)}
	for k, v := range sprig.TxtFuncMap() {
		rOpts = append(rOpts, gomarkdoc.WithTemplateFunc(k, v))
	}

	// Hack to get consistent ordering for template rendering.
	order := 1000
	rOpts = append(rOpts, gomarkdoc.WithTemplateFunc("order", func() int {
		order++
		return order
	}))

	out, err := gomarkdoc.NewRenderer(rOpts...)
	if err != nil {
		return nil, err
	}

	buildPkg, err := build.ImportDir("internal/codegen", build.ImportComment)
	if err != nil {
		return nil, err
	}

	// Create a documentation package from the build representation of our
	// package.
	pkg, err := lang.NewPackageFromBuild(logger.New(logger.ErrorLevel), buildPkg)
	if err != nil {
		return nil, err
	}
	pkgTypes := pkg.Types()

	// Sort based on name.
	sort.Slice(pkgTypes, func(i, j int) bool {
		return pkgTypes[i].Name() < pkgTypes[j].Name()
	})

	files := make([]file, 0, len(pkgTypes))
	for _, typ := range pkgTypes {
		// Skip non-template types
		if !strings.HasPrefix(typ.Name(), "Tpl") {
			continue
		}

		if typ.Name() == "TplError" {
			continue
		}

		for _, f := range typ.Methods() {
			txt, err := out.Func(f)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to generate documentation for %s", f.Name())
			}

			files = append(files, file{
				Name:     strings.ToLower(strings.TrimPrefix(typ.Name(), "Tpl")) + "." + f.Name(),
				Contents: txt,
			})
		}
	}

	return files, nil
}

// saveMarkdown writes the markdown files to disk.
func saveMarkdown(log slogext.Logger, files []file) error {
	for _, f := range files {
		log.Infof(" -> Writing %s", f.Name)
		if err := os.WriteFile(filepath.Join("docs", "functions", f.Name+".md"), []byte(f.Contents), 0o600); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	log := slogext.New()

	log.Info("Generating documentation")
	files, err := generateMarkdown(log)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := saveMarkdown(log, files); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
