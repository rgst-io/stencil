package main

import (
	"testing"

	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

func TestModuleTestSmokeTest(t *testing.T) {
	cmd := NewModuleTestCommand(slogext.NewTestLogger(t))

	assert.NilError(t, testRunCommand(t, cmd, "cmd/stencil/testdata/module_test"), "expected command to not fail")
}
