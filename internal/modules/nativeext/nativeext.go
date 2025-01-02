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
	"syscall"

	"github.com/jaredallard/archives"
	"github.com/jaredallard/vcs/releases"
	"github.com/jaredallard/vcs/resolver"
	"github.com/rogpeppe/go-internal/lockedfile"
	"go.rgst.io/stencil/internal/modules/nativeext/apiv1"
	"go.rgst.io/stencil/pkg/slogext"
)

// generatedTemplateFunc is the underlying type of a function
// generated by createFunctionFromTemplateFunction that's used
// to wrap the go plugin call to invoke said function
type generatedTemplateFunc func(...interface{}) (interface{}, error)

// Host implements an extension host that handles
// registering extensions and executing them.
type Host struct {
	// mu guards the [Host] against all possible instances of stencil
	// being ran on the current host under the current user. It is
	// important to ensure it is locked whenever storing extensions in the
	// global cache.
	mu         *lockedfile.Mutex
	r          *resolver.Resolver
	log        slogext.Logger
	extensions map[string]extension
}

// extension is an extension stored on an extension host
type extension struct {
	impl   apiv1.Implementation
	closer func() error
}

// getCacheDir returns the directory where extensions are cached in
func getCacheDir() (string, error) {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" { // default to $HOME/.cache as per XDG spec
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	return filepath.Join(cacheDir, "stencil"), nil
}

// NewHost creates a new extension host
func NewHost(log slogext.Logger) (*Host, error) {
	cacheDir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	mu := &lockedfile.Mutex{Path: filepath.Join(cacheDir, "cache.lock")}
	return &Host{
		mu:         mu,
		r:          resolver.NewResolver(),
		log:        log,
		extensions: make(map[string]extension),
	}, nil
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

// RegisterExtension registers a ext from a given source
// and compiles/downloads it. A client is then created
// that is able to communicate with the ext.
func (h *Host) RegisterExtension(ctx context.Context, source, name string, version *resolver.Version) error { //nolint:lll // Why: OK length.
	h.log.With("extension", name).With("source", source).Debug("Registered extension")

	var extPath string
	var err error
	if version.Virtual == "local" {
		extPath = filepath.Join(source, "bin", "plugin")
	} else {
		extPath, err = h.downloadFromRemote(ctx, source, name, version)
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
	cacheDir, err := getCacheDir()
	if err != nil {
		return "", err
	}

	// Example:
	//
	// $XDG_CACHE_HOME/stencil/nativeexts/github.com--rgst-io--plugin/v1.3.0/plugin
	path := filepath.Join(
		cacheDir, "nativeexts",
		strings.ReplaceAll(name, "/", "--"), version.Commit,
		filepath.Base(name),
	)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return path, nil
}

// downloadFromRemote downloads a release from github and extracts it to
// disk
func (h *Host) downloadFromRemote(ctx context.Context, source, name string, version *resolver.Version) (string, error) {
	if unlock, err := h.mu.Lock(); err != nil {
		h.log.WithError(err).Warn("failed to lock extension cache")
	} else {
		defer unlock()
	}

	// Check if the version we're pulling already exists on disk
	dlPath, err := h.getExtensionPath(version, name)
	if err != nil {
		return "", fmt.Errorf("failed to get extension path: %w", err)
	}
	if info, err := os.Stat(dlPath); err == nil && info.Mode() == 0o755 {
		h.log.With("name", name, "path", dlPath).Debug("using cached extension binary")
		return dlPath, nil
	}

	h.log.With("version", version).With("repo", source).Debug("Downloading native extension")
	resp, fi, err := releases.Fetch(ctx, &releases.FetchOptions{
		AssetNames: []string{
			filepath.Base(name) + "_*_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.*",
			filepath.Base(name) + "_*_" + runtime.GOOS + "_" + runtime.GOARCH + ".zip",
		},
		RepoURL: source,
		Tag:     version.Tag,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Close()

	a, err := archives.Open(resp, archives.OpenOptions{
		Extension: archives.Ext(fi.Name()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer a.Close()

	bin, err := archives.Pick(a, archives.PickFilterByName(filepath.Base(name)))
	if err != nil {
		return "", fmt.Errorf("failed to grab binary from archive: %w", err)
	}

	// Lock ForkLock whenever we are writing to a file that will execute
	// shortly after, to prevent its FD from leaking into a forked process
	// and thus making exec fail with ETXTBSY.
	//
	// See: https://github.com/golang/go/issues/22315
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()

	f, err := os.OpenFile(dlPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, bin); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
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
