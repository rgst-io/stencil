package codegen

import (
	"context"
	"os"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
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
			wantErr: os.ErrNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)
			s := &TplStencil{log: log}
			got, err := s.ReadBlocks(tt.args.fpath)

			if (tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("TplStencil.ReadBlocks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (tt.wantErr == nil) && err != nil {
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
	moduleHookName := "test"
	tests := []struct {
		name     string
		inserts  [][]any
		manifest *configuration.TemplateRepositoryManifest
		want     []any
		wantErr  bool
	}{
		{
			name: "should be able to insert and read",
			inserts: [][]any{
				{"abc"},
				{"def"},
				{map[string]any{
					"abc": "def",
				}},
				{"abc"},
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
		{
			name: "should support valid data with schema",
			inserts: [][]any{
				{map[string]any{
					"hello": "world",
				}},
			},
			manifest: &configuration.TemplateRepositoryManifest{
				ModuleHooks: map[string]configuration.ModuleHook{
					moduleHookName: {
						Schema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"hello": map[string]any{
									"type": "string",
								},
							},
						},
					},
				},
			},
			want: []any{map[string]any{"hello": "world"}},
		},
		{
			name: "should fail invalid data with schema",
			inserts: [][]any{
				{map[string]string{
					"hello": "world",
				}},
			},
			manifest: &configuration.TemplateRepositoryManifest{
				ModuleHooks: map[string]configuration.ModuleHook{
					moduleHookName: {
						Schema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"hello": map[string]any{
									"type": "number",
								},
							},
						},
					},
				},
			},
			want:    []any{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		// TODO(jaredallard): We need a better test initializer instead
		// of... this nightmare.
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)

			mods := make([]*modules.Module, 0)
			if tt.manifest != nil {
				mods = append(mods, &modules.Module{Name: "test", Manifest: tt.manifest})
			}

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
						nil,
					),
				),
				s:   &Stencil{sharedState: newSharedState(), modules: mods},
				log: log,
			}

			for _, insert := range tt.inserts {
				_, err := s.AddToModuleHook(s.t.Module.Name, moduleHookName, insert...)
				if err != nil {
					if tt.wantErr {
						continue
					}

					t.Errorf("TplStencil.GetModuleHook() error = %v", err)
					return
				} else if tt.wantErr {
					t.Errorf("TplStencil.GetModuleHook() wanted error, got nil")
				}
			}

			s.s.sharedState.hash() // Sorts the module hooks

			if got := s.GetModuleHook(moduleHookName); !reflect.DeepEqual(got, tt.want) {
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
						nil,
					),
				),
				s:   &Stencil{sharedState: newSharedState()},
				log: log,
			}

			s.SetGlobal(tt.args.name, tt.args.data)

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
				nil,
			),
		),
		s:   &Stencil{sharedState: newSharedState()},
		log: log,
	}

	actualContents, err := os.ReadFile("testdata/blocks-test.txt")
	assert.NilError(t, err, "expected os.ReadFile to succeed")

	contents, err := s.ReadFile("testdata/blocks-test.txt")
	assert.NilError(t, err, "expected ReadFile to succeed when file exists")

	assert.Equal(t, string(actualContents), contents)

	// fails when file doesn't exist
	_, err = s.ReadFile("file/that/doesnt/exist")
	assert.Equal(t, true, errors.Is(err, os.ErrNotExist))
}

func TestTplStencil_Include(t *testing.T) {
	type args struct {
		name    string
		dataSli []interface{}
	}
	tests := []struct {
		name        string
		args        args
		subTemplate string
		want        string
		wantErr     bool
	}{
		{
			name: "should error on non-existent template",
			args: args{
				name: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "should render a template",
			args: args{
				name: "hello-world",
			},
			subTemplate: `{{- define "hello-world" -}}{{ "Hello, world!" }}{{- end -}}`,
			want:        "Hello, world!",
		},
		{
			name: "should support data",
			args: args{
				name:    "hello-world",
				dataSli: []any{"xyz"},
			},
			subTemplate: `{{- define "hello-world" -}}{{ .Data }}{{- end -}}`,
			want:        "xyz",
		},
		{
			name: "should pass through data from parent template",
			args: args{
				name:    "hello-world",
				dataSli: []any{"xyz"},
			},
			subTemplate: `{{- define "hello-world" -}}{{ .Config.Name }}{{- end -}}`,
			want:        "TestTplStencil_Include/should_pass_through_data_from_parent_template",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)

			// create stencil
			st := &Stencil{sharedState: newSharedState()}

			// create module
			module, err := modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
				Name: "test",
			})
			assert.NilError(t, err, "expected NewModuleFromTemplates to succeed")

			// create template for module
			tpl, err := NewTemplate(
				module,
				"not-a-real-template.tpl",
				0o755,
				time.Now(),
				[]byte(tt.subTemplate),
				log,
				nil,
			)
			assert.NilError(t, err, "expected NewTemplate to succeed")

			// render template to register it
			tpl.Render(st,
				NewValues(context.Background(), &configuration.Manifest{
					Name: t.Name(),
				}, []*modules.Module{module}),
			)

			s := &TplStencil{
				t:   tpl,
				s:   st,
				log: log,
			}
			got, err := s.Include(tt.args.name, tt.args.dataSli...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TplStencil.Include() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TplStencil.Include() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTplStencil_ReadDir(t *testing.T) {
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
				nil,
			),
		),
		s:   &Stencil{sharedState: newSharedState()},
		log: log,
	}

	entries, err := s.ReadDir("../dsfdsfsd")
	assert.Equal(t, true, errors.Is(err, billy.ErrCrossedBoundary))
	assert.Equal(t, 0, len(entries))

	entries, err = s.ReadDir("/root")
	assert.Equal(t, true, errors.Is(err, os.ErrNotExist))
	assert.Equal(t, 0, len(entries))

	entries, err = s.ReadDir("testdata")
	assert.NilError(t, err)
	assert.Equal(t, true, len(entries) > 5)
	assert.Equal(t, true, slices.ContainsFunc(entries, func(entry ReadDirEntry) bool {
		return entry.Name() == "blocks-test.txt" && !entry.IsDir()
	}))
	assert.Equal(t, true, slices.ContainsFunc(entries, func(entry ReadDirEntry) bool {
		return entry.Name() == "args" && entry.IsDir()
	}))
}
