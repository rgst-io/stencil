// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file implements the rpc client transport for go-plugin

package apiv1

import (
	"net/rpc"

	"go.rgst.io/stencil/v2/pkg/slogext"
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
