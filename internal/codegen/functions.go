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

package codegen

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"go.rgst.io/stencil/v2/internal/yaml"
)

// TplError is a wrapper for an [error] that can be returned by function
// templates. This is due to [template] turning them into runtime panics
// by default.
type TplError struct {
	err error
}

// Error returns the error message.
func (e TplError) Error() string {
	return e.err.Error()
}

// Unwrap returns the underlying error.
func (e TplError) Unwrap() error {
	return e.err
}

// dereference dereferences a pointer returning the
// referenced data type. If the provided value is not
// a pointer, it is returned.
func dereference(i any) any {
	infType := reflect.TypeOf(i)

	// If not a pointer, noop
	if infType.Kind() != reflect.Ptr {
		return i
	}

	return reflect.ValueOf(i).Elem().Interface()
}

// quotejoinstrings takes a slice of strings and joins
// them with the provided separator, sep, while quoting all
// values
func quotejoinstrings(elems []string, sep string) string {
	for i := range elems {
		elems[i] = fmt.Sprintf("%q", elems[i])
	}
	return strings.Join(elems, sep)
}

// toYAML is a clone of the helm toYaml function, which takes
// an any and turns it into yaml
//
// Based on:
// https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/engine/funcs.go#L83
func toYAML(v any) (string, error) {
	// If no data, return an empty string
	if v == nil {
		return "", nil
	}

	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(data), "\n"), nil
}

// fromYAML converts a YAML document into [any].
//
// Based on: https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/engine/funcs.go#L98
func fromYAML(str string) (any, error) {
	var m any

	if err := yaml.Unmarshal([]byte(str), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// toJSON converts a any into a JSON document.
func toJSON(v any) (string, error) {
	// If no data, return an empty string
	if v == nil {
		return "", nil
	}

	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(data), "\n"), nil
}

// fromJSON converts a JSON document into a any.
func fromJSON(str string) (any, error) {
	var m any

	if err := json.Unmarshal([]byte(str), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// tplError creates a new [TplError] with the provided text
func tplError(text string) TplError {
	return TplError{errors.New(text)}
}

// Default are stock template functions that don't impact
// the generation of a file. Anything that does that should be located
// in the scope of the file renderer function instead
var Default = template.FuncMap{
	"Dereference":      dereference,
	"QuoteJoinStrings": quotejoinstrings,
	"toYaml":           toYAML,
	"fromYaml":         fromYAML,
	"toJson":           toJSON,
	"fromJson":         fromJSON,
	"error":            tplError,
}
