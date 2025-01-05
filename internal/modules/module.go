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

// Description: This file implements module specific code.

package modules

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	giturls "github.com/chainguard-dev/git-urls"
	gogit "github.com/go-git/go-git/v5"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/jaredallard/vcs/git"
	"github.com/jaredallard/vcs/resolver"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/v2/internal/modules/nativeext"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"gopkg.in/yaml.v3"
)

// Module is a stencil module that contains template files.
type Module struct {
	// t is a shared go-template that is used for this module. This is
	// important because this allows us to call shared templates across a
	// single module.
	//
	// Note: We don't currently support sharing templates across modules.
	t *template.Template

	// Manifest is the module's manifest information/configuration
	Manifest *configuration.TemplateRepositoryManifest

	// Name is the name of a module. This should be a valid go
	// import path. For example: github.com/rgst-io/stencil-base
	Name string

	// URI is the location of the module for fetching purposes. By
	// default, it's equal to the name with the HTTPS scheme.
	URI string

	// Version is the version of the module to use.
	Version *resolver.Version

	// fs is underlying filesystem for this module
	fs billy.Filesystem

	// dirReplacementsRendered is a rendered list of dirReplacements from the manifest,
	// ready to be used for immediate replacements.  It's a mapping of relative paths
	// to just the replacement name for the last path segment.
	dirReplacementsRendered map[string]string
}

// uriIsLocal returns true if the URI is a local file path
func uriIsLocal(uri string) bool {
	return !strings.Contains(uri, "://") || strings.HasPrefix(uri, "file://")
}

// uriForModule returns the URI for a module. If replacement is an
// empty string, the default URI is used.
func uriForModule(name, replacement string) string {
	if replacement == "" {
		return "https://" + name
	}

	return replacement
}

type NewModuleOpts struct {
	// ImportPath is the import path of the module. This should be the
	// Name field of [configuration.TemplateRepository].
	ImportPath string

	// Version is the version of the module to use. This should be a
	// parsed version of the Version field of
	// [configuration.TemplateRepository].
	Version *resolver.Version

	// FS is an optional filesystem to use for the module. When set, it
	// will be used instead of fetching the module from the network/disk.
	FS billy.Filesystem
}

// New creates a new module from a TemplateRepository. Version must be
// set and can be obtained via the internal/modules/resolver package, or
// by using the FetchModules function.
//
// uri is the URI for the module. If it is an empty string
// https://+name is used instead.
func New(ctx context.Context, uri string, opts NewModuleOpts) (*Module, error) {
	if opts.ImportPath == "" {
		return nil, fmt.Errorf("import path must be specified")
	}

	// Handle local modules if the URI is a local file path
	uri = uriForModule(opts.ImportPath, uri)
	if uriIsLocal(uri) {
		opts.Version = &resolver.Version{
			Virtual: "local",
		}
	}
	if opts.Version == nil {
		return nil, fmt.Errorf("version must be specified for module %q", opts.ImportPath)
	}

	m := Module{
		t:       template.New(opts.ImportPath).Funcs(sprig.TxtFuncMap()),
		Name:    opts.ImportPath,
		URI:     uri,
		Version: opts.Version,
		fs:      opts.FS,
	}

	mani, err := m.getManifest(ctx)
	if err != nil {
		return nil, err
	}
	m.Manifest = mani

	return &m, nil
}

// GetTemplate returns the go template for this module
func (m *Module) GetTemplate() *template.Template {
	return m.t
}

// RegisterExtensions registers all extensions provided by the given
// module. If the module is a local file URI then extensions will be
// sourced from the `./bin` directory of the base of the path.
func (m *Module) RegisterExtensions(ctx context.Context, ext *nativeext.Host) error {
	// Only register extensions if this repository declares extensions explicitly in its type.
	if !m.Manifest.Type.Contains(configuration.TemplateRepositoryTypeExt) {
		return nil
	}
	return ext.RegisterExtension(ctx, m.URI, m.Name, m.Version)
}

// getManifest downloads the module if not already downloaded and
// returns a parsed configuration.TemplateRepositoryManifest of this module.
func (m *Module) getManifest(ctx context.Context) (*configuration.TemplateRepositoryManifest, error) {
	fs, err := m.GetFS(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to download fs")
	}

	mf, err := fs.Open("manifest.yaml")
	if err != nil {
		return nil, err
	}
	defer mf.Close()

	var manifest configuration.TemplateRepositoryManifest
	if err := yaml.NewDecoder(mf).Decode(&manifest); err != nil {
		return nil, err
	}

	// ensure that the manifest name is equal to the import path
	if manifest.Name != m.Name {
		return nil, fmt.Errorf(
			"module declares its import path as %q but was imported as %q",
			manifest.Name, m.Name,
		)
	}

	return &manifest, nil
}

// GetFS returns a billy.Filesystem that contains the contents of this
// module. If we've already fetched the filesystem, it will not be
// fetched again.
func (m *Module) GetFS(ctx context.Context) (billy.Filesystem, error) {
	// If we've already fetched the filesystem, don't do it again.
	if m.fs != nil {
		return m.fs, nil
	}

	u, err := giturls.Parse(m.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module URI: %w", err)
	}

	var storageDir string
	if u.Scheme == "file" {
		// File URLs are already on disk, so just use that path.
		storageDir = strings.TrimPrefix(m.URI, "file://")
	} else {
		var err error
		storageDir, err = git.Clone(ctx, m.Version.GitRef(), m.URI, &git.CloneOptions{UseArchive: true})
		if err != nil {
			return nil, fmt.Errorf("failed to clone module: %w", err)
		}

		// Because this is a new feature, don't fail if we can't open the
		// repo for now. If we have a commit, ensure it matches the commit
		// we expect.
		if m.Version.Commit != "" {
			if r, err := gogit.PlainOpen(storageDir); err == nil {
				if ref, err := r.Head(); err == nil {
					if ref.Hash().String() != m.Version.Commit {
						return nil, fmt.Errorf("expected commit %q but got %q", m.Version.Commit, ref.Hash().String())
					}
				}
			}
		}
	}

	m.fs = osfs.New(storageDir)
	return m.fs, nil
}

// StoreDirReplacements pokes the template-rendered output from the stencil render
// function for use by the module rendering later on via ApplyDirReplacements.
func (m *Module) StoreDirReplacements(reps map[string]string) {
	m.dirReplacementsRendered = reps
}

// ApplyDirReplacements hops through the incoming path dir by dir, starting at the end
// (because the raw paths won't match if you replace the earlier path segments first),
// and see if there's any replacements to apply
func (m *Module) ApplyDirReplacements(path string) string {
	pp := strings.Split(path, string(os.PathSeparator))
	for i := len(pp) - 1; i >= 0; i-- {
		pathPart := strings.Join(pp[0:i+1], string(os.PathSeparator))
		if drepseg, has := m.dirReplacementsRendered[pathPart]; has {
			pp[i] = drepseg
		}
	}
	return strings.Join(pp, string(os.PathSeparator))
}
