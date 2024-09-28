package codegen

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

// TestTplModule_Tpl tests [TplModule] in a template context.
func TestTplModule_Tpl(t *testing.T) {
	tests := []struct {
		name             string
		functionTemplate string
		callingTemplate  string
		want             string
		wantErrContains  string
	}{
		{
			name:            "should error on non-existent template",
			callingTemplate: `{{- module.Call "caller.HelloWorld" }}`,
			wantErrContains: `calling Call: module "caller" did not register any functions or was not imported`,
		},
		{
			name: "should support calling exported function",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return "Hello, world!" }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ module.Call "function.HelloWorld" }}`,
			want:            "Hello, world!",
		},
		{
			name: "should support returning errors",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return "" (error "Failed!") }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ module.Call "function.HelloWorld" }}`,
			wantErrContains: "Failed!",
		},
		{
			name: "should not continue processing after error",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return "" (error "Failed!") }}
{{ fail "should not continue processing" }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ module.Call "function.HelloWorld" }}`,
			wantErrContains: "Failed!",
		},
		{
			name: "should support passing concrete types",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return (fromYaml "hello: world") }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ (module.Call "function.HelloWorld").hello }}`,
			want:            "world",
		},
		{
			name: "should pass data between calls",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return .Data }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ module.Call "function.HelloWorld" "hello" }}`,
			want:            "hello",
		},
		{
			name: "should use data from caller",
			functionTemplate: `{{- define "HelloWorld" -}}
{{ return (file.Path) }}
{{- end -}}
{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{ module.Call "function.HelloWorld" }}`,
			// This is "caller" because the path should be from the template
			// we are calling from. Otherwise, this would be "function".
			want: "caller",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)

			// create stencil
			st := &Stencil{sharedData: &sharedData{globals: make(map[string]global)}, moduleCaller: NewModuleCaller(), log: log}

			// create calling module
			callerModule, err := modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
				Name: "caller",
			})
			assert.NilError(t, err, "expected NewModuleFromTemplates to succeed")

			// create function template for module
			functionModule, err := modulestest.NewModuleFromTemplates(&configuration.TemplateRepositoryManifest{
				Name: "function",
			})
			assert.NilError(t, err, "expected NewModuleFromTemplates to succeed")

			mods := []*modules.Module{callerModule, functionModule}
			st.modules = mods

			// create calling template
			callerTpl, err := NewTemplate(
				callerModule,
				"caller.tpl",
				0o755,
				time.Now(),
				[]byte(tt.callingTemplate),
				log,
			)
			assert.NilError(t, err, "expected NewTemplate to succeed")

			// create function template
			functionTpl, err := NewTemplate(
				functionModule,
				"function.library.tpl",
				0o755,
				time.Now(),
				[]byte(tt.functionTemplate),
				log,
			)
			assert.NilError(t, err, "expected NewTemplate to succeed")

			vals := NewValues(context.Background(), &configuration.Manifest{Name: t.Name()}, mods)

			// render template to register it
			st.isFirstPass = true
			assert.NilError(t, functionTpl.Render(st, vals), "expected Render to succeed")

			// We already registered the function template, so we can render
			// the caller template now.
			st.isFirstPass = false

			err = callerTpl.Render(st, vals)
			if err != nil {
				if tt.wantErrContains == "" {
					t.Errorf("unexpected error: %v", err)
					return
				}

				assert.ErrorContains(t, err, tt.wantErrContains, "expected error to match")
				return
			}

			got := string(callerTpl.Files[0].contents)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
