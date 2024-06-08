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
	"github.com/getoutreach/gobox/pkg/cli/github"
	"github.com/getoutreach/gobox/pkg/cli/updater/resolver"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	giturls "github.com/whilp/git-urls"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/extensions"
	"gopkg.in/yaml.v3"
)

// localModuleVersion is the version string used for local modules
const localModuleVersion = "local"

// Module is a stencil module that contains template files.
type Module struct {
	// t is a shared go-template that is used for this module. This is important
	// because this allows us to call shared templates across a single module.
	// Note: We don't currently support sharing templates across modules. Instead
	// the data passing system should be used for cases like this.
	t *template.Template

	// Manifest is the module's manifest information/configuration
	Manifest *configuration.TemplateRepositoryManifest

	// Name is the name of a module. This should be a valid go
	// import path. For example: github.com/getoutreach/stencil-base
	Name string

	// URI is the underlying URI being used to download this module
	URI string

	// Version is the version of this module
	Version string

	// fs is a cached filesystem
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

// New creates a new module from a TemplateRepository. Version must be set and can
// be obtained via the gobox/pkg/cli/updater/resolver package, or by using the
// GetModulesForProject function.
//
// uri is the URI for the module. If it is an empty string https://+name is used
// instead.
func New(ctx context.Context, uri string, tr *configuration.TemplateRepository, fs billy.Filesystem) (*Module, error) {
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

// RegisterExtensions registers all extensions provided
// by the given module. If the module is a local file
// URI then extensions will be sourced from the `./bin`
// directory of the base of the path.
func (m *Module) RegisterExtensions(ctx context.Context, ext *extensions.Host) error {
	// Only register extensions if this repository declares extensions explicitly in its type.
	if !m.Manifest.Type.Contains(configuration.TemplateRepositoryTypeExt) {
		return nil
	}

	version := &resolver.Version{
		Tag: m.Version,
	}
	return ext.RegisterExtension(ctx, m.URI, m.Name, version)
}

// getManifest downloads the module if not already downloaded and returns a parsed
// configuration.TemplateRepositoryManifest of this module.
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

// GetFS returns a billy.Filesystem that contains the contents
// of this module.
func (m *Module) GetFS(ctx context.Context) (billy.Filesystem, error) {
	if m.fs != nil {
		return m.fs, nil
	}

	u, err := giturls.Parse(m.URI)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse module URI")
	}

	if u.Scheme == "file" {
		m.fs = osfs.New(strings.TrimPrefix(m.URI, "file://"))
		return m.fs, nil
	}

	m.fs = memfs.New()
	opts := &git.CloneOptions{
		URL:               m.URI,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Depth:             1,
	}

	if token, err := github.GetToken(); err == nil {
		opts.Auth = &githttp.BasicAuth{
			Username: "x-access-token",
			Password: string(token),
		}
	} else {
		logrus.WithError(err).Warn("failed to get github token, will use an unauthenticated client")
	}

	if m.Version != "" {
		opts.ReferenceName = plumbing.NewTagReferenceName(m.Version)
		opts.SingleBranch = true
	}

	// We don't use the git object here because all we care about is
	// the underlying filesystem object, which was created earlier
	if _, err := git.CloneContext(ctx, memory.NewStorage(), m.fs, opts); err != nil {
		// if tag not found try as a branch
		if !errors.Is(err, git.NoMatchingRefSpecError{}) {
			return nil, err
		}

		opts.ReferenceName = plumbing.NewBranchReferenceName(m.Version)
		if _, err := git.CloneContext(ctx, memory.NewStorage(), m.fs, opts); err != nil {
			return nil, errors.Wrap(err, "failed to find version as branch/tag")
		}
	}

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
