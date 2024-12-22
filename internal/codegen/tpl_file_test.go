package codegen

import (
	"os"
	"testing"

	"go.rgst.io/stencil/pkg/slogext"
	"go.rgst.io/stencil/pkg/stencil"
	"gotest.tools/v3/assert"
)

// TestTplFile_DeleteNoLockfile tests the file.Delete command when there's no lockfile history at all
func TestTplFile_DeleteNoLockfile(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
	}

	fo, err := tplf.Delete()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Deleted)
}

// TestTplFile_DeleteLockfileNoHistory tests the file.Delete command when there's a lockfile with no history of the file
func TestTplFile_DeleteLockfileNoHistory(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{},
		},
	}

	fo, err := tplf.Delete()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Deleted)
	assert.Equal(t, 0, len(tplf.lock.Files))
}

// TestTplFile_DeleteLockfileHistory tests the file.Delete command when there's a lockfile with history of the file
func TestTplFile_DeleteLockfileHistory(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "test.go"},
			},
		},
	}

	fo, err := tplf.Delete()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Deleted)
	assert.Equal(t, 0, len(tplf.lock.Files))
}

// TestTplFile_DeleteLockfileHistoryOfOther tests the file.Delete command when there's a lockfile with history of another file
func TestTplFile_DeleteLockfileHistoryOfOther(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "foo.go"},
			},
		},
	}

	fo, err := tplf.Delete()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Deleted)
	assert.Equal(t, 1, len(tplf.lock.Files))
}

// TestTplFile_OnceFileAlreadyExists tests the file.Once command when the target file already exists
func TestTplFile_OnceFileAlreadyExists(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: "test.go"},
		t:   &Template{},
		log: slogext.NewTestLogger(t),
	}

	assert.NilError(t, os.WriteFile("test.go", []byte("test"), 0o644))
	defer os.Remove("test.go")

	fo, err := tplf.Once()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Skipped)
}

// TestTplFile_OnceNoLockfile tests the file.Once command when there's no lockfile history at all
func TestTplFile_OnceNoLockfile(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
	}

	fo, err := tplf.Once()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, false, tplf.f.Skipped)
}

// TestTplFile_OnceLockNoHistory tests the file.Once command when there's no history of the file existing
func TestTplFile_OnceLockNoHistory(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{},
		},
	}

	fo, err := tplf.Once()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, false, tplf.f.Skipped)
}

// TestTplFile_OnceLockHasHistoryOfOtherFile tests the file.Once command when there's history of another file existing but not this one
func TestTplFile_OnceLockHasHistoryOfOtherFile(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "foo.go"},
			},
		},
	}

	fo, err := tplf.Once()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, false, tplf.f.Skipped)
}

// TestTplFile_OnceLockHasHistoryOfFile tests the file.Once command when there's history of the file existing
func TestTplFile_OnceLockHasHistoryOfFile(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: "test.go"},
		t:   &Template{},
		log: slogext.NewTestLogger(t),
		lock: &stencil.Lockfile{
			Files: []*stencil.LockfileFileEntry{
				{Name: "test.go"},
			},
		},
	}

	fo, err := tplf.Once()
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	assert.Equal(t, true, tplf.f.Skipped)
}
