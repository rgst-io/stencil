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

// Description: Implements the stencil function passed to templates
package codegen

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5/util"
	"github.com/jaredallard/cmdexec"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/nativeext"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/extensions/apiv1"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
)

// NewStencil creates a new, fully initialized Stencil renderer function
func NewStencil(m *configuration.Manifest, lock *stencil.Lockfile, mods []*modules.Module, log slogext.Logger, adopt bool) *Stencil {
	return &Stencil{
		log:                 log,
		m:                   m,
		ext:                 nativeext.NewHost(log),
		lock:                lock,
		modules:             mods,
		preRenderStageLimit: 20,
		sharedState:         newSharedState(),
		adoptMode:           adopt,
	}
}

// renderStage denotes the stage that a [Stencil] struct is in. See the
// comments on the const values for more information.
type renderStage int

const (
	// renderStagePre is the initial render stage. This stage denotes that
	// [sharedState] is still changing and more iterations may be needed.
	renderStagePre renderStage = iota

	// renderStageFinal is the final render stage. This stage denotes that
	// [sharedState] is stable and we can move on to the final render
	// stage. This is the stage where files are actually written.
	renderStageFinal
)

// Stencil provides the basic functions for
// stencil templates
type Stencil struct {
	log slogext.Logger
	m   *configuration.Manifest

	ext       *nativeext.Host
	extCaller *nativeext.ExtensionCaller

	lock *stencil.Lockfile

	// modules is a list of modules used in this stencil render
	modules []*modules.Module

	// renderStage is the current [renderStage] that the stencil is in.
	renderStage renderStage

	// preRenderStageLimit is the number of iterations to allow the
	// pre-render stage to run for. This is used to prevent infinite
	// loops in templates.
	preRenderStageLimit int

	// sharedState is the shared state between all templates.
	sharedState *sharedState

	// adoptMode denotes if we should use heuristics to detect code that should go
	// into blocks to assist with first-time adoption of templates
	adoptMode bool
}

// RegisterExtensions registers all extensions on the currently loaded
// modules.
func (s *Stencil) RegisterExtensions(ctx context.Context) error {
	for _, m := range s.modules {
		if err := m.RegisterExtensions(ctx, s.ext); err != nil {
			return errors.Wrapf(err, "failed to load extensions from module %q", m.Name)
		}
	}

	return nil
}

// RegisterInprocExtensions registers the input ext extension directly. This API is used in
// unit tests to render modules with templates that invoke native extensions: input 'ext' can be
// either an actual extension or a mock one (feeding fake data into the template).
func (s *Stencil) RegisterInprocExtensions(name string, ext apiv1.Implementation) {
	s.ext.RegisterInprocExtension(name, ext)
}

// GenerateLockfile generates a stencil.Lockfile based
// on a list of templates.
func (s *Stencil) GenerateLockfile(tpls []*Template) *stencil.Lockfile {
	l := &stencil.Lockfile{
		Version: version.Version.GitVersion,
	}

	for _, tpl := range tpls {
		for _, f := range tpl.Files {
			// Don't write files we skipped, or deleted, to the lockfile
			if f.Skipped || f.Deleted {
				continue
			}

			l.Files = append(l.Files, &stencil.LockfileFileEntry{
				Name:     f.Name(),
				Template: tpl.Path,
				Module:   tpl.Module.Name,
			})
		}
	}

	for _, m := range s.modules {
		l.Modules = append(l.Modules, &stencil.LockfileModuleEntry{
			Name:    m.Name,
			URL:     m.URI,
			Version: m.Version,
		})
	}

	l.Sort()

	return l
}

// Render renders all templates using the Manifest that was
// provided to stencil at creation time, returned is the templates
// that were produced and their associated files.
func (s *Stencil) Render(ctx context.Context, log slogext.Logger) ([]*Template, error) {
	tplfiles, err := s.getTemplates(ctx, log)
	if err != nil {
		return nil, err
	}

	if s.extCaller, err = s.ext.GetExtensionCaller(ctx); err != nil {
		return nil, err
	}

	log.Debug("Creating values for template")
	vals := NewValues(ctx, s.m, s.modules)
	log.Debug("Finished creating values")

	// Add the templates to their modules template to allow them to be able to access
	// functions declared in the same module
	for _, t := range tplfiles {
		log.Debugf("Parsing template %s", t.ImportPath())
		if err := t.Parse(s); err != nil {
			return nil, errors.Wrapf(err, "failed to parse template %q", t.ImportPath())
		}
	}

	// Render until we limit or state is stable
	var lastHash uint64
	var i int
	for {
		if i > (s.preRenderStageLimit - 1) {
			return nil, fmt.Errorf("failed to stabilize shared state within %d iterations", i)
		}

		log.Debug("Render stage", "iteration", i)
		for _, t := range tplfiles {
			log.Debugf("Render template %s", t.ImportPath())
			if err := t.Render(s, vals); err != nil {
				return nil, errors.Wrapf(err, "failed to render template %q", t.ImportPath())
			}

			// Don't keep files, we only need the shared state modifications.
			t.Files = nil
		}

		// Calculate the hash of the shared state
		hash, err := s.sharedState.hash()
		if err != nil {
			return nil, fmt.Errorf("failed to determine a stable hash for shared state: %w", err)
		}
		if hash == lastHash {
			log.Debugf("First pass render stable after %d iterations", i)
			break
		}

		lastHash = hash
		i++
	}

	// We're at the final render stage now.
	s.renderStage = renderStageFinal

	if err := s.calcDirReplacements(vals); err != nil {
		return nil, err
	}

	tpls := make([]*Template, 0)
	for _, t := range tplfiles {
		log.Debugf("Final render of template %s", t.ImportPath())
		if err := t.Render(s, vals); err != nil {
			return nil, errors.Wrapf(err, "failed to render template %q", t.ImportPath())
		}

		// append the rendered template to our list of templates processed
		tpls = append(tpls, t)
	}

	return tpls, nil
}

// calcDirReplacements calculates all of the final rendered paths for dirReplacements for each module
// It needs to be in stencil because it uses rendering, which needs the Values object from codegen,
// so we poke the rendered replacements into the module object for applying later in various ways.
func (s *Stencil) calcDirReplacements(vals *Values) error {
	for _, m := range s.modules {
		reps := map[string]string{}
		for dsrc, dtmp := range m.Manifest.DirReplacements {
			// Render replacement
			nn, err := s.renderDirReplacement(dtmp, m, vals)
			if err != nil {
				return err
			}
			reps[dsrc] = nn
		}
		m.StoreDirReplacements(reps)
	}
	return nil
}

// renderDirReplacement breaks out the actual rendering for calcDirReplacements to make it unit testable
func (s *Stencil) renderDirReplacement(template string, m *modules.Module, vals *Values) (string, error) {
	rt, err := NewTemplate(m, "dirReplace", 0o000, time.Time{}, []byte(template), s.log, nil)
	if err != nil {
		return "", err
	}

	if err := rt.Render(s, vals); err != nil {
		return "", err
	}

	nn := rt.Files[0].String()
	if strings.Contains(nn, string(os.PathSeparator)) {
		return "", fmt.Errorf("directory replacement of %s to %s contains path separator in output", template, nn)
	}

	return nn, nil
}

// PostRun runs all post run commands specified in the modules that
// this project depends on
func (s *Stencil) PostRun(ctx context.Context, log slogext.Logger) error {
	log.Info("Running post-run command(s)")

	type postRunCommand struct {
		Module string
		Spec   *configuration.PostRunCommandSpec
	}

	postRunCommands := []*postRunCommand{}

	// Check if the project has a '.mise.toml' and that 'mise' is
	// installed. If so, automatically trust the config.
	//
	// Note: This is to fix a special case with mise being used across
	// various language's modules. If anything else needs to be added
	// here, consider adding it to the post run commands array for a
	// specific module first. Otherwise, consider changing this system to
	// prevent more cases from being added to the code.
	if _, err := os.Stat(".mise.toml"); err == nil {
		// Check if 'mise' is in path to ensure this is less likely to fail.
		//
		//nolint:errcheck // Why: Checking path only
		if misePath, _ := exec.LookPath("mise"); misePath != "" {
			postRunCommands = append(postRunCommands, &postRunCommand{
				Module: "stencil",
				Spec: &configuration.PostRunCommandSpec{
					Name:    "mise: trust config",
					Command: "mise trust --quiet",
				},
			})
		}
	}

	// Get all post run commands from all modules
	for _, m := range s.modules {
		for _, prc := range m.Manifest.PostRunCommand {
			postRunCommands = append(postRunCommands, &postRunCommand{
				Module: m.Name,
				Spec:   prc,
			})
		}
	}

	for _, prc := range postRunCommands {
		log.Infof(" - %s (source: %s)", prc.Spec.Name, prc.Module)
		cmd := cmdexec.CommandContext(ctx, "/usr/bin/env", "bash", "-c", prc.Spec.Command)
		cmd.UseOSStreams(true)
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "failed to run post run command for module %q", prc.Module)
		}
	}

	return nil
}

// getTemplates takes all modules attached to this stencil
// struct and returns all templates exposed by it.
func (s *Stencil) getTemplates(ctx context.Context, log slogext.Logger) ([]*Template, error) {
	tpls := make([]*Template, 0)
	for _, m := range s.modules {
		log.Debugf("Fetching module %q", m.Name)
		fs, err := m.GetFS(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read module filesystem %q", m.Name)
		}

		// Note: This error should never really fire since we already fetched the FS above
		// that being said, we handle it here. Skip native extensions as they cannot have templates.
		if !m.Manifest.Type.Contains(configuration.TemplateRepositoryTypeTemplates) {
			log.Debugf("Skipping template discovery for module %q, not a template module (type %s)", m.Name, m.Manifest.Type)
			continue
		}

		log.Debugf("Discovering templates from module %q", m.Name)

		// Only find templates in the templates/ directory
		fs, err = fs.Chroot("templates")
		if err != nil {
			return nil, errors.Wrap(err, "failed to chroot module filesystem to templates/ (does it exist?)")
		}

		err = util.Walk(fs, "", func(path string, inf os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// add binary check here

			// Skip files without a .tpl extension
			isTemplate := filepath.Ext(path) == ".tpl"
			isBinary := filepath.Ext(path) == ".nontpl"
			if !isTemplate && !isBinary {
				return nil
			}

			f, err := fs.Open(path)
			if err != nil {
				return errors.Wrapf(err, "failed to open template %q from module %q", path, m.Name)
			}
			defer f.Close()

			tplContents, err := io.ReadAll(f)
			if err != nil {
				return errors.Wrapf(err, "failed to read template %q from module %q", path, m.Name)
			}

			log.Debugf("Discovered template %q", path)
			tpl, err := NewTemplate(m, path, inf.Mode(), inf.ModTime(), tplContents, log, &NewTemplateOpts{
				Adopt:  s.adoptMode,
				Binary: isBinary,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to create template %q from module %q", path, m.Name)
			}
			tpls = append(tpls, tpl)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	log.Debug("Finished discovering templates")

	// Shuffle the templates to prevent accidental file order guarantees
	// from being relied upon.
	//nolint:gosec // Why: We don't need that much entropy.
	rand.Shuffle(len(tpls), func(i, j int) {
		tpls[i], tpls[j] = tpls[j], tpls[i]
	})

	return tpls, nil
}

// Close closes all resources that should be closed when done
// rendering templates.
func (s *Stencil) Close() error {
	return errors.Wrap(s.ext.Close(), "failed to close native extensions")
}
