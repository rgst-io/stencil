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

// Description: This file implements the rpc server transport for go-plugin

package apiv1

import (
	"go.rgst.io/jaredallard/slogext/v2"
)

// rpcTransportServer implements a rpc backed implementation
// of implementationTransport.
type rpcTransportServer struct {
	log  slogext.Logger
	impl implementationTransport
}

// GetConfig returns the config for this extension
func (s *rpcTransportServer) GetConfig(_ any, resp **Config) error {
	v, err := s.impl.GetConfig()
	*resp = v
	return err
}

// GetTemplateFunctions returns the template functions for this extension
func (s *rpcTransportServer) GetTemplateFunctions(_ any, resp *[]*TemplateFunction) error {
	v, err := s.impl.GetTemplateFunctions()
	*resp = v
	return err
}

// ExecuteTemplateFunction executes a template function for this extension
//
//nolint:gocritic // Why: go-plugin wants this
func (s *rpcTransportServer) ExecuteTemplateFunction(t *TemplateFunctionExec, resp *[]byte) error {
	v, err := s.impl.ExecuteTemplateFunction(t)
	s.log.With("name", t.Name).WithError(err).Debugf("Extension function called: %s", string(v))
	*resp = v
	return err
}
