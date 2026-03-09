package stencil

import (
	"fmt"

	"github.com/codeglyph/go-dotignore/v2"
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
	*dotignore.PatternMatcher
}

// LoadIgnore creates a new [Ignore] from the provided path. If path is
// empty, [StencilIgnoreName] is used.
func LoadIgnore(path string) (*Ignore, error) {
	if path == "" {
		path = StencilIgnoreName
	}

	pm, err := dotignore.NewPatternMatcherFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open stencilignore %q: %w", path, err)
	}

	return &Ignore{pm}, nil
}
