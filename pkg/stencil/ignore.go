// Copyright (C) 2026 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package stencil

import (
	"fmt"
	"io"
	"os"

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

// LoadIgnoreFromReader parses a stencilignore from the given reader.
func LoadIgnoreFromReader(r io.Reader) (*Ignore, error) {
	pm, err := dotignore.NewPatternMatcherFromReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse stencilignore: %w", err)
	}

	return &Ignore{pm}, nil
}

// LoadIgnore creates a new [Ignore] from the provided path. If path is
// empty, [StencilIgnoreName] is used.
func LoadIgnore(path string) (*Ignore, error) {
	if path == "" {
		path = StencilIgnoreName
	}

	//nolint:gosec // Why: By design.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ignore, err := LoadIgnoreFromReader(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load %q: %w", path, err)
	}
	return ignore, nil
}
