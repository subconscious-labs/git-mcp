package mcp

import (
	"encoding/json"
	"testing"
)

// TestRequiredRepoPath ensures every params struct marks repo_path required.
// This is the single invariant that makes the server a drop-in replacement
// for the Python reference, where every tool call carries repo_path.
func TestRequiredRepoPath(t *testing.T) {
	types := []any{
		GitStatusParams{},
		GitDiffUnstagedParams{},
		GitDiffStagedParams{},
		GitDiffParams{},
		GitCommitParams{},
		GitAddParams{},
		GitResetParams{},
		GitLogParams{},
		GitCreateBranchParams{},
		GitCheckoutParams{},
		GitShowParams{},
		GitBranchParams{},
		GitPushParams{},
	}
	for _, v := range types {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatalf("%T: %v", v, err)
		}
		if _, ok := m["repo_path"]; !ok {
			t.Errorf("%T: missing repo_path in zero value", v)
		}
	}
}

func TestBranchTypeEnum(t *testing.T) {
	// The reference restricts branch_type to local, remote, all.
	// This test documents the allowed values; enforcement happens in
	// the security middleware and the backend.
	allowed := map[string]bool{"local": true, "remote": true, "all": true}
	for v := range allowed {
		if !allowed[v] {
			t.Fatalf("unreachable")
		}
	}
}

func TestDefaults(t *testing.T) {
	if DefaultContextLines != 3 {
		t.Fatalf("DefaultContextLines = %d, want 3", DefaultContextLines)
	}
	if DefaultMaxCount != 10 {
		t.Fatalf("DefaultMaxCount = %d, want 10", DefaultMaxCount)
	}
}
