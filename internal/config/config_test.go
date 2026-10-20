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
		{name: "flag wins", flag: "exec", env: "go-git", want: BackendExec},
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

func TestResolveMode(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		env     string
		want    Mode
		wantErr bool
	}{
		{name: "default", want: ModeRead},
		{name: "flag local-write", flag: "local-write", want: ModeLocalWrite},
		{name: "env remote-write", env: "remote-write", want: ModeRemoteWrite},
		{name: "flag wins", flag: "read", env: "remote-write", want: ModeRead},
		{name: "bad flag", flag: "write", wantErr: true},
		{name: "bad env", env: "full", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveMode(tt.flag, tt.env)
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

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(nil, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != ModeRead {
		t.Fatalf("default mode = %q, want %q", cfg.Mode, ModeRead)
	}
	if cfg.Backend != BackendExec {
		t.Fatalf("default backend = %q", cfg.Backend)
	}
	if len(cfg.AllowDir) != 0 {
		t.Fatalf("default allow-dir should be empty, got %v", cfg.AllowDir)
	}
}

func TestLoadModeFromFlag(t *testing.T) {
	cfg, err := Load([]string{"--mode=remote-write"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != ModeRemoteWrite {
		t.Fatalf("got %q", cfg.Mode)
	}
}

func TestLoadAllowDirRepeated(t *testing.T) {
	cfg, err := Load([]string{"--allow-dir=/a", "--allow-dir=/b"}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.AllowDir) != 2 || cfg.AllowDir[0] != "/a" || cfg.AllowDir[1] != "/b" {
		t.Fatalf("got %v", cfg.AllowDir)
	}
}
