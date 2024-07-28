//go:build !test_no_internet

package stencil

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jaredallard/vcs/resolver"
	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/pkg/configuration"
	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"gotest.tools/v3/assert"
)

// TODO(jaredallard): We shouldn't fetch live data here, but right now
// `replacements` override the logic we want to test here.

// TestResolveModulesShouldUseModulesFromLockfile ensures that the
// lockfile is used to source modules when a lockfile is available.
func TestResolveModulesShouldUseModulesFromLockfile(t *testing.T) {
	log := slogext.NewTestLogger(t)

	s := NewCommand(log, &configuration.Manifest{
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
	}, false)
	s.lock = &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				Commit: "3c3213721335c53fd78f4fede1b3704801616615",
				Tag:    "v0.5.0",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 1, "expected exactly one module")
	assert.DeepEqual(t, mods[0].Version, s.lock.Modules[0].Version)
}

// TestResolveModulesShouldUpgradeWhenExplicitlyAsked ensures that when
// 'stencil.yaml' is modified, the version in the lockfile is not used.
func TestResolveModulesShouldUpgradeWhenExplicitlyAsked(t *testing.T) {
	log := slogext.NewTestLogger(t)

	s := NewCommand(log, &configuration.Manifest{
		Modules: []*configuration.TemplateRepository{{
			Name:    "github.com/rgst-io/stencil-golang",
			Version: "v0.5.0", // 3c3213721335c53fd78f4fede1b3704801616615
		}},
	}, false)
	s.lock = &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				Commit: "fc954774dd29f0505158e86afbd18771ac92d50e",
				Tag:    "v0.4.0",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 1, "expected exactly one module")
	assert.DeepEqual(t, mods[0].Version, &resolver.Version{
		Commit: "3c3213721335c53fd78f4fede1b3704801616615",
		Tag:    "v0.5.0",
	})
}

// TestResolveShouldNotUpgradeOtherModulesWhenUpgradingOne tests that
// other modules do not get upgraded when one is upgraded.
func TestResolveShouldNotUpgradeOtherModulesWhenUpgradingOne(t *testing.T) {
	log := slogext.NewTestLogger(t)

	s := NewCommand(log, &configuration.Manifest{
		Modules: []*configuration.TemplateRepository{
			{
				Name:    "github.com/rgst-io/stencil-golang",
				Version: "v0.5.0", // 3c3213721335c53fd78f4fede1b3704801616615
			},
			{
				// TODO(jaredallard): We need some more test live repos.
				Name: "github.com/getoutreach/devbase",
			},
		},
	}, false)
	s.lock = &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				Commit: "fc954774dd29f0505158e86afbd18771ac92d50e",
				Tag:    "v0.4.0",
			},
		}, {
			Name: "github.com/getoutreach/devbase",
			Version: &resolver.Version{
				Commit: "850cae0d50691772bd56267d2056b9dd1b246176",
				Tag:    "v2.27.1",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 2, "expected exactly two modules")

	modsHM := slicesMap(mods, func(m *modules.Module) string { return m.Name })

	assert.DeepEqual(t, modsHM["github.com/rgst-io/stencil-golang"].Version, &resolver.Version{
		Commit: "3c3213721335c53fd78f4fede1b3704801616615",
		Tag:    "v0.5.0",
	})

	// other module shouldn't be changed
	assert.DeepEqual(t, modsHM["github.com/getoutreach/devbase"].Version, s.lock.Modules[1].Version)
}

// TestResolveModulesShouldUpdateReplacements ensures that stencil will
// 'upgrade' modules when a replacement is added.
func TestResolveModulesShouldUpdateReplacements(t *testing.T) {
	log := slogext.NewTestLogger(t)

	s := NewCommand(log, &configuration.Manifest{
		Modules: []*configuration.TemplateRepository{{
			Name: "github.com/rgst-io/stencil-golang",
		}},
		Replacements: map[string]string{
			"github.com/rgst-io/stencil-golang": filepath.Join("testdata", "stub-module"),
		},
	}, false)
	s.lock = &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/rgst-io/stencil-golang",
			Version: &resolver.Version{
				Commit: "3c3213721335c53fd78f4fede1b3704801616615",
				Tag:    "v0.5.0",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 1, "expected exactly one module")
	assert.DeepEqual(t, mods[0].Version, &resolver.Version{Virtual: "local"})
}

// TestResolveModulesShouldAllowAdds ensures that stencil supports
// adding new modules without running 'upgrade'.
func TestResolveModulesShouldAllowAdds(t *testing.T) {
	log := slogext.NewTestLogger(t)

	s := NewCommand(log, &configuration.Manifest{
		Modules: []*configuration.TemplateRepository{
			{
				Name:    "github.com/rgst-io/stencil-golang",
				Version: "v0.5.0", // 3c3213721335c53fd78f4fede1b3704801616615
			}, {
				// TODO(jaredallard): We need some more test live repos.
				Name: "github.com/getoutreach/devbase",
			},
		},
	}, false)
	s.lock = &stencil.Lockfile{
		Modules: []*stencil.LockfileModuleEntry{{
			Name: "github.com/getoutreach/devbase",
			Version: &resolver.Version{
				Commit: "850cae0d50691772bd56267d2056b9dd1b246176",
				Tag:    "v2.27.1",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 2, "expected exactly one module")

	mod := slicesMap(mods, func(m *modules.Module) string { return m.Name })["github.com/rgst-io/stencil-golang"]
	assert.DeepEqual(t, mod.Version, &resolver.Version{Tag: "v0.5.0", Commit: "3c3213721335c53fd78f4fede1b3704801616615"})
}
