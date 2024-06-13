// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: See package description

// Package stencil implements the stencil command, which is
// essentially a thing wrapper around the codegen package
// which does most of the heavy lifting.
package stencil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.rgst.io/stencil/internal/codegen"
	"go.rgst.io/stencil/internal/git/vcs/github"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/resolver"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"gopkg.in/yaml.v3"
)

// Command is a thin wrapper around the codegen package that
// implements the "stencil" command.
type Command struct {
	// lock is the current stencil lockfile at command creation time
	lock *stencil.Lockfile

	// manifest is the project manifest that is being used
	// for this template render
	manifest *configuration.Manifest

	// log is the logger used for logging output
	log slogext.Logger

	// dryRun denotes if we should write files to disk or not
	dryRun bool

	// frozenLockfile denotes if we should use versions from the lockfile
	// or not
	frozenLockfile bool

	// allowMajorVersionUpgrade denotes if we should allow major version
	// upgrades without a prompt or not
	allowMajorVersionUpgrades bool

	// token is the github token used for fetching modules
	token string
}

// NewCommand creates a new stencil command
func NewCommand(log slogext.Logger, s *configuration.Manifest,
	dryRun, frozen, allowMajorVersionUpgrades bool) *Command {
	l, err := stencil.LoadLockfile("")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithError(err).Warn("failed to load lockfile")
	}
	token, err := github.Token()
	if err != nil {
		log.Warn("failed to get github token, using anonymous access")
	}

	return &Command{
		lock:                      l,
		manifest:                  s,
		log:                       log,
		dryRun:                    dryRun,
		frozenLockfile:            frozen,
		allowMajorVersionUpgrades: allowMajorVersionUpgrades,
		token:                     token,
	}
}

// Run fetches dependencies of the root modules and builds the layered filesystem,
// after that GenerateFiles is called to actually walk the filesystem and render
// the templates. This step also does minimal post-processing of the dependencies
// manifests
func (c *Command) Run(ctx context.Context) error {
	c.log.Info("Fetching dependencies")
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: c.manifest,
		Log:      c.log,
	})
	if err != nil {
		return errors.Wrap(err, "failed to process modules list")
	}

	for _, m := range mods {
		c.log.Infof(" -> %s %s", m.Name, func(v *resolver.Version) string {
			switch {
			case v.Tag != "":
				return fmt.Sprintf("%s (%s)", v.Tag, v.Commit)
			case v.Branch != "":
				return fmt.Sprintf("branch %s (%s)", v.Branch, v.Commit)
			}

			return v.Commit
		}(m.Version))
	}

	st := codegen.NewStencil(c.manifest, mods, c.log)
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

// writeFile writes a codegen.File to disk based on its current state
func (c *Command) writeFile(f *codegen.File) error {
	action := "Created"
	if f.Deleted {
		action = "Deleted"

		if !c.dryRun {
			os.Remove(f.Name())
		}
	} else if f.Skipped {
		action = "Skipped"
	} else if _, err := os.Stat(f.Name()); err == nil {
		action = "Updated"
	}

	if action == "Created" || action == "Updated" {
		if !c.dryRun {
			if err := os.MkdirAll(filepath.Dir(f.Name()), 0o755); err != nil {
				return errors.Wrapf(err, "failed to ensure directory for %q existed", f.Name())
			}

			if err := os.WriteFile(f.Name(), f.Bytes(), f.Mode()); err != nil {
				return errors.Wrapf(err, "failed to create %q", f.Name())
			}
		}
	}

	msg := fmt.Sprintf("  -> %s %s", action, f.Name())
	if c.dryRun {
		msg += " (dry-run)"
	}

	if !f.Skipped {
		c.log.Info(msg)
	} else {
		c.log.Debug(msg, "reason", f.SkippedReason)
	}
	return nil
}

// writeFiles writes the files to disk
func (c *Command) writeFiles(st *codegen.Stencil, tpls []*codegen.Template) error {
	c.log.Infof("Writing template(s) to disk")
	for _, tpl := range tpls {
		for i := range tpl.Files {
			if err := c.writeFile(tpl.Files[i]); err != nil {
				return err
			}
		}
	}

	// Don't generate a lockfile in dry-run mode
	if c.dryRun {
		return nil
	}

	l := st.GenerateLockfile(tpls)
	f, err := os.Create(stencil.LockfileName)
	if err != nil {
		return errors.Wrap(err, "failed to create lockfile")
	}
	defer f.Close()

	return errors.Wrap(yaml.NewEncoder(f).Encode(l),
		"failed to encode lockfile into yaml")
}
