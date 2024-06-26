// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file implements the file struct passed to
// templates in Stencil.

package codegen

import (
	"os"
	"time"

	"go.rgst.io/stencil/pkg/slogext"
)

// TplFile is the current file we're writing output to in a
// template. This can be changed via file.SetPath and written
// to by file.Install. When a template does not call file.SetPath
// a default file is created that matches the current template path
// with the extension '.tpl' removed from the path and operated on.
type TplFile struct {
	// f is the current file we're writing to
	f *File

	// t is the current template
	t *Template

	// log is the logger to use for debugging
	log slogext.Logger
}

// Block returns the contents of a given block
//
//	###Block(name)
//	Hello, world!
//	###EndBlock(name)
//
//	###Block(name)
//	{{- /* Only output if the block is set */}}
//	{{- if not (empty (file.Block "name")) }}
//	{{ file.Block "name" }}
//	{{- end }}
//	###EndBlock(name)
//
//	###Block(name)
//	{{ - /* Short hand syntax, but adds newline if no contents */}}
//	{{ file.Block "name" }}
//	###EndBlock(name)
func (f *TplFile) Block(name string) string {
	return f.f.Block(name)
}

// SetPath changes the path of the current file being rendered
//
//	{{- file.SetPath "new/path/to/file.txt" }}
func (f *TplFile) SetPath(path string) (out string, err error) {
	path = f.t.Module.ApplyDirReplacements(path)
	err = f.f.SetPath(path)
	return "", err
}

// SetContents sets the contents of file being rendered to the value
//
// This is useful for programmatic file generation within a template.
//
//	{{ file.SetContents "Hello, world!" }}
func (f *TplFile) SetContents(contents string) error {
	f.f.SetContents(contents)
	return nil
}

// Skip skips the current file being rendered
//
//	{{- file.Skip "A reason to skip this file" }}
func (f *TplFile) Skip(reason string) (output string, err error) {
	f.f.Skipped = true
	f.f.SkippedReason = reason
	return "", nil
}

// Delete deletes the current file being rendered
//
//	{{ file.Delete }}
func (f *TplFile) Delete() error {
	f.f.Deleted = true
	return nil
}

// Static marks the current file as static
//
// Marking a file is equivalent to calling file.Skip, but instead
// file.Skip is only called if the file already exists. This is useful
// for files you want to generate but only once. It's generally
// recommended that you do not do this as it limits your ability to change
// the file in the future.
//
//	{{ file.Static }}
func (f *TplFile) Static() (out string, err error) {
	// if the file already exists, skip it
	if _, err := os.Stat(f.f.path); err == nil {
		f.log.With("template", f.t.Path, "path", f.f.path).
			Debug("Skipping static file because it already exists")
		return f.Skip("Static file, output already exists")
	}

	return "", nil
}

// Path returns the current path of the file we're writing to
//
//	{{ file.Path }}
func (f *TplFile) Path() string {
	return f.f.path
}

// Create creates a new file that is rendered by the current template
//
// If the template has a single file with no contents
// this file replaces it.
//
//	{{- define "command" }}
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
//	# Generate a "<commandName>.go" file for each command in .arguments.commands
//	{{- range $_, $commandName := (stencil.Arg "commands") }}
//	{{- file.Create (printf "cmd/%s.go" $commandName) 0600 now }}
//	{{- stencil.ApplyTemplate "command" | file.SetContents }}
//	{{- end }}
func (f *TplFile) Create(path string, mode os.FileMode, modTime time.Time) (out, err error) {
	f.f, err = NewFile(path, mode, modTime)
	if err != nil {
		return err, err
	}

	f.t.Files = append(f.t.Files, f.f)
	return nil, nil
}

// RemoveAll deletes all the contents in the provided path
//
//	{{ file.RemoveAll "path" }}
func (f *TplFile) RemoveAll(path string) (out, err error) {
	if err := os.RemoveAll(path); err != nil {
		return err, err
	}
	return nil, nil
}
