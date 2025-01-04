package codegen

import (
	"context"
	"os"
	"testing"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/internal/modules/modulestest"
	"go.rgst.io/stencil/internal/version"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
)

func TestValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp(t.TempDir(), "stencil-values-test")
	assert.NilError(t, err, "expected os.MkdirTemp() not to fail")

	env.ChangeWorkingDir(t, tmpDir)

	r, err := gogit.PlainInit(tmpDir, false)
	assert.NilError(t, err, "expected gogit.PlainInit() not to fail")

	wrk, err := r.Worktree()
	assert.NilError(t, err, "expected gogit.(Repository).Worktree() not to fail")

	cmt, err := wrk.Commit("initial commit", &gogit.CommitOptions{
		AllowEmptyCommits: true,
		Author: &object.Signature{
			Name:  "Stencil",
			Email: "email@example.com",
			When:  time.Now(),
		},
	})
	assert.NilError(t, err, "expected worktree.Commit() not to fail")

	err = wrk.Checkout(&gogit.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName("main"),
	})
	assert.NilError(t, err, "expected worktree.Checkout() not to fail")

	sm := &configuration.Manifest{
		Name: "testing",
	}

	vals := NewValues(context.Background(), sm, []*modules.Module{
		{
			Name: "testing",
			Version: &resolver.Version{
				Tag:    "1.2.3",
				Commit: "abc",
			},
		},
	})
	assert.DeepEqual(t, &Values{
		Git: git{
			Ref:           plumbing.NewBranchReferenceName("main").String(),
			Commit:        cmt.String(),
			Dirty:         false,
			DefaultBranch: "main",
		},
		Runtime: runtimeVals{
			Generator:        "stencil",
			GeneratorVersion: version.Version.GitVersion,
			Modules: modulesSlice{
				{
					Name:    "testing",
					Version: &resolver.Version{Tag: "1.2.3", Commit: "abc"},
				},
			},
		},
		Config: config{
			Name: sm.Name,
		},
	}, vals)
}

func TestGeneratedValues(t *testing.T) {
	log := slogext.NewTestLogger(t)

	man := &configuration.TemplateRepositoryManifest{
		Name: "testing",
	}
	m, err := modulestest.NewModuleFromTemplates(man, "testdata/values/values.tpl")
	assert.NilError(t, err, "failed to create module")

	st := NewStencil(&configuration.Manifest{
		Name:      "testing",
		Arguments: map[string]interface{}{},
	}, nil, []*modules.Module{m}, log, false)
	tpls, err := st.Render(context.Background(), log)
	assert.NilError(t, err, "failed to render templates")
	assert.Equal(t,
		tpls[0].Files[0].String(),
		"virtual (source: vfs) virtual (source: vfs) virtual (source: vfs) testdata/values/values.tpl\n",
	)
}
