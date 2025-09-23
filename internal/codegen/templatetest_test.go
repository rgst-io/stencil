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
	"path/filepath"
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

	m, err := modulestest.NewModuleFromTemplates(t, trmf)
	assert.NilError(t, err, "expected module creation to not fail")

	tpl, err := NewTemplate(m, "test.tpl", 0o755, time.Now(), []byte(contents), log, &NewTemplateOpts{})
	assert.NilError(t, err, "expected template creation to not fail")

	st := NewStencil(mf, nil, []*modules.Module{m}, log, false)
	assert.NilError(t, tpl.Render(st, NewValues(t.Context(), mf, []*modules.Module{m})), "expected render to not fail")

	return tpl
}

type quickTemplate struct {
	Filename             string
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

	m, err := modulestest.NewModuleFromTemplates(t, trmf)
	assert.NilError(t, err, "expected module creation to not fail")

	st := NewStencil(mf, nil, []*modules.Module{m}, log, false)

	tpls := make([]*Template, len(contents))

	fs, err := m.GetFS(t.Context())
	assert.NilError(t, err, "expected module to have fs")

	for i, qt := range contents {
		if qt.ExistingFileContents != "" {
			f, err := fs.Create(qt.Filename)
			assert.NilError(t, err, "expected create to work")
			_, err = f.Write([]byte(qt.ExistingFileContents))
			assert.NilError(t, err, "expected write to work")
			assert.NilError(t, f.Close(), "expected close to work")
		}
		tpl, err := NewTemplate(m, qt.Filename+".tpl", 0o755, time.Now(), []byte(qt.TemplateContents), log, &NewTemplateOpts{})
		assert.NilError(t, err, "expected template creation to not fail")
		assert.NilError(t, tpl.Render(st, NewValues(t.Context(), mf, []*modules.Module{m})), "expected render to not fail")
		tpls[i] = tpl
	}

	return tpls
}

func MoveFileToVFS(t *testing.T, tpl *Template, fpath string) []byte {
	contents, err := os.ReadFile(fpath)
	assert.NilError(t, err, "expected os.ReadFile to succeed")

	fs, err := tpl.Module.GetFS(t.Context())
	assert.NilError(t, err, "expected GetFS to succeed")
	fpb := filepath.Base(fpath)
	if fpb != "" {
		assert.NilError(t, fs.MkdirAll(fpb, 0o755), "expected MkdirAll to succeed")
	}
	btf, err := fs.Create(fpath)
	assert.NilError(t, err, "expected OpenFile to succeed")
	_, err = btf.Write(contents)
	assert.NilError(t, err, "expected Write to succeed")
	assert.NilError(t, btf.Close(), "expected Close to succeed")

	return contents
}

func MoveDirToVFS(t *testing.T, tpl *Template, fpath string) {
	contents, err := os.ReadDir(fpath)
	assert.NilError(t, err, "expected os.ReadDir to succeed")

	for _, entry := range contents {
		if entry.IsDir() {
			MoveDirToVFS(t, tpl, filepath.Join(fpath, entry.Name()))
			continue
		}
		MoveFileToVFS(t, tpl, filepath.Join(fpath, entry.Name()))
	}
}

func NewTestTemplate(t *testing.T, m *modules.Module, fpath string, mode os.FileMode,
	modTime time.Time, contents []byte, log slogext.Logger, opts *NewTemplateOpts) (*Template, error) {
	tp, err := NewTemplate(m, fpath, mode, modTime, contents, log, opts)
	tp.args = &Values{Context: t.Context()}
	tp.Module = m
	assert.NilError(t, err, "expected template creation to not fail")
	return tp, err
}
