// Copyright 2022 Outreach Corporation. All Rights Reserved.

// Description: This file contains helpers for git

// Package git implements helpers for interacting with git
package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"go.rgst.io/stencil/internal/testing/cmdexec"
)

// This block contains errors and regexes
var (
	// ErrNoHeadBranch is returned when a repository's HEAD (aka default) branch cannot
	// be determine
	ErrNoHeadBranch = errors.New("failed to find a head branch, does one exist?")

	// ErrNoRemoteHeadBranch is returned when a repository's remote  default/HEAD branch
	// cannot be determined.
	ErrNoRemoteHeadBranch = errors.New("failed to get head branch from remote origin")

	// headPattern is used to parse git output to determine the head branch
	headPattern = regexp.MustCompile(`HEAD branch: ([[:alpha:]]+)`)
)

// GetDefaultBranch determines the default/HEAD branch for a given git
// repository.
func GetDefaultBranch(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "show", "origin")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get head branch from remote origin")
	}

	matches := headPattern.FindStringSubmatch(string(out))
	if len(matches) != 2 {
		return "", ErrNoRemoteHeadBranch
	}

	return matches[1], nil
}

// Clone clone a git repository to a temporary directory and returns the
// path to the repository. If ref is empty, the default branch will be
// used.
func Clone(ctx context.Context, ref, url string) (string, error) {
	tempDir, err := os.MkdirTemp("", "stencil-"+strings.ReplaceAll(url, "/", "-"))
	if err != nil {
		return "", errors.Wrap(err, "failed to create temporary directory")
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "remote", "add", "origin", url},
		{"git", "-c", "protocol.version=2", "fetch", "origin", ref},
		{"git", "reset", "--hard", "FETCH_HEAD"},
	}
	for _, cmd := range cmds {
		//nolint:gosec // Why: Commands are not user provided.
		c := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
		c.Dir = tempDir
		if err := c.Run(); err != nil {
			var execErr *exec.ExitError
			if errors.As(err, &execErr) {
				return "", fmt.Errorf("failed to run %q (%w): %s", cmd, err, string(execErr.Stderr))
			}

			return "", fmt.Errorf("failed to run %q: %w", cmd, err)
		}
	}

	return tempDir, nil
}

// ListRemote returns a list of all remotes as shown from running 'git
// ls-remote'.
func ListRemote(ctx context.Context, remote string) ([][]string, error) {
	cmd := cmdexec.CommandContext(ctx, "git", "ls-remote", remote)
	out, err := cmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get remote branches")
	}

	remotes := make([][]string, 0)
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}

		remotes = append(remotes, strings.Fields(line))
	}
	return remotes, nil
}
