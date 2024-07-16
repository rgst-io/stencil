// Copyright (C) 2024 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stencil

import (
	"context"
	"testing"

	"go.rgst.io/stencil/internal/modules/resolver"
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
				Commit: "bd265e16cf75c06e2569b6658735d38b025599e2",
				Branch: "main",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 1, "expected exactly one module")
	assert.Equal(t, mods[0].Version.String(), s.lock.Modules[0].Version.String(), "expected same version to be used")
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
	assert.Equal(t, mods[0].Version.Commit, "3c3213721335c53fd78f4fede1b3704801616615", "expected v0.5.0")
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
				Commit: "9395dd53daf6ba1b1e2c5fa04c49eceb4465f05d",
				Branch: "main",
			},
		}},
	}

	mods, err := s.resolveModules(context.Background(), false)
	assert.NilError(t, err, "failed to resolve modules")
	assert.Equal(t, len(mods), 2, "expected exactly two modules")
	assert.Equal(t, mods[0].Version.Commit, "3c3213721335c53fd78f4fede1b3704801616615", "expected v0.5.0")
	assert.Equal(t, mods[1].Version.String(), s.lock.Modules[1].Version.String(), "expected other module to not be mutated")
}
