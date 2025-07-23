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

// Package main implements the "update-mise-lock" CLI tool which
// regenerates a .mise.lock for a given set of architectures.
//
// Usage: update-mise-lock [GOOS/GOARCH]
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/jaredallard/cmdexec"
)

// DockerImage is the image to use for running mise.
const DockerImage = "docker.io/jdxcode/mise"

// BinFmtImage is the image to use for setting up QEMU.
const BinFmtImage = "docker.io/tonistiigi/binfmt"

// Platform is a parsed version of GOOS/GOARCH pairings.
type Platform struct {
	OS   string
	Arch string
}

// String returns the original GOOS/GOARCH string.
func (p *Platform) String() string {
	return p.OS + "/" + p.Arch
}

// entrypoint is the entrypoint for this CLI tool.
func entrypoint(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no platforms passed (expects at least one GOOS/GOARCH pair)")
	}

	var platforms []Platform
	for _, arg := range args {
		spl := strings.Split(arg, "/")
		if len(spl) != 2 {
			return fmt.Errorf("invalid GOOS/GOARCH supplied %q", arg)
		}
		platforms = append(platforms, Platform{
			OS:   spl[0],
			Arch: spl[1],
		})
	}

	cmd := cmdexec.CommandContext(ctx, "docker", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run %q: %w", cmd.String(), err)
	}

	if _, err := os.Stat(".mise.toml"); err != nil {
		return fmt.Errorf("failed to check for .mise.toml (are you in the repo root?): %w", err)
	}

	// Need to setup QEMU
	if runtime.GOOS == "linux" {
		fmt.Println(" [update-mise-lock] Ensuring QEMU emulation is configured (binfmt)")
		//nolint:gosec,errcheck // Why: Best effort.
		cmdexec.CommandContext(ctx, "docker", "run", "--privileged", "--rm", "tonistiigi/binfmt", "--uninstall", "all").Run()
		cmd := cmdexec.CommandContext(ctx, "docker", "run", "--privileged", "--rm", "tonistiigi/binfmt", "--install", "all")
		cmd.UseOSStreams(true)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install QEMU emulation: %w", err)
		}
	}

	for i := range platforms {
		p := &platforms[i]

		fmt.Printf(" [%s] Updating lock file...\n", p.String())

		var cmd cmdexec.Cmd
		switch p.OS {
		case "darwin":
			if runtime.GOOS != "darwin" {
				return fmt.Errorf("updating darwin lockfiles is not implemented on non-darwin hosts")
			}

			switch {
			case runtime.GOARCH == "amd64" && p.Arch == "arm64":
				return fmt.Errorf("updating arm64 on amd64 mac is not implemented")
			case runtime.GOARCH == "arm64" && p.Arch == "amd64":
				return fmt.Errorf("updating amd64 on arm64 mac is not implemented")
			default:
				cmd = cmdexec.CommandContext(ctx, "mise", "install")
			}
		case "windows":
			return fmt.Errorf("windows is not implemented")
		default:
			//nolint:errcheck // Why: Best effort.
			ghToken, _ := cmdexec.Command("gh", "auth", "token").Output()
			if len(ghToken) == 0 {
				ghToken = []byte("")
			}

			cmd = cmdexec.CommandContext(ctx,
				"docker", "run", "-it", "--rm", "-v", "./:/app",
				"-e", "GITHUB_TOKEN="+strings.TrimSpace(string(ghToken)),
				"--platform", p.String(), "-w", "/app",
				"--entrypoint", "bash",
				DockerImage, "-oeu", "pipefail", "-c",
				"mise settings experimental=true && mise trust && exec mise install",
			)
		}

		cmd.UseOSStreams(true)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update lock for %s: %w", p.String(), err)
		}
	}

	return nil
}

// main calls [entrypoint] with error handling
func main() {
	ctx := context.Background()
	if err := entrypoint(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run: %v\n", err)
		os.Exit(1)
	}
}
