// Copyright 2022 Outreach Corporation. All Rights Reserved.

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
	"github.com/getoutreach/gobox/pkg/cli/updater/resolver"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/internal/git"
	"go.rgst.io/stencil/internal/modules/nativeext"
	"go.rgst.io/stencil/pkg/configuration"
	"gopkg.in/yaml.v3"
)

// localModuleVersion is the version string used for local modules
const localModuleVersion = "local"

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
	// import path. For example: github.com/getoutreach/stencil-base
	Name string

	// URI is the location of the module for fetching purposes. By
	// default, it's equal to the name with the HTTPS scheme.
	URI string

	// Version is the version of this module
	Version string

	// fs is underlying filesystem for this module
	fs billy.Filesystem
}

// uriIsLocal returns true if the URI is a local file path
func uriIsLocal(uri string) bool {
	return !strings.Contains(uri, "://") || strings.HasPrefix(uri, "file://")
}

// New creates a new module from a TemplateRepository. Version must be
// set and can be obtained via the gobox/pkg/cli/updater/resolver
// package, or by using the GetModulesForProject function.
//
// uri is the URI for the module. If it is an empty string
// https://+name is used instead.
func New(ctx context.Context, uri string, tr *configuration.TemplateRepository,
	fs billy.Filesystem) (*Module, error) {
	if uri == "" {
		uri = "https://" + tr.Name
	}

	// check if a url based on if :// is in the uri, this is kinda hacky
	// but the only way to do this with a URL+file path being allowed.
	// We also support the "older" file:// scheme.
	if uriIsLocal(uri) { // Assume it's a path.
		osPath := strings.TrimPrefix(uri, "file://")
		if _, err := os.Stat(osPath); err != nil {
			return nil, errors.Wrapf(err, "failed to find module %s at path %q", tr.Name, osPath)
		}

		// translate the path into a file:// URI
		uri = "file://" + osPath
		tr.Version = localModuleVersion
	}
	if tr.Version == "" {
		return nil, fmt.Errorf("version must be specified for module %q", tr.Name)
	}

	m := Module{
		t:       template.New(tr.Name).Funcs(sprig.TxtFuncMap()),
		Name:    tr.Name,
		URI:     uri,
		Version: tr.Version,
		fs:      fs,
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

	version := &resolver.Version{
		Tag: m.Version,
	}
	return ext.RegisterExtension(ctx, m.URI, m.Name, version)
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
		storageDir, err = git.Clone(ctx, m.Version, m.URI)
		if err != nil {
			return nil, fmt.Errorf("failed to clone module: %w", err)
		}
	}

	m.fs = osfs.New(storageDir)
	return m.fs, nil
}
