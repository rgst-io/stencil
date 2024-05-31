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

package github

import (
	"fmt"
	"os"
)

// envProvider implements the provider interface using the environment
// variables to retrieve a token.
type envProvider struct{}

// Token returns a valid token or an error if no token is found.
func (p *envProvider) Token() (string, error) {
	envVars := []string{"GITHUB_TOKEN"}
	for _, env := range envVars {
		if token := os.Getenv(env); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no token found in environment variables: %v", envVars)
}
