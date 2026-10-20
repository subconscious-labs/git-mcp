package config

import (
	"io"
	"testing"
)

func TestResolveBackend(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		env     string
		want    Backend
		wantErr bool
	}{
		{name: "default", want: BackendExec},
		{name: "flag only", flag: "go-git", want: BackendGoGit},
		{name: "env only", env: "go-git", want: BackendGoGit},
		{name: "flag wins over env", flag: "exec", env: "go-git", want: BackendExec},
		{name: "bad flag", flag: "svn", wantErr: true},
		{name: "bad env", env: "svn", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBackend(tt.flag, tt.env)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestLoadFromArgs(t *testing.T) {
	cfg, err := Load([]string{"--backend=go-git"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Backend != BackendGoGit {
		t.Fatalf("got %q", cfg.Backend)
	}
}
