package testmemfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

func WithManifest(manifest string) (billy.Filesystem, error) {
	fs := memfs.New()
	f, err := fs.Create("manifest.yaml")
	if err != nil {
		return nil, err
	}
	if _, err := f.Write([]byte(manifest)); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return fs, nil
}
