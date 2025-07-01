// Package yaml implements a thin wrapper around YAML parsing
// specifically with support for JSON schema parsing.
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
