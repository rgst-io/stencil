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

// Description: Implements the plugin RPC logic for the extension host

package apiv1

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"github.com/jaredallard/slogext"
)

// ExtensionPlugin is the high level plugin used by go-plugin
// it stores both the server and client implementation
type ExtensionPlugin struct {
	log  slogext.Logger
	impl implementationTransport
}

// Server serves a implementationTransport over net/rpc
func (p *ExtensionPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &rpcTransportServer{p.log, p.impl}, nil
}

// Client serves a Implementation over net/rpc
func (p *ExtensionPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &rpcTransportClient{p.log, c}, nil
}
