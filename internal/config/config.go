// Package config resolves runtime configuration from flags and environment
// variables. The precedence is: CLI flag > environment variable > default.
package config

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// Backend selects the Git implementation used to execute tool calls.
type Backend string

const (
	BackendExec  Backend = "exec"
	BackendGoGit Backend = "go-git"
)

var KnownBackends = []Backend{BackendExec, BackendGoGit}

// Mode controls which tools are exposed to the client.
type Mode string

const (
	// ModeRead exposes only read-only tools (7 total).
	ModeRead Mode = "read"

	// ModeLocalWrite additionally exposes git_add, git_commit, git_reset,
	// git_create_branch, git_checkout (12 total).
	ModeLocalWrite Mode = "local-write"

	// ModeRemoteWrite additionally exposes git_push (13 total).
	// This is the drop-in equivalent of the Python reference.
	ModeRemoteWrite Mode = "remote-write"
)

var KnownModes = []Mode{ModeRead, ModeLocalWrite, ModeRemoteWrite}

const (
	EnvBackend  = "GIT_MCP_BACKEND"
	EnvMode     = "GIT_MCP_MODE"
	EnvAllowDir = "GIT_MCP_ALLOW_DIR"
)

const DefaultBackend = BackendExec

// DefaultMode is read-only. Set --mode=local-write or --mode=remote-write
// to expose write tools. This diverges from the Python reference, where
// write tools are always present, in favor of a safer default.
const DefaultMode = ModeRead

// Config is the fully resolved server configuration.
type Config struct {
	Backend  Backend
	Mode     Mode
	AllowDir []string
}

// repeatedString accumulates repeated string flags.
type repeatedString []string

func (r *repeatedString) String() string { return strings.Join(*r, ",") }
func (r *repeatedString) Set(v string) error {
	*r = append(*r, v)
	return nil
}

// Load resolves configuration from the provided CLI arguments and the
// process environment. Diagnostics (usage, errors) are written to stderr.
func Load(args []string, stderr io.Writer) (*Config, error) {
	fs := flag.NewFlagSet("git-mcp", flag.ContinueOnError)
	fs.SetOutput(stderr)

	backendFlag := fs.String(
		"backend", "",
		fmt.Sprintf("Git backend: %s (default %q, overrides $%s)",
			joinStrings(backendStrings(KnownBackends)), DefaultBackend, EnvBackend),
	)
	modeFlag := fs.String(
		"mode", "",
		fmt.Sprintf("Tool exposure: %s (default %q, overrides $%s)",
			joinStrings(modeStrings(KnownModes)), DefaultMode, EnvMode),
	)
	var allowDirs repeatedString
	fs.Var(&allowDirs, "allow-dir",
		fmt.Sprintf("Restrict repo_path to this directory (repeatable, overrides $%s)", EnvAllowDir))

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	// Distinguish "flag not passed" from "flag passed as default value",
	// so env vars still take effect when the flag is absent.
	flagSet := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { flagSet[f.Name] = true })

	backend, err := resolveBackend(*backendFlag, os.Getenv(EnvBackend))
	if err != nil {
		return nil, err
	}
	mode, err := resolveMode(*modeFlag, os.Getenv(EnvMode))
	if err != nil {
		return nil, err
	}

	allowDir := []string(allowDirs)
	if len(allowDir) == 0 && !flagSet["allow-dir"] {
		if env := os.Getenv(EnvAllowDir); env != "" {
			for _, d := range strings.Split(env, string(os.PathListSeparator)) {
				if d != "" {
					allowDir = append(allowDir, d)
				}
			}
		}
	}

	return &Config{Backend: backend, Mode: mode, AllowDir: allowDir}, nil
}

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

func resolveMode(flagValue, envValue string) (Mode, error) {
	switch {
	case flagValue != "":
		return parseMode(flagValue, "flag -mode")
	case envValue != "":
		return parseMode(envValue, "env "+EnvMode)
	default:
		return DefaultMode, nil
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
		v, source, joinStrings(backendStrings(KnownBackends)))
}

func parseMode(v, source string) (Mode, error) {
	m := Mode(v)
	for _, known := range KnownModes {
		if m == known {
			return m, nil
		}
	}
	return "", fmt.Errorf("unknown mode %q from %s (want one of %s)",
		v, source, joinStrings(modeStrings(KnownModes)))
}

func backendStrings(bs []Backend) []string {
	out := make([]string, len(bs))
	for i, b := range bs {
		out[i] = string(b)
	}
	return out
}

func modeStrings(ms []Mode) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = string(m)
	}
	return out
}

func joinStrings(ss []string) string {
	return strings.Join(ss, ", ")
}
