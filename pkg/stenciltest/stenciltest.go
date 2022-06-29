// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file implements the stenciltest framework
// for testing templates generated by stencil.

// Package stenciltest contains code for testing templates
package stenciltest

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/getoutreach/stencil/internal/codegen"
	"github.com/getoutreach/stencil/internal/modules"
	"github.com/getoutreach/stencil/internal/modules/modulestest"
	"github.com/getoutreach/stencil/pkg/configuration"
	"github.com/getoutreach/stencil/pkg/extensions/apiv1"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/assert"
)

// Template is a template that is being tested by the stenciltest framework.
type Template struct {
	// path is the path to the template.
	path string

	// aditionalTemplates is a list of additional templates to add to the renderer,
	// but not to snapshot.
	additionalTemplates []string

	// m is the template repository manifest for this test
	m *configuration.TemplateRepositoryManifest

	// t is a testing object.
	t *testing.T

	// args are the arguments to the template.
	args map[string]interface{}

	// exts holds the inproc extensions
	exts map[string]apiv1.Implementation

	// errStr is the string an error should contain, if this is set then the template
	// MUST error.
	errStr string

	// persist denotes if we should save a snapshot or not
	// This is meant for tests.
	persist bool
}

// New creates a new test for a given template.
func New(t *testing.T, templatePath string, additionalTemplates ...string) *Template {
	// GOMOD: <module path>/go.mod
	b, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("failed to determine path to manifest: %v", err)
	}
	basepath := strings.TrimSuffix(strings.TrimSpace(string(b)), "/go.mod")

	b, err = os.ReadFile(filepath.Join(basepath, "manifest.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var m configuration.TemplateRepositoryManifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	return &Template{
		t:                   t,
		m:                   &m,
		path:                templatePath,
		additionalTemplates: additionalTemplates,
		persist:             true,
		exts:                map[string]apiv1.Implementation{},
	}
}

// Args sets the arguments to the template.
func (t *Template) Args(args map[string]interface{}) *Template {
	t.args = args
	return t
}

// Ext registers an in-proc extension with the current stencil template. The stenciltest library
// does not load the real extensions (because extensions can invoke outbound network calls).
// It is up to the unit test to provide each extension used by their template with this API.
// Unit tests can decide if they can use the real implementation of the extension AS IS or if a
// mock extension is needed to feed fake data per test case.
//
// Note: even though input extension is registered inproc, its response to ExecuteTemplateFunction
// will be encoded as JSON and decoded back as a plain inteface{} to simulate the GRPC transport
// layer between stencil and the same extension. Refer to the inprocExt struct docs for details.
func (t *Template) Ext(name string, ext apiv1.Implementation) *Template {
	t.exts[name] = inprocExt{ext: ext}
	return t
}

// ErrorContains denotes that this test run should fail, and the message
// should contain the provided string.
//
//   t.ErrorContains("i am an error")
func (t *Template) ErrorContains(msg string) {
	t.errStr = msg
}

// Run runs the test.
func (t *Template) Run(save bool) {
	t.t.Run(t.path, func(got *testing.T) {
		m, err := modulestest.NewModuleFromTemplates(t.m.Arguments, "modulestest", append([]string{t.path}, t.additionalTemplates...)...)
		if err != nil {
			got.Fatalf("failed to create module from template %q", t.path)
		}

		mf := &configuration.ServiceManifest{Name: "testing", Arguments: t.args,
			Modules: []*configuration.TemplateRepository{{Name: m.Name}}}
		st := codegen.NewStencil(mf, []*modules.Module{m}, logrus.New())

		for name, ext := range t.exts {
			st.RegisterInprocExtensions(name, ext)
		}

		tpls, err := st.Render(context.Background(), logrus.New())
		if err != nil {
			if t.errStr != "" {
				// if t.errStr was set then we expected an error, since that
				// was set via t.ErrorContains()
				if err == nil {
					got.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t.t, err, t.errStr, "expected render to fail with error containing %q", t.errStr)
			} else {
				got.Fatalf("failed to render: %v", err)
			}
		}

		for _, tpl := range tpls {
			// skip templates that aren't the one we are testing
			if tpl.Path != t.path {
				continue
			}

			for _, f := range tpl.Files {
				// skip the snapshot
				if !t.persist {
					continue
				}

				// Create snapshots with a .snapshot ext to keep them away from linters, see Jira for more details.
				// TODO(jaredallard)[DTSS-2086]: figure out what to do with the snapshot codegen.File directive
				snapshotName := f.Name() + ".snapshot"
				success := got.Run(snapshotName, func(got *testing.T) {
					snapshot := cupaloy.New(cupaloy.ShouldUpdate(func() bool { return save }), cupaloy.CreateNewAutomatically(true))
					snapshot.SnapshotT(got, f)
				})
				if !success {
					got.Fatalf("Generated file %q did not match snapshot", f.Name())
				}
			}

			// only ever process one template
			break
		}
	})
}
