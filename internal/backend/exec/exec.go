// Package exec implements GitBackend by shelling out to the system git.
// It is the reference backend: its output formats match the Python
// mcp-server-git implementation byte-for-byte wherever practical.
//
// Every positional argument that comes from the client is validated by
// security.ValidatePositional before it reaches git, closing the
// CVE-2025-68144 class of flag injections.
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

func (b *Backend) Status(ctx context.Context, repoPath string) (string, error) {
	return b.git.Run(ctx, repoPath, "status")
}

func (b *Backend) DiffUnstaged(ctx context.Context, repoPath string, contextLines int) (string, error) {
	return b.git.Run(ctx, repoPath, "diff", unified(contextLines))
}

func (b *Backend) DiffStaged(ctx context.Context, repoPath string, contextLines int) (string, error) {
	return b.git.Run(ctx, repoPath, "diff", "--cached", unified(contextLines))
}

func (b *Backend) Diff(ctx context.Context, repoPath, target string, contextLines int) (string, error) {
	sha, err := b.git.ResolveCommit(ctx, repoPath, target)
	if err != nil {
		return "", err
	}
	return b.git.Run(ctx, repoPath, "diff", sha, unified(contextLines))
}

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
		if err := security.ValidatePositional("path", path); err != nil {
			return "", err
		}
		target = sha + ":" + path
	}
	return b.git.Run(ctx, repoPath, "show", target)
}

func (b *Backend) Branch(ctx context.Context, repoPath string, opts backend.BranchOptions) (string, error) {
	args := []string{"branch", "--format=%(refname:short)%(if)%(HEAD)%(then) (current)%(end)"}
	switch opts.Type {
	case "local", "":
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

func (b *Backend) Add(ctx context.Context, repoPath string, paths []string) (string, error) {
	for i, p := range paths {
		if err := security.ValidatePositional(fmt.Sprintf("files[%d]", i), p); err != nil {
			return "", err
		}
	}
	args := append([]string{"add", "--"}, paths...)
	if _, err := b.git.Run(ctx, repoPath, args...); err != nil {
		return "", err
	}
	return "Files staged successfully", nil
}

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

func (b *Backend) Reset(ctx context.Context, repoPath string, _ []string) (string, error) {
	if _, err := b.git.Run(ctx, repoPath, "reset"); err != nil {
		return "", err
	}
	return "All staged changes reset", nil
}

func (b *Backend) CreateBranch(ctx context.Context, repoPath, name, baseBranch string) (string, error) {
	if err := security.ValidatePositional("branch_name", name); err != nil {
		return "", err
	}
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

func (b *Backend) Checkout(ctx context.Context, repoPath, branch string) (string, error) {
	if _, err := b.git.ResolveCommit(ctx, repoPath, branch); err != nil {
		return "", err
	}
	if _, err := b.git.Run(ctx, repoPath, "checkout", branch, "--"); err != nil {
		return "", err
	}
	return fmt.Sprintf("Switched to branch '%s'", branch), nil
}

// Fetch updates remote-tracking refs from the given remote. It does not
// touch local branches, tags, or the working tree.
func (b *Backend) Fetch(ctx context.Context, repoPath, remote string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	if err := security.ValidatePositional("remote", remote); err != nil {
		return "", err
	}
	stdout, stderr, err := b.git.RunWithStderr(ctx, repoPath, "fetch", remote)
	if err != nil {
		return "", err
	}
	if msg := summarizeStreams(stdout, stderr); msg != "" {
		return fmt.Sprintf("Fetched from '%s': %s", remote, msg), nil
	}
	return fmt.Sprintf("Fetched from '%s'", remote), nil
}

// Pull fast-forwards the current branch from the given remote. It uses
// --ff-only: no merge commits are ever created, no rebase is performed.
// If the branch cannot be fast-forwarded, git errors and the caller is
// expected to surface it.
func (b *Backend) Pull(ctx context.Context, repoPath, remote, branch string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	if err := security.ValidatePositional("remote", remote); err != nil {
		return "", err
	}
	args := []string{"pull", "--ff-only", remote}
	if branch != "" {
		if err := security.ValidatePositional("branch", branch); err != nil {
			return "", err
		}
		args = append(args, branch)
	}
	stdout, stderr, err := b.git.RunWithStderr(ctx, repoPath, args...)
	if err != nil {
		return "", err
	}
	if msg := summarizeStreams(stdout, stderr); msg != "" {
		return msg, nil
	}
	return "Already up to date.", nil
}

// Push pushes the given refspec. Both remote and branch are validated as
// positional arguments before reaching git. When git reports success but
// writes "Everything up-to-date" to stderr, that message is returned
// instead of an empty string.
func (b *Backend) Push(ctx context.Context, repoPath, remote, branch string) (string, error) {
	if remote == "" {
		remote = "origin"
	}
	if err := security.ValidatePositional("remote", remote); err != nil {
		return "", err
	}
	args := []string{"push", remote}
	if branch != "" {
		if err := security.ValidatePositional("branch", branch); err != nil {
			return "", err
		}
		args = append(args, branch)
	}
	stdout, stderr, err := b.git.RunWithStderr(ctx, repoPath, args...)
	if err != nil {
		return "", err
	}
	if msg := summarizeStreams(stdout, stderr); msg != "" {
		return msg, nil
	}
	return "Everything up-to-date", nil
}

// summarizeStreams returns a compact, single-line summary of the combined
// stdout and stderr of a git command. Used for fetch/pull/push where git
// may report useful information on either stream.
func summarizeStreams(stdout, stderr string) string {
	combined := strings.TrimSpace(stdout) + "\n" + strings.TrimSpace(stderr)
	combined = strings.TrimSpace(combined)
	if combined == "\n" {
		return ""
	}
	lines := strings.Split(combined, "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func unified(n int) string {
	if n <= 0 {
		n = 3
	}
	return fmt.Sprintf("--unified=%d", n)
}
