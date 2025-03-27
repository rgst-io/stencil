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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// _ ensures that we implement the os.FileInfo interface
var _ os.FileInfo = &File{}

// File is a file that was created by a rendered template
type File struct {
	// blocks is the contents of all known blocks. Blocks are arbitrary
	// comments that encompass data that is persisted across runs of stencil. These
	// are then exposed via the Block method to be re-injected at template runtime.
	// This enables users to persist their changes in certain areas.
	blocks map[string]*blockInfo

	// contents is the contents of the file rendered for this template
	contents []byte

	// path is the full path to the file
	path string

	// mode is the mode of the file
	mode os.FileMode

	// modTime is the modification time of this file (when it was last modified)
	// Note: This is not always reliable currently.
	modTime time.Time

	// sourceTemplate is the template that is currently acting on this file
	sourceTemplate *Template

	// Below are public fields that are useful for determining
	// how to process this file.

	// Deleted denotes this file as being deleted, if this is
	// true then f.contents should not be used.
	Deleted bool

	// Skipped denotes this file as being skipped, if this is
	// true then f.contents should not be used.
	Skipped bool

	// SkippedReason is the reason why this file was skipped
	SkippedReason string

	// Warnings is an array of warnings that were created
	// while rendering this template
	Warnings []string
}

// NewFile creates a new file, an existing file at the given path is
// parsed to read blocks from, if it exists. An error is returned if
// the file is unable to be read for a reason other than not existing.
func NewFile(path string, mode os.FileMode, modTime time.Time, sourceTemplate *Template) (*File, error) {
	blocks, err := parseBlocks(path, sourceTemplate)
	if err != nil {
		return nil, err
	}

	return &File{path: path, mode: mode, modTime: modTime, blocks: blocks, sourceTemplate: sourceTemplate}, nil
}

// Block returns the contents of a given block.
func (f *File) Block(name string) string {
	bi, ok := f.blocks[name]
	if !ok {
		return ""
	}

	// If a deprecated block is loaded, warn
	if bi.Version == BlockVersion1 {
		f.sourceTemplate.log.Warnf(
			"Deprecated V1 block (%s) found at %s:%d-%d",
			name,
			f.path,
			bi.StartLine, bi.EndLine,
		)
	}

	return bi.Contents
}

// AddDeprecationNotice adds a deprecation notice to a file
func (f *File) AddDeprecationNotice(msg string) {
	if f.Warnings == nil {
		f.Warnings = []string{msg}
		return
	}

	f.Warnings = append(f.Warnings, msg)
}

// SetPath updates the path of this file. This causes
// the blocks to be parsed again.
func (f *File) SetPath(path string) error {
	blocks, err := parseBlocks(path, f.sourceTemplate)
	if err != nil {
		return err
	}
	f.blocks = blocks
	f.path = path

	return nil
}

// SetMode updates the mode of the file
func (f *File) SetMode(mode os.FileMode) {
	f.mode = mode
}

// SetContents updates the contents of the current file
func (f *File) SetContents(contents string) {
	f.contents = []byte(contents)
}

// Bytes returns the contents of this file as bytes
func (f *File) Bytes() []byte {
	return f.contents
}

// String returns the contents of this file as a string
func (f *File) String() string {
	return string(f.Bytes())
}

// The below functions implement the os.FileInfo
// interface

// Name returns the name of the file
func (f *File) Name() string {
	return f.path
}

// IsDir returns if this is file is a directory or not
// Note: We only support rendering files currently
func (f *File) IsDir() bool {
	return false
}

// ModTime returns the last modification time for the file
func (f *File) ModTime() time.Time {
	return f.modTime
}

// Mode returns the file mode
func (f *File) Mode() os.FileMode {
	return f.mode
}

// Size returns the size of the file
func (f *File) Size() int64 {
	return int64(len(f.contents))
}

// Sys implements the os.FileInfo.Sys method. This does
// not do anything.
func (f *File) Sys() interface{} {
	return nil
}

// Write writes a [codegen.File] to disk based on its current state, logging appropriately
func (f *File) Write(log slogext.Logger, dryRun bool) error {
	action := "Created"
	if f.Deleted {
		action = "Deleted"

		if !dryRun {
			if err := os.Remove(f.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	} else if f.Skipped {
		action = "Skipped"
	} else if _, err := os.Stat(f.Name()); err == nil {
		action = "Updated"
	}

	if action == "Created" || action == "Updated" {
		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(f.Name()), 0o750); err != nil {
				return fmt.Errorf("failed to create directory %q: %w", filepath.Dir(f.Name()), err)
			}

			//nolint:gosec // Why: By design.
			if err := os.WriteFile(f.Name(), f.Bytes(), f.Mode()); err != nil {
				return fmt.Errorf("failed to write file %q: %w", f.Name(), err)
			}
		}
	}

	msg := fmt.Sprintf("  -> %s %s", action, f.Name())
	if dryRun {
		msg += " (dry-run)"
	}

	if !f.Skipped {
		log.Info(msg)
	} else {
		// For skipped files, we only log at debug level
		log.Debug(msg, "reason", f.SkippedReason)
	}
	return nil
}
