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

// Package apiv1 exports the native extension API for Go extensions to
// implement.
package apiv1

import "go.rgst.io/stencil/v2/internal/modules/nativeext/apiv1"

// Contains aliases to make them accessible to consumers.
const (
	// Version is an alias to [apiv1.Version]
	Version = apiv1.Version

	// Name is an alias to [apiv1.Name].
	Name = apiv1.Name

	// CookieKey is an alias to [apiv1.CookieKey].
	CookieKey = apiv1.CookieKey

	// CookieValue is an alias to [apiv1.CookieValue].
	CookieValue = apiv1.CookieValue
)

// Implementation is the interface that must be implemented by a native
// extension.
type Implementation = apiv1.Implementation

// TemplateFunction is a request to create a new template function.
type TemplateFunction = apiv1.TemplateFunction

// TemplateFunctionExec executes a template function
type TemplateFunctionExec = apiv1.TemplateFunctionExec

// Config is configuration returned by an extension to the extension
// host.
type Config = apiv1.Config

// NewExtensionImplementation is an alias to
// [apiv1.NewExtensionImplementation] to allow usage of it by consumers.
var NewExtensionImplementation = apiv1.NewExtensionImplementation
