// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: Contains tests for the stencil package

package stencil_test

import (
	"fmt"
	"testing"

	"go.rgst.io/stencil/pkg/stencil"
	"gotest.tools/v3/assert"
)

func ExampleLoadLockfile() {
	// Load the lockfile
	l, err := stencil.LoadLockfile("testdata")
	if err != nil {
		// handle the error
		fmt.Println(err)
		return
	}

	fmt.Println(l.Version)

	// Output:
	// v1.6.2
}

func TestLockfilePruneEmpty(t *testing.T) {
	l, err := stencil.LoadLockfile("testdata")
	assert.NilError(t, err)

	ret := l.Prune()
	assert.Equal(t, 0, len(ret))
}

func TestLockfilePruneMissingFile(t *testing.T) {
	l := stencil.Lockfile{Files: []*stencil.LockfileFileEntry{
		{Name: "foo.bar"},
	}}

	ret := l.Prune()
	assert.DeepEqual(t, []string{"foo.bar"}, ret)

	l.Sort()
	assert.Equal(t, 0, len(l.Files))
}

func TestLockfileSortList(t *testing.T) {
	l := stencil.Lockfile{Files: []*stencil.LockfileFileEntry{
		{Name: "foo.bar"},
		{Name: "bar.foo"},
	}}

	l.Sort()
	assert.Equal(t, "bar.foo", l.Files[0].Name)
	assert.Equal(t, "foo.bar", l.Files[1].Name)
}
