package stencil

import (
	"path/filepath"

	gitignore "github.com/denormal/go-gitignore"
)

const (
	// StencilIgnoreName is the default name of the file used for [Ignore].
	StencilIgnoreName = ".stencilignore"
)

// Ignore represents a .stencilignore file. This file is semantically
// similar to a .gitignore, except it configures stencil to ignore
// certain files from being modified post-render.
//
// This mode is primarily intended for temporary deviations of files
// from their corresponding templates. Stencil itself will warn when
// files are ignored and, optional exit with a non-zero exit code.
type Ignore struct {
	gitignore.GitIgnore
}

// LoadIgnore creates a new [Ignore].
func LoadIgnore(path string) (*Ignore, error) {
	gi, err := gitignore.NewFromFile(filepath.Join(path, StencilIgnoreName))
	if err != nil {
		return nil, err
	}
	return &Ignore{gi}, nil
}
