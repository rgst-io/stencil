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

// TestLockfilePruneFilesEmpty tests lockfile's prunefiles function for an empty file list
func TestLockfilePruneFilesEmpty(t *testing.T) {
	l, err := stencil.LoadLockfile("testdata")
	assert.NilError(t, err)

	ret := l.PruneFiles([]string{})
	assert.Equal(t, 0, len(ret))
}

// TestLockfilePruneFilesMissingFile tests lockfile's prunefiles function for a missing file
func TestLockfilePruneFilesMissingFile(t *testing.T) {
	l := stencil.Lockfile{Files: []*stencil.LockfileFileEntry{
		{Name: "foo.bar"},
	}}

	ret := l.PruneFiles([]string{})
	assert.DeepEqual(t, []string{"foo.bar"}, ret)

	l.Sort()
	assert.Equal(t, 0, len(l.Files))
}

// TestLockfilePruneFilesMissingFileSpecified tests lockfile's prunefiles function for a missing file that's specified
func TestLockfilePruneFilesMissingFileSpecified(t *testing.T) {
	l := stencil.Lockfile{Files: []*stencil.LockfileFileEntry{
		{Name: "foo.bar"},
	}}

	ret := l.PruneFiles([]string{"foo.bar"})
	assert.DeepEqual(t, []string{"foo.bar"}, ret)

	l.Sort()
	assert.Equal(t, 0, len(l.Files))
}

// TestLockfilePruneFilesMissingFileSpecifiedWrong tests lockfile's prunefiles function for a missing file that was specified wrong
func TestLockfilePruneFilesMissingFileSpecifiedWrong(t *testing.T) {
	l := stencil.Lockfile{Files: []*stencil.LockfileFileEntry{
		{Name: "foo.bar"},
	}}

	ret := l.PruneFiles([]string{"foo.barrio"})
	assert.DeepEqual(t, []string{}, ret)

	l.Sort()
	assert.Equal(t, 1, len(l.Files))
}

// TestLockfilPruneModulesEmpty tests lockfile's prunemodules function for an empty module list
func TestLockfilPruneModulesEmpty(t *testing.T) {
	l, err := stencil.LoadLockfile("testdata")
	assert.NilError(t, err)

	ret := l.PruneModules([]string{}, []string{})
	assert.Equal(t, 0, len(ret))
}

// TestLockfilePruneModulesMissingModule tests lockfile's prunemodules function for a missing module
func TestLockfilePruneModulesMissingModule(t *testing.T) {
	l := stencil.Lockfile{Modules: []*stencil.LockfileModuleEntry{
		{Name: "foo"},
	}}

	ret := l.PruneModules([]string{}, []string{})
	assert.DeepEqual(t, []string{"foo"}, ret)

	l.Sort()
	assert.Equal(t, 0, len(l.Modules))
}

// TestLockfilePruneModulesMissingModuleSpecified tests lockfile's prunemodules function for a missing module that's specified
func TestLockfilePruneModulesMissingModuleSpecified(t *testing.T) {
	l := stencil.Lockfile{Modules: []*stencil.LockfileModuleEntry{
		{Name: "foo"},
	}}

	ret := l.PruneModules([]string{}, []string{"foo"})
	assert.DeepEqual(t, []string{"foo"}, ret)

	l.Sort()
	assert.Equal(t, 0, len(l.Modules))
}

// TestLockfilePruneModulesMissingModuleSpecifiedWrong tests lockfile's prunemodules function for a missing module that was specified wrong
func TestLockfilePruneModulesMissingModuleSpecifiedWrong(t *testing.T) {
	l := stencil.Lockfile{Modules: []*stencil.LockfileModuleEntry{
		{Name: "foo.bar"},
	}}

	ret := l.PruneModules([]string{}, []string{"foo.barrio"})
	assert.DeepEqual(t, []string{}, ret)

	l.Sort()
	assert.Equal(t, 1, len(l.Modules))
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
