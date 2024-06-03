// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Contains tests for the template file

package codegen

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "embed"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

//go:embed testdata/multi-file.tpl
var multiFileTemplate string

//go:embed testdata/multi-file-input.tpl
var multiFileInputTemplate string

//go:embed testdata/apply-template-passthrough.tpl
var applyTemplatePassthroughTemplate string

//go:embed testdata/generated-block/template.txt.tpl
var generatedBlockTemplate string

//go:embed testdata/generated-block/fake.txt
var fakeGeneratedBlockFile string

func memfsWithManifest(manifest string) billy.Filesystem {
	fs := memfs.New()
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte(manifest))
	f.Close()
	return fs
}

func TestSingleFileRender(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest("name: testing\n")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "virtual-file.tpl", 0o644, time.Now(), []byte("hello world!"), log)
	assert.NilError(t, err, "failed to create basic template")

	sm := &configuration.Manifest{Name: "testing"}

	st := NewStencil(sm, []*modules.Module{m}, log)
	err = tpl.Render(st, NewValues(context.Background(), sm, nil))
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, tpl.Files[0].String(), "hello world!", "expected Render() to modify first created file")
}

func TestMultiFileRender(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest("name: testing\narguments:\n  commands:\n    type: list")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "multi-file.tpl", 0o644,
		time.Now(), []byte(multiFileTemplate), log)
	assert.NilError(t, err, "failed to create template")

	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]interface{}{
		"commands": []string{"hello", "world", "command"},
	}}

	st := NewStencil(sm, []*modules.Module{m}, log)
	err = tpl.Render(st, NewValues(context.Background(), sm, nil))
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpl.Files), 3, "expected Render() to create 3 files")

	for i, f := range tpl.Files {
		assert.Equal(t, f.String(), "command", "rendered template %d contents differred", i)
	}
}

func TestMultiFileWithInputRender(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest("name: testing\narguments:\n  commands:\n    type: list")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "multi-file-input.tpl", 0o644,
		time.Now(), []byte(multiFileInputTemplate), log)
	assert.NilError(t, err, "failed to create template")

	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]interface{}{
		"commands": []string{"hello", "world", "command"},
	}}

	st := NewStencil(sm, []*modules.Module{m}, log)
	err = tpl.Render(st, NewValues(context.Background(), sm, nil))
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpl.Files), 3, "expected Render() to create 3 files")

	for i, f := range tpl.Files {
		assert.Equal(t, (sm.Arguments["commands"].([]string))[i], f.String(), "rendered template %d contents differred", i)
	}
}

func TestApplyTemplateArgumentPassthrough(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest("name: testing\narguments:\n  commands:\n    type: list")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "apply-template-passthrough.tpl", 0o644,
		time.Now(), []byte(applyTemplatePassthroughTemplate), log)
	assert.NilError(t, err, "failed to create template")

	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]interface{}{
		"commands": []string{"hello", "world", "command"},
	}}

	st := NewStencil(sm, []*modules.Module{m}, log)
	err = tpl.Render(st, NewValues(context.Background(), sm, nil))
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpl.Files), 1, "expected Render() to create 1 files")

	assert.Equal(t, "testing", tpl.Files[0].String(), "rendered template contents differed")
}

func TestGeneratedBlock(t *testing.T) {
	log := slogext.NewTestLogger(t)
	tempDir := t.TempDir()
	fakeFilePath := filepath.Join(tempDir, "generated-block.txt")

	fs := memfsWithManifest("name: testing\n")
	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]interface{}{}}
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, []*modules.Module{m}, log)
	assert.NilError(t, os.WriteFile(fakeFilePath, []byte(fakeGeneratedBlockFile), 0o644),
		"failed to write generated file")

	tpl, err := NewTemplate(m, "generated-block/template.tpl", 0o644,
		time.Now(), []byte(generatedBlockTemplate), log)
	assert.NilError(t, err, "failed to create template")

	tplf, err := NewFile(fakeFilePath, 0o644, time.Now())
	assert.NilError(t, err, "failed to create file")

	// Add the file (fake) to the template so that the template uses it for blocks
	tpl.Files = []*File{tplf}
	tpl.Render(st, NewValues(context.Background(), sm, nil))

	assert.Equal(t, tpl.Files[0].String(), fakeGeneratedBlockFile, "expected fake to equal rendered output")
}

// TestLibraryTemplate ensures that library templates don't generate
// files as well as that the library flag is set correctly.
func TestLibraryTemplate(t *testing.T) {
	fs := memfsWithManifest("name: testing\n")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	log := slogext.NewTestLogger(t)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "hello.library.tpl", 0o644, time.Now(), []byte("hello world!"), log)
	assert.NilError(t, err, "failed to create basic template")
	assert.Equal(t, tpl.Library, true, "expected library template to be marked as such")

	assert.NilError(t, tpl.Render(
		NewStencil(&configuration.Manifest{Name: "testing"}, []*modules.Module{m},
			log), NewValues(context.Background(), &configuration.Manifest{Name: "testing"}, nil)),
		"expected library template to not fail on render")

	assert.Equal(t, len(tpl.Files), 0, "expected library template to not generate files")
}

// TestLibraryCantAccessFileFunctions ensures that library templates
// can't access file functions in the template.
func TestLibraryCantAccessFileFunctions(t *testing.T) {
	fs := memfsWithManifest("name: testing\n")
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	log := slogext.NewTestLogger(t)
	assert.NilError(t, err, "failed to NewWithFS")

	tpl, err := NewTemplate(m, "hello.library.tpl", 0o644, time.Now(), []byte("{{ file.Create }}"), log)
	assert.NilError(t, err, "failed to create basic template")
	assert.Equal(t, tpl.Library, true, "expected library template to be marked as such")

	err = tpl.Render(NewStencil(&configuration.Manifest{Name: "testing"}, []*modules.Module{m}, log),
		NewValues(context.Background(), &configuration.Manifest{Name: "testing"}, nil))
	assert.ErrorContains(t, err,
		"attempted to use file in a template that doesn't support file rendering",
		"expected library template to fail on render",
	)
}

func TestSimpleDirReplacement(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest("name: testing\ndirReplacements:\n  a: 'b'\n")
	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]interface{}{}}
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, []*modules.Module{m}, log)
	tpl, err := NewTemplate(m, "a/base.tpl", 0o644, time.Now(), []byte("out"), log)
	assert.NilError(t, err, "failed to create template")

	tplf, err := NewFile("a/base", 0o644, time.Now())
	assert.NilError(t, err, "failed to create file")

	assert.Equal(t, tplf.path, "a/base")
	vals := NewValues(context.Background(), sm, nil)
	err = tpl.applyDirReplacements(tplf, st, vals)
	assert.NilError(t, err, "failed to apply dir replacements")
	assert.Equal(t, tplf.path, "b/base")
}

func TestNestedAndTemplatedDirReplacements(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest(
		"name: testing\ndirReplacements:\n  a: 'b'\n  a/c: '{{ stencil.Arg \"x\" }}'\narguments:\n  x:\n    type: string\n")
	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]any{"x": "d"}}
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, []*modules.Module{m}, log)
	tpl, err := NewTemplate(m, "a/c/base.tpl", 0o644, time.Now(), []byte("out"), log)
	assert.NilError(t, err, "failed to create template")

	tplf, err := NewFile("a/c/base", 0o644, time.Now())
	assert.NilError(t, err, "failed to create file")

	assert.Equal(t, tplf.path, "a/c/base")
	vals := NewValues(context.Background(), sm, nil)
	err = tpl.applyDirReplacements(tplf, st, vals)
	assert.NilError(t, err, "failed to apply dir replacements")
	assert.Equal(t, tplf.path, "b/d/base")
}

func TestBadDirReplacement(t *testing.T) {
	log := slogext.NewTestLogger(t)
	fs := memfsWithManifest(
		"name: testing\ndirReplacements:\n  a: 'b/c'\n")
	sm := &configuration.Manifest{Name: "testing"}
	m, err := modulestest.NewWithFS(context.Background(), "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, []*modules.Module{m}, log)
	tpl, err := NewTemplate(m, "a/base.tpl", 0o644, time.Now(), []byte("out"), log)
	assert.NilError(t, err, "failed to create template")

	tplf, err := NewFile("a/base", 0o644, time.Now())
	assert.NilError(t, err, "failed to create file")

	assert.Equal(t, tplf.path, "a/base")
	vals := NewValues(context.Background(), sm, nil)
	err = tpl.applyDirReplacements(tplf, st, vals)
	assert.ErrorContains(t, err, "contains path separator in output")
	assert.Equal(t, tplf.path, "a/base")
}
