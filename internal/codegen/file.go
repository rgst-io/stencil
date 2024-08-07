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
	"os"
	"time"
)

// _ ensures that we implement the os.FileInfo interface
var _ os.FileInfo = &File{}

// File is a file that was created by a rendered template
type File struct {
	// blocks is the contents of all known blocks. Blocks are arbitrary
	// comments that encompass data that is persisted across runs of stencil. These
	// are then exposed via the Block method to be re-injected at template runtime.
	// This enables users to persist their changes in certain areas.
	blocks map[string]string

	// contents is the contents of the file rendered for this template
	contents []byte

	// path is the full path to the file
	path string

	// mode is the mode of the file
	mode os.FileMode

	// modTime is the modification time of this file (when it was last modified)
	// Note: This is not always reliable currently.
	modTime time.Time

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
func NewFile(path string, mode os.FileMode, modTime time.Time) (*File, error) {
	blocks, err := parseBlocks(path)
	if err != nil {
		return nil, err
	}

	return &File{path: path, mode: mode, modTime: modTime, blocks: blocks}, nil
}

// Block returns the contents of a given block.
func (f *File) Block(name string) string {
	return f.blocks[name]
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
	blocks, err := parseBlocks(path)
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
