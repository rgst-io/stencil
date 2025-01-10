// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package dotnotation

import "strings"

// PrefixNode is a node in a [PrefixTree]. The only reason this isn't
// merged with the [PrefixTree] type is to reduce memory usage by only
// setting adjacent when children have been added.
type PrefixNode struct {
	// adjacent stores the adjacent nodes.
	//
	// Because a tree can `n` children and is a string, we use a map to make
	// it easier to look up children for duplicate node handling.
	adjacent map[string]*PrefixNode
}

// NewPrefixTree returns a new [PrefixNode] with an empty root.
func NewPrefixTree() *PrefixNode {
	return &PrefixNode{}
}

// Add adds a string to the tree does nothing if the prefix already
// exists.
func (pn *PrefixNode) Add(str string) {
	spl := strings.Split(strings.TrimPrefix(str, "."), ".")

	last := pn
	for _, k := range spl {
		if last.adjacent == nil {
			last.adjacent = make(map[string]*PrefixNode)
		}

		// Node already exists, continue
		if _, ok := last.adjacent[k]; ok {
			last = last.adjacent[k]
			continue
		}

		// Otherwise, create another node.
		last.adjacent[k] = &PrefixNode{}
		last = last.adjacent[k]
	}
}

// Has returns true if this prefix is known, this includes partial
// matches.
func (pn *PrefixNode) Has(str string) bool {
	spl := strings.Split(strings.TrimPrefix(str, "."), ".")

	last := pn
	for _, k := range spl {
		// No children, can't match.
		if last.adjacent == nil {
			return false
		}

		// No match
		if _, ok := last.adjacent[k]; !ok {
			return false
		}

		last = last.adjacent[k]
	}

	// Nothing left to look for, must've matched.
	return true
}
