package codegen

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

// TestTplModule_Tpl tests [TplModule] in a template context.
func TestTplModule_Tpl(t *testing.T) {
	tests := []struct {
		name                string
		functionTemplate    string
		callingTemplate     string
		want                string
		renderStage         renderStage
		wantFuncErrContains string
		wantErrContains     string
	}{
		{
			name:            "should error on non-existent template",
			callingTemplate: `{{- module.Call "caller.HelloWorld" }}`,
			wantErrContains: `function "HelloWorld" in module "caller" was not registered`,
		},
		{
			name:            "should error on invalid function name",
			callingTemplate: `{{- module.Call "blah" }}`,
			wantErrContains: `expected format module.function, got "blah"`,
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
			name:        "should break on duplicate export",
			renderStage: renderStageFinal,
			functionTemplate: `{{- define "HelloWorld" -}}
		{{ return (fromYaml "hello: world") }}
		{{- end -}}
		{{- module.Export "HelloWorld" -}}
		{{- module.Export "HelloWorld" -}}`,
			callingTemplate:     ``,
			wantFuncErrContains: "already exported",
		},
		{
			name: "use context from module being called",
			functionTemplate: `{{- stencil.SetGlobal "a" "func" -}}
		{{- define "HelloWorld" -}}
		{{ return (stencil.GetGlobal "a") }}
		{{- end -}}
		{{- module.Export "HelloWorld" -}}`,
			callingTemplate: `{{- stencil.SetGlobal "a" "caller" -}}
		{{ module.Call "function.HelloWorld" }}`,
			want: "func",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slogext.NewTestLogger(t)

			// create stencil
			st := &Stencil{sharedState: newSharedState(), log: log}

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
				nil,
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
				nil,
			)
			assert.NilError(t, err, "expected NewTemplate to succeed")

			vals := NewValues(t.Context(), &configuration.Manifest{Name: t.Name()}, mods)

			// render template to register it
			st.renderStage = tt.renderStage
			err = functionTpl.Render(st, vals)
			if err != nil {
				if tt.wantFuncErrContains == "" {
					t.Errorf("unexpected function error: %v", err)
					return
				}

				assert.ErrorContains(t, err, tt.wantFuncErrContains, "expected function error to match")
				return
			} else if tt.wantFuncErrContains != "" {
				t.Errorf("expected function error to contain %q, got nil", tt.wantFuncErrContains)
			}

			// We already registered the function template, so we can render
			// the caller template now.
			st.renderStage = renderStageFinal

			err = callerTpl.Render(st, vals)
			if err != nil {
				if tt.wantErrContains == "" {
					t.Errorf("unexpected error: %v", err)
					return
				}

				assert.ErrorContains(t, err, tt.wantErrContains, "expected error to match")
				return
			} else if tt.wantErrContains != "" {
				t.Errorf("expected error to contain %q, got nil", tt.wantErrContains)
			}

			got := string(callerTpl.Files[0].contents)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
