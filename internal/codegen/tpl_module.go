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

package codegen

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jaredallard/slogext"
	"go.rgst.io/stencil/v2/internal/modules"
)

// ErrStopProcessingTemplate is an error that can be returned to
// immediately stop processing a template.
//
// Currently, only used in the [ModuleCaller].
var ErrStopProcessingTemplate = errors.New("stop processing template")

// TplModule is a per-template struct for executing functions
// registered onto [ModuleCaller].
type TplModule struct {
	s   *Stencil
	t   *Template
	log slogext.Logger
}

// exportChecks is a map of already exported functions to prevent
// duplicate exports.
var exportChecks = make(map[string]struct{})

// executorScope is used by [TplModule.Export] to determine if the
// function should receive the calling template's scope or use the
// function's template's scope. See [TplModule.Export] for more
// information.
type executorScope string

var (
	// executorScopeCaller is scoped to the caller
	executorScopeCaller executorScope = "caller"

	// executorScopeFunction is scoped to the function
	executorScopeFunction executorScope = "function"
)

// Export registers a function to allow it to be called by other
// templates.
//
// This is only able to be called in library templates and the
// function's name must start with a capital letter. Function names are
// also only eligible to be exported once, if a function is exported
// twice the second call will be a runtime error.
//
// The second argument to "module.Export" controls what "stencil",
// "file" and other template-scoped functions target. Valid options are
// "caller" or "function":
//
//   - caller: stencil, file and other functions target the caller's
//     template.
//   - function (default): stencil, file and other functions target the
//     function's template.
//
// Note: The default of "function" will be changed to "caller" in v3.
//
// Example:
//
//	{{- define "HelloWorld" }}
//	{{- return (printf "Hello, %s!" .Data) }}
//	{{- end }}
//
//	{{ module.Export "HelloWorld" }}
//
// Or:
//
//	{{ module.Export "HelloWorld" "caller" }}
func (tm *TplModule) Export(name string, scopeSli ...executorScope) (string, error) {
	// We only allow functions to be exported before the final pass.
	if tm.s.renderStage == renderStageFinal {
		// In the final pass, though, check to make sure there's not dupes
		checkName := fmt.Sprintf("%s.%s", tm.t.Module.Name, name)
		if _, ok := exportChecks[checkName]; ok {
			return "", fmt.Errorf("function %q in module %q was already exported", name, tm.t.Module.Name)
		}
		exportChecks[checkName] = struct{}{}

		return "", nil
	}

	var scope executorScope
	switch len(scopeSli) {
	case 1:
		scope = scopeSli[0]
	case 0:
		tm.log.With(
			"module", tm.t.Module.Name,
			"template", tm.t.Path,
		).Warn("module.Export scope default will change to 'caller' in v3")
		scope = executorScopeFunction
	default:
		return "", fmt.Errorf("got %d arguments, expected max 2", len(scopeSli))
	}

	switch scope {
	case executorScopeCaller, executorScopeFunction:
	default:
		return "", fmt.Errorf(
			"unknown value for scope: %s (expected %s or %s)", scope,
			executorScopeCaller, executorScopeFunction,
		)
	}

	if !tm.t.Library {
		return "", fmt.Errorf("only library templates can export functions")
	}

	if !strings.HasPrefix(name, strings.ToUpper(name[:1])) {
		return "", fmt.Errorf("exported function names must start with a capital letter")
	}

	moduleName := tm.t.Module.Name
	key := tm.s.sharedState.key(moduleName, name)

	ef := exportedFunction{Template: tm.t, Scope: scope}
	tm.s.sharedState.Functions.Store(key, ef)
	tm.log.Debug("Exported function", "module.name", moduleName, "function.name", name)

	return "", nil
}

// Call executes a template function by name with the provided
// arguments. The function must have been exported by the module that
// provides it through the [Export] function.
//
// Attempting to call a function that does not exist will return an
// error outside of the first pass where it will return `nil` instead.
//
// The template that is called must return a value using the `return`
// template function, which is only available in this context.
//
// In addition, all of the file, stencil and other functions are in the
// context of the owning template, not the template calling the function.
//
// `.` in a template function acts the same way as it does for
// [TplStencil.Include] (`stencil.Include`). Meaning, it points to
// [Values]. The caller passed data is accessible on `.Data`.
//
// Example:
//
//	// module-a
//	{{- define "HelloWorld" }}
//	{{- return (printf "Hello, %s!" .Data) }}
//	{{- end }}
//	{{ module.Export "HelloWorld" }}
//
//	// module-b
//	{{ module.Call "github.com/rgst-io/module-a.HelloWorld" "Jared" }}
//	// Output: Hello, Jared
//
// **WARNING**: Functions cannot call other functions due to a race
// condition in how templates are rendered, this will be fixed in v3 of
// stencil.
func (tm *TplModule) Call(name string, args ...any) (any, error) {
	// Allows args to not be set.
	if len(args) > 1 {
		return nil, fmt.Errorf("Call() only takes max two arguments, name and data")
	}

	// We don't allow calling functions during the pre-render stage
	// because we don't know what functions are available yet.
	if tm.s.renderStage == renderStagePre {
		return nil, nil
	}

	// Currently the export/call system has race conditions, so we do not
	// allow calling functions in templates. Eventually, this will be
	// fixed and allowed.
	if tm.t.Library {
		return "", fmt.Errorf("library templates cannot call functions")
	}

	// Get the module name and function name by splitting the name by the
	// last period.
	var moduleName, functionName string
	lastPeriodIdx := strings.LastIndex(name, ".")
	if lastPeriodIdx == -1 {
		moduleName = tm.t.Module.Name
		functionName = name
	} else {
		moduleName, functionName = name[:lastPeriodIdx], name[lastPeriodIdx+1:]
	}

	key := tm.s.sharedState.key(moduleName, functionName)
	ef, ok := tm.s.sharedState.Functions.Load(key)
	if !ok {
		return nil, fmt.Errorf("function %q in module %q was not exported", functionName, moduleName)
	}

	// Get the module that we requested. If the module name is the same as
	// the current one, we do not need to look it up since it's attached
	// to the template render context.
	var module *modules.Module
	if moduleName == tm.t.Module.Name {
		module = tm.t.Module
	} else {
		for _, m := range tm.s.modules {
			if m.Name == moduleName {
				module = m
				break
			}
		}
		if module == nil {
			return nil, fmt.Errorf("module %s was not found on stencil (this is a possible bug)", moduleName)
		}
	}

	// Create a copy of the current values so we can set the data on it
	// without mutating the original values.
	d := tm.t.args.Copy()
	if len(args) > 0 {
		d.Data = args[0]
	}

	var returnValsMu sync.Mutex
	var returnVal any
	var errVal TplError

	// We clone the template to allow us to reset what values are
	// accessible to the function we're going to call.
	tmpTpl, err := module.GetTemplate().Clone()
	if err != nil {
		return nil, err
	}

	tplScope := ef.Template
	if ef.Scope == executorScopeCaller {
		tplScope = tm.t
	}

	tmpTpl.Funcs(NewFuncMap(tm.s, tplScope, tm.log))
	tmpTpl.Funcs(map[string]any{
		// return captures a value from the executed function template and
		// returns it to the calling template.
		//
		// If an error is returned, it will be plumbed back up to Call
		// instead of being handled at the called template level.
		"return": func(v any, err ...TplError) (string, error) {
			returnValsMu.Lock()
			defer returnValsMu.Unlock()

			if len(err) > 1 {
				return "", fmt.Errorf("return() only takes one error argument")
			}

			if len(err) > 0 {
				errVal = err[0]
				return "", ErrStopProcessingTemplate
			}

			returnVal = v
			return "", nil
		},
	})

	if err := tmpTpl.ExecuteTemplate(io.Discard, functionName, d); err != nil && !errors.Is(err, ErrStopProcessingTemplate) {
		return nil, err
	}

	return returnVal, errVal.err
}
