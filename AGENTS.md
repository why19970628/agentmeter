# AGENTS

[中文](AGENTS.zh-CN.md)

This file guides AI coding agents working in this repository.

## Project Mission

AgentMeter is an offline-first Go tool for reading local AI agent usage logs and reporting token usage across Codex, Claude Code, Cursor, Gemini CLI, and compatible local records.

The project must stay local-first:

- Do not upload usage data.
- Do not read or print API keys.
- Do not add hidden telemetry or background sync.
- Do not include real user paths, prompts, or private log samples in docs or tests.

## Project Shape

- Go module at the repository root.
- CLI and web server entrypoint: `main.go`.
- Usage scanning and aggregation: `internal/usage`.
- CLI report rendering: `internal/report`.
- Web dashboard server: `internal/web`.
- Web UI: `templates` and `static`.
- Sample data: `examples`.

## Common Commands

Use Makefile targets when possible:

```bash
make test
make build
make validate
make run ARGS="summary -period all -group model -paths examples"
```

Targeted checks are preferred while iterating:

```bash
go test ./internal/usage
go test ./internal/web
```

## Development Rules

- Keep parsing tolerant of missing, nested, or provider-specific fields.
- Add tests for parser changes, especially missing fields, nested records, malformed logs, and session-context inference.
- Preserve stable CLI output where practical. If output columns or JSON fields change, update tests and docs.
- Web changes should keep English as the default UI and support Chinese switching.
- Keep release assets self-contained enough to run the web dashboard, including templates and static files.
- Avoid new dependencies unless they clearly improve reliability without hurting offline behavior or binary simplicity.

## Verification

- Parser or aggregation changes: run `go test ./internal/usage`.
- Web UI/server changes: run `go test ./internal/web`.
- Report/CLI changes: run `go test ./internal/report` and a small CLI smoke when useful.
- Release or CI changes: run `make validate` when available, or at least `go test ./...`.

Before handoff or commit, run:

```bash
go test ./...
```

## CI and Submission Protocol

- CI must run `make validate` and `make test-harness` when those targets exist. If `test-harness` is not available in this repository, keep CI on `make validate` and document the exception before adding a harness target.
- PRs must include Summary, Requirement Classification, Acceptance Criteria, Changed Areas, Release Decision, TDD / Test Evidence, Validation, and Risk and Rollback.
- Feature and bugfix PRs must mark that a release is required after merge, or state that the user explicitly approved deferring the release.
- Harness/tooling, docs, CI, and other engineering-process-only PRs should mark release not required unless they affect shipped behavior.
- Commit messages should follow `{emoji} {type}{scope}: {subject}` and the emoji must match the commit type:
  - `✨ feat`
  - `🐛 fix`
  - `📝 docs`
  - `👷 ci`
  - `💄 style`
  - `♻️ refactor`
  - `🔖 release`
  - `⚡️ perf`
  - `✅ test`
  - `🔧 chore`
  - `🏗️ build`
- If commit hooks or commitlint are present, run `make setup` once to enable `.githooks/commit-msg`, or run `make commitlint COMMIT_MSG_FILE=<commit-msg-file>` directly. PR CI should validate commit messages in the PR range when commitlint exists.
- Do not claim completion after skipping required local validation.
- Stage and commit only files that belong to the current iteration. Do not include unrelated dirty files.

## Release Notes

Release descriptions should describe user-visible changes, not internal local paths or private test commands. Do not include real machine paths in README, release notes, or examples.
