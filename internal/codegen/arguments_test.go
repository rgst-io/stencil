package codegen

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"gotest.tools/v3/assert"
)

func TestArguments(t *testing.T) {
	args, err := BuildArguments(&configuration.Manifest{
		Arguments: map[string]any{
			"a": map[string]any{
				"b": map[string]any{
					"c": "hello",
				},
				"d": map[string]any{
					"e": "world",
				},
			},
		},
	}, []configuration.TemplateRepositoryManifest{{
		Name: "testing",
		Arguments: map[string]configuration.Argument{
			"a.b": {
				Schema: map[string]any{
					"type": "string",
				},
			},
		},
	}})
	assert.NilError(t, err, "expected BuildArguments() to not fail")

	spew.Dump(args)

	t.Fail()
}
