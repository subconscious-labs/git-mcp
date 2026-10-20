// Package security hosts the middleware that guards every tool invocation.
// This file provides SafeGit, the only sanctioned way to spawn the system
// git binary.
package security

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DefaultGitTimeout bounds every spawned git process.
const DefaultGitTimeout = 30 * time.Second

// SafeGit runs git commands under a hardened environment:
//   - no shell interpolation (arguments passed as a slice)
//   - no interactive prompts (GIT_TERMINAL_PROMPT=0)
//   - no pager (GIT_PAGER="")
//   - explicit stdin closure
//   - hard timeout
//   - per-invocation safe.directory for the target repo, without touching
//     the user's global gitconfig
type SafeGit struct {
	Binary  string
	Timeout time.Duration
}

// NewSafeGit locates the git binary on PATH and returns a hardened runner.
func NewSafeGit() (*SafeGit, error) {
	bin, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git binary not found in PATH: %w", err)
	}
	return &SafeGit{Binary: bin, Timeout: DefaultGitTimeout}, nil
}

// Run executes `git -c safe.directory=<abs> -C <abs> args...` and returns
// stdout. stderr is captured and returned in the error, so callers can
// propagate git's diagnostics to the client.
func (g *SafeGit) Run(ctx context.Context, repoPath string, args ...string) (string, error) {
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("resolve repo path %q: %w", repoPath, err)
	}

	ctx, cancel := context.WithTimeout(ctx, g.Timeout)
	defer cancel()

	full := append([]string{
		"-c", "safe.directory=" + abs,
		"-C", abs,
	}, args...)
	cmd := exec.CommandContext(ctx, g.Binary, full...)
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_PAGER=",
		"GIT_ASKPASS=/bin/false",
		"SSH_ASKPASS=/bin/false",
	)
	cmd.Stdin = nil

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], msg)
	}
	return stdout.String(), nil
}

// ResolveCommit validates rev and returns its canonical SHA.
// It is the only safe way to accept a revision from untrusted input
// before passing it to a git command that would otherwise interpret
// it as a flag.
func (g *SafeGit) ResolveCommit(ctx context.Context, repoPath, rev string) (string, error) {
	if rev == "" {
		return "", fmt.Errorf("empty revision")
	}
	out, err := g.Run(ctx, repoPath,
		"rev-parse", "--verify", "--end-of-options", rev+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("invalid revision %q: %w", rev, err)
	}
	return strings.TrimSpace(out), nil
}
