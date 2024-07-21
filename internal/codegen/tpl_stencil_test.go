// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains the public API for templates
// for stencil

package codegen

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

func TestTplStencil_ReadBlocks(t *testing.T) {
	type fields struct {
	}
	type args struct {
		fpath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr error
	}{
		{
			name: "should read blocks from a file",
			args: args{
				fpath: "testdata/blocks-test.txt",
			},
			want: map[string]string{
				"helloWorld": "Hello, world!",
				"e2e":        "content",
			},
		},
		{
			name: "should error on out of chroot path",
			args: args{
				fpath: "../testdata/blocks-test.txt",
			},
			wantErr: billy.ErrCrossedBoundary,
		},
		{
			name: "should return no data on non-existent file",
			args: args{
				fpath: "testdata/does-not-exist.txt",
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)
			s := &TplStencil{log: log}
			got, err := s.ReadBlocks(tt.args.fpath)

			// String checking because errors.Is isn't working
			if (tt.wantErr != nil) && err.Error() != tt.wantErr.Error() {
				t.Errorf("TplStencil.ReadBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TplStencil.ReadBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestTplStencil_GetModuleHook(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		inserts []any
		args    args
		want    []any
	}{
		{
			inserts: []any{
				[]string{"abc"},
				[]string{"def"},
				[]any{map[string]any{
					"abc": "def",
				}},
				[]string{"abc"},
			},
			args: args{
				name: "name",
			},
			want: []any{
				// This is what the hashing resulted in
				map[string]any{
					"abc": "def",
				},
				"def",
				"abc",
				"abc",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)
			s := &TplStencil{
				t: must(
					NewTemplate(
						must(modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
							Name: "test",
						})),
						"not-a-real-template.tpl",
						0o755,
						time.Now(),
						[]byte(""),
						log,
					),
				),
				s: &Stencil{sharedData: &sharedData{
					moduleHooks: make(map[string]*moduleHook),
					globals:     make(map[string]global),
				}},
				log: log,
			}

			s.s.isFirstPass = true
			for _, insert := range tt.inserts {
				if _, err := s.AddToModuleHook(s.t.Module.Name, tt.args.name, insert); err != nil {
					t.Errorf("TplStencil.GetModuleHook() error = %v", err)
					return
				}
			}

			// Ensure that GetModuleHook never returns anything other than
			// `[]any` during the first pass.
			if got := s.GetModuleHook(tt.args.name); !reflect.DeepEqual(got, []any{}) {
				t.Errorf("TplStencil.GetModuleHook() = %v, want %v", got, []any{})
			}
			s.s.isFirstPass = false

			// Sort the module hooks, which should be called by stencil before
			// the second pass
			s.s.sortModuleHooks()

			if got := s.GetModuleHook(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TplStencil.GetModuleHook() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGlobals contains tests for ensuring that the Set/GetGlobal
// functions work as expected.
func TestGlobals(t *testing.T) {
	type args struct {
		name string
		data any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "can insert data",
			args: args{
				name: "hello-world",
				data: "abc",
			},
			want: "abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)
			s := &TplStencil{
				t: must(
					NewTemplate(
						must(modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
							Name: "test",
						})),
						"not-a-real-template.tpl",
						0o755,
						time.Now(),
						[]byte{},
						log,
					),
				),
				s:   &Stencil{sharedData: &sharedData{globals: make(map[string]global)}},
				log: log,
			}

			s.s.isFirstPass = true
			s.SetGlobal(tt.args.name, tt.args.data)

			// Ensure we return nothing during the first pass.
			assert.Equal(t, s.GetGlobal(tt.args.name), nil)
			s.s.isFirstPass = false

			// Ensure we return data after the first pass. SetGlobal should
			// still take effect during the first pass.
			assert.DeepEqual(t, s.GetGlobal(tt.args.name), tt.want)
		})
	}
}

func TestTplStencil_ReadFile(t *testing.T) {
	log := slogext.NewTestLogger(t)
	s := &TplStencil{
		t: must(
			NewTemplate(
				must(modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
					Name: "test",
				})),
				"not-a-real-template.tpl",
				0o755,
				time.Now(),
				[]byte{},
				log,
			),
		),
		s:   &Stencil{sharedData: &sharedData{globals: make(map[string]global)}},
		log: log,
	}

	actualContents, err := os.ReadFile("testdata/blocks-test.txt")
	assert.NilError(t, err, "expected os.ReadFile to succeed")

	contents, err := s.ReadFile("testdata/blocks-test.txt")
	assert.NilError(t, err, "expected ReadFile to succeed when file exists")

	assert.Equal(t, string(actualContents), contents)

	// fails when file doesn't exist
	_, err = s.ReadFile("file/that/doesnt/exist")
	assert.Error(t, err, `file "file/that/doesnt/exist" does not exist`)
}
