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
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"go.rgst.io/stencil/internal/testing/cmdexec"
)

// ghProvider implements the provider interface using the Github CLI to
// retrieve a token.
type ghProvider struct{}

// Token returns a valid token or an error if no token is found.
func (p *ghProvider) Token() (string, error) {
	cmd := cmdexec.Command("gh", "auth", "token")
	token, err := cmd.Output()
	if err != nil {
		var execErr *exec.ExitError
		if errors.As(err, &execErr) {
			return "", fmt.Errorf("gh failed: %s (%w)", string(execErr.Stderr), execErr)
		}

		return "", fmt.Errorf("gh failed: %w (no stderr)", err)
	}

	return strings.TrimSpace(string(token)), nil
}
