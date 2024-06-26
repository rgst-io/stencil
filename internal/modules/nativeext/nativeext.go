// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package description.

// Package nativeext contains the logic for interacting with native
// extensions in stencil.
package nativeext

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	giturls "github.com/chainguard-dev/git-urls"
	"github.com/getoutreach/gobox/pkg/cfg"
	"github.com/getoutreach/gobox/pkg/cli/updater/archive"
	"github.com/getoutreach/gobox/pkg/cli/updater/release"
	"go.rgst.io/stencil/internal/git/vcs/github"
	"go.rgst.io/stencil/internal/modules/nativeext/apiv1"
	"go.rgst.io/stencil/internal/modules/resolver"
	"go.rgst.io/stencil/pkg/slogext"
)

// generatedTemplateFunc is the underlying type of a function
// generated by createFunctionFromTemplateFunction that's used
// to wrap the go plugin call to invoke said function
type generatedTemplateFunc func(...interface{}) (interface{}, error)

// Host implements an extension host that handles
// registering extensions and executing them.
type Host struct {
	r          *resolver.Resolver
	log        slogext.Logger
	extensions map[string]extension
}

// extension is an extension stored on an extension host
type extension struct {
	impl   apiv1.Implementation
	closer func() error
}

// NewHost creates a new extension host
func NewHost(log slogext.Logger) *Host {
	return &Host{
		r:          resolver.NewResolver(),
		log:        log,
		extensions: make(map[string]extension),
	}
}

// createFunctionFromTemplateFunction takes a given
// TemplateFunction and turns it into a callable function
func (h *Host) createFunctionFromTemplateFunction(extName string, ext apiv1.Implementation,
	fn *apiv1.TemplateFunction) generatedTemplateFunc {
	extPath := extName + "." + fn.Name

	return func(args ...interface{}) (interface{}, error) {
		if len(args) > fn.NumberOfArguments {
			return nil, fmt.Errorf("too many arguments, expected %d, got %d", fn.NumberOfArguments, len(args))
		}

		resp, err := ext.ExecuteTemplateFunction(&apiv1.TemplateFunctionExec{
			Name:      fn.Name,
			Arguments: args,
		})
		if err != nil {
			// return an error if the extension returns an error
			return nil, fmt.Errorf("failed to execute template function %q: %w", extPath, err)
		}

		// return the response, and a nil error
		return resp, nil
	}
}

// GetExtensionCaller returns an extension caller that's
// aware of all extension functions
func (h *Host) GetExtensionCaller(_ context.Context) (*ExtensionCaller, error) {
	// funcMap stores the extension functions discovered
	funcMap := map[string]map[string]generatedTemplateFunc{}

	// Call all extensions to get the template functions provided
	for extName, ext := range h.extensions {
		funcs, err := ext.impl.GetTemplateFunctions()
		if err != nil {
			return nil, fmt.Errorf("failed to get template functions from plugin %q: %w", extName, err)
		}

		for _, f := range funcs {
			h.log.With("extension", extName).With("function", f.Name).Debug("Registering extension function")
			tfunc := h.createFunctionFromTemplateFunction(extName, ext.impl, f)

			if _, ok := funcMap[extName]; !ok {
				funcMap[extName] = make(map[string]generatedTemplateFunc)
			}
			funcMap[extName][f.Name] = tfunc
		}
	}

	// return the lookup function, used via Call()
	return &ExtensionCaller{funcMap}, nil
}

// TODO(jaredallard)[DTSS-1926]: Refactor a lot of this RegisterExtension code.

// RegisterExtension registers a ext from a given source
// and compiles/downloads it. A client is then created
// that is able to communicate with the ext.
func (h *Host) RegisterExtension(ctx context.Context, source, name string, version *resolver.Version) error { //nolint:lll // Why: OK length.
	h.log.With("extension", name).With("source", source).Debug("Registered extension")

	u, err := giturls.Parse(source)
	if err != nil {
		return fmt.Errorf("failed to parse extension URL: %w", err)
	}

	var extPath string
	if u.Scheme == "file" {
		extPath = filepath.Join(strings.TrimPrefix(source, "file://"), "bin", "plugin")
	} else {
		extPath, err = h.downloadFromRemote(ctx, name, version)
	}
	if err != nil {
		return fmt.Errorf("failed to setup extension: %w", err)
	}

	ext, closer, err := apiv1.NewExtensionClient(ctx, extPath, h.log)
	if err != nil {
		return err
	}

	// Right now we don't have any configuration, so this serves as test
	// to ensure that the extension is working.
	if _, err := ext.GetConfig(); err != nil {
		return fmt.Errorf("failed to get config from extension: %w", err)
	}
	h.extensions[name] = extension{ext, closer}

	return nil
}

// RegisterInprocExtension registers an extension that is implemented
// within the same process directly with the host. Please limit the use
// of this API for unit testing only!
func (h *Host) RegisterInprocExtension(name string, ext apiv1.Implementation) {
	h.log.With("extension", name).Debug("Registered inproc extension")
	h.extensions[name] = extension{ext, func() error { return nil }}
}

// getExtensionPath returns the path to an extension binary
func (h *Host) getExtensionPath(version *resolver.Version, name string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Example:
	// $HOME/.cache/stencil/nativeexts/github.com/rgst-io/plugin/@v1.3.0/plugin
	path := filepath.Join(
		// TODO(jaredallard): Support XDG_CACHE_HOME.
		homeDir, ".cache", "stencil", "nativeexts",
		name, fmt.Sprintf("@%s", version.Commit), filepath.Base(name),
	)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return path, nil
}

// downloadFromRemote downloads a release from github and extracts it to disk
//
// using the example extension module: go.rgst.io/stencil-plugin
//
//	org: getoutreach
//	repo: stencil-plugin
//	name: go.rgst.io/stencil-plugin
func (h *Host) downloadFromRemote(ctx context.Context, name string,
	version *resolver.Version) (string, error) {
	token, err := github.Token()
	if err != nil {
		h.log.WithError(err).Warn("Failed to get github token, falling back to anonymous")
	}

	repoURL := "https://" + name

	// Check if the version we're pulling already exists on disk
	dlPath, err := h.getExtensionPath(version, name)
	if err != nil {
		return "", fmt.Errorf("failed to get extension path: %w", err)
	}
	if info, err := os.Stat(dlPath); err == nil && info.Mode() == 0o755 {
		return dlPath, nil
	}

	h.log.With("version", version).With("repo", repoURL).Debug("Downloading native extension")
	a, archiveName, _, err := release.Fetch(ctx, cfg.SecretData(token), &release.FetchOptions{
		AssetName: filepath.Base(name) + "_*_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz",
		RepoURL:   repoURL,
		Tag:       version.Tag,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch release: %w", err)
	}

	bin, _, err := archive.Extract(ctx, archiveName, a, archive.WithFilePath(filepath.Base(name)))
	if err != nil {
		return "", fmt.Errorf("failed to extract archive: %w", err)
	}

	f, err := os.Create(dlPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, bin); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	f.Close()

	// Ensure the file is executable.
	if err := os.Chmod(dlPath, 0o755); err != nil {
		return "", fmt.Errorf("failed to ensure plugin is executable: %w", err)
	}

	return dlPath, nil
}

// Close terminates the extension host, which in turn stops
// all current native extensions
func (h *Host) Close() error {
	var errs []error
	for _, ext := range h.extensions {
		if err := ext.closer(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
