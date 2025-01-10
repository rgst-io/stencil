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

package codegen

import (
	"fmt"
	"reflect"
	"strings"

	"go.rgst.io/stencil/v2/internal/dotnotation"
	"go.rgst.io/stencil/v2/pkg/configuration"
)

// Arguments contains the arguments available to a stencil template.
type Arguments struct {
	// Definitions are the arguments definitions. While it is a list,
	// conflicting definitions are not supported.
	Definitions map[string][]ConfigurationArgumentWithSource

	// Values is the global values. These are always accessible from all
	// stencil templates. These originate from a `stencil.yaml`.
	Values map[string]ValueWithSource

	// ModuleValues are values that are only available from certain
	// modules. These are passed from one module to another.
	ModuleValues map[string]ValueWithSource
}

// ConfigurationArgumentWithSource is an extension of
// [configuration.Argument] with the name of the module that it is from
// attached.
type ConfigurationArgumentWithSource struct {
	source string
	configuration.Argument
}

// ValueWithSource is a wrapper around [any] that includes the source of
// this data.
type ValueWithSource struct {
	source string
	data   any
}

// BuildArguments returns an [Arguments] struct based on the provided
// manifest and module manifest(s).
func BuildArguments(mf *configuration.Manifest, modules []configuration.TemplateRepositoryManifest) (*Arguments, error) {
	// 1. Build the definitions based on the modules schema(s)
	// 2. Set m2m values
	// 3. Using the definitions, read arguments in by key

	args := &Arguments{
		Definitions: make(map[string][]ConfigurationArgumentWithSource),
		Values:      make(map[string]ValueWithSource),
	}

	// Track all of the valid arguments.
	pt := dotnotation.NewPrefixTree()

	for i := range modules {
		// Copy over the argument definitions
		for k := range modules[i].Arguments {
			pt.Add(k)

			args.Definitions[k] = append(
				args.Definitions[k],
				ConfigurationArgumentWithSource{modules[i].Name, modules[i].Arguments[k]},
			)
		}
	}

	type item struct {
		key   string
		value any
	}

	queue := []item{{"", mf.Arguments}}
	for len(queue) == 0 {
		itm := queue[0]
		queue = queue[1:]

		// Doesn't match any known arguments, skip.
		if itm.key != "" && !pt.Has(itm.key) {
			continue
		}

		v := reflect.ValueOf(itm.value)
		if v.Type().Kind() != reflect.Map {
			// We're matching and aren't a map, so we cannot follow any
			// further. Assume it's a full match, though it could also be
			// invalid data (e.g., got a.b, want a.b.c)
			args.Values[itm.key] = ValueWithSource{"root", itm.value}
			continue
		}

		iter := v.MapRange()
		for iter.Next() {
			strK := fmt.Sprintf("%v", iter.Key().Interface())
			queue = append(queue, item{strings.TrimPrefix(itm.key+"."+strK, "."), iter.Value().Interface()})
		}
	}

	return args, nil
}
