package codegen

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/sirupsen/logrus"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/stencil"
	"gotest.tools/v3/assert"
)

func TestBasicE2ERender(t *testing.T) {
	fs := memfs.New()
	ctx := context.Background()

	// create stub manifest
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte("name: testing"))
	f.Close()

	// create a stub template
	f, err := fs.Create("templates/test-template.tpl")
	assert.NilError(t, err, "failed to create stub template")
	f.Write([]byte("{{ .Config.Name }}"))
	f.Close()

	st := NewStencil(&configuration.Manifest{
		Name:      "test",
		Arguments: map[string]interface{}{},
	}, []*modules.Module{
		modules.NewWithFS(ctx, "testing", fs),
	}, logrus.New())

	tpls, err := st.Render(ctx, logrus.New())
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpls), 1, "expected Render() to return a single template")
	assert.Equal(t, len(tpls[0].Files), 1, "expected Render() template to return a single file")
	assert.Equal(t, tpls[0].Files[0].String(), "test", "expected Render() to return correct output")

	lock := st.GenerateLockfile(tpls)
	assert.DeepEqual(t, lock, &stencil.Lockfile{
		Version: version.Version,
		Modules: []*stencil.LockfileModuleEntry{
			{
				Name:    "testing",
				URL:     "vfs://testing",
				Version: "vfs",
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
	}, []*modules.Module{m1, m2}, logrus.New())

	tpls, err := st.Render(ctx, logrus.New())
	assert.NilError(t, err, "expected Render() to not fail")
	assert.Equal(t, len(tpls), 2, "expected Render() to return two templates")
	// template return order is randomized to prevent order dependencies
	slices.SortFunc(tpls, func(a, b *Template) int {
		if a.Module.Name < b.Module.Name {
			return -1
		}
		if a.Module.Name > b.Module.Name {
			return 1
		}
		return 0
	})
	assert.Equal(t, len(tpls[1].Files), 1, "expected Render() m2 template to return a single file")
	assert.Equal(t, strings.TrimSpace(tpls[1].Files[0].String()), "a", "expected Render() m2 to return correct output")
}

func ExampleStencil_PostRun() {
	fs := memfs.New()
	ctx := context.Background()

	// create a stub manifest
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte("name: testing\npostRunCommand:\n- command: echo \"hello\""))
	f.Close()

	nullLog := logrus.New()
	nullLog.SetOutput(io.Discard)

	st := NewStencil(&configuration.Manifest{
		Name:      "test",
		Arguments: map[string]interface{}{},
	}, []*modules.Module{
		modules.NewWithFS(ctx, "testing", fs),
	}, logrus.New())
	err := st.PostRun(ctx, nullLog)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// hello
}
