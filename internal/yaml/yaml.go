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

package yaml

import (
	"fmt"

	"sigs.k8s.io/yaml"

	stdyaml "go.yaml.in/yaml/v3"
)

// Marshal is an alias to [stdyaml.Marshal].
var Marshal = stdyaml.Marshal

// Unmarshal wraps [stdyaml.Unmarshal] but ensures the provided data is
// compatible with JSON serialization.
func Unmarshal(b []byte, obj any) error {
	var err error
	if b, err = yaml.YAMLToJSON(b); err != nil {
		return fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	return stdyaml.Unmarshal(b, obj)
}
