package main

import (
	"testing"

	"go.rgst.io/stencil/pkg/slogext"
	"gotest.tools/v3/assert"
)

func TestFailsWhenUnknownArgumentsArePassed(t *testing.T) {
	app := NewStencil(slogext.NewTestLogger(t))
	assert.Assert(t, app != nil, "expected NewStencil() to not return nil")

	err := testRunApp(t, "", app, "im-not-a-command")
	assert.ErrorContains(t, err, "unexpected arguments: [im-not-a-command]")
}
