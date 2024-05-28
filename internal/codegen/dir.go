package codegen

import (
	"io"
	"path/filepath"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/sirupsen/logrus"
	"go.rgst.io/stencil/internal/modules"
	"sigs.k8s.io/yaml"
)

// rawDirManifest is a helper to unmarshal data from a directory manifest yaml
type rawDirManifest struct {
	RenameTemplate string `yaml:"renameTemplate"`
}

// DirManifest encompasses data from optional directory-based manifests
type DirManifest struct {
	OriginalPathBase string

	Template *Template

	ReplaceName string
}

// LoadDirManifest unmarshals a directory manifest from a directory
func LoadDirManifest(fs billy.Filesystem, path string, mod *modules.Module, log logrus.FieldLogger) (*DirManifest, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conts, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var rm rawDirManifest
	if err := yaml.Unmarshal(conts, &rm); err != nil {
		return nil, err
	}

	rt, err := NewTemplate(mod, path, 0o000, time.Time{}, []byte(rm.RenameTemplate), log)
	if err != nil {
		return nil, err
	}

	dm := DirManifest{
		OriginalPathBase: filepath.Dir(path),
		Template:         rt,
	}

	return &dm, nil
}

func (m *DirManifest) Render(st *Stencil, vals *Values) error {
	if err := m.Template.Render(st, vals, nil); err != nil {
		return err
	}

	m.ReplaceName = m.Template.Files[0].String()
	return nil
}
