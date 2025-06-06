// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file implements the rpc server transport for go-plugin

package apiv1

import (
	"go.rgst.io/stencil/v2/pkg/slogext"
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
