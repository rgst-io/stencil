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

// Description: This file contains the public API for templates
// for stencil

package codegen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/pkg/slogext"
)

// TplStencil contains the global functions available to a template for
// interacting with stencil.
type TplStencil struct {
	// s is the underlying stencil object that this is attached to
	s *Stencil

	// t is the current template in the context of our render
	t *Template

	log slogext.Logger
}

// GetModuleHook returns a module block in the scope of this module
//
// This is incredibly useful for allowing other modules to write
// to files that your module owns. Think of them as extension points
// for your module. The value returned by this function is always a
// []any, aka a list.
//
//	{{- /* This returns a []any */}}
//	{{ $hook := stencil.GetModuleHook "myModuleHook" }}
//	{{- range $hook }}
//	  {{ . }}
//	{{- end }}
func (s *TplStencil) GetModuleHook(name string) []any {
	// On the first pass, never return any data. If we did, the data would
	// be unreliably set because we don't sort the templates in any way or
	// guarantee that they will be rendered in specific any order.
	if s.s.isFirstPass {
		return []any{}
	}

	k := s.s.sharedData.key(s.t.Module.Name, name)
	v := s.s.sharedData.moduleHooks[k]
	if v == nil {
		// No data, return nothing
		return []any{}
	}

	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(v)).
		Debug("getting module hook")

	return v.values
}

// SetGlobal sets a global to be used in the context of the current template module
// repository. This is useful because sometimes you want to define variables inside
// of a helpers template file after doing manifest argument processing and then use
// them within one or more template files to be rendered; however, go templates limit
// the scope of symbols to the current template they are defined in, so this is not
// possible without external tooling like this function.
//
// This template function stores (and its inverse, GetGlobal, retrieves) data that is
// not strongly typed, so use this at your own risk and be averse to panics that could
// occur if you're using the data it returns in the wrong way.
//
//	{{- /* This writes a global into the current context of the template module repository */}}
//	{{- stencil.SetGlobal "IsGeorgeCool" true -}}
func (s *TplStencil) SetGlobal(name string, data any) string {
	// Only modify on first pass
	if !s.s.isFirstPass {
		return ""
	}

	k := s.s.sharedData.key(s.t.Module.Name, name)
	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(data)).
		Debug("adding to global store")

	s.s.sharedData.globals[k] = global{
		template: s.t.Path,
		value:    data,
	}

	return ""
}

// GetGlobal retrieves a global variable set by SetGlobal. The data returned from this function
// is unstructured so by averse to panics - look at where it was set to ensure you're dealing
// with the proper type of data that you think it is.
//
//	{{- /* This retrieves a global from the current context of the template module repository */}}
//	{{ $isGeorgeCool := stencil.GetGlobal "IsGeorgeCool" }}
func (s *TplStencil) GetGlobal(name string) any {
	// Never return any data during the first pass because that would be
	// non-deterministic.
	if s.s.isFirstPass {
		return nil
	}

	k := s.s.sharedData.key(s.t.Module.Name, name)

	v, ok := s.s.sharedData.globals[k]
	if !ok {
		s.log.With("template", s.t.ImportPath(), "path", k).
			Warn("failed to retrieved data from global store")
		return nil
	}

	s.log.With(
		"template", s.t.ImportPath(),
		"path", k,
		"data", spew.Sdump(v),
		"definingTemplate", v.template,
	).Debug("retrieved data from global store")
	return v.value
}

// AddToModuleHook adds to a hook in another module
//
// This functions write to module hook owned by another module for
// it to operate on. These are not strongly typed so it's best practice
// to look at how the owning module uses it for now. Module hooks must always
// be written to with a list to ensure that they can always be written to multiple
// times.
//
//	{{- /* This writes to a module hook */}}
//	{{- stencil.AddToModuleHook "github.com/myorg/repo" "myModuleHook" (list "myData") }}
func (s *TplStencil) AddToModuleHook(module, name string, data interface{}) (out string, err error) {
	// Only modify on first pass
	if !s.s.isFirstPass {
		return "", nil
	}

	k := s.s.sharedData.key(module, name)
	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(data)).
		Debug("adding to module hook")

	v := reflect.ValueOf(data)
	if !v.IsValid() {
		err := fmt.Errorf("third parameter, data, must be set")
		return "", err
	}

	// we only allow slices or maps to allow multiple templates to
	// write to the same block
	if v.Kind() != reflect.Slice {
		err := fmt.Errorf("unsupported module block data type %q, supported type is slice", v.Kind())
		return "", err
	}

	// convert the slice into a []any
	interfaceSlice := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		interfaceSlice[i] = v.Index(i).Interface()
	}

	// if set, append, otherwise assign
	if _, ok := s.s.sharedData.moduleHooks[k]; ok {
		s.s.sharedData.moduleHooks[k].values = append(s.s.sharedData.moduleHooks[k].values, interfaceSlice...)
	} else {
		s.s.sharedData.moduleHooks[k] = &moduleHook{values: interfaceSlice}
	}

	return "", nil
}

// ReadFile reads a file from the current directory and returns it's contents
//
//	{{ stencil.ReadFile "myfile.txt" }}
func (s *TplStencil) ReadFile(name string) (string, error) {
	f, ok := s.exists(name)
	if !ok {
		return "", errors.Errorf("file %q does not exist", name)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Exists returns true if the file exists in the current directory
//
//	{{- if stencil.Exists "myfile.txt" }}
//	{{ stencil.ReadFile "myfile.txt" }}
//	{{- end }}
func (s *TplStencil) Exists(name string) bool {
	f, ok := s.exists(name)
	if ok {
		f.Close() // close the file handle, since we don't need it
	}
	return ok
}

// exists returns a billy.File if the file exists, and true. If it doesn't,
// nil is returned and false.
func (s *TplStencil) exists(name string) (billy.File, bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, false
	}

	f, err := osfs.New(cwd).Open(name)
	if err != nil {
		return nil, false
	}
	return f, true
}

// ApplyTemplate executes a named template (defined through the `define`
// function) with the provided optional data.
//
// The provided data can be accessed within the defined template under
// the `.Data` key.
//
// `.` is a copy of [Values] for the calling template, meaning it is not
// mutated to reflect that of the template being rendered.
//
// ## Examples
//
// ### Without Data
//
//	{{- define "command"}}
//	package main
//
//	import "fmt"
//
//	func main() {
//	  fmt.Println("hello, world!")
//	}
//
//	{{- end }}
//
//	{{- stencil.ApplyTemplate "command" | file.SetContents }}
//
// ### With Data
//
//	{{- define "command"}}
//	{{- $cliName := .Data }}
//	package main
//
//	import "fmt"
//
//	func main() {
//			fmt.Println("hello from {{ $cliName }}!")
//	}
//
//	{{- end }}
//
//	{{- range $cliName := stencil.Arg "clis" }}
//	{{- stencil.ApplyTemplate "command" $cliName | file.SetContents }}
//	{{- end }}
func (s *TplStencil) ApplyTemplate(name string, dataSli ...any) (string, error) {
	// We check for dataSli here because we had to set it to a range of arguments
	// to allow it to be not set.
	if len(dataSli) > 1 {
		return "", fmt.Errorf("ApplyTemplate() only takes max two arguments, name and data")
	}

	// Create a copy of the current values so we can set the data on it
	// without mutating the original values.
	d := s.t.args.Copy()
	if len(dataSli) > 0 {
		d.Data = dataSli[0]
	}

	var buf bytes.Buffer
	if err := s.t.Module.GetTemplate().ExecuteTemplate(&buf, name, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ReadBlocks parses a file and attempts to read the blocks from it, and their data.
//
// As a special case, if the file does not exist, an empty map is returned instead of an error.
//
// **NOTE**: This function does not guarantee that blocks are able to be read during runtime.
// for example, if you try to read the blocks of a file from another module there is no guarantee
// that that file will exist before you run this function. Nor is there the ability to tell stencil
// to do that (stencil does not have any order guarantees). Keep that in mind when using this function.
//
//	{{- $blocks := stencil.ReadBlocks "myfile.txt" }}
//	{{- range $name, $data := $blocks }}
//	  {{- $name }}
//	  {{- $data }}
//	{{- end }}
func (s *TplStencil) ReadBlocks(fpath string) (map[string]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// ensure that the file is within the current directory
	// and not attempting to escape it
	if _, err := osfs.New(cwd).Stat(fpath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}

		return nil, err
	}

	data, err := parseBlocks(fpath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Debug logs the provided arguments under the DEBUG log level (must run stencil with --debug).
//
//	{{- $_ := stencil.Debug "I'm a log!" }}
func (s *TplStencil) Debug(args ...interface{}) error {
	s.log.With("path", s.t.Path).Debugf("%s", args...)

	// We have to return something...
	return nil
}
