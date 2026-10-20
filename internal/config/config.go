package config

import (
	"flag"
	"fmt"
	"io"
	"os"
)

type Backend string

const (
	BackendExec  Backend = "exec"
	BackendGoGit Backend = "go-git"
)

var KnownBackends = []Backend{BackendExec, BackendGoGit}

const (
	EnvBackend      = "GIT_MCP_BACKEND"
	EnvWriteAccess  = "GIT_MCP_WRITE_ACCESS"
)

const DefaultBackend = BackendExec

type Config struct {
	Backend     Backend
	WriteAccess bool
}

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
		false,
		fmt.Sprintf("Enable write tools such as git_push (overrides $%s)", EnvWriteAccess),
	)

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("parse flags: %w", err)
	}

	backend, err := resolveBackend(*backendFlag, os.Getenv(EnvBackend))
	if err != nil {
		return nil, err
	}

	writeAccess := *writeAccessFlag
	if !writeAccess {
		if env := os.Getenv(EnvWriteAccess); env == "1" || env == "true" {
			writeAccess = true
		}
	}

	return &Config{Backend: backend, WriteAccess: writeAccess}, nil
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
