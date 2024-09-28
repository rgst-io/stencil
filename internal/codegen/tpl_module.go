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
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.rgst.io/stencil/internal/modules"
	"go.rgst.io/stencil/pkg/slogext"
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

// Export registers a function to allow it to be called by other
// templates.
//
// This is only able to be called in library templates and the
// function's name must start with a capital letter. Function names are
// also only eligible to be exported once, if a function is exported
// twice the second call will be a runtime error.
//
// Example:
//
//	{{- define "HelloWorld" }}
//	{{- return (printf "Hello, %s!" .Data) }}
//	{{- end }}
//
//	{{ module.Export "HelloWorld" }}
func (tm *TplModule) Export(name string) (string, error) {
	// We only allow functions to be exported in the first pass.
	if tm.s.renderStage == renderStageFinal {
		return "", nil
	}

	if !tm.t.Library {
		return "", fmt.Errorf("only library templates can export functions")
	}

	if !strings.HasPrefix(name, strings.ToUpper(name[:1])) {
		return "", fmt.Errorf("exported function names must start with a capital letter")
	}

	moduleName := tm.t.Module.Name
	key := tm.s.sharedState.key(moduleName, name)
	if _, ok := tm.s.sharedState.Functions.Load(key); ok {
		return "", fmt.Errorf("function %s in module %s was already exported", name, moduleName)
	}

	tm.s.sharedState.Functions.Store(key, struct{}{})
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
// context of the parent template, not the template being called.
//
// `.` in a template function acts the same way as it does for
// [TplStencil.ApplyTemplate] (`stencil.ApplyTemplate`). Meaning, it
// points to [Values]. The caller passed data is accessible on `.Data`.
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
//	{{ module.Call "module-b.HelloWorld" "Jared" }}
//	// Output: Hello, Jared
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

	// Get the module name and function name by splitting the name by the
	// last period.
	lastPeriodIdx := strings.LastIndex(name, ".")
	if lastPeriodIdx == -1 {
		return nil, fmt.Errorf("expected format module.function, got %q", name)
	}
	moduleName, functionName := name[:lastPeriodIdx], name[lastPeriodIdx+1:]

	key := tm.s.sharedState.key(moduleName, functionName)
	if _, ok := tm.s.sharedState.Functions.Load(key); !ok {
		return nil, fmt.Errorf("function %q in module %q was not registered", functionName, moduleName)
	}

	// Find the module's template that we requested.
	var module *modules.Module
	for _, m := range tm.s.modules {
		if m.Name == moduleName {
			module = m
			break
		}
	}
	if module == nil {
		return nil, fmt.Errorf("module %s was not found on stencil (this is a possible bug)", moduleName)
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

	tmpTpl.Funcs(NewFuncMap(tm.s, tm.t, tm.log))
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
