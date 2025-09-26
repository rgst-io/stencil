// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

// Description: Contains high-level abstractions for testing template
// rendering.

package codegen

import (
	"os"
	"testing"
	"time"

	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

// RenderTemplate creates a template with the provided contents,
// executes it, and returns the post-rendered [Template].
//
// TODO(jaredallard): This shouldn't be exposed outside of tests, but
// right now we'd need a decently large test refactor so here it is.
func RenderTemplate(
	t *testing.T,
	mf *configuration.Manifest,
	trmf *configuration.TemplateRepositoryManifest,
	contents string,
) *Template {
	log := slogext.NewTestLogger(t)

	if mf == nil {
		mf = &configuration.Manifest{}
	}
	if mf.Name == "" {
		mf.Name = t.Name()
	}

	if trmf == nil {
		trmf = &configuration.TemplateRepositoryManifest{}
	}
	if trmf.Name == "" {
		trmf.Name = t.Name() + "Module"
	}

	m, err := modulestest.NewModuleFromTemplates(trmf)
	assert.NilError(t, err, "expected module creation to not fail")

	tpl, err := NewTemplate(m, "test.tpl", 0o755, time.Now(), []byte(contents), log, &NewTemplateOpts{})
	assert.NilError(t, err, "expected template creation to not fail")

	st := NewStencil(mf, nil, []*modules.Module{m}, log, false)
	assert.NilError(t, tpl.Render(st, NewValues(t.Context(), mf, []*modules.Module{m})), "expected render to not fail")

	return tpl
}

type quickTemplate struct {
	Filename string
	// TemplateContents is the contents of the template. If not set, only
	// ExistingFileContents will be written to Filename, if
	// ExistingFileContents is also set.
	TemplateContents     string
	ExistingFileContents string
}

func RenderTemplates(
	t *testing.T,
	mf *configuration.Manifest,
	trmf *configuration.TemplateRepositoryManifest,
	contents ...quickTemplate,
) []*Template {
	log := slogext.NewTestLogger(t)

	if mf == nil {
		mf = &configuration.Manifest{}
	}
	if mf.Name == "" {
		mf.Name = t.Name()
	}

	if trmf == nil {
		trmf = &configuration.TemplateRepositoryManifest{}
	}
	if trmf.Name == "" {
		trmf.Name = t.Name() + "Module"
	}

	m, err := modulestest.NewModuleFromTemplates(trmf)
	assert.NilError(t, err, "expected module creation to not fail")

	st := NewStencil(mf, nil, []*modules.Module{m}, log, false)
	tpls := make([]*Template, 0, len(contents))

	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	for _, qt := range contents {
		if qt.ExistingFileContents != "" {
			t.Logf("Wrote to %s", qt.Filename)
			assert.NilError(t, os.WriteFile(qt.Filename, []byte(qt.ExistingFileContents), 0o600), "expected WriteFile to work")
		}

		if qt.TemplateContents != "" {
			tpl, err := NewTemplate(m, qt.Filename+".tpl", 0o755, time.Now(), []byte(qt.TemplateContents), log, &NewTemplateOpts{})
			assert.NilError(t, err, "expected template creation to not fail")
			assert.NilError(t, tpl.Render(st, NewValues(t.Context(), mf, []*modules.Module{m})), "expected render to not fail")
			tpls = append(tpls, tpl)
		}
	}

	return tpls
}
