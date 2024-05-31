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
	"testing"

	"go.rgst.io/stencil/internal/testing/cmdexec"
	"golang.org/x/oauth2"
	"gotest.tools/v3/assert"
)

// TestEnvOverGH ensures that the GITHUB_TOKEN environment variable
// takes precedence over the token returned by the gh CLI.
func TestEnvOverGH(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "github_token")

	cmdexec.UseMockExecutor(t, cmdexec.NewMockExecutor(&cmdexec.MockCommand{
		Name:   "gh",
		Args:   []string{"auth", "token"},
		Stdout: []byte("gh_token\n"),
	}))

	cli, err := New()
	assert.NilError(t, err)
	token, err := cli.Client().Transport.(*oauth2.Transport).Source.Token()
	assert.NilError(t, err)
	assert.Equal(t, "github_token", token.AccessToken)
}

// TestReturnsErrors ensures that New returns underlying provider errors
// and that they can be found.
func TestReturnsErrors(t *testing.T) {
	cmdexec.UseMockExecutor(t, cmdexec.NewMockExecutor(&cmdexec.MockCommand{
		Name: "gh",
		Args: []string{"auth", "token"},
		Err:  fmt.Errorf("bad things happened"),
	}))

	token, err := Token()
	var tokenErr ErrNoToken
	assert.Assert(t, errors.As(err, &tokenErr))
	assert.Assert(t, token == "", "expected token to be empty")

	// Find a GH cli error
	var found bool
	for _, e := range tokenErr.errs {
		if e.Error() == "gh failed: bad things happened (no stderr)" {
			found = true
		}
	}
	assert.Assert(t, found, "expected error not found")
}
