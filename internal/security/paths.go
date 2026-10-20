package security

import (
	"path/filepath"
	"strings"
)

// CheckAllowed verifies that repoPath is under one of the allowed
// directories. An empty allowDir list means "no allowlist configured";
// in that case the caller is responsible for enforcing access by
// running the process under a restricted user account.
//
// Comparison is done on canonical absolute paths (symlinks resolved)
// so that a symlink inside an allowed root cannot escape it.
func CheckAllowed(allowDir []string, repoPath string) error {
	if len(allowDir) == 0 {
		return nil
	}
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return NewViolation(ViolationPathTraversal, "repo_path", err.Error())
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return NewViolation(ViolationPathTraversal, "repo_path", err.Error())
	}
	for _, root := range allowDir {
		rel, err := filepath.Rel(root, real)
		if err != nil {
			continue
		}
		if rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))) {
			return nil
		}
	}
	return NewViolation(ViolationPathTraversal, "repo_path",
		"not under any allowed directory")
}

// CanonicalizeAllowDir normalizes an allowlist entry once at startup.
// It is applied to each --allow-dir argument before the server begins
// serving requests.
func CanonicalizeAllowDir(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}
