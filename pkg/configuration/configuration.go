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

// Package configuration implements configuration loading logic
// for stencil repositories and template repositories
package configuration

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// ValidateNameRegexp is the regex used to validate the project's name
const ValidateNameRegexp = `^[_a-z][_a-z0-9-]*$`

// NewManifest reads a manifest from disk at the
// specified path, parses it, and returns the output.
func NewManifest(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s *Manifest
	if err := yaml.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	if !ValidateName(s.Name) {
		return nil, fmt.Errorf("name field in %q was invalid", path)
	}

	return s, nil
}

// NewDefaultManifest returns a parsed project manifest
// from a set default path on disk.
func NewDefaultManifest() (*Manifest, error) {
	manifestFiles := []string{"stencil.yaml", "service.yaml"}
	for _, file := range manifestFiles {
		if _, err := os.Stat(file); err == nil {
			return NewManifest(file)
		}
	}

	return nil, fmt.Errorf("no manifest found (searched %v)", manifestFiles)
}

// Manifest is a manifest used to describe a project and impact
// what files are included
type Manifest struct {
	// Name is the name of the project
	Name string `yaml:"name" jsonschema:"required"`

	// Modules are the template modules that this project depends
	// on and utilizes
	Modules []*TemplateRepository `yaml:"modules,omitempty"`

	// Versions is a map of versions of certain tools, this is used by templates
	// and will likely be replaced with something better in the future.
	Versions map[string]string `yaml:"versions,omitempty"`

	// Arguments is a map of arbitrary arguments to pass to the generator
	Arguments map[string]interface{} `yaml:"arguments"`

	// Replacements is a list of module names to replace their URI.
	//
	// Expected format:
	// - local file: path/to/module
	// - remote file: https://github.com/getoutreach/stencil-base
	Replacements map[string]string `yaml:"replacements,omitempty"`
}

// TemplateRepository is a repository of template files.
type TemplateRepository struct {
	// Name is the name of this module. This should be a valid go import path
	Name string `yaml:"name" jsonschema:"required"`

	// Version is a semantic version or branch of the template repository
	// that should be downloaded if not set then the latest version is used.
	//
	// Version can also be a constraint as supported by the underlying
	// resolver:
	// https://pkg.go.dev/go.rgst.io/stencil/internal/modules/resolver
	//
	// But note that constraints are currently not locked so the version
	// will change as the module is resolved on subsequent runs.
	// Eventually, this will be changed to use the lockfile by default.
	Version string `yaml:"version,omitempty"`
}

// TemplateRepositoryManifest is a manifest of a template repository
type TemplateRepositoryManifest struct {
	// Name is the name of this template repository.
	// This must match the import path.
	Name string `yaml:"name" jsonschema:"required"`

	// Modules are template repositories that this manifest requires
	Modules []*TemplateRepository `yaml:"modules,omitempty"`

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
	Default interface{} `yaml:"default,omitempty"`

	// Schema is a JSON schema, in YAML, for the argument.
	Schema map[string]any `yaml:"schema"`

	// From is a reference to an argument in another module, if this is
	// set, all other fields are ignored and instead the module referenced
	// field's are used instead. The name of the argument, the key in the map,
	// must be the same across both modules.
	From string `yaml:"from,omitempty"`
}

// ValidateName ensures that the name of a project in the manifest
// fits the criteria we require.
func ValidateName(name string) bool {
	// This is more restrictive than the actual spec.  We're artificially
	// restricting ourselves to non-Unicode names because (in practice) we
	// probably don't support international characters very well, either.
	//
	// See also:
	// 	https://golang.org/ref/spec#Identifiers
	acceptableName := regexp.MustCompile(ValidateNameRegexp)
	return acceptableName.MatchString(name)
}
