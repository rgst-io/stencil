// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains the public API for templates
// for stencil

package codegen

import (
	"testing"

	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
)

func TestTplFile_OnceEz(t *testing.T) {
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

func TestTplFile_OnceLockHasHistory(t *testing.T) {
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
