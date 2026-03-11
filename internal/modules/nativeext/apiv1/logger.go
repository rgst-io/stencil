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

// Description: This file implements a simple io.Writer that writes
// to a function with a fmt.Print signature.

package apiv1

import "io"

// _ is a implementation check
var _ io.Writer = &logger{}

// logger implements io.Writer to write to a function with a fmt.Print signature
type logger struct {
	fn func(args ...any)
}

// Write writes the data to the logger
func (l *logger) Write(p []byte) (n int, err error) {
	l.fn("[go-plugin] ", string(p))
	return len(p), nil
}
