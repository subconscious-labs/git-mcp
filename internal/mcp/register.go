package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/subconscious-labs/git-mcp/internal/backend"
	"github.com/subconscious-labs/git-mcp/internal/config"
	"github.com/subconscious-labs/git-mcp/internal/security"
)

// RegisterTools installs tools according to cfg.Mode:
//
//	read          → 7 read-only tools
//	local-write   → + git_add, git_commit, git_reset,
//	                  git_create_branch, git_checkout (12 total)
//	remote-write  → + git_fetch, git_pull, git_push (15 total)
func RegisterTools(s *mcp.Server, b backend.GitBackend, cfg *config.Config) {
	registerReadTools(s, b)
	if cfg.Mode == config.ModeLocalWrite || cfg.Mode == config.ModeRemoteWrite {
		registerLocalWriteTools(s, b)
	}
	if cfg.Mode == config.ModeRemoteWrite {
		registerRemoteWriteTools(s, b)
	}
}

func registerReadTools(s *mcp.Server, b backend.GitBackend) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_status",
		Description: "Shows the working tree status",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitStatusParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Status(ctx, p.RepoPath)
		if err != nil {
			return nil, nil, err
		}
		return text(fmt.Sprintf("Repository status:\n%s", out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_diff_unstaged",
		Description: "Shows changes in the working directory that are not yet staged",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitDiffUnstagedParams) (*mcp.CallToolResult, any, error) {
		out, err := b.DiffUnstaged(ctx, p.RepoPath, deref(p.ContextLines))
		if err != nil {
			return nil, nil, err
		}
		return text(fmt.Sprintf("Unstaged changes:\n%s", out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_diff_staged",
		Description: "Shows changes that are staged for commit",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitDiffStagedParams) (*mcp.CallToolResult, any, error) {
		out, err := b.DiffStaged(ctx, p.RepoPath, deref(p.ContextLines))
		if err != nil {
			return nil, nil, err
		}
		return text(fmt.Sprintf("Staged changes:\n%s", out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_diff",
		Description: "Shows differences between branches or commits",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitDiffParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Diff(ctx, p.RepoPath, p.Target, deref(p.ContextLines))
		if err != nil {
			return nil, nil, err
		}
		return text(fmt.Sprintf("Differences with %s:\n%s", p.Target, out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_log",
		Description: "Shows the commit logs",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitLogParams) (*mcp.CallToolResult, any, error) {
		maxCount := DefaultMaxCount
		if p.MaxCount != nil {
			maxCount = *p.MaxCount
		}
		opts := backend.LogOptions{MaxCount: maxCount}
		if p.StartTimestamp != nil {
			opts.StartTimestamp = *p.StartTimestamp
		}
		if p.EndTimestamp != nil {
			opts.EndTimestamp = *p.EndTimestamp
		}
		out, err := b.Log(ctx, p.RepoPath, opts)
		if err != nil {
			return nil, nil, err
		}
		return text(fmt.Sprintf("Commit history:\n%s", out)), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_show",
		Description: "Shows the contents of a commit, or of a file or directory given as <commit>:<path>",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitShowParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Show(ctx, p.RepoPath, p.Revision)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_branch",
		Description: "List Git branches",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitBranchParams) (*mcp.CallToolResult, any, error) {
		opts := backend.BranchOptions{Type: p.BranchType}
		if p.Contains != nil {
			opts.Contains = *p.Contains
		}
		if p.NotContains != nil {
			opts.NotContains = *p.NotContains
		}
		out, err := b.Branch(ctx, p.RepoPath, opts)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})
}

func registerLocalWriteTools(s *mcp.Server, b backend.GitBackend) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_add",
		Description: "Adds file contents to the staging area",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitAddParams) (*mcp.CallToolResult, any, error) {
		if len(p.Files) == 0 {
			return nil, nil, fmt.Errorf("no files provided to stage")
		}
		out, err := b.Add(ctx, p.RepoPath, p.Files)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_commit",
		Description: "Records changes to the repository",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitCommitParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Commit(ctx, p.RepoPath, p.Message)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_reset",
		Description: "Unstages all staged changes",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitResetParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Reset(ctx, p.RepoPath, nil)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_create_branch",
		Description: "Creates a new branch from an optional base branch",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitCreateBranchParams) (*mcp.CallToolResult, any, error) {
		base := ""
		if p.BaseBranch != nil {
			base = *p.BaseBranch
		}
		out, err := b.CreateBranch(ctx, p.RepoPath, p.BranchName, base)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_checkout",
		Description: "Switches branches",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitCheckoutParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Checkout(ctx, p.RepoPath, p.BranchName)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})
}

func registerRemoteWriteTools(s *mcp.Server, b backend.GitBackend) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_fetch",
		Description: "Fetches remote refs without touching local branches (extension)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitFetchParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Fetch(ctx, p.RepoPath, p.Remote)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_pull",
		Description: "Fast-forwards the current branch from its remote (extension, --ff-only)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitPullParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Pull(ctx, p.RepoPath, p.Remote, p.Branch)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "git_push",
		Description: "Pushes the current branch to its configured upstream",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, p GitPushParams) (*mcp.CallToolResult, any, error) {
		out, err := b.Push(ctx, p.RepoPath, p.Remote, p.Branch)
		if err != nil {
			return nil, nil, err
		}
		return text(out), nil, nil
	})
}

// text wraps a plain string as a successful MCP tool result and runs it
// through the output sanitizer. Redaction findings are logged to stderr;
// the matched data itself is never logged.
func text(s string) *mcp.CallToolResult {
	out, findings := security.SanitizeOutput(s)
	if len(findings) > 0 {
		log.Printf("sanitized output: %v", findings)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: out},
		},
	}
}

func deref(p *int) int {
	if p == nil {
		return DefaultContextLines
	}
	return *p
}

// ToolNamesByMode returns the tool names registered for the given mode.
// It is the single source of truth for what RegisterTools installs; the
// register functions above must stay in sync with it. Exported for tests.
func ToolNamesByMode(m config.Mode) []string {
	read := []string{
		"git_status",
		"git_diff_unstaged",
		"git_diff_staged",
		"git_diff",
		"git_log",
		"git_show",
		"git_branch",
	}
	if m == config.ModeRead {
		return read
	}
	local := append(read,
		"git_add",
		"git_commit",
		"git_reset",
		"git_create_branch",
		"git_checkout",
	)
	if m == config.ModeLocalWrite {
		return local
	}
	return append(local, "git_fetch", "git_pull", "git_push")
}
