// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains the public API for templates
// for stencil

package codegen

import (
	"testing"

	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
)

// TestTplFile_OnceNoLockfile tests the file.Once command when there's no lockfile history at all
func TestTplFile_OnceNoLockfile(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
	}

	fo, err := tplf.Once()
	if err != nil {
		t.Errorf("should not error")
	}
	if fo != "" {
		t.Errorf("should return empty string")
	}
	if tplf.f.Skipped {
		t.Errorf("should not be skipped")
	}
}

// TestTplFile_OnceLockNoHistory tests the file.Once command when there's no history of the file existing
func TestTplFile_OnceLockNoHistory(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{},
		},
	}

	fo, err := tplf.Once()
	if err != nil {
		t.Errorf("should not error")
	}
	if fo != "" {
		t.Errorf("should return empty string")
	}
	if tplf.f.Skipped {
		t.Errorf("should not be skipped")
	}
}

// TestTplFile_OnceLockHasHistoryOfOtherFile tests the file.Once command when there's history of another file existing but not this one
func TestTplFile_OnceLockHasHistoryOfOtherFile(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "foo.go"},
			},
		},
	}

	fo, err := tplf.Once()
	if err != nil {
		t.Errorf("should not error")
	}
	if fo != "" {
		t.Errorf("should return empty string")
	}
	if tplf.f.Skipped {
		t.Errorf("should not be skipped")
	}
}

// TestTplFile_OnceLockHasHistoryOfFile tests the file.Once command when there's history of the file existing
func TestTplFile_OnceLockHasHistoryOfFile(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: "test.go"},
		t:   &Template{},
		log: slogext.NewTestLogger(t),
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "test.go"},
			},
		},
	}

	fo, err := tplf.Once()
	if err != nil {
		t.Errorf("should not error")
	}
	if fo != "" {
		t.Errorf("should return empty string")
	}
	if !tplf.f.Skipped {
		t.Errorf("should be skipped")
	}
}
