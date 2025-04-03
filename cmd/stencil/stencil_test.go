package main

import (
	"testing"

	"go.rgst.io/stencil/v2/pkg/slogext"
	"gotest.tools/v3/assert"
)

func TestFailsWhenUnknownArgumentsArePassed(t *testing.T) {
	cmd := NewStencil(slogext.NewTestLogger(t))
	assert.Assert(t, cmd != nil, "expected NewStencil() to not return nil")

	err := testRunCommand(t, cmd, "", "im-not-a-command")
	assert.ErrorContains(t, err, "unexpected arguments: [im-not-a-command]")
}
