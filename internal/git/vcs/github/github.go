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

// Package github provides methods for creating a Github client for the
// purposes of interacting with the API. Provides support for retrieving
// authentication tokens from the following sources:
//
// - Environment variables ($GITHUB_TOKEN)
// - Github CLI
package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v63/github"
	"golang.org/x/oauth2"
)

// defaultProviders is a list of credential providers that are used to
// retrieve a token by default.
var defaultProviders = []provider{
	&envProvider{},
	&ghProvider{},
}

// Token returns a valid token from one of the configured credential
// providers. If no token is found, ErrNoToken is returned.
func Token() (string, error) {
	token := ""
	errors := []error{}
	for _, p := range defaultProviders {
		var err error
		token, err = p.Token()
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// Got a token, break out of the loop.
		if token != "" {
			break
		}
	}
	if token == "" {
		return "", ErrNoToken{errors}
	}
	return token, nil
}

// New returns a new [github.Client] using credentials from one of the
// configured credential providers. If no token is found, an
// unauthenticated client is returned.
func New() (*github.Client, error) {
	token, err := Token()
	if err != nil {
		return github.NewClient(http.DefaultClient), nil
	}

	// Note: background ctx is used here because we don't want the oauth2
	// client to pick up credentials from a provided context.
	return github.NewClient(oauth2.NewClient(context.Background(),
		oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)),
	), nil
}
