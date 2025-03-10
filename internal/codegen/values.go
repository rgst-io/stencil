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

// Description: This file contains global variables that are
// exposed to templates at the root of the template arguments.
// (e.g. {{ .Repository.Name }})

package codegen

import (
	"context"

	gogit "github.com/go-git/go-git/v5"
	vcsgit "github.com/jaredallard/vcs/git"
	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/version"
	"go.rgst.io/stencil/v2/pkg/configuration"
)

// runtimeVals contains information about the current state
// of an application. This includes things like Golang
// version, stencil version, and other tool information.
type runtimeVals struct {
	// Generator is the name of the tool that is generating this file
	// generally this would be "stencil", this value should be changed
	// if using stencil's codegen package outside of the stencil CLI.
	Generator string

	// GeneratorVersion is the current version of the generator being
	// used.
	GeneratorVersion string

	// Modules contains a list of all modules that are being rendered
	// in a stencil run
	Modules modulesSlice
}

// git contains information about the current git repository
// that is being ran in
type git struct {
	// Ref is the current ref of the Git repository, this
	// is in the refs/<type>/<name> format
	Ref string

	// Commit is the current commit that this git repository is at
	Commit string

	// Dirty denotes if the current git repository is dirty or not.
	// Dirty is determined by having untracked changes to the current
	// index.
	Dirty bool

	// DefaultBranch is the default branch to use for this repository
	// generally this is equal to "main", but some repositories
	// use other values.
	DefaultBranch string
}

// config contains a small amount of configuration that
// originates from the project manifest and is propagated
// here.
type config struct {
	// Name is the name of this repository
	Name string
}

// module contains information about the current module that
// is rendering a template.
type module struct {
	// Name is the name of the current module
	Name string

	// Version is the version of the current module
	Version *resolver.Version
}

// stencilTemplate contains information about the current template
type stencilTemplate struct {
	// Name is the name of the template
	Name string
}

// modulesSlice is a list of modules with helpers on top of it
type modulesSlice []module

// ByName returns a module by name
func (m modulesSlice) ByName(name string) module {
	for _, mod := range m {
		if mod.Name == name {
			return mod
		}
	}

	return module{}
}

// Values is the top level container for variables being passed to a
// stencil template. When updating this struct, ensure that the receiver
// functions are updated to reflect the new fields.
type Values struct {
	// Git is information about the current git repository, if there is one
	Git git

	// Runtime is information about the current runtime environment
	Runtime runtimeVals

	// Config is strongly typed values from the project manifest
	Config config

	// Module is information about the current module being rendered
	Module module

	// Template is the name of the template being rendered
	Template stencilTemplate

	// Data is only available when a template is being rendered through
	// stencil.Include. It contains the data passed through said
	// call.
	Data any
}

// Copy returns a copy of the current values
func (v *Values) Copy() *Values {
	nv := *v
	return &nv
}

// WithModule returns a copy of the current values with the
// provided module information being set.
func (v *Values) WithModule(name string, ver *resolver.Version) *Values {
	nv := v.Copy()
	nv.Module.Name = name
	nv.Module.Version = ver
	return nv
}

// WithTemplate returns a copy of the current values with the
// provided template information being set.
func (v *Values) WithTemplate(name string) *Values {
	nv := v.Copy()
	nv.Template.Name = name
	return nv
}

// NewValues returns a fully initialized Values
// based on the current runtime environment.
func NewValues(ctx context.Context, sm *configuration.Manifest, mods []*modules.Module) *Values {
	vals := &Values{
		Git: git{},
		Runtime: runtimeVals{
			Generator:        "stencil",
			GeneratorVersion: version.Version.GitVersion,
			Modules:          modulesSlice{},
		},
		Config: config{
			Name: sm.Name,
		},
		Module:   module{},
		Template: stencilTemplate{},
	}

	for _, m := range mods {
		vals.Runtime.Modules = append(vals.Runtime.Modules, module{
			Name:    m.Name,
			Version: m.Version,
		})
	}

	// If we're a repository, add repository information
	if r, err := gogit.PlainOpen(""); err == nil {
		db, err := vcsgit.GetDefaultBranch(ctx, "")
		if err != nil {
			db = "main"
		}
		vals.Git.DefaultBranch = db

		// Add HEAD information
		if pref, err := r.Head(); err == nil {
			vals.Git.Ref = pref.Name().String()
			vals.Git.Commit = pref.Hash().String()
		}

		// Check if the worktree is clean
		if wrk, err := r.Worktree(); err == nil {
			if stat, err := wrk.Status(); err == nil {
				vals.Git.Dirty = !stat.IsClean()
			}
		}
	}

	return vals
}
