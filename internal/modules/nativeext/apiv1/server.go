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

// Description: Implements a plugin Implementation
// for the extensions host.

package apiv1

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/jaredallard/slogext"
)

// NewHandshake returns a plugin.HandshakeConfig for
// this extension api version.
func NewHandshake() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  Version,
		MagicCookieKey:   CookieKey,
		MagicCookieValue: CookieValue,
	}
}

// NewExtensionImplementation implements a new extension
// and starts serving it.
func NewExtensionImplementation(impl Implementation, log slogext.Logger) error {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:       hclog.Trace,
		Output:      &logger{fn: func(args ...any) { log.Debugf("%s", args...) }},
		DisableTime: true,
	})

	plugin.Serve(&plugin.ServeConfig{
		Logger:          logger,
		HandshakeConfig: NewHandshake(),
		Plugins: map[string]plugin.Plugin{
			Name: &ExtensionPlugin{log, newImplementationToImplementationTransport(impl)},
		},
	})

	return nil
}
