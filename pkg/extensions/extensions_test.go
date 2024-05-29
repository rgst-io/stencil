package extensions_test

import (
	"context"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/getoutreach/gobox/pkg/cli/updater/resolver"
	"go.rgst.io/stencil/internal/slogext"
	"go.rgst.io/stencil/pkg/extensions"
	"gotest.tools/v3/assert"
)

func TestCanImportNativeExtension(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ext := extensions.NewHost(slogext.NewTestLogger(t))
	defer ext.Close()

	version := &resolver.Version{
		Tag: "v1.3.0",
	}
	err := ext.RegisterExtension(ctx, "https://github.com/getoutreach/stencil-golang", "github.com/getoutreach/stencil-golang", version)
	assert.NilError(t, err, "failed to register extension")

	caller, err := ext.GetExtensionCaller(ctx)
	assert.NilError(t, err, "failed to get extension caller")

	resp, err := caller.Call("github.com/getoutreach/stencil-golang.ParseGoMod", "go.mod", "module test\n\ngo 1.19")
	assert.NilError(t, err, "failed to call extension")

	moduleMap := resp.(map[string]interface{})["Module"].(map[string]interface{})
	spew.Dump(moduleMap)
	assert.Equal(t, moduleMap["Syntax"].(map[string]interface{})["Token"].([]interface{})[1], "test", "failed to parse go.mod")
}
