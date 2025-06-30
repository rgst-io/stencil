// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TemplateRepositoryManifest is a manifest of a template repository
type TemplateRepositoryManifest struct {
	// Name is the name of this template repository.
	// This must match the import path.
	Name string `yaml:"name" jsonschema:"required"`

	// Modules are template repositories that this manifest requires
	Modules []*TemplateRepository `yaml:"modules,omitempty"`

	// MinStencilVersion is the minimum version of stencil that is required to
	// render this module.
	MinStencilVersion string `yaml:"minStencilVersion,omitempty"`

	// Type stores a comma-separated list of template repository types served by the current module.
	// Use the TemplateRepositoryTypes.Contains method to check.
	Type TemplateRepositoryTypes `yaml:"type,omitempty"`

	// PostRunCommand is a command to be ran after rendering and post-processors
	// have been ran on the project
	PostRunCommand []*PostRunCommandSpec `yaml:"postRunCommand,omitempty"`

	// Arguments are a declaration of arguments to the template generator
	Arguments map[string]Argument `yaml:"arguments,omitempty"`

	// DirReplacements is a list of directory name replacement templates to render
	DirReplacements map[string]string `yaml:"dirReplacements,omitempty"`

	// ModuleHooks contains configuration for module hooks, keyed by their
	// name.
	ModuleHooks map[string]ModuleHook `yaml:"moduleHooks,omitempty"`
}

// PostRunCommandSpec is the spec of a command to be ran and its
// friendly name
type PostRunCommandSpec struct {
	// Name is the name of the command being ran, used for UX
	Name string `yaml:"name,omitempty"`

	// Command is the command to be ran, note: this is ran inside
	// of a bash shell.
	Command string `yaml:"command" jsonschema:"required"`
}

// Argument is a user-input argument that can be passed to
// templates
type Argument struct {
	// Description is a description of this argument.
	Description string `yaml:"description"`

	// Required denotes this argument as required.
	Required bool `yaml:"required,omitempty"`

	// Default is the default value for this argument if it's not set.
	// This cannot be set when required is true.
	Default any `yaml:"default,omitempty"`

	// Schema is a JSON schema, in YAML, for the argument.
	Schema map[string]any `yaml:"schema"`

	// From is a reference to an argument in another module, if this is
	// set, all other fields are ignored and instead the module referenced
	// field's are used instead. The name of the argument, the key in the map,
	// must be the same across both modules.
	From string `yaml:"from,omitempty"`
}

// ModuleHook contains configuration for a module hook.
type ModuleHook struct {
	// Schema is a JSON schema. When set this is used to validate all
	// module hook data as it is inserted.
	Schema map[string]any `yaml:"schema,omitempty"`
}

// LoadTemplateRepositoryManifest reads a template repository manifest
// from disk and returns it.
//
// In most cases, you should use LoadDefaultTemplateRepositoryManifest
// instead as it contains the standard locations for a manifest.
func LoadTemplateRepositoryManifest(path string) (*TemplateRepositoryManifest, error) {
	//nolint:gosec // Why: Not user input.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s TemplateRepositoryManifest
	if err := yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// LoadDefaultTemplateRepositoryManifest reads a template repository
// manifest from disk and returns it, using a standard set of locations.
func LoadDefaultTemplateRepositoryManifest() (*TemplateRepositoryManifest, error) {
	manifestFiles := []string{"manifest.yaml"}
	for _, file := range manifestFiles {
		if _, err := os.Stat(file); err == nil {
			return LoadTemplateRepositoryManifest(file)
		}
	}

	return nil, fmt.Errorf("no manifest found (searched %v)", manifestFiles)
}
