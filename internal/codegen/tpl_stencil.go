// Copyright (C) 2024-2025 stencil contributors
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

	"github.com/davecgh/go-spew/spew"
	"github.com/go-git/go-billy/v5"
	"github.com/puzpuzpuz/xsync/v4"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
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
	k := s.s.sharedState.key(s.t.Module.Name, name)
	v, _ := s.s.sharedState.ModuleHooks.Load(k)
	if v == nil {
		// No data, return nothing
		return []any{}
	}

	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(v)).
		Debug("getting module hook")

	return v
}

// SetGlobal sets a global to be used in the context of the current
// template module repository. This is useful because sometimes you want
// to define variables inside of a helpers template file after doing
// manifest argument processing and then use them within one or more
// template files to be rendered; however, go templates limit the scope
// of symbols to the current template they are defined in, so this is
// not possible without external tooling like this function.
//
// This template function stores (and its inverse, GetGlobal, retrieves)
// data that is not strongly typed, so use this at your own risk and be
// averse to panics that could  occur if you're using the data it
// returns in the wrong way.
//
//	{{- /* This writes a global into the current context of the template module repository */}}
//	{{- stencil.SetGlobal "IsGeorgeCool" true -}}
func (s *TplStencil) SetGlobal(name string, data any) string {
	k := s.s.sharedState.key(s.t.Module.Name, name)
	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(data)).
		Debug("adding to global store")

	s.s.sharedState.Globals.Store(k, global{
		Template: s.t.Path,
		Value:    data,
	})

	return ""
}

// GetGlobal retrieves a global variable set by SetGlobal. The data
// returned from this function is unstructured so by averse to panics -
// look at where it was set to ensure you're dealing with the proper
// type of data that you think it is.
//
//	{{- /* This retrieves a global from the current context of the template module repository */}}
//	{{ $isGeorgeCool := stencil.GetGlobal "IsGeorgeCool" }}
func (s *TplStencil) GetGlobal(name string) any {
	k := s.s.sharedState.key(s.t.Module.Name, name)
	v, ok := s.s.sharedState.Globals.Load(k)
	if !ok {
		if s.s.renderStage == renderStageFinal {
			s.log.With("template", s.t.ImportPath(), "path", k).
				Warn("failed to retrieve data from global store")
		} else {
			s.log.With("template", s.t.ImportPath(), "path", k).
				Debug("failed to retrieve data from global store on pre-final stage")
		}
		return nil
	}

	s.log.With(
		"template", s.t.ImportPath(),
		"path", k,
		"data", spew.Sdump(v),
		"definingTemplate", v.Template,
	).Debug("retrieved data from global store")
	return v.Value
}

// AddToModuleHook adds to a hook in another module
//
// This functions write to module hook owned by another module for
// it to operate on. These are not strongly typed so it's best practice
// to look at how the owning module uses it for now.
//
//	{{- /* This writes to a module hook */}}
//	{{- stencil.AddToModuleHook "github.com/myorg/repo" "myModuleHook" "myData" }}
func (s *TplStencil) AddToModuleHook(module, name string, data ...any) (out string, err error) {
	// Attempt to read the destined module's manifest for extra features.
	var mcfg *configuration.TemplateRepositoryManifest
	for _, m := range s.s.modules {
		if m.Name == module {
			mcfg = m.Manifest
			break
		}
	}
	if mcfg != nil {
		// Check if we have a schema. If we do, use it to validate our module
		// hook data.
		mhcfg, ok := mcfg.ModuleHooks[name]
		if ok && mhcfg.Schema != nil {
			for _, d := range data {
				if err := validateJSONSchema(module+"/moduleHooks/"+name, mhcfg.Schema, d); err != nil {
					return "", err
				}
			}
		}
	}

	k := s.s.sharedState.key(module, name)
	s.log.With("template", s.t.ImportPath(), "path", k, "data", spew.Sdump(data)).
		Debug("adding to module hook")

	s.s.sharedState.ModuleHooks.Compute(k, func(old moduleHook, _ bool) (moduleHook, xsync.ComputeOp) {
		return append(old, data...), xsync.UpdateOp
	})

	return "", nil
}

// ReadFile reads a file from the current directory and returns it's
// contents
//
//	{{ stencil.ReadFile "myfile.txt" }}
func (s *TplStencil) ReadFile(name string) (string, error) {
	f, err := s.existsAndMaybeOpen(name, true)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ReadDirEntry is a partial of [os.DirEntry] returned by
// [TplStencil.ReadDir]
type ReadDirEntry interface {
	Name() string
	IsDir() bool
}

// ReadDir reads the contents of a directory and returns a list of
// files/directories
//
//	{{ range $entry := stencil.ReadDir "/tests" }}
//	  {{ if $entry.IsDir }}
//	    {{ $entry.Name }}
//	  {{ end }}
//	{{ end }}
func (s *TplStencil) ReadDir(name string) ([]ReadDirEntry, error) {
	_, err := s.existsAndMaybeOpen(name, false)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}

	rv := make([]ReadDirEntry, 0, len(entries))
	for _, entry := range entries {
		rv = append(rv, entry)
	}

	return rv, nil
}

// Exists returns true if the file exists in the current directory
//
//	{{- if stencil.Exists "myfile.txt" }}
//	{{ stencil.ReadFile "myfile.txt" }}
//	{{- end }}
func (s *TplStencil) Exists(name string) bool {
	_, err := s.existsAndMaybeOpen(name, false)
	return err == nil
}

// existsAndMaybeOpen returns a billy.File if the file existsAndMaybeOpen, and nil if it doesn't,
// also returning an error if one is applicable.
func (s *TplStencil) existsAndMaybeOpen(name string, andOpen bool) (billy.File, error) {
	fs, err := s.t.Module.GetFS(s.t.args.Context)
	if err != nil {
		return nil, err
	}

	if _, err := fs.Stat(name); err != nil {
		return nil, err
	}

	if !andOpen {
		return nil, nil
	}

	f, err := fs.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Include executes a named template (defined through the `define`
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
//	{{- stencil.Include "command" | file.SetContents }}
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
//	{{- stencil.Include "command" $cliName | file.SetContents }}
//	{{- end }}
func (s *TplStencil) Include(name string, dataSli ...any) (string, error) {
	// We check for dataSli here because we had to set it to a range of arguments
	// to allow it to be not set.
	if len(dataSli) > 1 {
		return "", fmt.Errorf("Include() only takes max two arguments, name and data")
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

// ApplyTemplate is an alias to stencil.Include.
//
// Deprecated: Use stencil.Include instead.
func (s *TplStencil) ApplyTemplate(name string, dataSli ...any) (string, error) {
	s.log.With(
		"module", s.t.Module.Name,
		"template", s.t.Path,
	).Warn("stencil.ApplyTemplate is deprecated, use stencil.Include instead")
	return s.Include(name, dataSli...)
}

// ReadBlocks parses a file and attempts to read the blocks from it, and
// their data.
//
// As a special case, if the file does not exist, an empty map is
// returned instead of an error.
//
// **NOTE**: This function does not guarantee that blocks are able to be
// read during runtime. For example, if you try to read the blocks of a
// file from another module there is no guarantee that that file will
// exist before you run this function. Nor is there the ability to tell
// stencil to do that (stencil does not have any order guarantees).
// Keep that in mind when using this function.
//
//	{{- $blocks := stencil.ReadBlocks "myfile.txt" }}
//	{{- range $name, $data := $blocks }}
//	  {{- $name }}
//	  {{- $data }}
//	{{- end }}
func (s *TplStencil) ReadBlocks(fpath string) (map[string]string, error) {
	_, err := s.existsAndMaybeOpen(fpath, false)
	if err != nil {
		return nil, err
	}

	data, err := parseBlocks(fpath, s.t)
	if err != nil {
		return nil, err
	}

	rv := make(map[string]string)
	for k, v := range data {
		rv[k] = v.Contents
	}
	return rv, nil
}

// Debug logs the provided arguments under the DEBUG log level (must run
// stencil with --debug).
//
//	{{- stencil.Debug "I'm a log!" }}
func (s *TplStencil) Debug(args ...any) string {
	s.log.With("path", s.t.Path).Debugf("%s", args...)

	return ""
}
