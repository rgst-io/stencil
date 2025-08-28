// Copyright (C) 2025 stencil contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"errors"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// validateJSONSchema validates the provided data against the provided
// schema.
//
// If "identifier" is provided its used as a friendly name for the JSON
// schema in error messages.
func validateJSONSchema(identifier string, schemaMap map[string]any, data any) error {
	schemaBuf := new(bytes.Buffer)
	if err := json.NewEncoder(schemaBuf).Encode(schemaMap); err != nil {
		return fmt.Errorf("failed to encode schema into JSON: %w", err)
	}

	jsc := jsonschema.NewCompiler()
	jsc.DefaultDraft(jsonschema.Draft7)

	doc, err := jsonschema.UnmarshalJSON(schemaBuf)
	if err != nil {
		return fmt.Errorf("failed to decode (re)encoded JSON schema: %w", err)
	}
	if err := jsc.AddResource(identifier, doc); err != nil {
		return fmt.Errorf("failed to parse JSON schema (%s): %w", identifier, err)
	}

	schema, err := jsc.Compile(identifier)
	if err != nil {
		return fmt.Errorf("failed to compile JSON schema (%s): %w", identifier, err)
	}

	if err := schema.Validate(data); err != nil {
		errs := make([]error, 0)

		var validationError *jsonschema.ValidationError
		if errors.As(err, &validationError) {
			for _, validationErr := range validationError.DetailedOutput().Errors {
				pth := strings.TrimPrefix(validationErr.InstanceLocation, "/")

				errStrs := recurseError(&validationErr, []string{})

				//nolint:errcheck // Why: Best effort way to get the error.
				if pth == "" {
					// Can't provide detailed field information. Wrapped error
					// will provide the top-level location.
					errs = append(errs, fmt.Errorf("%s", strings.Join(errStrs, ", ")))
				} else {
					errs = append(errs, fmt.Errorf("%s: %s", pth, strings.Join(errStrs, ", ")))
				}
			}
		} else {
			errs = append(errs, err)
		}

		return fmt.Errorf("data failed json schema validation (%s): %w", identifier, errors.Join(errs...))
	}

	return nil
}

// recurseError recursively traverses the output unit struct which has both singular and plural error fields,
// returning a slice of error strings that it's adding onto.
func recurseError(ou *jsonschema.OutputUnit, errStrs []string) []string {
	if ou.Error != nil {
		//nolint:errcheck // Why: Best effort way to get the error.
		b, _ := ou.Error.MarshalJSON()

		errStr := strings.TrimSuffix(strings.TrimPrefix(string(b), "\""), "\"")
		errStrs = append(errStrs, errStr)
	} else if len(ou.Errors) > 0 {
		for _, ne := range ou.Errors {
			recurseError(&ne, errStrs)
		}
	}
	return errStrs
}
