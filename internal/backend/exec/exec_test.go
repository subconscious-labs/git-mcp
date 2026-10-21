package exec

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/subconscious-labs/git-mcp/internal/security"
)

// testEnv returns the environment for test git commands: no system or
// global gitconfig (so commit.gpgsign, hooks, and user.name/email from
// the developer's machine never interfere), no interactive prompts.
func testEnv() []string {
	return append(os.Environ(),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_TERMINAL_PROMPT=0",
		"GIT_PAGER=",
		"GIT_ASKPASS=/bin/false",
	)
}

// mustGit runs git in dir with a hermetic environment. Commit-related
// flags disable GPG signing so that systems with commit.gpgsign=true do
// not prompt for a passphrase during tests.
func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = testEnv()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// setup creates a bare "remote", a working "work" repo that pushes the
// first commit, and a "clone". Returns remote and clone paths.
func setup(t *testing.T) (remote, clone string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	remote = filepath.Join(dir, "remote.git")
	clone = filepath.Join(dir, "clone")
	work := filepath.Join(dir, "work")

	mustGit(t, dir, "init", "--bare", remote)
	mustGit(t, dir, "init", "-b", "main", work)
	writeFile(t, filepath.Join(work, "README.md"), "hello\n")
	mustGit(t, work, "add", "README.md")
	mustGit(t, work,
		"-c", "user.email=t@e", "-c", "user.name=T", "-c", "commit.gpgsign=false",
		"commit", "-m", "init")
	mustGit(t, work, "remote", "add", "origin", remote)
	mustGit(t, work, "push", "-u", "origin", "main")

	mustGit(t, dir, "clone", remote, clone)
	mustGit(t, clone, "checkout", "main")
	return remote, clone
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func newBackend(t *testing.T) *Backend {
	t.Helper()
	git, err := security.NewSafeGit()
	if err != nil {
		t.Fatal(err)
	}
	return New(git)
}

func TestFetch(t *testing.T) {
	remote, clone := setup(t)
	work := filepath.Join(filepath.Dir(remote), "work")
	writeFile(t, filepath.Join(work, "new.txt"), "x\n")
	mustGit(t, work, "add", "new.txt")
	mustGit(t, work,
		"-c", "user.email=t@e", "-c", "user.name=T", "-c", "commit.gpgsign=false",
		"commit", "-m", "new")
	mustGit(t, work, "push", "origin", "main")

	out, err := newBackend(t).Fetch(context.Background(), clone, "origin")
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if out == "" {
		t.Fatal("empty output")
	}
	ref := exec.Command("git", "-C", clone, "rev-parse", "origin/main")
	ref.Env = testEnv()
	if err := ref.Run(); err != nil {
		t.Fatalf("origin/main missing after fetch: %v", err)
	}
}

func TestPull_FFOnly(t *testing.T) {
	remote, clone := setup(t)
	work := filepath.Join(filepath.Dir(remote), "work")
	writeFile(t, filepath.Join(work, "ff.txt"), "y\n")
	mustGit(t, work, "add", "ff.txt")
	mustGit(t, work,
		"-c", "user.email=t@e", "-c", "user.name=T", "-c", "commit.gpgsign=false",
		"commit", "-m", "ff")
	mustGit(t, work, "push", "origin", "main")

	out, err := newBackend(t).Pull(context.Background(), clone, "origin", "main")
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if out == "" {
		t.Fatal("empty output")
	}
}

func TestPull_DivergedRejected(t *testing.T) {
	remote, clone := setup(t)
	work := filepath.Join(filepath.Dir(remote), "work")
	writeFile(t, filepath.Join(work, "w.txt"), "w\n")
	mustGit(t, work, "add", "w.txt")
	mustGit(t, work,
		"-c", "user.email=t@e", "-c", "user.name=T", "-c", "commit.gpgsign=false",
		"commit", "-m", "w")
	mustGit(t, work, "push", "origin", "main")

	writeFile(t, filepath.Join(clone, "c.txt"), "c\n")
	mustGit(t, clone, "add", "c.txt")
	mustGit(t, clone,
		"-c", "user.email=t@e", "-c", "user.name=T", "-c", "commit.gpgsign=false",
		"commit", "-m", "c")

	_, err := newBackend(t).Pull(context.Background(), clone, "origin", "main")
	if err == nil {
		t.Fatal("expected divergence error")
	}
	if !strings.Contains(err.Error(), "fast-forward") && !strings.Contains(err.Error(), "Not possible") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPush_EverythingUpToDate(t *testing.T) {
	_, clone := setup(t)
	out, err := newBackend(t).Push(context.Background(), clone, "origin", "main")
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if out == "" {
		t.Fatal("expected non-empty output, got empty (regression from v0.3.0)")
	}
	low := strings.ToLower(out)
	if !strings.Contains(low, "up-to-date") && !strings.Contains(low, "up to date") {
		t.Fatalf("expected up-to-date message, got %q", out)
	}
}

func TestFetch_RejectsFlagInjection(t *testing.T) {
	_, clone := setup(t)
	_, err := newBackend(t).Fetch(context.Background(), clone, "--upload-pack=/tmp/pwn")
	if err == nil {
		t.Fatal("expected flag-injection rejection")
	}
	if !strings.Contains(err.Error(), "flag_injection") {
		t.Fatalf("wrong error: %v", err)
	}
}
