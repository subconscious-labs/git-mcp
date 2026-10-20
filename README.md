# git-mcp

A drop-in compatible MCP server for Git, written in Go.

The wire contract — tool names, argument schemas (including the per-call
`repo_path` field), and output formats — matches the Python reference
`mcp-server-git`, so existing clients work unchanged when the server is
run with `--mode=remote-write`.

## Status

Phase 3 complete.

- **12 tools** from the Python reference, plus `git_push` as an extension.
- **Tool exposure is mode-gated**: `read` (7), `local-write` (12),
  `remote-write` (13). Default is `read`.
- **Path allowlist** via `--allow-dir` (repeatable).
- **Flag-injection protection** on every positional argument.
- **Output sanitization**: secrets (AWS, GitHub tokens, JWTs, PEM keys,
  key=value credentials) are redacted; prompt-injection patterns are
  wrapped in `<<<UNTRUSTED>>>` markers.
- **`safe.directory`** is scoped per invocation; the user's global
  gitconfig is never modified.

The `go-git` backend is planned for Phase 4. `git_fetch` and `git_pull`
are also planned for Phase 4 as extensions over the reference contract.

## Build

```sh
make build
```

The binary is emitted at `bin/git-mcp`. It is statically linked
(`CGO_ENABLED=0`) so it can be shipped as a single file.

## Run

```sh
bin/git-mcp --mode=remote-write --allow-dir=/opt/subconscious
```

The server speaks the MCP protocol over stdio. **All diagnostics are
written to stderr**; stdout is reserved for the protocol.

## Configuration

| Precedence | Source | Variable |
|---|---|---|
| 1 (highest) | CLI flag | `--mode=read\|local-write\|remote-write` |
| 2 | Environment | `GIT_MCP_MODE` |
| 3 (lowest) | Default | `read` |

Same precedence for `--backend` / `GIT_MCP_BACKEND` and
`--allow-dir` / `GIT_MCP_ALLOW_DIR` (repeatable flag; env is
`PATH_LIST_SEPARATOR`-joined).

### Modes

| Mode | Tools exposed | Write endpoints reachable |
|---|---|---|
| `read` | 7 | none |
| `local-write` | 12 | add, commit, reset, create_branch, checkout |
| `remote-write` | 13 | + push |

## Security

- **Path allowlist**: every `repo_path` must resolve under one of the
  `--allow-dir` roots (symlinks are resolved before comparison). An
  empty allowlist disables the check; run the process under a restricted
  user in that case.
- **Positional validation**: any argument that looks like a git flag
  (leading `-`) or contains control characters is rejected before
  reaching git.
- **Output sanitization**: known secret patterns are redacted;
  prompt-injection markers are wrapped.

## Development

```sh
make lint           # golangci-lint
make test           # unit tests
make security-test  # CVE / traversal / redaction regressions
make ci             # full gate before merge
```

## License

See `LICENSE`.
