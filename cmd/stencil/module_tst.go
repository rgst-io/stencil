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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jaredallard/cmdexec"
	"github.com/samber/lo"
	"github.com/urfave/cli/v3"
	"go.rgst.io/stencil/v2/internal/cmd/stencil"
	"go.rgst.io/stencil/v2/internal/yaml"
	"go.rgst.io/stencil/v2/pkg/configuration"
	"go.rgst.io/stencil/v2/pkg/slogext"
)

// Test represents a stencil test.
type Test struct {
	// Name is the name of the test
	Name string

	// Dir is the directory where the test lives
	Dir string

	// Failed denotes if the test failed or not
	Failed bool

	// FailureReason is the optional reason that a test failed
	FailureReason error

	// TemplateRepoManifest is the manifest of the template repository
	// this test was testing.
	TemplateRepoManifest *configuration.TemplateRepositoryManifest
}

// runTest runs a test in the provided directory with the provided test
// name
func runTest(ctx context.Context, log slogext.Logger, dir string, t *Test) error {
	tmpTestDir, err := os.MkdirTemp("", "stencil-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory for test: %w", err)
	}

	if err := os.Chdir(tmpTestDir); err != nil {
		return fmt.Errorf("failed to change working directory: %w", err)
	}

	// Restore original work dir and clean up the temporary test directory
	defer func() {
		if err := os.Chdir(dir); err != nil {
			log.WithError(err).Warn("failed to restore previous working directory")
		}

		if err := os.RemoveAll(tmpTestDir); err != nil {
			log.WithError(err).Warn("failed to clean up temporary working directory")
		}
	}()

	manifestPath := filepath.Join(t.Dir, "stencil.yaml")
	serializedMfPath := filepath.Join(tmpTestDir, "stencil.yaml")

	mf, err := configuration.LoadManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load stencil.yaml: %w", err)
	}

	// Ensure we're using a name that matches the test name and that we're
	// using a local version of the module that we're trying to test.
	mf.Name = t.Name

	modules := lo.SliceToMap(mf.Modules,
		func(m *configuration.TemplateRepository) (string, *configuration.TemplateRepository) {
			return m.Name, m
		},
	)

	// Ensure that we control the version of ourself when testing, but
	// also allow other modules to be imported.
	modules[t.TemplateRepoManifest.Name] = &configuration.TemplateRepository{
		Name:    t.TemplateRepoManifest.Name,
		Version: "=0.0.0", // We replace it below.
	}
	mf.Modules = lo.MapToSlice(modules,
		func(_ string, m *configuration.TemplateRepository) *configuration.TemplateRepository {
			return m
		},
	)

	mf.Replacements = map[string]string{
		t.TemplateRepoManifest.Name: dir,
	}

	b, err := yaml.Marshal(mf)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest modifications: %w", err)
	}

	if err := os.WriteFile(serializedMfPath, b, 0o600); err != nil {
		return fmt.Errorf("failed to write manifest modifications to temp dir: %w", err)
	}

	tlog, tlogbuf := slogext.NewCapturedLogger()
	defer tlogbuf.Reset()

	tlog = tlog.With("test.name", t.Name)

	cmd := stencil.NewCommand(tlog, mf, false, false, false, false)
	if err := cmd.Run(ctx); err != nil {
		fmt.Print(tlogbuf.String())
		return err
	}

	for _, validator := range mf.Testing.Validators {
		cmd := cmdexec.CommandContext(ctx, "bash", "-euo", "pipefail", "-c", validator)
		if err := cmd.Run(); err != nil {
			fmt.Print(tlogbuf.String())
			return fmt.Errorf("validator failed (%s): %w", cmd.String(), err)
		}
	}

	return nil
}

// ModuleTestAction implements [cli.ActionFunc] for
// [NewModuleTestCommand].
func ModuleTestAction(log slogext.Logger) cli.ActionFunc {
	return func(ctx context.Context, _ *cli.Command) error {
		moduleDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		trmf, err := configuration.LoadDefaultTemplateRepositoryManifest()
		if err != nil {
			return fmt.Errorf("failed to load template repository manifest: %w", err)
		}

		testDir := "tests"
		if !filepath.IsAbs(testDir) {
			var err error
			testDir, err = filepath.Abs(testDir)
			if err != nil {
				return fmt.Errorf("failed to make test directory absolute: %w", err)
			}
		}

		dirs, err := os.ReadDir(testDir)
		if err != nil {
			return fmt.Errorf("failed to read tests directory: %w", err)
		}

		tests := make([]*Test, 0)
		for _, dir := range dirs {
			if dir.Name() == "." || dir.Name() == ".." {
				continue
			}

			tests = append(tests, &Test{
				Name:                 dir.Name(),
				Dir:                  filepath.Join(testDir, dir.Name()),
				TemplateRepoManifest: trmf,
			})
		}

		if len(tests) == 0 {
			log.Warn("no tests to run")
			return nil
		}

		log.Info("running tests", "tests", len(tests))
		success := 0
		failed := 0
		for _, t := range tests {
			log := log.With("test.name", t.Name)

			if err := runTest(ctx, log, moduleDir, t); err != nil {
				failed++
				t.Failed = true
				t.FailureReason = err

				log.WithError(err).Error("test failed")
				continue
			}

			success++

			log.Info("test succeeded")
		}

		log.Info("test summary", "success", success, "failed", failed, "total", len(tests))

		for _, t := range tests {
			if !t.Failed {
				continue
			}

			return fmt.Errorf("tests failed")
		}

		return nil
	}
}

// NewModuleTestCommand returns a new [cli.Command] for the module test
// command.
func NewModuleTestCommand(log slogext.Logger) *cli.Command {
	return &cli.Command{
		Name:        "test",
		Usage:       "Test a stencil module",
		Description: "Runs tests against a stencil module",
		UsageText:   "test",
		Action:      ModuleTestAction(log),
	}
}
