package main

import (
	"bytes"
	"testing"

	"go.rgst.io/jaredallard/slogext/v2"
	"go.rgst.io/stencil/v2/internal/testing/stdouttest"
	"gotest.tools/v3/assert"
)

func TestModuleTestSmokeTest(t *testing.T) {
	cmd := NewModuleTestCommand(slogext.NewTestLogger(t))

	assert.NilError(t, testRunCommand(t, cmd, "cmd/stencil/testdata/module_test"), "expected command to not fail")
}

func TestModuleTestSmokeTestFailure(t *testing.T) {
	cmd := NewModuleTestCommand(slogext.NewTestLogger(t))

	assert.Error(t, testRunCommand(t, cmd, "cmd/stencil/testdata/module_test_failure"), "tests failed", "expected command to fail")
}

func TestModuleTestFailureShowsOutput(t *testing.T) {
	// Some output is written directly to stdout, so we need to capture it
	// here.
	out := stdouttest.Run(t, func() {
		cmd := NewModuleTestCommand(slogext.NewTestLogger(t))

		err := testRunCommand(t, cmd, "cmd/stencil/testdata/module_test_failure")
		assert.Error(t, err, "tests failed", "expected command to fail")
	})
	assert.Assert(t, bytes.Contains(out, []byte("failed!")), "expected output to be included")
}
