package codegen

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

type testTpl struct {
	s   *Stencil
	t   *Template
	log slogext.Logger
}

// fakeTemplate returns a faked struct suitable for testing
// template functions
func fakeTemplate(t *testing.T, args map[string]interface{},
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

	fs, err := m.GetFS(context.Background())
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

	test.s = NewStencil(&configuration.Manifest{
		Name:      "testing",
		Arguments: args,
		Modules:   []*configuration.TemplateRepository{{Name: m.Name}},
	}, nil, []*modules.Module{m}, log, false)

	// use the first template from the module
	// which we've created earlier after loading the module in the
	// NewModuleFromTemplates call. This won't be used, but it's
	// enough to set up the correct environment for running template test functions.
	tpls, err := test.s.getTemplates(context.Background(), log)
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
func fakeTemplateMultipleModules(t *testing.T, manifestArgs map[string]interface{},
	args ...map[string]configuration.Argument) *testTpl {
	test := &testTpl{}
	log := slogext.NewTestLogger(t)

	mods := make([]*modules.Module, len(args))
	importList := []string{}
	for i := range args {
		if i == 0 {
			continue
		}

		man := &configuration.TemplateRepositoryManifest{
			Name:      fmt.Sprintf("test-%d", i),
			Arguments: args[i],
		}
		m, err := modulestest.NewModuleFromTemplates(man, "testdata/args/test.tpl")
		if err != nil {
			t.Fatal(err)
		}
		importList = append(importList, m.Name)
		mods[i] = m
	}

	var trs []*configuration.TemplateRepository
	for _, imp := range importList {
		trs = append(trs, &configuration.TemplateRepository{Name: imp})
	}
	man := &configuration.TemplateRepositoryManifest{
		Name:      "test-0",
		Arguments: args[0],
		Modules:   trs,
	}
	var err error
	mods[0], err = modulestest.NewModuleFromTemplates(man, "testdata/args/test.tpl")
	if err != nil {
		t.Fatal(err)
	}

	moduleTr := make([]*configuration.TemplateRepository, len(mods))
	for i := range mods {
		moduleTr[i] = &configuration.TemplateRepository{Name: mods[i].Name}
	}

	test.s = NewStencil(&configuration.Manifest{
		Name:      "testing",
		Arguments: manifestArgs,
		Modules:   moduleTr,
	}, nil, mods, log, false)

	// use the first template from the module
	// which we've created earlier after loading the module in the
	// NewModuleFromTemplates call. This won't be used, but it's
	// enough to set up the correct environment for running template test functions.
	tpls, err := test.s.getTemplates(context.Background(), log)
	if err != nil {
		t.Fatal(err)
	}
	test.t = tpls[0]
	test.log = log

	return test
}

func TestTplStencil_Arg(t *testing.T) {
	type args struct {
		pth string
	}
	tests := []struct {
		name    string
		fields  *testTpl
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "should support basic argument",
			fields: fakeTemplate(t, map[string]interface{}{
				"hello": "world",
			}, map[string]configuration.Argument{
				"hello": {},
			}),
			args: args{
				pth: "hello",
			},
			want:    "world",
			wantErr: false,
		},
		{
			name: "should fail when an argument is not defined",
			fields: fakeTemplate(t, map[string]interface{}{
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
			fields: fakeTemplate(t, map[string]interface{}{
				"hello": "world",
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]interface{}{
						"type": "string",
					},
				},
			}),
			args: args{
				pth: "hello",
			},
			want:    "world",
			wantErr: false,
		},
		{
			name: "should fail when provided value doesn't match json schema",
			fields: fakeTemplate(t, map[string]interface{}{
				"hello": 1,
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]interface{}{
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
			fields: fakeTemplate(t, map[string]interface{}{
				"hello": map[string]interface{}{
					"world": map[string]interface{}{
						"abc": []interface{}{"def"},
					},
				},
			}, map[string]configuration.Argument{
				"hello": {
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"world": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"abc": map[string]interface{}{
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
			want:    map[string]interface{}{"world": map[string]interface{}{"abc": []interface{}{"def"}}},
			wantErr: false,
		},
		{
			name: "should return default type when arg is not provided",
			fields: fakeTemplate(t, map[string]interface{}{},
				map[string]configuration.Argument{
					"hello": {
						Schema: map[string]interface{}{
							"type": "string",
						},
					},
				}),
			args: args{
				pth: "hello",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "should support from",
			fields: fakeTemplateMultipleModules(t,
				map[string]interface{}{
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
						Schema: map[string]interface{}{
							"type": "string",
						},
					},
				},
			),
			args: args{
				pth: "hello",
			},
			want:    "world",
			wantErr: false,
		},
		{
			name: "should support from schema fail",
			fields: fakeTemplateMultipleModules(t,
				map[string]interface{}{
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
						Schema: map[string]interface{}{
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
				t.Errorf("TplStencil.Arg() = %v, want %v", got, tt.want)
			}
		})
	}
}
