## [0.2.0] - 2026-10-20

### ⚠️ Breaking changes
- Remove `--write-access` flag; use `--mode={read,local-write,remote-write}`.
- Default mode is now `read` (7 tools), not all 13. Pass
  `--mode=remote-write` for full drop-in parity with the Python reference.
- `GIT_MCP_WRITE_ACCESS` env var is ignored; use `GIT_MCP_MODE`.

### 🚀 Features
- `--mode` flag and `GIT_MCP_MODE` env (precedence: flag > env > default).
- `--allow-dir` flag (repeatable) and `GIT_MCP_ALLOW_DIR` env; symlink-resolved
  path confinement for every `repo_path`.
- Output sanitization: AWS keys, GitHub tokens, JWTs, PEM blocks and
  `key=value` credentials are redacted; prompt-injection patterns are wrapped
  in `<<<UNTRUSTED>>>` markers.

### 🐛 Bug fixes
- `git_push` and `git_checkout` no longer pass user-supplied refspec/branch
  to git without validation (flag-injection class CVE-2025-68144).
- `git_show` accepts `<commit>:<path>` and lists directory trees.
- Root-owned repositories are readable without touching global gitconfig
  (`-c safe.directory=<repo>` per invocation).

### 🔒 Security
- Every positional argument (branch, target, path, remote, refspec, revision)
  is validated against leading `-` and control characters before reaching git.

