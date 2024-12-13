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

// Description: This file contains the logic and type for a template
// that is being rendered by stencil.
package codegen

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/pkg/slogext"
)

// Template is a file that has been processed by stencil
type Template struct {
	// parsed denotes if this template has been parsed or not
	parsed bool

	// args are the arguments passed to the template
	args *Values

	// log is the logger to use for debug logging
	log slogext.Logger

	// mode is the os file mode of the template, this is used
	// for the default file if not modified during render time
	mode os.FileMode

	// modTime is the modification time of the template, this is used
	// for the default file if not modified during render time
	modTime time.Time

	// adoptMode denotes if we should use heuristics to detect code that should go
	// into blocks to assist with first-time adoption of templates
	adoptMode bool

	// Module is the underlying module that's creating this template
	Module *modules.Module

	// Path is the path of this template relative to the owning module
	Path string

	// Files is a list of files that this template generated
	Files []*File

	// Contents is the content of this template
	Contents []byte

	// Binary denotes if a template is a binary "template" (not actually a
	// template) or not.  Binary "templates" are just copied verbatim to the
	// target path, used for embedding binary files into template repos for
	// uses like gradle-wrapper.jar.
	Binary bool

	// Library denotes if a template is a library template or not. Library
	// templates cannot generate files.
	Library bool
}

type NewTemplateOpts struct {
	// Enable the adoptMode option for the Template file (see [codegen.Template.adoptMode])
	Adopt bool

	// Enable the binary option for the Template file (see [codegen.Template.Binary])
	Binary bool
}

// NewTemplate creates a new Template with the current file being the same name
// with the extension .tpl being removed. If the provided template has
// the extension .library.tpl, then the Library field is set to true.
func NewTemplate(m *modules.Module, fpath string, mode os.FileMode,
	modTime time.Time, contents []byte, log slogext.Logger, opts *NewTemplateOpts) (*Template, error) {
	var library bool
	if filepath.Ext(strings.TrimSuffix(fpath, ".tpl")) == ".library" {
		library = true
	}

	t := &Template{
		log:      log,
		mode:     mode,
		modTime:  modTime,
		Module:   m,
		Path:     fpath,
		Contents: contents,
		Library:  library,
	}

	if opts != nil {
		t.adoptMode = opts.Adopt
		t.Binary = opts.Binary
	}

	return t, nil
}

// ImportPath returns the path to this template, this is meant to denote
// which module this template is attached to
func (t *Template) ImportPath() string {
	return path.Join(t.Module.Name, t.Path)
}

// Parse parses the provided template and makes it available to be Rendered
// in the context of the current module.
func (t *Template) Parse(_ *Stencil) error {
	if !t.Binary {
		// Add the current template to the template object on the module that we're
		// attached to. This enables us to call functions in other templates within our
		// 'module context'.
		if _, err := t.Module.GetTemplate().New(t.ImportPath()).Funcs(NewFuncMap(nil, nil, t.log)).
			Parse(string(t.Contents)); err != nil {
			return err
		}
	}

	t.parsed = true

	return nil
}

// Render renders the provided template, the produced files
// are rendered onto the Files field of the template struct.
func (t *Template) Render(st *Stencil, vals *Values) error {
	if len(t.Files) == 0 && !t.Library {
		var p string
		if t.Binary {
			p = strings.TrimSuffix(t.Path, ".nontpl")
		} else {
			p = strings.TrimSuffix(t.Path, ".tpl")
		}
		p = t.Module.ApplyDirReplacements(p)
		f, err := NewFile(p, t.mode, t.modTime, t)
		if err != nil {
			return err
		}
		t.Files = []*File{f}
	}

	// Parse the template if we haven't already
	if !t.parsed {
		if err := t.Parse(st); err != nil {
			return err
		}
	}

	if t.Binary {
		t.Files[0].SetContents(string(t.Contents))
		return nil
	}

	// Update the module values
	t.args = vals.WithModule(t.Module.Name, t.Module.Version).WithTemplate(t.Path)

	// Execute a specific file because we're using a shared template, if we attempt to render
	// the entire template we'll end up just rendering the base template (<module>) which is empty
	var buf bytes.Buffer
	if err := t.Module.GetTemplate().Funcs(NewFuncMap(st, t, t.log)).
		ExecuteTemplate(&buf, t.ImportPath(), t.args); err != nil {
		return err
	}

	// If we're a library template, we don't want to generate any files so
	// we can return early here.
	if t.Library {
		t.Files = nil
		return nil
	}

	// If we're writing only a single file, and the contents is empty
	// then we should write the output of the rendered template.
	//
	// This ensures that templates don't need to call file.Create
	// by default, only when they want to customize the output
	if len(t.Files) == 1 && len(t.Files[0].Bytes()) == 0 {
		t.Files[0].SetContents(buf.String())
	} else if len(t.Files) > 1 {
		// otherwise, remove the first file that was created when
		// we constructed the template. It's only used when we have
		// no calls to file.Create
		t.Files = t.Files[1:len(t.Files)]
	}

	return nil
}
