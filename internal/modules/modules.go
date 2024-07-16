// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file implements fetching modules for a given
// project manifest.

// Package modules implements all logic needed for interacting
// with stencil modules and their interaction with a project generated
// by stencil.
package modules

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"go.rgst.io/stencil/internal/modules/resolver"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
)

// resolvedModule is used to keep track of a module during the resolution
// stage, keeping track of the constraints that were used to resolve the
// module's version.
type resolvedModule struct {
	// Module is the underlying [Module] that was created for this module.
	*Module

	// history is the version resolution history for this module
	history []history

	// version is the version that was resolved for this module
	version *resolver.Version
}

type history struct {
	// parent is the name of the module that imported this module
	parent string

	// version is the version that was resolved for this module
	version *resolver.Version

	// criteria is the criteria that was used to resolve this module
	criteria *resolver.Criteria
}

// resolveModule is used to keep track of a module that needs to be resolved
type resolveModule struct {
	// conf is the configuration to be used to resolve the module
	conf *configuration.TemplateRepository

	// parent is the name of the module that imported this module
	parent string
}

// ModuleResolveOptions contains options for resolving modules
type ModuleResolveOptions struct {
	// Log is the logger to use
	Log slogext.Logger

	// Manifest is the project manifest to resolve modules for
	Manifest *configuration.Manifest

	// Replacements is a map of modules to use instead of ones specified
	// in the manifest. This is mainly used by tests but also in cases
	// where a specific version of a module should be used (e.g.,
	// lockfile).
	Replacements map[string]*Module
}

// criteriaForVersionString returns a resolver.Criteria for a given
// version string. This function will attempt to parse the version
// string as a constraint, then as a semver. If it's neither, it's
// assumed the input is a branch.
func criteriaForVersionString(version string) *resolver.Criteria {
	// Empty version, default to latest
	if version == "" {
		return &resolver.Criteria{
			Constraint: ">=0.0.0",
		}
	}

	// Attempt to parse as a constraint
	if c, err := semver.NewConstraint(version); err == nil {
		return &resolver.Criteria{
			Constraint: c.String(),
		}
	}

	// Attempt to parse as a version
	if _, err := semver.NewVersion(version); err == nil {
		return &resolver.Criteria{
			Constraint: "=" + version,
		}
	}

	// Otherwise, probably a branch.
	return &resolver.Criteria{
		Branch: version,
	}
}

// resolutionError returns an error for a failed module resolution
// with a given import path and history of constraints that were used
// to resolve the module.
func resolutionError(importPath string, history []history) error {
	errorString := ""
	for i := range history {
		h := &history[i]
		errorString += strings.Repeat(" ", i*2) + "└─ "

		wants := "*"
		switch {
		case h.criteria.Branch != "":
			wants = "branch " + h.criteria.Branch
		case h.criteria.Constraint != "":
			wants = h.criteria.Constraint
		}

		errorString += fmt.Sprintln(history[i].parent, "wants", wants)
	}

	return fmt.Errorf("failed to resolve module '%s' with constraints\n%s", importPath, errorString)
}

// FetchModules fetches modules for a given Manifest. See
// [ModuleResolveOptions] for more information on the various options
// that this function supports.
func FetchModules(ctx context.Context, opts *ModuleResolveOptions) ([]*Module, error) {
	// Used to track which modules to resolve and which one's have been
	// resolved, for returning later.
	resolveList := make([]resolveModule, 0)
	modules := make(map[string]*resolvedModule)

	// Create a new resolver
	r := resolver.NewResolver()

	// For each module in the manifest, add it to the list of modules
	// to be resolved.
	for _, m := range opts.Manifest.Modules {
		resolveList = append(resolveList, resolveModule{
			conf:   m,
			parent: opts.Manifest.Name + " (top-level)",
		})
	}

	// Resolve all versions, adding more to the stack as we go
	for len(resolveList) > 0 {
		mod := resolveList[0]
		importPath := mod.conf.Name
		wantedVerCriteria := criteriaForVersionString(mod.conf.Version)
		uri := uriForModule(importPath, opts.Manifest.Replacements[importPath])

		opts.Log.With("module", importPath).With("criteria", wantedVerCriteria).Debug("Resolving module")

		// version is the version to use for this module
		var version *resolver.Version

		// Initialize the module if it doesn't exist in the map.
		if _, ok := modules[importPath]; !ok {
			modules[importPath] = &resolvedModule{
				history: []history{},
			}
		}

		// Check if we've already attempted to resolve this module with this
		// criteria before. If we have, then we can skip resolving it again.
		var alreadyResolved bool
		for _, h := range modules[importPath].history {
			if h.criteria.Equal(wantedVerCriteria) {
				opts.Log.With("module", importPath).With("version", h.version).Debug("Already resolved module")
				// Log the attempt and remove the module from the list
				modules[importPath].history = append(modules[importPath].history, history{
					parent:   mod.parent,
					version:  h.version,
					criteria: wantedVerCriteria,
				})
				alreadyResolved = true
				break
			}
		}
		if alreadyResolved {
			resolveList = resolveList[1:]
			continue
		}

		// If we're using a local module or a replacement, we don't need to
		// resolve the version.
		if uriIsLocal(uri) {
			version = &resolver.Version{Virtual: "local"}
		} else if opts.Replacements[importPath] != nil {
			version = &resolver.Version{Virtual: "in-memory"}
		}

		// Add an entry to the history for this module. We add this before
		// looking up the version so that we know what requested this module
		// at resolve time.
		modules[importPath].history = append(modules[importPath].history, history{
			parent:   mod.parent,
			version:  version,
			criteria: wantedVerCriteria,
		})

		// No version, need to resolve it.
		if version == nil {
			// Use our criteria along with the previous criteria to resolve
			// the module version, if we have any.
			criteria := []*resolver.Criteria{wantedVerCriteria}
			for _, h := range modules[importPath].history {
				criteria = append(criteria, h.criteria)
			}

			var err error
			version, err = r.Resolve(ctx, uri, criteria...)
			if err != nil {
				return nil, resolutionError(importPath, modules[importPath].history)
			}

			// Track that we got this version for this module
			// TODO(jaredallard): Do this better.
			modules[importPath].history[len(modules[importPath].history)-1].version = version

			// log the attempt
			opts.Log.With("module", importPath).With("version", version).Debug("Resolved module")
		}

		// Use a module from the replacements if set, otherwise create one
		// from the resolved version.
		var m *Module
		if opts.Replacements[importPath] != nil {
			m = opts.Replacements[importPath]
			opts.Log.Debug("Using forced module version", "module", importPath, "version", m.Version)
		} else {
			var err error
			m, err = New(ctx, uri, NewModuleOpts{
				ImportPath: importPath,
				Version:    version,
			})
			if err != nil {
				return nil, err
			}

			opts.Log.With("module", importPath).With("version", version).Debug("Created module")
		}

		// Add the dependencies of this module to the stack to be resolved
		for _, mfm := range m.Manifest.Modules {
			opts.Log.With("module", importPath).With("dependency", mfm.Name).Debug("Adding dependency")
			resolveList = append(resolveList, resolveModule{
				conf:   mfm,
				parent: importPath + "@" + version.String(),
			})
		}

		// Update the module with the new version we found.
		modules[importPath].Module = m
		modules[importPath].version = version

		// Resolve the next module
		resolveList = resolveList[1:]
	}

	// Convert the resolved modules to a list of modules
	modulesList := make([]*Module, 0, len(modules))
	for _, m := range modules {
		modulesList = append(modulesList, m.Module)
	}
	return modulesList, nil
}
