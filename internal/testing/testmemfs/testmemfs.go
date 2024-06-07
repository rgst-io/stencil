package testmemfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

func WithManifest(manifest string) billy.Filesystem {
	fs := memfs.New()
	f, _ := fs.Create("manifest.yaml")
	f.Write([]byte(manifest))
	f.Close()
	return fs
}
