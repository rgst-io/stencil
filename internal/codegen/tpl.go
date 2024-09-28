// Copyright (C) 2024 stencil contributors
// Copyright (C) 2022-2023 Outreach Corporation
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

// Description: This file contains helpers for creating
// functions exposed to stencil codegen files.

package codegen

import (
	"fmt"
	"text/template"

	"go.rgst.io/stencil/internal/modules/nativeext"
	"go.rgst.io/stencil/pkg/slogext"
)

// NewFuncMap returns the standard func map for a template
func NewFuncMap(st *Stencil, t *Template, log slogext.Logger) template.FuncMap {
	// At first look it might be confusing why we allow these to be nil,
	// this is because when we call Parse() on a template, these values
	// will be nil but not actually used until Execute() is called.
	var tplst *TplStencil
	var tplf *TplFile
	var tplm *TplModule
	if st != nil {
		tplst = &TplStencil{st, t, log}
	}
	if t != nil && len(t.Files) > 0 {
		tplf = &TplFile{t.Files[0], t, st.lock, log}
	}
	if t != nil && st != nil {
		tplm = &TplModule{st, t, log}
	}

	// build the function map
	funcs := Default
	funcs["stencil"] = func() *TplStencil { return tplst }
	funcs["file"] = func() *TplFile {
		if tplf == nil {
			panic(fmt.Errorf("attempted to use file in a template that doesn't support file rendering"))
		}
		return tplf
	}
	funcs["extensions"] = func() *nativeext.ExtensionCaller { return st.extCaller }
	funcs["module"] = func() *TplModule { return tplm }

	// Only valid in the "module" context. This is overwritten in the
	// [TplModule.Call].
	funcs["return"] = func() (string, error) {
		return "", fmt.Errorf("'return' can only be called during a module.Call")
	}

	return funcs
}
