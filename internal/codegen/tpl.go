// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains helpers for creating
// functions exposed to stencil codegen files.

package codegen

import (
	"fmt"
	"text/template"

	"go.rgst.io/stencil/internal/modules/nativeext"
	"go.rgst.io/stencil/internal/slogext"
)

// NewFuncMap returns the standard func map for a template
func NewFuncMap(st *Stencil, t *Template, log slogext.Logger) template.FuncMap {
	// We allow tplst & tplf to be nil in the case of
	// .Parse() of a template, where they need to be present
	// but aren't actually executed by the template
	// (execute is the one that renders it)
	var tplst *TplStencil
	var tplf *TplFile
	if st != nil {
		tplst = &TplStencil{st, t, log}
	}
	if t != nil && len(t.Files) > 0 {
		tplf = &TplFile{t.Files[0], t, log}
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
	return funcs
}
