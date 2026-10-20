// Package config resolves runtime configuration from flags and environment
// variables. The precedence is: CLI flag > environment variable > default.
package config

import (
	"flag"
	"fmt"
	"io"
	"os"
)

// Backend selects the Git implementation used to execute tool calls.
type Backend string

const (
	BackendExec  Backend = "exec"   // shell out to the system git binary
	BackendGoGit Backend = "go-git" // pure-Go implementation via go-git
)

// KnownBackends lists every backend value accepted by the server.
var KnownBackends = []Backend{BackendExec, BackendGoGit}

// Environment variables consulted by Load.
const (
	EnvBackend     = "GIT_MCP_BACKEND"
	EnvWriteAccess = "GIT_MCP_WRITE_ACCESS"
)

// DefaultBackend is used when neither a flag nor an env var is set.
const DefaultBackend = BackendExec

// DefaultWriteAccess mirrors the Python reference, which always exposes
// write tools. Set --write-access=false or GIT_MCP_WRITE_ACCESS=false to
// run in read-only mode (no git_add, git_commit, git_reset,
// git_create_branch, git_checkout, git_push).
const DefaultWriteAccess = true

// Config is the fully resolved server configuration.
type Config struct {
	Backend     Backend
	WriteAccess bool
}

// Load resolves configuration from the provided CLI arguments and the
// process environment. Diagnostics (usage, errors) are written to stderr.
func Load(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("git-mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)

	backendFlag := fs.String(
		"backend",
		"",
		fmt.Sprintf("Git backend: %s (default %q, overrides $%s)",
			joinBackends(KnownBackends), DefaultBackend, EnvBackend),
	)
	writeAccessFlag := fs.Bool(
		"write-access",
		DefaultWriteAccess,
		fmt.Sprintf("Enable write tools (default %t, overrides $%s)",
			DefaultWriteAccess, EnvWriteAccess),
	)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	// Distinguish "flag not passed" from "flag passed as default value",
	// so that env vars still take effect when the flag is absent.
	flagSet := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { flagSet[f.Name] = true })

	backend, err := resolveBackend(*backendFlag, os.Getenv(EnvBackend))
	if err != nil {
		return nil, err
	}

	writeAccess := resolveWriteAccess(*writeAccessFlag, flagSet["write-access"], os.Getenv(EnvWriteAccess))

	return &Config{Backend: backend, WriteAccess: writeAccess}, nil
}

// resolveBackend implements the precedence chain: flag > env > default.
// Empty strings are treated as "not set".
func resolveBackend(flagValue, envValue string) (Backend, error) {
	switch {
	case flagValue != "":
		return parseBackend(flagValue, "flag -backend")
	case envValue != "":
		return parseBackend(envValue, "env "+EnvBackend)
	default:
		return DefaultBackend, nil
	}
}

// resolveWriteAccess implements the precedence chain for booleans:
// explicit flag > env > default. An empty env string means "not set".
func resolveWriteAccess(flagValue, flagSet bool, envValue string) bool {
	switch {
	case flagSet:
		return flagValue
	case envValue != "":
		return envValue == "1" || envValue == "true"
	default:
		return DefaultWriteAccess
	}
}

func parseBackend(v, source string) (Backend, error) {
	b := Backend(v)
	for _, known := range KnownBackends {
		if b == known {
			return b, nil
		}
	}
	return "", fmt.Errorf("unknown backend %q from %s (want one of %s)",
		v, source, joinBackends(KnownBackends))
}

func joinBackends(bs []Backend) string {
	out := ""
	for i, b := range bs {
		if i > 0 {
			out += ", "
		}
		out += string(b)
	}
	return out
}
