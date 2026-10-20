// Command git-mcp is a drop-in compatible MCP server for Git,
// exposing the same tool contract as the Python reference implementation.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/subconscious-labs/git-mcp/internal/backend"
	execbackend "github.com/subconscious-labs/git-mcp/internal/backend/exec"
	"github.com/subconscious-labs/git-mcp/internal/config"
	mcpserver "github.com/subconscious-labs/git-mcp/internal/mcp"
	"github.com/subconscious-labs/git-mcp/internal/security"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetPrefix("git-mcp: ")

	if err := run(os.Args[1:], os.Stderr); err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	cfg, err := config.Load(args, stderr)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	log.Printf("starting git-mcp version=%s backend=%s mode=%s",
		version, cfg.Backend, cfg.Mode)

	git, err := security.NewSafeGit()
	if err != nil {
		return fmt.Errorf("init safe git: %w", err)
	}

	if len(cfg.AllowDir) > 0 {
		canon := make([]string, 0, len(cfg.AllowDir))
		for _, d := range cfg.AllowDir {
			c, err := security.CanonicalizeAllowDir(d)
			if err != nil {
				return fmt.Errorf("allow-dir %q: %w", d, err)
			}
			canon = append(canon, c)
		}
		git.AllowDir = canon
		log.Printf("allow-dir: %v", canon)
	} else {
		log.Printf("WARNING: no --allow-dir configured; " +
			"any repo_path readable by this process is permitted")
	}

	var b backend.GitBackend
	switch cfg.Backend {
	case config.BackendExec:
		b = execbackend.New(git)
	case config.BackendGoGit:
		return fmt.Errorf("go-git backend is not implemented yet (Phase 4)")
	default:
		return fmt.Errorf("unsupported backend %q", cfg.Backend)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "git-mcp",
		Version: version,
	}, nil)

	mcpserver.RegisterTools(srv, b, cfg)

	if err := srv.Run(ctx, &mcp.StdioTransport{}); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("server: %w", err)
	}

	log.Printf("shutdown complete")
	return nil
}
