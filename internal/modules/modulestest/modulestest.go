// Copyright (C) 2024 stencil contributors
// Copyright (C) 2022-2023 Outreach Corporation
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

// Description: This file implements basic helpers for module
// test interaction

// Package modulestest contains code for interacting with modules
// in tests.
package modulestest

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/jaredallard/vcs/resolver"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"gopkg.in/yaml.v3"
)

// addTemplateToFS adds a template to a billy.Filesystem
func addTemplateToFS(fs billy.Filesystem, tpl string) error {
	srcFile, err := os.Open(tpl)
	if err != nil {
		return errors.Wrapf(err, "failed to open template file %q", tpl)
	}
	defer srcFile.Close()

	destF, err := fs.Create(filepath.Join("templates", tpl))
	if err != nil {
		return errors.Wrapf(err, "failed to create template %q in memfs", tpl)
	}
	defer destF.Close()

	// Copy the template file to the fs
	_, err = io.Copy(destF, srcFile)
	return errors.Wrapf(err, "failed to copy template %q to memfs", tpl)
}

// NewModuleFromTemplate creates a module with the provided template
// being the only file in the module.
func NewModuleFromTemplates(manifest *configuration.TemplateRepositoryManifest,
	templates ...string) (*modules.Module, error) {
	fs := memfs.New()
	for _, tpl := range templates {
		if err := addTemplateToFS(fs, tpl); err != nil {
			return nil, err
		}
	}

	mf, err := fs.Create("manifest.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in memory manifest file")
	}
	defer mf.Close()

	// write a manifest file so that we can handle arguments
	enc := yaml.NewEncoder(mf)
	if err := enc.Encode(manifest); err != nil {
		return nil, errors.Wrap(err, "failed to encode generated module manifest")
	}
	if err := enc.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close generated module manifest")
	}

	// create the module
	return NewWithFS(context.Background(), manifest.Name, fs)
}

// NewWithFS creates a module with the specified file system.
func NewWithFS(ctx context.Context, name string, fs billy.Filesystem) (*modules.Module, error) {
	return modules.New(ctx, "vfs://"+name, modules.NewModuleOpts{
		ImportPath: name,
		Version:    &resolver.Version{Virtual: "vfs"},
		FS:         fs,
	})
}
