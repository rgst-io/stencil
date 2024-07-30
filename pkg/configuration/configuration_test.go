package configuration_test

import (
	"fmt"
	"os"
	"testing"

	"go.rgst.io/stencil/pkg/configuration"
	"gotest.tools/v3/assert"
)

func ExampleValidateName() {
	// Normal name
	success := configuration.ValidateName("test")
	fmt.Println("success:", success)

	// Invalid name
	success = configuration.ValidateName("test.1234")
	fmt.Println("success:", success)

	// Output:
	// success: true
	// success: false
}

func ExampleNewManifest() {
	sm, err := configuration.NewManifest("testdata/stencil.yaml")
	if err != nil {
		// handle the error
		fmt.Println("error:", err)
		return
	}

	fmt.Println(sm.Name)
	fmt.Println(sm.Arguments)

	// Output:
	// testing
	// map[hello:world]
}

func TestShouldSupportServiceYaml(t *testing.T) {
	// Change into test directory
	curDir, err := os.Getwd()
	assert.NilError(t, err)
	defer os.Chdir(curDir)
	assert.NilError(t, os.Chdir("testdata/interop/service-if-not-found"))

	sm, err := configuration.NewDefaultManifest()
	assert.NilError(t, err)

	assert.Equal(t, sm.Name, "service")
}

func TestShouldUseStencilOverServiceYaml(t *testing.T) {
	// Change into test directory
	curDir, err := os.Getwd()
	assert.NilError(t, err)
	defer os.Chdir(curDir)
	assert.NilError(t, os.Chdir("testdata/interop/stencil-over-service"))

	sm, err := configuration.NewDefaultManifest()
	assert.NilError(t, err)

	assert.Equal(t, sm.Name, "stencil")
}
