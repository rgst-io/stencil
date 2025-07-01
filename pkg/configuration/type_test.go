package configuration_test

import (
	"testing"

	"go.rgst.io/stencil/v2/internal/yaml"
	"gotest.tools/v3/assert"

	"go.rgst.io/stencil/v2/pkg/configuration"
)

func TestTemplateRepositoryType(t *testing.T) {
	assert.NilError(t, nil)
	tests := []struct {
		Name           string
		In             string
		Contains       []configuration.TemplateRepositoryType
		DoesNotContain []configuration.TemplateRepositoryType
	}{
		{
			Name:           "empty",
			In:             "",
			Contains:       []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeTemplates},
			DoesNotContain: []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeExt},
		},
		{
			Name:           "string templates",
			In:             "templates",
			Contains:       []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeTemplates},
			DoesNotContain: []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeExt},
		},
		{
			Name:           "string extension",
			In:             "extension",
			Contains:       []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeExt},
			DoesNotContain: []configuration.TemplateRepositoryType{configuration.TemplateRepositoryTypeTemplates},
		},
		{
			Name: "legacy csv both",
			In:   "extension,templates",
			Contains: []configuration.TemplateRepositoryType{
				configuration.TemplateRepositoryTypeExt,
				configuration.TemplateRepositoryTypeTemplates,
			},
		},
		{
			Name: "slice both",
			In:   "- extension\n- templates",
			Contains: []configuration.TemplateRepositoryType{
				configuration.TemplateRepositoryTypeExt,
				configuration.TemplateRepositoryTypeTemplates,
			},
		},
	}

	for i := range tests {
		test := &tests[i]

		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			var ts configuration.TemplateRepositoryTypes
			assert.NilError(t, yaml.Unmarshal([]byte(test.In), &ts))

			for _, typ := range test.Contains {
				assert.Assert(t, ts.Contains(typ))
			}

			for _, typ := range test.DoesNotContain {
				assert.Assert(t, !ts.Contains(typ))
			}
		})
	}
}
