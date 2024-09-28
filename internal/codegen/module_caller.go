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
	"sync"
)

// ModuleCaller is a struct that interacts with registering and calling
// functions across templates.
type ModuleCaller struct {
	// mu protects the functions map
	mu sync.RWMutex

	// functions is a map of all functions that can be called. The format
	// is:
	//   "module-name":
	//     "function-name": "template-name"
	functions map[string]map[string]string
}

// NewModuleCaller creates a new [ModuleCaller].
func NewModuleCaller() *ModuleCaller {
	return &ModuleCaller{
		functions: make(map[string]map[string]string),
	}
}
