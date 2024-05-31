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
	"testing"

	"go.rgst.io/stencil/internal/testing/cmdexec"
	"gotest.tools/v3/assert"
)

// TestGhProviderTrimsSpace ensures that the token returned by the
// ghProvider is trimmed of any leading or trailing whitespace.
func TestGhProviderTrimsSpace(t *testing.T) {
	t.Parallel()

	p := &ghProvider{}

	cmdexec.UseMockExecutor(t, cmdexec.NewMockExecutor(&cmdexec.MockCommand{
		Name:   "gh",
		Args:   []string{"auth", "token"},
		Stdout: []byte(" token\n"),
	}))

	token, err := p.Token()
	assert.NilError(t, err)
	assert.Equal(t, "token", token)
}
