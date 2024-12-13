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

// Description: See package description

// Package stencil implements the stencil command, which is
// essentially a thing wrapper around the codegen package
// which does most of the heavy lifting.
package stencil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/internal/codegen"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"golang.org/x/mod/semver"
)

// Command is a thin wrapper around the codegen package that implements
// the "stencil" command. It is responsible for fetching dependencies,
// rendering templates, and writing files to disk using the underlying
// codegen package.
type Command struct {
	// lock is the current stencil lockfile at command creation time
	lock *stencil.Lockfile

	// manifest is the project manifest that is being used for this
	// template render
	manifest *configuration.Manifest

	// log is the logger used for logging output
	log slogext.Logger

	// dryRun denotes if we should write files to disk or not
	dryRun bool

	// adopt denotes if we should use heuristics to detect code that should go
	// into blocks to assist with first-time adoption of templates
	adopt bool
}

// printVersion is a command line friendly version of
// resolver.Version.String()
func printVersion(v *resolver.Version) string {
	switch {
	case v.Tag != "":
		return fmt.Sprintf("%s (%s)", v.Tag, v.Commit)
	case v.Branch != "":
		return fmt.Sprintf("branch %s (%s)", v.Branch, v.Commit)
	}

	return v.Commit
}

// NewCommand creates a new stencil command
func NewCommand(log slogext.Logger, s *configuration.Manifest, dryRun, adopt bool) *Command {
	l, err := stencil.LoadLockfile("")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warn("failed to load lockfile")
	}

	return &Command{
		lock:     l,
		manifest: s,
		log:      log,
		dryRun:   dryRun,
		adopt:    adopt,
	}
}

// useModulesFromLockfile returns a list of modules from the lockfile
// that should be used for this run of the stencil command.
//
// Modules import paths provided in 'skip' will be skipped and not
// returned in the modules slice.
func (c *Command) useModulesFromLockfile(ctx context.Context, skip map[string]struct{}) ([]*modules.Module, error) {
	if skip == nil {
		skip = make(map[string]struct{})
	}

	mods := make([]*modules.Module, 0, len(c.lock.Modules))
	for _, me := range c.lock.Modules {
		if _, ok := skip[me.Name]; ok {
			continue
		}

		m, err := modules.New(ctx, me.URL, modules.NewModuleOpts{
			ImportPath: me.Name,
			Version:    me.Version,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create module: %w", err)
		}

		mods = append(mods, m)
	}

	return mods, nil
}

// resolveModules fetches the modules for the project and returns them.
// If a lockfile is present, it will use the modules from the lockfile
// instead of resolving them. If ignoreLockfile is true, it will ignore
// the lockfile and resolve the modules anyways.
func (c *Command) resolveModules(ctx context.Context, ignoreLockfile bool) ([]*modules.Module, error) {
	// replacements contains module versions that should be used instead
	// of being resolved.
	replacements := make([]*modules.Module, 0)

	// If we have a lockfile, we also need to check if the modules list
	// has changed since the last run. If it has, we need to re-resolve
	// the changed modules.
	if c.lock != nil && !ignoreLockfile {
		manifestModulesHM := make(map[string]string)
		for _, m := range c.manifest.Modules {
			manifestModulesHM[m.Name] = m.Version
		}

		// Compare the modules from the lockfile vs the manifest to
		// determine which ones have changed.
		changed := make(map[string]struct{})
		for _, m := range c.lock.Modules {
			// If a version was changed to be replaced with a local version,
			// we also need to re-resolve it. We check before determining if
			// we're a "latest" module because we do want to allow
			// replacements to override those.
			if c.manifest.Replacements[m.Name] != "" && m.Version.Virtual != "local" {
				changed[m.Name] = struct{}{}
				continue
			}

			manifestEntryVer, ok := manifestModulesHM[m.Name]
			if manifestEntryVer == "" {
				// We shouldn't automatically re-resolve modules that don't ask
				// for a version, since that would be an unintended upgrade.
				continue
			}

			// If it doesn't exist anymore, ignore it. Ideally, we'd trigger a
			// re-resolve of the entire project, but because we can't track
			// direct dependencies we can't do this yet.
			if !ok {
				continue
			}

			// Check if the version changed. Because we can't create a
			// resolver.Version right now (we don't know if it's a branch or
			// version at this stage), we do a lame string check against all
			// of the version types.
			if !slices.Contains([]string{m.Version.Commit, m.Version.Tag, m.Version.Branch}, manifestEntryVer) {
				changed[m.Name] = struct{}{}
				continue
			}
		}

		var err error
		replacements, err = c.useModulesFromLockfile(ctx, changed)
		if err != nil {
			return nil, fmt.Errorf("failed to use modules from lock: %w", err)
		}
	}

	replacementsHM := slicesMap(replacements, func(m *modules.Module) string { return m.Name })

	// On first run, we need to resolve the modules. Otherwise, the user
	// will be expected to run 'stencil upgrade' to update the lockfile.
	return modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest:     c.manifest,
		Log:          c.log,
		Replacements: replacementsHM,
	})
}

// Upgrade checks for upgrades to the modules in the project and
// upgrades them if necessary. If no lockfile is present, it will
// log a message and return without doing anything.
func (c *Command) Upgrade(ctx context.Context) error {
	if c.lock == nil {
		c.log.Info("No lockfile found, run 'stencil' to fetch dependencies first")
		return nil
	}

	c.log.Info("Checking for upgrades")
	mods, err := c.resolveModules(ctx, true)
	if err != nil {
		return err
	}

	// Convert the lockfile modules to an easy importPath -> version
	// lookup.
	lockModules := make(map[string]*resolver.Version)
	if c.lock != nil {
		for _, m := range c.lock.Modules {
			lockModules[m.Name] = m.Version
		}
	}

	var hadChanges bool
	for _, new := range mods {
		c.log.Debug("Checking", "module", new.Name, "version", printVersion(new.Version))
		// If the module is in the lockfile, check if the version has
		// changed. If it has, log the change.
		if old, ok := lockModules[new.Name]; ok {
			if old.Equal(new.Version) {
				continue
			}

			c.log.Infof(" -> %s (%s -> %s)", new.Name, printVersion(old), printVersion(new.Version))
			hadChanges = true
		} else {
			c.log.Infof(" -> %s (%s)", new.Name, printVersion(new.Version))
			hadChanges = true
		}
	}
	if !hadChanges {
		c.log.Info("No new versions found")
		return nil
	}

	return c.runWithModules(ctx, mods)
}

// Run fetches dependencies of the root modules and builds the layered filesystem,
// after that GenerateFiles is called to actually walk the filesystem and render
// the templates. This step also does minimal post-processing of the dependencies
// manifests
func (c *Command) Run(ctx context.Context) error {
	c.log.Info("Fetching dependencies")
	mods, err := c.resolveModules(ctx, false)
	if err != nil {
		return err
	}

	for _, m := range mods {
		c.log.Infof(" -> %s %s", m.Name, printVersion(m.Version))

		if m.Manifest.MinStencilVersion != "" {
			// semver.Compare expects the version to be prefixed with "v"
			if semver.Compare("v"+version.Version.GitVersion, "v"+m.Manifest.MinStencilVersion) < 0 {
				return fmt.Errorf("stencil version %s is less than the required version %s for %s",
					version.Version.GitVersion, m.Manifest.MinStencilVersion, m.Name)
			}
		}
	}

	return c.runWithModules(ctx, mods)
}

// runWithModules runs the stencil command with the given modules
func (c *Command) runWithModules(ctx context.Context, mods []*modules.Module) error {
	st := codegen.NewStencil(c.manifest, c.lock, mods, c.log, c.adopt)
	defer st.Close()

	c.log.Info("Loading native extensions")
	if err := st.RegisterExtensions(ctx); err != nil {
		return err
	}

	c.log.Info("Rendering templates")
	tpls, err := st.Render(ctx, c.log)
	if err != nil {
		return err
	}

	if err := c.writeFiles(st, tpls); err != nil {
		return err
	}

	// Can't dry run post run yet
	if c.dryRun {
		c.log.Info("Skipping post-run commands, dry-run")
		return nil
	}

	return st.PostRun(ctx, c.log)
}

// writeFiles writes the files to disk
func (c *Command) writeFiles(st *codegen.Stencil, tpls []*codegen.Template) error {
	c.log.Infof("Writing template(s) to disk")
	for _, tpl := range tpls {
		for i := range tpl.Files {
			if err := tpl.Files[i].Write(c.log, c.dryRun); err != nil {
				return err
			}
		}
	}

	// Don't generate a lockfile in dry-run mode
	if c.dryRun {
		return nil
	}

	l := st.GenerateLockfile(tpls)
	if c.lock != nil {
		// Pull in older missing files (if any) from the last lock file
		l.MergeMissingInfoFromOlderLockfile(c.lock)
	}

	return l.Write()
}
