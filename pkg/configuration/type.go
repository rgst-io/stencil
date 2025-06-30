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

// Description: types of stencil repos

package configuration

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// TemplateRepositoryType specifies what type of a stencil repository the current one is.
type TemplateRepositoryType string

// This block contains all of the TemplateRepositoryType values
const (
	// TemplateRepositoryTypeExt denotes a repository as being
	// an extension repository. This means that it contains
	// a go extension. This repository may also contain go-templates if
	// this type is used together with the TemplateRepositoryTypeTemplates.
	TemplateRepositoryTypeExt TemplateRepositoryType = "extension"

	// TemplateRepositoryTypeTemplates denotes a repository as being a standard template repository.
	// When the same module/repo serves more than one type, join this explicit value with other
	// types, e.g. "templates,extension".
	TemplateRepositoryTypeTemplates TemplateRepositoryType = "templates"
)

// TemplateRepositoryTypes specifies what type of a stencil repository the current one is.
// Use Contains to check for a type - it has special handling for the default case.
// Even though it is a struct, it is marshalled and unmarshalled as a string with comma separated
// values of TemplateRepositoryType.
type TemplateRepositoryTypes []TemplateRepositoryType

// UnmarshalYAML unmarshals TemplateRepositoryTypes from a string with comma-separated values.
func (ts *TemplateRepositoryTypes) UnmarshalYAML(value *yaml.Node) error {
	//nolint:exhaustive // Why: default catches ones we care about here.
	switch value.Kind {
	case yaml.ScalarNode:
		var csv string
		if err := value.Decode(&csv); err != nil {
			return fmt.Errorf("failed to parse as csv: %w", err)
		}

		for _, t := range strings.Split(csv, ",") {
			*ts = append(*ts, TemplateRepositoryType(t))
		}
	case yaml.SequenceNode:
		var types []TemplateRepositoryType
		if err := value.Decode(&types); err != nil {
			return fmt.Errorf("failed to parse as sequence: %w", err)
		}

		*ts = TemplateRepositoryTypes(types)
	default:
		return fmt.Errorf("unexpected yaml node kind: %v", value.Kind)
	}

	return nil
}

// Contains returns true if current repo needs to serve inpt type, with default assumed
// to be a templates-only repo (we do not support repos with no purpose).
func (ts TemplateRepositoryTypes) Contains(t TemplateRepositoryType) bool {
	if len(ts) == 0 {
		// empty types defaults to templates only (we do not support repos with no purpose)
		return t == TemplateRepositoryTypeTemplates
	}

	return slices.Contains(ts, t)
}
