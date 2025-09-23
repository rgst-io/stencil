//go:build !test_no_internet

package modules_test

import (
	"testing"

	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/v2/internal/modules"
	"go.rgst.io/stencil/v2/internal/modules/modulestest"
	"go.rgst.io/stencil/v2/internal/testing/testmemfs"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

// newLogger creates a new logger for testing
func newLogger(t *testing.T) slogext.Logger {
	return slogext.NewTestLogger(t)
}

func TestCanFetchModule(t *testing.T) {
	ctx := t.Context()
	m, err := modules.New(ctx, "", modules.NewModuleOpts{
		ImportPath: "github.com/rgst-io/stencil-module", Version: &resolver.Version{Branch: "main"},
	})
	assert.NilError(t, err, "failed to call New()")
	assert.Assert(t, m.Manifest.Type.Contains(configuration.TemplateRepositoryTypeTemplates), "failed to validate returned manifest")

	fs, err := m.GetFS(ctx)
	assert.NilError(t, err, "failed to call GetFS() on module")

	_, err = fs.Stat("manifest.yaml")
	assert.NilError(t, err, "failed to validate returned manifest from fs")
}

func TestReplacementLocalModule(t *testing.T) {
	sm := &configuration.Manifest{
		Name: "testing-project",
		Modules: []*configuration.TemplateRepository{
			{
				Name: "github.com/rgst-io/stencil-module",
			},
		},
		Replacements: map[string]string{
			"github.com/rgst-io/stencil-module": "file://testdata",
		},
	}

	mods, err := modules.FetchModules(t.Context(), &modules.ModuleResolveOptions{Manifest: sm, Log: newLogger(t)})
	assert.NilError(t, err, "expected GetModulesForProject() to not error")
	assert.Equal(t, len(mods), 1, "expected exactly one module to be returned")
	assert.Equal(t, mods[0].URI, sm.Replacements["github.com/rgst-io/stencil-module"],
		"expected module to use replacement URI")
}

func TestCanGetLatestVersion(t *testing.T) {
	ctx := t.Context()
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name: "github.com/rgst-io/stencil-module",
				},
			},
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")
	assert.Assert(t, len(mods) >= 1, "expected at least one module to be returned")
}

func TestHandleMultipleConstraints(t *testing.T) {
	ctx := t.Context()
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/getoutreach/stencil-base",
					Version: "=<0.5.0",
				},
				{
					Name: "nested_constraint",
				},
			},
			Replacements: map[string]string{
				"nested_constraint": "file://testdata/nested_constraint",
			},
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")
	assert.Equal(t, len(mods), 2, "expected exactly two modules to be returned")

	// find stencil-base to validate version
	index := -1
	for i, m := range mods {
		if m.Name == "github.com/getoutreach/stencil-base" {
			index = i
			break
		}
	}

	assert.DeepEqual(t,
		mods[index].Version,
		&resolver.Version{Tag: "v0.3.2", Commit: "91797167d0e48ae4c9640c0acbd7447eb9e1e5e4"},
	)
}

func TestHandleNestedModules(t *testing.T) {
	ctx := t.Context()
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name: "a",
				},
			},
			Replacements: map[string]string{
				"a": "file://testdata/nested_modules/a",
				"b": "file://testdata/nested_modules/b",
			},
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")

	// ensure that a resolved b
	assert.Equal(t, len(mods), 2, "expected exactly two modules to be returned")

	// ensure that we resolved both a and b
	found := 0
	for _, m := range mods {
		if m.Name == "a" || m.Name == "b" {
			found++
		}
	}

	assert.Equal(t, found, 2, "expected both modules to be returned")
}

func TestFailOnIncompatibleConstraints(t *testing.T) {
	ctx := t.Context()
	_, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/getoutreach/stencil-base",
					Version: ">=0.5.0",
				},
				{
					// wants patch of 0.3.0
					Name: "nested_constraint",
				},
			},
			Replacements: map[string]string{
				"nested_constraint": "file://testdata/nested_constraint",
			},
		},
		Log: newLogger(t),
	})
	assert.Error(t, err,
		"failed to resolve module 'github.com/getoutreach/stencil-base': no versions found that satisfy criteria\n"+
			"\n"+
			"Constraints:\n"+
			"└─ testing-project (top-level) wants >=0.5.0\n"+
			"  └─ nested_constraint@virtual (source: local) wants ~0.3.0\n",
		"expected GetModulesForProject() to error")
}

func TestCanUseBranch(t *testing.T) {
	ctx := t.Context()
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/rgst-io/stencil-module",
					Version: "main",
				},
			},
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")

	var mod *modules.Module
	for _, m := range mods {
		if m.Name == "github.com/rgst-io/stencil-module" {
			mod = m
			break
		}
	}
	if mod == nil {
		t.Fatal("failed to find module")
	}

	assert.Equal(t, mod.Version.Branch, "main", "expected module to match")
}

func TestBranchAlwaysUsedOverDependency(t *testing.T) {
	ctx := t.Context()

	// Create in-memory module that also requires stencil-base
	man := &configuration.TemplateRepositoryManifest{
		Name: "test",
		Modules: []*configuration.TemplateRepository{
			{
				Name:    "github.com/rgst-io/stencil-module",
				Version: ">=v0.0.0",
			},
		},
	}
	mDep, err := modulestest.NewModuleFromTemplates(t, man)
	assert.NilError(t, err, "failed to create dep module")

	// Resolve a fake project that requires a branch of a dependency that the in-memory module also requires
	// but with a different version constraint
	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Replacements: map[string]*modules.Module{"test-dep": mDep},
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/rgst-io/stencil-module",
					Version: "main",
				},
				{
					Name: "test-dep",
				},
			},
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")

	var mod *modules.Module
	for _, m := range mods {
		if m.Name == "github.com/rgst-io/stencil-module" {
			mod = m
			break
		}
	}
	if mod == nil {
		t.Fatal("failed to find module")
	}

	assert.Equal(t, mod.Version.Branch, "main", "expected module to match")
}

func TestShouldResolveInMemoryModule(t *testing.T) {
	ctx := t.Context()

	// require test-dep which is also an in-memory module to make sure that we can resolve at least once
	// an in-memory module
	man := &configuration.TemplateRepositoryManifest{
		Name: "test",
		Modules: []*configuration.TemplateRepository{
			{Name: "test-dep"},
		},
	}
	m, err := modulestest.NewModuleFromTemplates(t, man)
	assert.NilError(t, err, "failed to create module")

	// this relies on the top-level to ensure that re-resolving still picks
	// the in-memory module
	man = &configuration.TemplateRepositoryManifest{
		Name: "test-dep",
		Modules: []*configuration.TemplateRepository{
			{Name: "test"},
		},
	}
	mDep, err := modulestest.NewModuleFromTemplates(t, man)
	assert.NilError(t, err, "failed to create dep module")

	mods, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{Name: "test"},
			},
		},
		Replacements: map[string]*modules.Module{
			"test":     m,
			"test-dep": mDep,
		},
		Log: newLogger(t),
	})
	assert.NilError(t, err, "failed to call GetModulesForProject()")
	assert.Equal(t, len(mods), 2, "expected exactly two modules to be returned")

	var mod *modules.Module
	for _, m := range mods {
		if m.Name == "test" {
			mod = m
			break
		}
	}
	assert.Equal(t, mod.Name, m.Name, "expected module to match")
}

func TestShouldErrorOnTwoDifferentBranches(t *testing.T) {
	ctx := t.Context()
	_, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/rgst-io/stencil-module",
					Version: "main",
				},
				{
					Name:    "github.com/rgst-io/stencil-module",
					Version: "rc",
				},
			},
		},
		Log: newLogger(t),
	})
	assert.ErrorContains(t, err,
		"failed to resolve module 'github.com/rgst-io/stencil-module': unable to satisfy multiple branch constraints (rc, main)\n"+
			"\n"+
			"Constraints:\n"+
			"└─ testing-project (top-level) wants branch main\n"+
			"  └─ testing-project (top-level) wants branch rc\n",
		"expected GetModulesForProject() to error")
}

func TestSimpleDirReplacement(t *testing.T) {
	fs, err := testmemfs.WithManifest("name: testing\ndirReplacements:\n  a: 'b'\n")
	assert.NilError(t, err, "failed to testmemfs.WithManifest")
	m, err := modulestest.NewWithFS(t, "testing", fs)
	assert.NilError(t, err, "failed to NewWithFS")

	m.StoreDirReplacements(map[string]string{"a": "b"})

	assert.Equal(t, m.ApplyDirReplacements("a/base"), "b/base")
}

func TestShouldErrorOnNonExistentRepo(t *testing.T) {
	ctx := t.Context()
	_, err := modules.FetchModules(ctx, &modules.ModuleResolveOptions{
		Manifest: &configuration.Manifest{
			Name: "testing-project",
			Modules: []*configuration.TemplateRepository{
				{
					Name:    "github.com/rgst-io/i-am-not-a-real-repo",
					Version: "main",
				},
			},
		},
		Log: newLogger(t),
	})
	//nolint:lll // Why: Error message is long.
	assert.Error(t, err, "failed to resolve module 'github.com/rgst-io/i-am-not-a-real-repo': failed to get remote branches: exec failed (exit status 128): remote: Repository not found.\nfatal: repository 'https://github.com/rgst-io/i-am-not-a-real-repo/' not found\n\n\nThis error could be due to invalid credentials. Ensure your git configuration is correct.", "expected GetModulesForProject() to error")
}
