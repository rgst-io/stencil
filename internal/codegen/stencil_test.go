package codegen

import (
	"context"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

func TestBasicE2ERender(t *testing.T) {
	fs := memfs.New()
	ctx := context.Background()
	log := slogext.NewTestLogger(t)

	// create stub manifest
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte("name: testing"))
	f.Close()

	// create a stub template
	f, err := fs.Create("templates/test-template.tpl")
	assert.NilError(t, err, "failed to create stub template")
	f.Write([]byte("{{ .Config.Name }}"))
	f.Close()

	tp, err := modulestest.NewWithFS(ctx, "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")
	st := NewStencil(&configuration.Manifest{
		Name:      "test",
		Arguments: map[string]any{},
	}, nil, []*modules.Module{tp}, log, false)

	tpls, err := st.Render(ctx, log)
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpls), 1, "expected Render() to return a single template")
	assert.Equal(t, len(tpls[0].Files), 1, "expected Render() template to return a single file")
	assert.Equal(t, tpls[0].Files[0].String(), "test", "expected Render() to return correct output")

	lock := st.GenerateLockfile(tpls)
	assert.DeepEqual(t, lock, &stencil.Lockfile{
		Version: version.Version.GitVersion,
		Modules: []*stencil.LockfileModuleEntry{
			{
				Name:    "testing",
				URL:     "vfs://testing",
				Version: &resolver.Version{Virtual: "vfs"},
			},
		},
		Files: []*stencil.LockfileFileEntry{
			{
				Name:     "test-template",
				Template: "test-template.tpl",
				Module:   "testing",
			},
		},
	})
}

func TestModuleHookRender(t *testing.T) {
	ctx := context.Background()
	log := slogext.NewTestLogger(t)

	// create modules
	m1man := &configuration.TemplateRepositoryManifest{
		Name: "testing1",
	}
	m1, err := modulestest.NewModuleFromTemplates(m1man, "testdata/module-hook/m1.tpl")
	if err != nil {
		t.Errorf("failed to create module 1: %v", err)
	}
	m2man := &configuration.TemplateRepositoryManifest{
		Name: "testing2",
	}
	m2, err := modulestest.NewModuleFromTemplates(m2man, "testdata/module-hook/m2.tpl")
	if err != nil {
		t.Errorf("failed to create module 2: %v", err)
	}

	st := NewStencil(&configuration.Manifest{
		Name:      "test",
		Arguments: map[string]interface{}{},
	}, nil, []*modules.Module{m1, m2}, log, false)

	tpls, err := st.Render(ctx, log)
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpls), 2, "expected Render() to return two templates")
	// template return order is randomized to prevent order dependencies
	slices.SortFunc(tpls, func(a, b *Template) int {
		return strings.Compare(a.Module.Name, b.Module.Name)
	})
	assert.Equal(t, len(tpls[1].Files), 1, "expected Render() m2 template to return a single file")
	assert.Equal(t, strings.TrimSpace(tpls[1].Files[0].String()), "a", "expected Render() m2 to return correct output")
}

func TestDirReplacementRendering(t *testing.T) {
	log := slogext.NewTestLogger(t)
	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]any{"x": "d"}}
	m1man := &configuration.TemplateRepositoryManifest{
		Name: "testing1",
		DirReplacements: map[string]string{
			"testdata":             `bob`,
			"testdata/replacement": `{{ stencil.Arg "x" }}`,
		},
		Arguments: map[string]configuration.Argument{"x": {Schema: map[string]any{"type": "string"}}},
	}
	m1, err := modulestest.NewModuleFromTemplates(m1man, "testdata/replacement/m1.tpl")
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, nil, []*modules.Module{m1}, log, false)

	tps, err := st.Render(context.Background(), log)
	assert.NilError(t, err, "failed to render template")
	assert.Equal(t, len(tps), 1)
	assert.Equal(t, len(tps[0].Files), 1)
	assert.Equal(t, tps[0].Files[0].path, "bob/d/m1")
}

func TestBinaryRender(t *testing.T) {
	log := slogext.NewTestLogger(t)
	sm := &configuration.Manifest{Name: "testing", Arguments: map[string]any{"x": "d"}}
	m1man := &configuration.TemplateRepositoryManifest{
		Name: "testing1",
	}
	m1, err := modulestest.NewModuleFromTemplates(m1man, "testdata/binary.nontpl")
	assert.NilError(t, err, "failed to NewWithFS")

	st := NewStencil(sm, nil, []*modules.Module{m1}, log, false)

	tpls, err := st.Render(context.Background(), log)
	assert.NilError(t, err, "failed to render template")
	assert.Equal(t, len(tpls), 1)
	assert.Equal(t, len(tpls[0].Files), 1)

	wd, err := os.Getwd()
	assert.NilError(t, err, "failed to get working directory")
	td := os.TempDir()
	err = os.Chdir(td)
	assert.NilError(t, err, "failed to change working directory")
	defer os.Chdir(wd)

	err = tpls[0].Files[0].Write(log, false)
	assert.NilError(t, err, "failed to file out")

	// read entire binary file
	cont, err := os.ReadFile(path.Join(wd, "testdata/binary.nontpl"))
	assert.NilError(t, err, "failed to load binary file")
	cont2, err := os.ReadFile(tpls[0].Files[0].path)
	assert.NilError(t, err, "failed to load output binary file")
	assert.DeepEqual(t, cont, cont2)
}

func TestBadDirReplacement(t *testing.T) {
	log := slogext.NewTestLogger(t)
	sm := &configuration.Manifest{Name: "testing"}
	m1man := &configuration.TemplateRepositoryManifest{Name: "testing1"}
	m, err := modulestest.NewModuleFromTemplates(m1man, "testdata/replacement/m1.tpl")
	assert.NilError(t, err, "failed to NewModuleFromTemplates")

	st := NewStencil(sm, nil, []*modules.Module{m}, log, false)
	vals := NewValues(context.Background(), sm, nil)
	_, err = st.renderDirReplacement("b/c", m, vals)
	assert.ErrorContains(t, err, "contains path separator in output")
}

// TestIterations ensures that stencil properly handles iterations based
// on state, in this case through global variables.
func TestIterations(t *testing.T) {
	fs := memfs.New()
	ctx := context.Background()
	log := slogext.NewTestLogger(t)

	// create stub manifest
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte("name: testing"))
	f.Close()

	// create a stub template
	f, err := fs.Create("templates/test-template.tpl")
	assert.NilError(t, err, "failed to create stub template")
	f.Write([]byte(`{{- stencil.SetGlobal "x" 1 }}
{{- stencil.SetGlobal "z" (add (stencil.GetGlobal "y") 1) }}
{{- stencil.SetGlobal "y" (add (stencil.GetGlobal "x") 1) }}

x: {{ stencil.GetGlobal "x" }}
y: {{ stencil.GetGlobal "y" }}
z: {{ stencil.GetGlobal "z" }}`))
	assert.NilError(t, f.Close(), "failed to close stub template")

	tp, err := modulestest.NewWithFS(ctx, "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")
	st := NewStencil(&configuration.Manifest{
		Name:      "test",
		Arguments: map[string]any{},
	}, nil, []*modules.Module{tp}, log, false)

	tpls, err := st.Render(ctx, log)
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpls), 1, "expected Render() to return a single template")
	assert.Equal(t, len(tpls[0].Files), 1, "expected Render() template to return a single file")

	var resp map[string]int
	assert.NilError(t, yaml.Unmarshal(tpls[0].Files[0].contents, &resp), "failed to unmarshal response")
	assert.DeepEqual(t, resp, map[string]int{"x": 1, "y": 2, "z": 3})

	lock := st.GenerateLockfile(tpls)
	assert.DeepEqual(t, lock, &stencil.Lockfile{
		Version: version.Version.GitVersion,
		Modules: []*stencil.LockfileModuleEntry{
			{
				Name:    "testing",
				URL:     "vfs://testing",
				Version: &resolver.Version{Virtual: "vfs"},
			},
		},
		Files: []*stencil.LockfileFileEntry{
			{
				Name:     "test-template",
				Template: "test-template.tpl",
				Module:   "testing",
			},
		},
	})
}
