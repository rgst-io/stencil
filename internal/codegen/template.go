// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains the logic and type for a template
// that is being rendered by stencil.
package codegen

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/slogext"
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

	// Module is the underlying module that's creating this template
	Module *modules.Module

	// Path is the path of this template relative to the owning module
	Path string

	// Files is a list of files that this template generated
	Files []*File

	// Contents is the content of this template
	Contents []byte

	// Library denotes if a template is a library template or not. Library
	// templates cannot generate files.
	Library bool
}

// NewTemplate creates a new Template with the current file being the same name
// with the extension .tpl being removed. If the provided template has
// the extension .library.tpl, then the Library field is set to true.
func NewTemplate(m *modules.Module, fpath string, mode os.FileMode,
	modTime time.Time, contents []byte, log slogext.Logger) (*Template, error) {
	var library bool
	if filepath.Ext(strings.TrimSuffix(fpath, ".tpl")) == ".library" {
		library = true
	}

	return &Template{
		log:      log,
		mode:     mode,
		modTime:  modTime,
		Module:   m,
		Path:     fpath,
		Contents: contents,
		Library:  library,
	}, nil
}

// ImportPath returns the path to this template, this is meant to denote
// which module this template is attached to
func (t *Template) ImportPath() string {
	return path.Join(t.Module.Name, t.Path)
}

// Parse parses the provided template and makes it available to be Rendered
// in the context of the current module.
func (t *Template) Parse(_ *Stencil) error {
	// Add the current template to the template object on the module that we're
	// attached to. This enables us to call functions in other templates within our
	// 'module context'.
	if _, err := t.Module.GetTemplate().New(t.ImportPath()).Funcs(NewFuncMap(nil, nil, t.log)).
		Parse(string(t.Contents)); err != nil {
		return err
	}

	t.parsed = true

	return nil
}

// Render renders the provided template, the produced files
// are rendered onto the Files field of the template struct.
func (t *Template) Render(st *Stencil, vals *Values) error {
	if len(t.Files) == 0 && !t.Library {
		f, err := NewFile(strings.TrimSuffix(t.Path, ".tpl"), t.mode, t.modTime)
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

	if !st.isFirstPass {
		// Now that everything's been decided, see if we need to replace any file paths from directory manifests
		for _, tf := range t.Files {
			if err := t.applyDirReplacements(tf, st, vals); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Template) applyDirReplacements(tf *File, st *Stencil, vals *Values) error {
	// Hop through the path dir by dir, starting at the end (because the raw paths won't match if you replace the earlier
	// path segments first), and see if there's any replacements.
	pp := strings.Split(tf.path, string(os.PathSeparator))
	for i := len(pp) - 1; i >= 0; i-- {
		pathPart := strings.Join(pp[0:i+1], string(os.PathSeparator))
		if dr, has := t.Module.Manifest.DirReplacements[pathPart]; has {
			// Render replacement
			rt, err := NewTemplate(t.Module, "dirReplace", 0o000, time.Time{}, []byte(dr), t.log)
			if err != nil {
				return err
			}

			if err := rt.Render(st, vals); err != nil {
				return err
			}

			nn := rt.Files[0].String()
			if strings.Contains(nn, string(os.PathSeparator)) {
				return fmt.Errorf("directory replacement of %s to %s contains path separator in output", pp[i], nn)
			}
			pp[i] = nn
		}
	}
	tf.path = strings.Join(pp, string(os.PathSeparator))

	return nil
}
