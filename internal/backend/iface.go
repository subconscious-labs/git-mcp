// Package backend defines the GitBackend contract implemented by every
// supported Git backend (exec, go-git). The interface mirrors the tool
// surface of the Python reference mcp-server-git one-to-one, so that a
// client written for the reference implementation works unchanged.
package backend

import "context"

// LogOptions mirrors the optional arguments of git_log.
// An empty StartTimestamp or EndTimestamp means "not set".
type LogOptions struct {
	MaxCount       int
	StartTimestamp string
	EndTimestamp   string
}

// BranchOptions mirrors the arguments of git_branch.
// Type is one of "local", "remote", "all".
// Empty Contains / NotContains mean "not set".
type BranchOptions struct {
	Type        string
	Contains    string
	NotContains string
}

// GitBackend is the single abstraction every tool handler talks to.
// Implementations must be safe for concurrent use.
type GitBackend interface {
	// Read operations --------------------------------------------------

	Status(ctx context.Context, repoPath string) (string, error)
	DiffUnstaged(ctx context.Context, repoPath string, contextLines int) (string, error)
	DiffStaged(ctx context.Context, repoPath string, contextLines int) (string, error)
	Diff(ctx context.Context, repoPath, target string, contextLines int) (string, error)
	Log(ctx context.Context, repoPath string, opts LogOptions) (string, error)
	Show(ctx context.Context, repoPath, revision string) (string, error)
	Branch(ctx context.Context, repoPath string, opts BranchOptions) (string, error)

	// Write operations -------------------------------------------------

	Add(ctx context.Context, repoPath string, paths []string) (string, error)
	Commit(ctx context.Context, repoPath, message string) (string, error)
	Reset(ctx context.Context, repoPath string, paths []string) (string, error)
	CreateBranch(ctx context.Context, repoPath, name string, baseBranch string) (string, error)
	Checkout(ctx context.Context, repoPath, branch string) (string, error)
	Push(ctx context.Context, repoPath, remote, branch string) (string, error)
}
