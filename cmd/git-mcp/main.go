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

	log.Printf("starting git-mcp version=%s backend=%s write_access=%t",
		version, cfg.Backend, cfg.WriteAccess)

	git, err := security.NewSafeGit()
	if err != nil {
		return fmt.Errorf("init safe git: %w", err)
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
