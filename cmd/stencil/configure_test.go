package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rgst-io/stencil/pkg/configuration"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

func TestConfigureModule(t *testing.T) {
	tt := []struct {
		Name                      string
		RemoveNativeExtensionFlag bool
		Given                     configuration.Manifest
		Expected                  configuration.Manifest
		ShouldError               bool
	}{
		{
			Name:                      "EnsureProjectNoChangeWithoutFlag",
			RemoveNativeExtensionFlag: true,
			Given: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
					},
					"reportingTeam": "test_name",
				},
			},
			Expected: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
					},
					"reportingTeam": "test_name",
				},
			},
			ShouldError: true,
		}, {
			Name:                      "EnsureNativeExtensionAddition",
			RemoveNativeExtensionFlag: false,
			Given: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
					},
					"reportingTeam": "test_name",
				},
			},
			Expected: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
						"force":             true,
					},
					"plugin":        true,
					"reportingTeam": "test_name",
				},
			},
			ShouldError: false,
		}, {
			Name:                      "EnsureNativeExtensionReversion",
			RemoveNativeExtensionFlag: true,
			Given: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
						"force":             true,
					},
					"plugin":        true,
					"reportingTeam": "test_name",
				},
			},
			Expected: configuration.Manifest{
				Name: "test",
				Modules: []*configuration.TemplateRepository{
					{
						Name: "github.com/rgst-io/stencil-template-base",
					},
				},
				Arguments: map[string]interface{}{
					"description": "test module configure",
					"releaseOptions": map[string]bool{
						"enablePrereleases": true,
					},
					"reportingTeam": "test_name",
				},
			},
			ShouldError: false,
		},
	}

	for _, test := range tt {
		test := test

		t.Run(test.Name, func(t *testing.T) {
			var tm = &configuration.Manifest{}
			var comp = &configuration.Manifest{}

			// Create temporary stencil.yaml with valid values
			tempFile := filepath.Join(t.TempDir(), "stencil.yaml")
			b, err := yaml.Marshal(test.Given)
			assert.NilError(t, err, "failed to marshal given yaml")
			assert.NilError(t, os.WriteFile(tempFile, b, 0o777), "failed to write file")

			// Setup expected values
			b, err = yaml.Marshal(test.Expected)
			assert.NilError(t, err, "failed to marshal expected yaml")
			err = yaml.Unmarshal(b, tm)
			assert.NilError(t, err, "failed to unmarshal expected yaml")

			// configure the stencil.yaml and compare to expected
			err = readAndMergeStencilYaml(tempFile, test.RemoveNativeExtensionFlag, "")
			if test.ShouldError == true {
				assert.Error(t, err, "no action")
			} else {
				assert.NilError(t, err, "failed to read and configure stencil.yaml")
			}

			b, err = os.ReadFile(tempFile)
			assert.NilError(t, err, "failed to read stencil.yaml")

			err = yaml.Unmarshal(b, comp)
			assert.NilError(t, err, "failed to unmarshal stencil.yaml")

			assert.DeepEqual(t, tm, comp)
		})
	}
}
