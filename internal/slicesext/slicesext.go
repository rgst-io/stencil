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

// Package slicesext contains helpers for interacting with slices
package slicesext

// Map creates a map from a given slice. The key is determined
// from the result of the key function being ran on the given value
// type.
func Map[K comparable, V any](src []V, key func(V) K) map[K]V {
	result := make(map[K]V)
	for _, v := range src {
		result[key(v)] = v
	}
	return result
}

// FromMap collects the values from a map into a slice.
func FromMap[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}
