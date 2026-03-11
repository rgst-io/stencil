// Copyright (C) 2026 stencil contributors
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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v3"
	"go.rgst.io/stencil/v2/pkg/configuration"

	// We don't use internal/yaml here because of the JSON conversion.
	"go.yaml.in/yaml/v3"
)

// NewUseCommand returns a new urfave/cli.Command for the use command
func NewUseCommand() *cli.Command {
	return &cli.Command{
		Name:    "use",
		Aliases: []string{"add"},
		Usage:   "Use new/modify existing modules in use",
		Description: strings.Join([]string{
			"Add new modules, modify existing versions, or replace an existing module.",
			"",
			"Replace a local module: stencil use ../../path-to-module",
			"Use a new module: stencil use github.com/rgst-io/stencil-golang",
			"Use a specific version: stencil use github.com/rgst-io/stencil-golang@v1.0.0",
		}, "\n"),
		Arguments: []cli.Argument{&cli.StringArg{
			Name:      "module",
			UsageText: "<module>",
		}},
		// Empty function means the file-system will be looked at instead of
		// trying to generate our own completion.
		ShellComplete: func(_ context.Context, _ *cli.Command) {},
		Action: func(_ context.Context, c *cli.Command) error {
			if c.StringArg("module") == "" {
				return errors.New("expected exactly one argument, module")
			}

			return useModule(c.StringArg("module"))
		},
	}
}

// useModule mutates `stencil.yaml` to use a new module or replace an
// existing one (if the provided module is a file path).
func useModule(filePathOrModulePath string) error {
	var moduleName string
	var moduleVer string
	var isReplacement bool

	stencilPath := "stencil.yaml"
	mfPath := filepath.Join(filePathOrModulePath, "manifest.yaml")
	if tmf, err := configuration.LoadTemplateRepositoryManifest(mfPath); err == nil {
		isReplacement = true
		moduleName = tmf.Name
	} else {
		moduleName = filePathOrModulePath
		if strings.HasPrefix("/", moduleName) {
			return fmt.Errorf("module name cannot start with a / unless using a replacement")
		}

		spl := strings.SplitN(moduleName, "@", 2)
		if len(spl) != 1 {
			moduleVer = spl[1]
			moduleName = strings.TrimSuffix(moduleName, "@"+moduleVer)
		}
	}

	b, err := os.ReadFile(stencilPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", stencilPath, err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(b, &node); err != nil {
		return fmt.Errorf("failed to decode %s as yaml: %w", stencilPath, err)
	}

	root := &node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		root = node.Content[0]
	}

	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("expected %s root to be a mapping (is it empty?)", stencilPath)
	}

	if isReplacement {
		replacements := findOrCreateMapKey(root, "replacements")
		if replacements.Kind != yaml.MappingNode {
			replacements.Kind = yaml.MappingNode
			replacements.Content = make([]*yaml.Node, 0)
		}

		setMapValue(replacements, moduleName, filePathOrModulePath)
	} else {
		modules := findOrCreateMapKey(root, "modules")
		if modules.Kind != yaml.SequenceNode {
			modules.Kind = yaml.SequenceNode
			modules.Content = make([]*yaml.Node, 0)
		}

		moduleEntry := findOrCreateModuleEntry(modules, moduleName)
		if moduleVer != "" {
			setMapValue(moduleEntry, "version", moduleVer)
		} else {
			deleteMapKey(moduleEntry, "version")
		}
	}

	f, err := os.Create(stencilPath)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", stencilPath, err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	defer enc.Close()
	enc.SetIndent(2)
	if err := enc.Encode(&node); err != nil {
		return fmt.Errorf("failed to encode mutated %s as yaml: %w", stencilPath, err)
	}

	if isReplacement {
		fmt.Fprintf(os.Stdout,
			"Replaced %s => %s\n", moduleName, filePathOrModulePath,
		)
	} else {
		fmt.Fprintf(os.Stdout,
			"Added %s\n", moduleName,
		)
	}

	return nil
}

// findOrCreateMapKey finds a key in a mapping node and returns its
// value node. If the key doesn't exist, it creates it with an
// uninitialized value node.
func findOrCreateMapKey(n *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}

	kn := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	vn := &yaml.Node{}
	n.Content = append(n.Content, kn, vn)
	return vn
}

// setMapValue sets or updates a key-value pair in a mapping node.
func setMapValue(n *yaml.Node, key, value string) {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			n.Content[i+1].Kind = yaml.ScalarNode
			n.Content[i+1].Value = value
			return
		}
	}

	kn := &yaml.Node{Kind: yaml.ScalarNode, Value: key}
	vn := &yaml.Node{Kind: yaml.ScalarNode, Value: value}
	n.Content = append(n.Content, kn, vn)
}

// deleteMapKey removes a key-value pair from a mapping node.
func deleteMapKey(n *yaml.Node, key string) {
	for i := 0; i < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			n.Content = append(n.Content[:i], n.Content[i+2:]...)
			return
		}
	}
}

// findOrCreateModuleEntry finds a module entry in a sequence by name.
// If the module doesn't exist, it creates and appends a new module entry.
func findOrCreateModuleEntry(modules *yaml.Node, name string) *yaml.Node {
	for _, item := range modules.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}

		for i := 0; i < len(item.Content); i += 2 {
			if item.Content[i].Value == "name" && item.Content[i+1].Value == name {
				return item
			}
		}
	}

	newModule := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: name},
		},
	}
	modules.Content = append(modules.Content, newModule)
	return newModule
}
