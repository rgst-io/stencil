package codegen

import (
	"fmt"
	"reflect"
	"testing"

	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/internal/yaml"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

type testTpl struct {
	s   *Stencil
	t   *Template
	log slogext.Logger
}

// fakeTemplate returns a faked struct suitable for testing
// template functions
func fakeTemplate(t *testing.T, args map[string]any,
	requestArgs map[string]configuration.Argument) *testTpl {
	test := &testTpl{}
	log := slogext.NewTestLogger(t)

	man := &configuration.TemplateRepositoryManifest{
		Name:      "test",
		Arguments: requestArgs,
	}
	m, err := modulestest.NewModuleFromTemplates(man)
	if err != nil {
		t.Fatal(err)
	}

	fs, err := m.GetFS(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	f, err := fs.Create("templates/test.tpl")
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if args != nil {
		args = emulateYAMLParsing(t, args).(map[string]any)
	}

	test.s = NewStencil(&configuration.Manifest{
		Name:      "testing",
		Arguments: args,
		Modules:   []*configuration.TemplateRepository{{Name: m.Name}},
	}, nil, []*modules.Module{m}, log, false)

	// use the first template from the module
	// which we've created earlier after loading the module in the
	// NewModuleFromTemplates call. This won't be used, but it's
	// enough to set up the correct environment for running template test functions.
	tpls, err := test.s.getTemplates(t.Context(), log)
	if err != nil {
		t.Fatal(err)
	}
	test.t = tpls[0]
	test.log = log

	return test
}

// fakeTemplateMultipleModules returns a faked struct suitable for testing
// that has multiple modules in the project manifest, the first arguments list
// is for the first module, the second is for the second module, and so forth.
// the first module will import all other modules
func fakeTemplateMultipleModules(t *testing.T, manifestArgs map[string]any,
	args ...map[string]configuration.Argument) *testTpl {
	test := &testTpl{}
	log := slogext.NewTestLogger(t)

	mods := make([]*modules.Module, len(args))
	for i := range args {
		// Depend on the first module after us, if it exists.
		var deps []*configuration.TemplateRepository
		if i+1 < len(args) {
			deps = append(deps, &configuration.TemplateRepository{Name: fmt.Sprintf("test-%d", i+1)})
		}

		var err error
		mods[i], err = modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
			Name:      fmt.Sprintf("test-%d", i),
			Arguments: args[i],
			Modules:   deps,
		}, "testdata/args/test.tpl")
		assert.NilError(t, err)
	}

	if manifestArgs != nil {
		manifestArgs = emulateYAMLParsing(t, manifestArgs).(map[string]any)
	}

	test.s = NewStencil(&configuration.Manifest{
		Name:      "testing",
		Arguments: manifestArgs,
		// The first module brings in all the others.
		Modules: []*configuration.TemplateRepository{{Name: mods[0].Name}},
	}, nil, mods, log, false)

	// use the first template from the module
	// which we've created earlier after loading the module in the
	// NewModuleFromTemplates call. This won't be used, but it's
	// enough to set up the correct environment for running template test functions.
	tpls, err := test.s.getTemplates(t.Context(), log)
	assert.NilError(t, err)
	test.t = tpls[0]
	test.log = log

	return test
}

// emulateYAMLParsing takes the provided data, encodes it to yaml and
// then decodes it. This is done to allow for emulation of real yaml
// encoding.
func emulateYAMLParsing(t *testing.T, i any) (o any) {
	b, err := yaml.Marshal(i)
	assert.NilError(t, err, "expected marshal to succeed")
	assert.NilError(t, yaml.Unmarshal(b, &o), "expected unmarshal to succeed")
	return
}

func TestTplStencil_Arg(t *testing.T) {
	type args struct {
		pth string
	}
	tests := []struct {
		name    string
		fields  *testTpl
		args    args
		want    any
		wantErr bool
	}{
		{
			name: "should support basic argument",
			fields: fakeTemplate(t, map[string]any{
				"hello": "world",
			}, map[string]configuration.Argument{
				"hello": {},
			}),
			args: args{
				pth: "hello",
			},
			want: "world",
		},
		{
			name: "should fail when an argument is not defined",
			fields: fakeTemplate(t, map[string]any{
				"hello": "world",
			}, map[string]configuration.Argument{}),
			args: args{
				pth: "hello",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "should support basic JSON schema",
			fields: fakeTemplate(t, map[string]any{
				"hello": "world",
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]any{
						"type": "string",
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want: "world",
		},
		{
			name: "should fail when provided value doesn't match json schema",
			fields: fakeTemplate(t, map[string]any{
				"hello": 1,
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]any{
						"type": "string",
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should support nested json schema",
			fields: fakeTemplate(t, map[string]any{
				"hello": map[string]any{
					"world": map[string]any{
						"abc": []any{"def"},
					},
				},
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"world": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"abc": map[string]any{
										"type": "array",
									},
								},
							},
						},
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want:    map[string]any{"world": map[string]any{"abc": []any{"def"}}},
			wantErr: false,
		},
		{
			name: "should return default type when arg is not provided",
			fields: fakeTemplate(t, map[string]any{},
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]any{
							"type": "string",
						},
					},
				}),
			args: args{
				pth: "hello",
			},
			want: "",
		},
		{
			name: "should support from",
			fields: fakeTemplateMultipleModules(t,
				map[string]any{
					"hello": "world",
				},
				// test-0
				map[string]configuration.Argument{
					"hello": {
						From: "test-1",
					},
				},
				// test-1
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]any{
							"type": "string",
						},
					},
				},
			),
			args: args{
				pth: "hello",
			},
			want: "world",
		},
		{
			name: "should support from schema fail",
			fields: fakeTemplateMultipleModules(t,
				map[string]any{
					"hello": "world",
				},
				// test-0
				map[string]configuration.Argument{
					"hello": {
						From: "test-1",
					},
				},
				// test-1
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]any{
							"type": "number",
						},
					},
				},
			),
			args: args{
				pth: "hello",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should support recursive from",
			fields: fakeTemplateMultipleModules(t,
				map[string]any{
					"hello": "world",
				},
				// test-0
				map[string]configuration.Argument{
					"hello": {
						From: "test-1",
					},
				},
				// test-1
				map[string]configuration.Argument{
					"hello": {
						From: "test-2",
					},
				},
				// test-2
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]any{
							"type": "string",
						},
					},
				},
			),
			args: args{
				pth: "hello",
			},
			want: "world",
		},
		{
			name: "should support recursive from schema fail",
			fields: fakeTemplateMultipleModules(t,
				map[string]any{
					"hello": "world",
				},
				// test-0
				map[string]configuration.Argument{
					"hello": {
						From: "test-1",
					},
				},
				// test-1
				map[string]configuration.Argument{
					"hello": {
						From: "test-2",
					},
				},
				// test-2
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]any{
							"type": "number",
						},
					},
				},
			),
			args: args{
				pth: "hello",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "should return default value when not set and support json schema",
			fields: fakeTemplate(t, map[string]any{}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]any{
						"type": "object",
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want: map[string]any{},
		},
		{
			name: "should not return map[any]any",
			fields: fakeTemplate(t, map[string]any{
				"hello": map[int]string{0: "1"},
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]any{
						"type": "object",
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want: map[string]any{"0": "1"},
		},
		{
			name: "should retain argument default (int)",
			fields: fakeTemplate(t, map[string]any{}, map[string]configuration.Argument{
				"number": {
					Default: 100,
					Schema: map[string]any{
						"type": "number",
					},
				},
			}),
			args: args{
				pth: "number",
			},
			want: 100,
		},
		{
			name: "should retain argument default (float)",
			fields: fakeTemplate(t, map[string]any{}, map[string]configuration.Argument{
				"float": {
					Default: 0.1,
					Schema: map[string]any{
						"type": "number",
					},
				},
			}),
			args: args{
				pth: "float",
			},
			want: 0.1,
		},
		{
			name: "should retain argument default (bool)",
			fields: fakeTemplate(t, map[string]any{}, map[string]configuration.Argument{
				"boolean": {
					Default: true,
					Schema: map[string]any{
						"type": "boolean",
					},
				},
			}),
			args: args{
				pth: "boolean",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &TplStencil{
				s:   tt.fields.s,
				t:   tt.fields.t,
				log: tt.fields.log,
			}
			got, err := s.Arg(tt.args.pth)
			if (err != nil) != tt.wantErr {
				t.Errorf("TplStencil.Arg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TplStencil.Arg() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}
