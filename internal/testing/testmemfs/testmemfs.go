// Copyright (C) 2024 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testmemfs is a helper for tests that rely on the filesystem.
package testmemfs

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
)

// WithManifest creates a new in-memory filesystem with a manifest.yaml
// in the root.
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
