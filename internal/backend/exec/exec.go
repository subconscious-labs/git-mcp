// Package exec implements GitBackend by shelling out to the system git.
// It is the reference backend: its output formats match the Python
// mcp-server-git implementation byte-for-byte wherever practical.
package exec

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/subconscious-labs/git-mcp/internal/backend"
	"github.com/subconscious-labs/git-mcp/internal/security"
)

// Backend implements backend.GitBackend via the system git binary.
type Backend struct {
	git *security.SafeGit
}

// New constructs an exec backend using the given SafeGit runner.
func New(git *security.SafeGit) *Backend {
	return &Backend{git: git}
}

// compile-time interface check
var _ backend.GitBackend = (*Backend)(nil)

// Status returns `git status` output, unwrapped.
// The tool layer adds the "Repository status:\n" prefix.
func (b *Backend) Status(ctx context.Context, repoPath string) (string, error) {
	return b.git.Run(ctx, repoPath, "status")
}

// DiffUnstaged mirrors git diff --unified=N (working tree vs index).
func (b *Backend) DiffUnstaged(ctx context.Context, repoPath string, contextLines int) (string, error) {
	return b.git.Run(ctx, repoPath, "diff", unified(contextLines))
}

// DiffStaged mirrors git diff --cached --unified=N (index vs HEAD).
func (b *Backend) DiffStaged(ctx context.Context, repoPath string, contextLines int) (string, error) {
	return b.git.Run(ctx, repoPath, "diff", "--cached", unified(contextLines))
}

// Diff mirrors git diff <target> --unified=N.
// The target is validated to a commit SHA first so it can never be
// interpreted as a flag by git.
func (b *Backend) Diff(ctx context.Context, repoPath, target string, contextLines int) (string, error) {
	sha, err := b.git.ResolveCommit(ctx, repoPath, target)
	if err != nil {
		return "", err
	}
	return b.git.Run(ctx, repoPath, "diff", sha, unified(contextLines))
}

// Log mirrors git log with the reference pretty format.
func (b *Backend) Log(ctx context.Context, repoPath string, opts backend.LogOptions) (string, error) {
	args := []string{
		"log",
		"-n", strconv.Itoa(opts.MaxCount),
		"--pretty=format:Commit: %H%nAuthor: %an <%ae>%nDate: %aI%nMessage: %B%n---",
	}
	if opts.StartTimestamp != "" {
		args = append(args, "--after="+opts.StartTimestamp)
	}
	if opts.EndTimestamp != "" {
		args = append(args, "--before="+opts.EndTimestamp)
	}
	out, err := b.git.Run(ctx, repoPath, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(out, "\n"), nil
}

// Show mirrors git show <revision>.
// Accepts both object form (SHA, tag, HEAD) and path form
// (<commit>:<path>), matching the reference tool description.
func (b *Backend) Show(ctx context.Context, repoPath, revision string) (string, error) {
	commit, path, hasPath := strings.Cut(revision, ":")
	sha, err := b.git.ResolveCommit(ctx, repoPath, commit)
	if err != nil {
		return "", err
	}
	target := sha
	if hasPath {
		if path == "" {
			return "", fmt.Errorf("empty path in revision %q", revision)
		}
		// Defense in depth until Phase 3 lands ValidatePositional.
		if strings.HasPrefix(path, "-") {
			return "", fmt.Errorf("invalid path %q: must not start with '-'", path)
		}
		target = sha + ":" + path
	}
	return b.git.Run(ctx, repoPath, "show", target)
}

// Branch mirrors the reference branch listing. Output is one branch per
// line, with a trailing " (current)" marker on the checked-out branch.
func (b *Backend) Branch(ctx context.Context, repoPath string, opts backend.BranchOptions) (string, error) {
	args := []string{"branch", "--format=%(refname:short)%(if)%(HEAD)%(then) (current)%(end)"}
	switch opts.Type {
	case "local", "":
		// default: local branches only
	case "remote":
		args = append(args, "-r")
	case "all":
		args = append(args, "-a")
	default:
		return "", fmt.Errorf("invalid branch_type %q (want local, remote or all)", opts.Type)
	}
	if opts.Contains != "" {
		sha, err := b.git.ResolveCommit(ctx, repoPath, opts.Contains)
		if err != nil {
			return "", fmt.Errorf("invalid --contains: %w", err)
		}
		args = append(args, "--contains="+sha)
	}
	if opts.NotContains != "" {
		sha, err := b.git.ResolveCommit(ctx, repoPath, opts.NotContains)
		if err != nil {
			return "", fmt.Errorf("invalid --not-contains: %w", err)
		}
		args = append(args, "--no-contains="+sha)
	}
	out, err := b.git.Run(ctx, repoPath, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(out, "\n"), nil
}

// Add stages the given paths. The trailing "--" prevents any path from
// being interpreted as a flag.
func (b *Backend) Add(ctx context.Context, repoPath string, paths []string) (string, error) {
	args := append([]string{"add", "--"}, paths...)
	if _, err := b.git.Run(ctx, repoPath, args...); err != nil {
		return "", err
	}
	return "Files staged successfully", nil
}

// Commit mirrors git commit -m <message>. The message is passed as a
// single argv value, so it can never be interpreted as a flag by git.
func (b *Backend) Commit(ctx context.Context, repoPath, message string) (string, error) {
	if _, err := b.git.Run(ctx, repoPath, "commit", "-m", message); err != nil {
		return "", err
	}
	sha, err := b.git.Run(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Changes committed successfully with hash %s", strings.TrimSpace(sha)), nil
}

// Reset mirrors the reference: unstages everything.
func (b *Backend) Reset(ctx context.Context, repoPath string, _ []string) (string, error) {
	if _, err := b.git.Run(ctx, repoPath, "reset"); err != nil {
		return "", err
	}
	return "All staged changes reset", nil
}

// CreateBranch creates and immediately checks out a new branch.
// If baseBranch is empty, the new branch points at the current HEAD.
func (b *Backend) CreateBranch(ctx context.Context, repoPath, name, baseBranch string) (string, error) {
	if baseBranch != "" {
		sha, err := b.git.ResolveCommit(ctx, repoPath, baseBranch)
		if err != nil {
			return "", err
		}
		if _, err := b.git.Run(ctx, repoPath, "checkout", "-b", name, sha); err != nil {
			return "", err
		}
	} else {
		if _, err := b.git.Run(ctx, repoPath, "checkout", "-b", name); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("Created and switched to new branch '%s'", name), nil
}

// Checkout switches branches. The target is validated as a commit-ish
// first; the checkout itself uses the user-provided name (a SHA would
// detach HEAD, which the reference does not do).
func (b *Backend) Checkout(ctx context.Context, repoPath, branch string) (string, error) {
	if _, err := b.git.ResolveCommit(ctx, repoPath, branch); err != nil {
		return "", err
	}
	if _, err := b.git.Run(ctx, repoPath, "checkout", branch, "--"); err != nil {
		return "", err
	}
	return fmt.Sprintf("Switched to branch '%s'", branch), nil
}

// Push is the write-access extension over the Python reference.
func (b *Backend) Push(ctx context.Context, repoPath, remote, branch string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	args := []string{"push", remote}
	if branch != "" {
		args = append(args, branch)
	}
	return b.git.Run(ctx, repoPath, args...)
}

// unified formats a context_lines value as a git flag.
// The reference default is 3.
func unified(n int) string {
	if n <= 0 {
		n = 3
	}
	return fmt.Sprintf("--unified=%d", n)
}
