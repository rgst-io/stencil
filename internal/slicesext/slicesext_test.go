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

import (
	"reflect"
	"slices"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	type args struct {
		src []int
		key func(int) string
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{
			name: "converts slice into map",
			args: args{
				src: []int{0, 1, 2, 3},
				key: strconv.Itoa,
			},
			want: map[string]int{"0": 0, "1": 1, "2": 2, "3": 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Map(tt.args.src, tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestFromMap(t *testing.T) {
	type args struct {
		m map[int]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "converts map into slice",
			args: args{m: map[int]string{1: "1", 2: "2", 3: "3"}},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromMap(tt.args.m)

			// Special case for tests. We can't sort within [FromMap] because
			// we're not guaranteed to be [cmp.Ordered].
			slices.Sort(got)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
