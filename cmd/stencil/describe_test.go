// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains code for the describe command

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"go.rgst.io/stencil/internal/modules/resolver"
	"go.rgst.io/stencil/pkg/stencil"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

func Test_cleanPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should support relative path",
			args: args{
				path: "foo/bar",
			},
			want:    "foo/bar",
			wantErr: false,
		},
		{
			name: "should support relative path with .",
			args: args{
				path: "./foo/bar",
			},
			want:    "foo/bar",
			wantErr: false,
		},
		{
			name: "should support relative path with ..",
			args: args{
				path: "../stencil/foo/bar",
			},
			want:    "foo/bar",
			wantErr: false,
		},
		{
			name: "should support absolute path",
			args: args{
				path: filepath.Join(cwd, "foo", "bar"),
			},
			want:    "foo/bar",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanPath(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleanPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_describeFile_shouldFunction ensures that describeFile can read a
// file out of a lockfile, when present.
func Test_describeFile_shouldFunction(t *testing.T) {
	tmpDir := t.TempDir()

	chdir(t, tmpDir)

	lock := &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "test-module",
			URL:  "vfs://",
			Version: &resolver.Version{
				Commit: "xyz",
			},
		}},
		Files: []*stencil.LockfileFileEntry{{
			Name:     "hello-world",
			Template: "hello-world.tpl",
			Module:   "test-module",
		}},
	}

	// write the lockfile
	b, err := yaml.Marshal(lock)
	assert.NilError(t, err)
	assert.NilError(t, os.WriteFile(stencil.LockfileName, b, 0o600))
	assert.NilError(t, os.WriteFile("hello-world", []byte{}, 0o644))
	out := &bytes.Buffer{}

	assert.NilError(t, describeFile("hello-world", out))
	assert.Equal(t,
		out.String(),
		"hello-world was created by module https://test-module (template: hello-world.tpl)\n",
	)
}
