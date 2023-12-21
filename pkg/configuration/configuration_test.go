// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains tests for the configuration pac

package configuration_test

import (
	"fmt"

	"github.com/rgst-io/stencil/pkg/configuration"
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
