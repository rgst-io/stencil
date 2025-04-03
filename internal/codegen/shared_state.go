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

package codegen

import (
	"fmt"
	"path"
	"sort"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/puzpuzpuz/xsync/v4"
)

// hashModuleHookValue hashes the module hook value using the
// hashstructure library. This function panics if hashing fails because
// there is currently no known case where this would fail.
func hashModuleHookValue(m any) uint64 {
	hash, err := hashstructure.Hash(m, hashstructure.FormatV2, nil)
	if err != nil {
		panic(fmt.Errorf("failed to hash module hook value: %w", err))
	}
	return hash
}

// moduleHook is a wrapper type for module hook values that
// contains the values for module hooks
type moduleHook []any

// Sort sorts the module hook values by their hash
func (m moduleHook) Sort() {
	sort.Slice(m, func(i, j int) bool {
		return hashModuleHookValue(m[i]) < hashModuleHookValue(m[j])
	})
}

// global is an explicit type used to define global variables in the sharedData
// type (specifically the globals struct field) so that we can track not only the
// value of the global but also the template it came from.
type global struct {
	// Template is the template that defined this global (and is scoped too)
	Template string

	// Value is the underlying value
	Value any
}

// exportedFunction stores data about a function that has been exported with
// module.Export for later retrieval/usage at module.Call-time.
type exportedFunction struct {
	Template *Template
}

// sharedState stores data that is injected by templates and shared
// between them.
//
// Note: Fields should be exported so that [sharedState.hash] can hash
// them.
type sharedState struct {
	// functions is a map of modules to templates that have been exported
	// through [TplModule].
	Functions   *xsync.Map[string, exportedFunction]
	Globals     *xsync.Map[string, global]
	ModuleHooks *xsync.Map[string, moduleHook]
}

// newSharedState returns an initialized (empty underlying maps)
// sharedData type.
func newSharedState() *sharedState {
	return &sharedState{
		Functions:   xsync.NewMap[string, exportedFunction](),
		ModuleHooks: xsync.NewMap[string, moduleHook](),
		Globals:     xsync.NewMap[string, global](),
	}
}

// hash returns a hash of the current sharedState. This is used to
// determine if the sharedState has changed.
func (s *sharedState) hash() (uint64, error) {
	// Ensure that our slices are sorted so that the hash is consistent.
	for _, v := range s.ModuleHooks.Range {
		v.Sort()
	}

	return hashstructure.Hash(s, hashstructure.FormatV2, nil)
}

// key returns the key name to use for any of the maps on [sharedState].
//
// The module parameter should just be the name of the module. Key
// should be the unique identifier for the value.
func (*sharedState) key(module, key string) string {
	return path.Join(module, key)
}
