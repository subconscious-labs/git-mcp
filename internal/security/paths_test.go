package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckAllowed_EmptyAllowlist(t *testing.T) {
	if err := CheckAllowed(nil, "/anywhere"); err != nil {
		t.Fatalf("empty allowlist should permit anything: %v", err)
	}
}

func TestCheckAllowed_AllowsUnderRoot(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "repo")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckAllowed([]string{realRoot}, sub); err != nil {
		t.Fatalf("should be allowed: %v", err)
	}
}

func TestCheckAllowed_RejectsOutside(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckAllowed([]string{realRoot}, other); err == nil {
		t.Fatal("outside path should be rejected")
	}
}

func TestCheckAllowed_RejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip("symlink not supported:", err)
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckAllowed([]string{realRoot}, link); err == nil {
		t.Fatal("symlink escape should be rejected")
	}
}

func TestCanonicalizeAllowDir(t *testing.T) {
	root := t.TempDir()
	got, err := CanonicalizeAllowDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got == "" {
		t.Fatal("empty canonical path")
	}
}
