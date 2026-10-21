# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

While the project is in the `0.x` series, a breaking change increments the
minor version.

## [Unreleased]

## [0.3.1] - 2026-10-21

### Added
- `--version` flag; prints `git-mcp <version>` to stderr and exits.
- Integration tests for `git_fetch`, `git_pull` (`--ff-only`) and
  `git_push`, using a real bare remote in `t.TempDir()`.
- Mode-gating tests that pin the tool count per mode (7/12/15).

### Fixed
- `git_push` no longer returns an empty string when git reports
  "Everything up-to-date"; the message is now surfaced from stderr.
- `git_fetch` and `git_pull` surface git's progress and summary lines
  instead of returning an empty body.
- `--help` exits with status 0 instead of failing with
  `fatal: config: parse flags: flag: help requested`.

## [0.3.0] - 2026-10-21

### Added
- `git_fetch` and `git_pull` extension tools, registered only in
  `--mode=remote-write`. `git_pull` uses `--ff-only` and never creates
  merge commits or rebases.
- `SafeGit.RunWithStderr` (internal) for commands whose stderr carries
  meaningful output on success.

### Changed
- Remote and branch arguments of `git_push` are validated through
  `security.ValidatePositional`, closing the refspec-injection surface
  identified in the v0.2.0 audit.
- Tool count in `--mode=remote-write` is now 15 (was 13).

## [0.2.0] - 2026-10-20

### Breaking changes
- Remove `--write-access` flag; use `--mode={read,local-write,remote-write}`.
- Default mode is now `read` (7 tools), not all 13. Pass
  `--mode=remote-write` for full drop-in parity with the Python reference.
- `GIT_MCP_WRITE_ACCESS` env var is ignored; use `GIT_MCP_MODE`.

### Added
- `--mode` flag and `GIT_MCP_MODE` env (precedence: flag > env > default).
- `--allow-dir` flag (repeatable) and `GIT_MCP_ALLOW_DIR` env;
  symlink-resolved path confinement for every `repo_path`.
- Output sanitization: AWS keys, GitHub tokens, JWTs, PEM blocks and
  `key=value` credentials are redacted; prompt-injection patterns are
  wrapped in `<<<UNTRUSTED>>>` markers.

### Fixed
- `git_push` and `git_checkout` no longer pass user-supplied
  refspec/branch to git without validation (CVE-2025-68144 class).
- `git_show` accepts `<commit>:<path>` and lists directory trees.
- Root-owned repositories are readable without touching global gitconfig
  (`-c safe.directory=<repo>` per invocation).

## [0.1.0] - 2026-10-20

### Added
- Initial release: MCP server with stdio transport, 12 tools from the
  Python reference, exec backend, hardened SafeGit runner, contract
  tests.
