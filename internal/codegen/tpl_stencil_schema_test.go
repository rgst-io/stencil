package codegen

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func Test_validateJSONSchema(t *testing.T) {
	testCases := []struct {
		name       string
		identifier string
		schema     map[string]any
		data       any
		expected   error
	}{
		{
			name: "should be able to validate a basic schema",
			schema: map[string]any{
				"type": "string",
			},
			data:     "hello world",
			expected: nil,
		},
		{
			name: "should be able to fail a basic schema",
			schema: map[string]any{
				"type": "number",
			},
			data: "hello world",
			expected: fmt.Errorf(
				"data failed json schema validation " +
					"(gotest/Test_validateJSONSchema/should_be_able_to_fail_a_basic_schema): " +
					"got string, want number",
			),
		},
		{
			name: "should return non obscure error",
			schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"hello": map[string]any{
						"type": "number",
					},
				},
			},
			data: map[string]any{"hello": "world"},
			expected: fmt.Errorf(
				"data failed json schema validation " +
					"(gotest/Test_validateJSONSchema/should_return_non_obscure_error): " +
					"hello: got string, want number",
			),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.identifier == "" {
				tc.identifier = "gotest/" + t.Name()
			}

			err := validateJSONSchema(tc.identifier, tc.schema, tc.data)
			if tc.expected == nil {
				assert.NilError(t, err)
			} else {
				assert.Equal(t, err.Error(), tc.expected.Error())
			}
		})
	}
}
