package codegen

import (
	"bytes"
	"os"
	"path"
	"testing"

	"go.rgst.io/stencil/v2/pkg/slogext"
	"go.rgst.io/stencil/v2/pkg/stencil"
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

func TestTplFile_MigrateToSrcFileExistsNoDestFile(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: path.Join(t.TempDir(), "test.go")},
		log: slogext.NewTestLogger(t),
	}

	// Set up the initial state
	contents := []byte("test")
	assert.NilError(t, os.WriteFile(tplf.f.path, contents, 0o644))

	newPath := path.Join(t.TempDir(), "testnew.go")
	os.Remove(newPath)

	fo, err := tplf.MigrateTo(newPath)
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	// Deleted should be true but file should still exist (deleted is processed later)
	assert.Equal(t, true, tplf.f.Deleted)
	_, err = os.Stat(tplf.f.path)
	assert.NilError(t, err)

	_, err = os.Stat(newPath)
	assert.NilError(t, err)

	contentsNew, err := os.ReadFile(newPath)
	assert.NilError(t, err)
	assert.Equal(t, string(contentsNew), string(contents))
}

func TestTplFile_MigrateToSrcFileExistsDestFileExists(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: path.Join(t.TempDir(), "test.go")},
		log: slogext.NewTestLogger(t),
	}

	// Set up the initial state
	contents := []byte("test")
	assert.NilError(t, os.WriteFile(tplf.f.path, contents, 0o644))

	newPath := path.Join(t.TempDir(), "testnew.go")
	contentsNew := []byte("testnew")
	assert.NilError(t, os.WriteFile(newPath, contentsNew, 0o644))

	fo, err := tplf.MigrateTo(newPath)
	assert.NilError(t, err)
	assert.Equal(t, "", fo)
	// Deleted should be true but file should still exist (deleted is processed later)
	assert.Equal(t, true, tplf.f.Deleted)
	_, err = os.Stat(tplf.f.path)
	assert.NilError(t, err)

	_, err = os.Stat(newPath)
	assert.NilError(t, err)

	contentsNewNew, err := os.ReadFile(newPath)
	assert.NilError(t, err)
	assert.Equal(t, string(contentsNewNew), string(contents))
}

func TestTplFile_MigrateToSrcFileNoExists(t *testing.T) {
	tplf := TplFile{
		f:   &File{path: path.Join(t.TempDir(), "test.go")},
		t:   &Template{},
		log: slogext.NewTestLogger(t),
	}

	// Set up the initial state
	os.Remove(tplf.f.path)

	newPath := path.Join(t.TempDir(), "testnew.go")

	fo, err := tplf.MigrateTo(newPath)
	assert.NilError(t, err)
	assert.Equal(t, "", fo)

	assert.Equal(t, false, tplf.f.Deleted)
	assert.Equal(t, true, tplf.f.Skipped)

	_, err = os.Stat(tplf.f.path)
	assert.ErrorContains(t, err, "no such file")
	_, err = os.Stat(newPath)
	assert.ErrorContains(t, err, "no such file")
}

func TestTplFile_RemoveAll(t *testing.T) {
	tplf := TplFile{
		f: &File{path: "test.go"},
	}

	wd, err := os.Getwd()
	assert.NilError(t, err, "failed to get working directory")
	td := os.TempDir()
	err = os.Chdir(td)
	assert.NilError(t, err, "failed to change working directory")
	defer os.Chdir(wd)

	os.MkdirAll("test", 0o755)
	os.WriteFile("test/test.go", []byte("test"), 0o644)
	os.WriteFile("test/test2.go", []byte("test2"), 0o644)
	_, err = os.Stat("test/test.go")
	assert.NilError(t, err)
	_, err = os.Stat("test/test2.go")
	assert.NilError(t, err)

	fo, err := tplf.RemoveAll("test")
	assert.NilError(t, err)
	assert.Equal(t, "", fo)

	_, err = os.Stat("test")
	assert.ErrorContains(t, err, "no such file")
	_, err = os.Stat("test/test.go")
	assert.ErrorContains(t, err, "no such file")
	_, err = os.Stat("test/test2.go")
	assert.ErrorContains(t, err, "no such file")
}

func TestTplFile_SetMode(t *testing.T) {
	tpl := RenderTemplate(t, nil, nil, `{{- file.SetMode 0o777 }}`)
	assert.Equal(t, tpl.Files[0].Mode(), os.FileMode(0o777), "expected file mode to be 0o777")
}

func TestTplFile_SetPath(t *testing.T) {
	tpl := RenderTemplate(t, nil, nil, `{{- file.SetPath "xyz" }}`)
	assert.Equal(t, tpl.Files[0].path, "xyz", "expected file path to be xyz")
}

func TestTplFile_SetContents(t *testing.T) {
	tpl := RenderTemplate(t, nil, nil, `{{- file.SetContents "xyz" }}`)
	assert.Assert(t, bytes.Equal(tpl.Files[0].contents, []byte("xyz")), "expected file contents to be xyz")
}

func TestTplFile_Static(t *testing.T) {
	tpl := RenderTemplate(t, nil, nil, `{{- file.Static }}`)
	assert.Equal(t, tpl.Files[0].Skipped, false, "expected file to not be skipped when doesn't exist")
}

func TestTplFile_MigrateFrom_ShouldPersistBlocks(t *testing.T) {
	qts := []quickTemplate{
		{
			Filename:             "old.go",
			ExistingFileContents: "## <<Stencil::Block(custom)>>\nold\n## <</Stencil::Block>>",
		},
		{
			Filename: "new.go",
			TemplateContents: `{{ file.MigrateFrom "old.go" -}}` +
				"## <<Stencil::Block(custom)>>\n{{ file.Block \"custom\" }}\n## <</Stencil::Block>>",
		},
	}

	tpls := RenderTemplates(t, nil, nil, qts...)

	assert.Equal(t, len(tpls), 1, "expected 1 templates to be rendered")
	assert.Equal(t, len(tpls[0].Files), 2, "expected 2 files to be created")

	for _, f := range tpls[0].Files {
		switch f.Name() {
		case "old.go":
			assert.Equal(t, f.Deleted, true, "expected old file (old.go) to be deleted")
		case "new.go":
			assert.Equal(t, f.String(), "## <<Stencil::Block(custom)>>\nold\n## <</Stencil::Block>>",
				"expected old to be migrated to new with old contents")
		default:
			t.Errorf("expected old.go & new.go, got: %s", f.Name())
		}
	}
}

func TestTplFile_MigrateFrom_ShouldUseNewOverOldBlocks(t *testing.T) {
	qts := []quickTemplate{
		{
			Filename:             "old.go",
			ExistingFileContents: "## <<Stencil::Block(custom)>>\nold\n## <</Stencil::Block>>",
		},
		{
			Filename:             "new.go",
			ExistingFileContents: "## <<Stencil::Block(custom)>>\nnew\n## <</Stencil::Block>>",
		},
		{
			Filename: "new.go",
			TemplateContents: `{{ file.MigrateFrom "old.go" -}}` +
				"## <<Stencil::Block(custom)>>\n{{ file.Block \"custom\" }}\n## <</Stencil::Block>>",
		},
	}

	tpls := RenderTemplates(t, nil, nil, qts...)

	assert.Equal(t, len(tpls), 1, "expected 1 templates to be rendered")
	assert.Equal(t, len(tpls[0].Files), 2, "expected 2 files to be created")

	for _, f := range tpls[0].Files {
		switch f.Name() {
		case "old.go":
			assert.Equal(t, f.Deleted, true, "expected old file (old.go) to be deleted")
		case "new.go":
			assert.Equal(t, f.String(), "## <<Stencil::Block(custom)>>\nnew\n## <</Stencil::Block>>",
				"expected new file blocks to be used over old")
		default:
			t.Errorf("expected old.go & new.go, got: %s", f.Name())
		}
	}
}

func TestTplFile_MigrateFrom_ShouldNotDeleteOldIfNotExist(t *testing.T) {
	qts := []quickTemplate{
		{
			Filename:             "new.go",
			ExistingFileContents: "## <<Stencil::Block(custom)>>\nnew\n## <</Stencil::Block>>",
		},
		{
			Filename: "new.go",
			TemplateContents: `{{ file.MigrateFrom "old.go" -}}` +
				"## <<Stencil::Block(custom)>>\n{{ file.Block \"custom\" }}\n## <</Stencil::Block>>",
		},
	}

	tpls := RenderTemplates(t, nil, nil, qts...)

	assert.Equal(t, len(tpls), 1, "expected 1 templates to be rendered")
	assert.Equal(t, len(tpls[0].Files), 1, "expected 2 files to be created")

	for _, f := range tpls[0].Files {
		switch f.Name() {
		case "new.go":
			assert.Equal(t, f.String(), "## <<Stencil::Block(custom)>>\nnew\n## <</Stencil::Block>>",
				"expected new file blocks to be used over old")
		default:
			t.Errorf("expected new.go, got: %s", f.Name())
		}
	}
}

func TestTplFile_MigrateFrom_ShouldNoopIfNewAndOldNoExist(t *testing.T) {
	qts := []quickTemplate{
		{
			Filename: "new.go",
			TemplateContents: `{{ file.MigrateFrom "old.go" -}}` +
				"## <<Stencil::Block(custom)>>\n{{ file.Block \"custom\" }}\n## <</Stencil::Block>>",
		},
	}

	tpls := RenderTemplates(t, nil, nil, qts...)

	assert.Equal(t, len(tpls), 1, "expected 1 templates to be rendered")
	assert.Equal(t, len(tpls[0].Files), 1, "expected 2 files to be created")

	for _, f := range tpls[0].Files {
		switch f.Name() {
		case "new.go":
			assert.Equal(t, f.String(), "## <<Stencil::Block(custom)>>\n\n## <</Stencil::Block>>",
				"expected new file blocks to be used over old")
		default:
			t.Errorf("expected new.go, got: %s", f.Name())
		}
	}
}
