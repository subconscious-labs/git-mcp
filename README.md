# git-mcp

A drop-in compatible MCP server for Git, written in Go.

The wire contract — tool names, argument schemas (including the per-call
`repo_path` field), and output formats — matches the Python reference
`mcp-server-git`, so existing clients work unchanged.

## Status

Phase 0 of the development plan is complete: the skeleton, configuration
loader, backend interface and security error types are in place. Tool
handlers land in Phase 2.

## Build

```sh
make build
```

The binary is emitted at `bin/git-mcp`. It is statically linked
(`CGO_ENABLED=0`) so it can be shipped as a single file.

## Run

```sh
bin/git-mcp --backend=exec
```

The server speaks the MCP protocol over stdio. **All diagnostics are
written to stderr**; stdout is reserved for the protocol.

## Configuration

| Precedence | Source | Variable |
|---|---|---|
| 1 (highest) | CLI flag | `--backend=exec\|go-git` |
| 2 | Environment | `GIT_MCP_BACKEND` |
| 3 (lowest) | Default | `exec` |

## Development

```sh
make lint           # golangci-lint
make test           # unit tests
make security-test  # CVE / traversal / redaction regressions
make ci             # full gate before merge
```

## License

See `LICENSE`.
