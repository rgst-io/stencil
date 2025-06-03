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

// Description: This file contains the code for the stencil.Arg
// template function.

package codegen

import (
	"fmt"
	"slices"

	"go.rgst.io/stencil/v2/internal/dotnotation"
	"go.rgst.io/stencil/v2/pkg/configuration"
)

// Arg returns the value of an argument in the project's manifest
//
//	{{- stencil.Arg "name" }}
func (s *TplStencil) Arg(pth string) (any, error) {
	if pth == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	mf := s.t.Module.Manifest
	if _, ok := mf.Arguments[pth]; !ok {
		return "", fmt.Errorf("module %q doesn't list argument %q as an argument in its manifest", s.t.Module.Name, pth)
	}
	arg := mf.Arguments[pth]

	// If there's a "from" we should handle that now before anything else,
	// so that its definition is used.
	if arg.From != "" {
		fromArg, err := s.resolveFrom(pth, s.t.Module.Manifest, &arg)
		if err != nil {
			return "", err
		}
		// Guaranteed to not be nil
		arg = *fromArg
	}

	mapInf := make(map[any]any)
	for k, v := range s.s.m.Arguments {
		mapInf[k] = v
	}

	// if not set then we return a default value based on the denoted type
	v, err := dotnotation.Get(mapInf, pth)
	if err != nil {
		v, err = s.resolveDefault(pth, &arg)
		if err != nil {
			return "", err
		}
	}

	// validate the data
	if arg.Schema != nil {
		if err := s.validateArg(pth, &arg, v); err != nil {
			return nil, err
		}
	}

	return v, nil
}

// resolveDefault resolves the default value of an argument from the manifest
func (s *TplStencil) resolveDefault(pth string, arg *configuration.Argument) (any, error) {
	if arg.Default != nil {
		return arg.Default, nil
	}

	if arg.Required {
		return nil, fmt.Errorf("module %q requires argument %q but is not set", s.t.Module.Name, pth)
	}

	// json schema convention is to define "type" as the top level key.
	typ, ok := arg.Schema["type"]
	if !ok {
		// We don't know what type this should be so return nothing.
		return nil, nil
	}
	typs, ok := typ.(string)
	if !ok {
		return nil, fmt.Errorf("module %q argument %q has invalid type: %v", s.t.Module.Name, pth, typ)
	}

	var v any
	switch typs {
	case "map", "object":
		v = make(map[any]any)
	case "list", "array":
		v = make([]any, 0)
	case "boolean", "bool":
		v = false
	case "integer", "int", "number":
		v = 0
	case "string":
		v = ""
	default:
		return "", fmt.Errorf("module %q argument %q has invalid type %q", s.t.Module.Name, pth, typs)
	}

	return v, nil
}

// resolveFrom resoles the "from" field of an argument
func (s *TplStencil) resolveFrom(pth string, mf *configuration.TemplateRepositoryManifest, arg *configuration.Argument) (*configuration.Argument, error) {
	foundModuleInDeps := slices.ContainsFunc(mf.Modules, func(m *configuration.TemplateRepository) bool {
		return m.Name == arg.From
	})
	if !foundModuleInDeps {
		return nil, fmt.Errorf(
			"module %q argument %q references an argument in module %q, but doesn't list it as a dependency",
			s.t.Module.Name, pth, arg.From,
		)
	}

	// Get the manifest for the referenced module
	var fromMf *configuration.TemplateRepositoryManifest
	for _, m := range s.s.modules {
		if m.Name == arg.From {
			fromMf = m.Manifest

			// Found the module, break
			break
		}
	}
	if fromMf == nil {
		return nil, fmt.Errorf(
			"module %q argument %q references an argument in module %q, but wasn't imported by stencil (this is a bug)",
			s.t.Module.Name, pth, arg.From,
		)
	}

	// Ensure that the module imported exposes that argument
	fromArg, ok := fromMf.Arguments[pth]
	if !ok {
		return nil, fmt.Errorf(
			"module %q argument %q references an argument in module %q, but the module does not expose that argument",
			s.t.Module.Name, pth, arg.From,
		)
	}

	// If we are, ourselves, a from then we need to resolve it again.
	if fromArg.From != "" {
		// Reusing 'pth' is safe because from key's must be equal.
		recurFromArg, err := s.resolveFrom(pth, fromMf, &fromArg)
		if err != nil {
			return nil, fmt.Errorf("recursive from resolve failed for module %s -> %s: %w", arg.From, fromArg.From, err)
		}
		return recurFromArg, nil
	}

	return &fromArg, nil
}

// validateArg validates an argument against the schema
func (s *TplStencil) validateArg(pth string, arg *configuration.Argument, v any) error {
	return validateJSONSchema(s.t.Module.Name+"/arguments/"+pth, arg.Schema, v)
}
