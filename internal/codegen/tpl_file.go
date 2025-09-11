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

// Description: This file implements the file struct passed to
// templates in Stencil.

package codegen

import (
	"os"
	"slices"
	"time"

	"github.com/go-git/go-billy/v5/osfs"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"go.rgst.io/stencil/v2/pkg/stencil"
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

	// lock is the original lockfile passed in
	lock *stencil.Lockfile

	// log is the logger to use for debugging
	log slogext.Logger
}

// Block returns the contents of a given block
//
//	## <<Stencil::Block(name)>>
//	Hello, world!
//	## <</Stencil::Block>>
//
//	## <<Stencil::Block(name)>>
//	{{- /* Only output if the block is set */}}
//	{{- if not (empty (file.Block "name")) }}
//	{{ file.Block "name" }}
//	{{- end }}
//	## <</Stencil::Block>>
//
//	## <<Stencil::Block(name)>>
//	{{- /* Short hand syntax. Adds newline if no contents */}}
//	{{ file.Block "name" }}
//	## <</Stencil::Block>>
func (f *TplFile) Block(name string) string {
	return f.f.Block(name)
}

// BlockI returns the contents of a given block with an indentation nuance that if
// the block is empty, it will not add a newline to the output.
//
//	## <<Stencil::Block(name)>>
//	## <</Stencil::Block>>
//
//	## <<Stencil::Block(name)>>
//	{{- /* Has no newline if no contents, due to the - at the block prefix */ }}
//	{{- file.BlockI "name" }}
//	## <</Stencil::Block>>
func (f *TplFile) BlockI(name string) string {
	contents := f.f.Block(name)
	if contents == "" {
		return ""
	}

	return "\n" + contents
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
//	{{- file.SetContents "Hello, world!" }}
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
//	{{- file.Delete }}
func (f *TplFile) Delete() (out string, err error) {
	if f.lock != nil {
		f.lock.Files = slices.DeleteFunc(f.lock.Files, func(ff *stencil.LockfileFileEntry) bool { return ff.Name == f.f.path })
	}
	f.f.Deleted = true
	return "", nil
}

// Static marks the current file as static
//
// Marking a file is equivalent to calling file.Skip, but instead
// file.Skip is only called if the file already exists. This is useful
// for files you want to generate but only once. It's generally
// recommended that you do not do this as it limits your ability to change
// the file in the future.
//
//	{{- file.Static }}
func (f *TplFile) Static() (out string, err error) {
	// if the file already exists, skip it
	if _, err := osfs.Default.Stat(f.f.path); err == nil {
		f.log.With("template", f.t.Path, "path", f.f.path).
			Debug("Skipping static file because it already exists")
		return f.Skip("Static file, output already exists")
	}

	return "", nil
}

// Once will only generate this file a single time, if it doesn't already
// exist, and store that fact in the stencil.lock file.
//
// The first time a Once file is generated, it has its provenance stored
// in the stencil.lock file.  Going forward, Once checks the lock file
// for history of the file, and if it finds it, it performs the same
// action as file.Skip.
//
//	{{- file.Once }}
func (f *TplFile) Once() (out string, err error) {
	// if the file exists at all, skip it
	if _, err := osfs.Default.Stat(f.f.path); err == nil {
		f.log.With("template", f.t.Path, "path", f.f.path).
			Debug("Skipping once file because it already exists on FS")
		return f.Skip("Once file, output already exists")
	}

	// if the file already exists in the lockfile, skip it
	if f.lock != nil && slices.ContainsFunc(f.lock.Files, func(ff *stencil.LockfileFileEntry) bool { return ff.Name == f.f.path }) {
		f.log.With("template", f.t.Path, "path", f.f.path).
			Debug("Skipping once file because it already exists in the lockfile")
		return f.Skip("Once file, already in lockfile")
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
//	{{- /* Skip the file that generates other files */}
//	{{- file.Skip }}
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
//	{{- stencil.Include "command" | file.SetContents }}
//	{{- end }}
func (f *TplFile) Create(path string, mode os.FileMode, modTime time.Time) (out string, err error) {
	f.f, err = NewFile(path, mode, modTime, f.t)
	if err != nil {
		return "", err
	}

	f.t.Files = append(f.t.Files, f.f)
	return "", nil
}

// RemoveAll deletes all the contents in the provided path
//
//	{{- file.RemoveAll "path" }}
func (f *TplFile) RemoveAll(path string) (out string, err error) {
	if err := os.RemoveAll(path); err != nil {
		return "", err
	}
	return "", nil
}

// MigrateTo migrates the current file to a new path.  If the old file doesn't exist, it is
// treated as a `file.Skip`.  If the old file still exists, then it is moved to the new
// path (via a create and write, then delete of the old path, not a filesystem move).  If
// the MigrateTo target file already exists, it is overwritten.
//
//	{{- file.MigrateTo "new/path/to/file.txt" }}
func (f *TplFile) MigrateTo(path string) (out string, err error) {
	if _, err := osfs.Default.Stat(f.f.path); err != nil {
		f.log.With("template", f.t.Path, "path", f.f.path).
			Debug("Skipping MigrateTo because the file doesn't exist")
		return f.Skip("MigrateTo file input doesn't exist")
	}

	f.log.With("path", f.f.path).With("to", path).
		Debug("Migrating file to new path")
	contents, err := os.ReadFile(f.f.path)
	if err != nil {
		return "", err
	}

	fn, err := osfs.Default.Create(path)
	if err != nil {
		return "", err
	}
	defer fn.Close()

	if _, err := fn.Write(contents); err != nil {
		return "", err
	}

	f.log.With("path", f.f.path).
		Debug("Deleting original file after migration")
	f.f.Deleted = true

	return "", nil
}

// SetMode sets the provided mode on the current output file.
//
//	{{- file.SetMode 0o755 }}
func (f *TplFile) SetMode(mode os.FileMode) (string, error) {
	f.f.mode = mode
	return "", nil
}
