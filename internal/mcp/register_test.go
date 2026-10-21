package mcp

import (
	"testing"

	"github.com/subconscious-labs/git-mcp/internal/config"
)

// TestRegisterTools_ByMode pins the tool count per mode. This is the
// gate that catches mode-gating regressions.
func TestRegisterTools_ByMode(t *testing.T) {
	cases := []struct {
		mode config.Mode
		want int
	}{
		{config.ModeRead, 7},
		{config.ModeLocalWrite, 12},
		{config.ModeRemoteWrite, 15},
	}
	for _, c := range cases {
		t.Run(string(c.mode), func(t *testing.T) {
			if got := len(ToolNamesByMode(c.mode)); got != c.want {
				t.Fatalf("mode %q: got %d tools, want %d", c.mode, got, c.want)
			}
		})
	}
}

// TestRegisterTools_ExtensionGating verifies that extension tools appear
// only in remote-write.
func TestRegisterTools_ExtensionGating(t *testing.T) {
	type tc struct {
		mode    config.Mode
		present []string
		absent  []string
	}
	cases := []tc{
		{
			mode:    config.ModeRead,
			present: []string{"git_status", "git_show", "git_branch"},
			absent:  []string{"git_commit", "git_fetch", "git_push"},
		},
		{
			mode:    config.ModeLocalWrite,
			present: []string{"git_commit", "git_add", "git_checkout"},
			absent:  []string{"git_fetch", "git_pull", "git_push"},
		},
		{
			mode:    config.ModeRemoteWrite,
			present: []string{"git_fetch", "git_pull", "git_push"},
			absent:  []string{},
		},
	}
	for _, c := range cases {
		t.Run(string(c.mode), func(t *testing.T) {
			have := map[string]bool{}
			for _, n := range ToolNamesByMode(c.mode) {
				have[n] = true
			}
			for _, name := range c.present {
				if !have[name] {
					t.Errorf("mode %q: missing %q", c.mode, name)
				}
			}
			for _, name := range c.absent {
				if have[name] {
					t.Errorf("mode %q: %q should be absent", c.mode, name)
				}
			}
		})
	}
}

// TestToolNamesByMode_NoDuplicates guards against accidental duplication
// when RegisterTools and ToolNamesByMode drift apart.
func TestToolNamesByMode_NoDuplicates(t *testing.T) {
	for _, m := range config.KnownModes {
		seen := map[string]bool{}
		for _, n := range ToolNamesByMode(m) {
			if seen[n] {
				t.Errorf("mode %q: duplicate tool %q", m, n)
			}
			seen[n] = true
		}
	}
}
