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

// Package version contains the current version of the stencil CLI.
package version

import (
	// Used for the ASCII art.
	_ "embed"

	goversion "github.com/caarlos0/go-version"
)

// These variables are set at build time via ldflags.
var (
	version   = ""
	commit    = ""
	treeState = ""
	date      = ""
	builtBy   = ""
)

//go:embed embed/stencil.ascii.txt
var asciiName string

// Version is the current version of the stencil CLI.
var Version = goversion.GetVersionInfo(
	goversion.WithAppDetails("stencil", "A modern living-template engine for evolving repositories", "\033[4mhttps://stencil.rgst.io\033[0m"),
	goversion.WithASCIIName("\033[90m"+asciiName+"\033[0m\n"),
	func(i *goversion.Info) {
		if commit != "" {
			i.GitCommit = commit
		}
		if treeState != "" {
			i.GitTreeState = treeState
		}
		if date != "" {
			i.BuildDate = date
		}
		if version != "" {
			i.GitVersion = version
		}
		if builtBy != "" {
			i.BuiltBy = builtBy
		}
	},
)
