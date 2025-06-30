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

// Package main implements a CLI for regenerating JSON schemas used by
// stencil.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	"go.rgst.io/stencil/v2/pkg/configuration"
)

type schema struct {
	// Type should be a Go struct.
	Type any

	// FileName is the name of the file to write the schema to.
	FileName string
}

func main() {
	schemas := []schema{
		{
			Type:     &configuration.Manifest{},
			FileName: "stencil",
		},
		{
			Type:     &configuration.TemplateRepositoryManifest{},
			FileName: "manifest",
		},
	}

	for _, s := range schemas {
		fileName := fmt.Sprintf("%s.jsonschema.json", s.FileName)

		r := new(jsonschema.Reflector)
		r.FieldNameTag = "yaml"

		// Add comments to the schema.
		if err := r.AddGoComments("go.rgst.io/stencil/v2", "pkg/configuration"); err != nil {
			fmt.Printf("error adding comments for %s: %v\n", s.FileName, err)
			os.Exit(1)
		}

		schema := r.Reflect(s.Type)
		// VSCode doesn't handle above draft-07 right now, so we force it.
		schema.Version = "https://json-schema.org/draft-07/schema#"

		// Apply manifest-specific overrides.
		if s.FileName == "manifest" {
			subSchema, ok := schema.Definitions["Argument"].Properties.Get("schema")
			if !ok {
				fmt.Println("failed to update schema property for manifest to include $ref")
				os.Exit(1)
			}

			subSchema.Ref = schema.Version
			subSchema.Type = "" // Don't set type.
		}

		b, err := schema.MarshalJSON()
		if err != nil {
			fmt.Printf("error generating schema for %s: %v\n", s.FileName, err)
			os.Exit(1)
		}

		if err := os.WriteFile(filepath.Join("schemas", fileName), b, 0o600); err != nil {
			fmt.Printf("error writing schema for %s: %v\n", s.FileName, err)
			os.Exit(1)
		}
	}
}
