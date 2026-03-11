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

// Description: This file implements the rpc client transport for go-plugin

package apiv1

import (
	"net/rpc"

	"github.com/jaredallard/slogext"
)

// _ is a compile time assertion we implement the interface
var _ implementationTransport = &rpcTransportClient{}

// rpcTransportClient implements the plugin client over
// rpc. This is a low level interface responsible for transmitting
// the implementationTransport over the wire.
type rpcTransportClient struct {
	log    slogext.Logger
	client *rpc.Client
}

// GetConfig returns the config for the extension
func (g *rpcTransportClient) GetConfig() (*Config, error) {
	var resp *Config
	err := g.client.Call("Plugin.GetConfig", new(any), &resp)
	return resp, err
}

// GetTemplateFunctions returns the template functions for this extension
func (g *rpcTransportClient) GetTemplateFunctions() ([]*TemplateFunction, error) {
	var resp []*TemplateFunction
	err := g.client.Call("Plugin.GetTemplateFunctions", new(any), &resp)
	return resp, err
}

// ExecuteTemplateFunction exectues a template function for this extension
func (g *rpcTransportClient) ExecuteTemplateFunction(t *TemplateFunctionExec) ([]byte, error) {
	// IDEA(jaredallard): Actually stream this data in the future
	var resp []byte
	err := g.client.Call("Plugin.ExecuteTemplateFunction", t, &resp)
	g.log.With("data", string(resp)).WithError(err).With("name", t.Name).Debug("Extension function returned data")
	return resp, err
}
