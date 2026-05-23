# AgentMeter

[中文](README.zh-CN.md)

AgentMeter is an offline-first local usage meter for AI agent tools. It reads local logs, summarizes token usage, estimates cost, and provides both a script-friendly CLI and a built-in Web dashboard.

It does not upload usage data, read API keys, or require a backend service. Everything is computed from files on your machine.

## Features

- Local usage summaries for Codex, Claude Code, Cursor, Gemini CLI, and generic JSON/JSONL logs.
- CLI output in `table`, `json`, and `markdown` formats.
- Built-in Web dashboard with token trends, daily API request counts, rankings, and model usage tables.
- Token breakdowns for input, output, cache read, cache write, reasoning, and tool tokens.
- Model usage table grouped by day, tool, and model.
- Estimated USD cost based on a built-in offline model price catalog.
- CSV export from the Web dashboard.
- No server-side dependency, account connection, or telemetry upload.

## Quick Start

Run the Web dashboard with sample data:

```bash
cd agentmeter
go run . serve -paths examples
```

Then open:

```text
http://127.0.0.1:8787
```

Run a CLI summary:

```bash
cd agentmeter
go run . summary -period all -group model -paths examples
```

> Note: if your local Go environment has a custom `GOROOT` or cache policy, set those variables in your shell before running the examples.

## CLI Usage

Summary by source:

```bash
go run . summary -period month -paths ~/.codex,~/.claude
```

Summary by model:

```bash
go run . summary -period all -group model -paths ~/.codex,~/.claude,~/.gemini
```

JSON output for scripts:

```bash
go run . summary -period week -format json -paths ~/.codex
```

Chinese CLI headers:

```bash
go run . summary -period all -group model -lang zh-CN -paths examples
```

Markdown output:

```bash
go run . summary -period month -format markdown -group model -paths examples
```

Scan-only summary:

```bash
go run . scan -paths examples
```

Start the Web dashboard:

```bash
go run . serve -addr 127.0.0.1:8787 -paths ~/.codex,~/.claude,~/.cursor,~/.gemini
```

## Periods

AgentMeter currently supports:

- `today`
- `week`
- `month`
- `all`

## Output Formats

The CLI supports:

- `table`
- `json`
- `markdown`

Example table output:

```text
Model            Input Tokens  Output Tokens  Cache Read  Cache Write  Reasoning  Tool Tokens  Total Tokens  Est. USD
gpt-5            142,612       32,110         8,100       1,600        3,420      900          180,642       $0.499
claude-sonnet-4  121,000       29,200         23,100      4,200        0          0            154,400       $0.801
Total            263,612       61,310         31,200      5,800        3,420      900          335,042       $1.300
```

## Web Dashboard

The Web dashboard is designed for local inspection:

- Summary cards for total tokens, input/output tokens, cache tokens, reasoning tokens, request count, and model count.
- Token usage trend with compact daily values.
- Daily API request count line chart with per-day request labels.
- Tool and project rankings; project ranking shows project names only and scrolls when many projects exist.
- Model usage details grouped by `day + tool + model`.
- CSV export for further analysis.
- English UI by default, with a one-click Chinese/English toggle. You can also use `?lang=zh-CN` or `?lang=en`.

## Data Sources

By default, AgentMeter looks at common local agent directories:

```text
~/.codex
~/.claude
~/.cursor
~/.gemini
```

You can override paths explicitly:

```bash
go run . summary -paths ~/.codex,~/logs/agent-usage
```

AgentMeter recursively reads:

- `.json`
- `.jsonl`
- `.log`

## Supported Usage Fields

AgentMeter recognizes flat and nested usage records, including:

```json
{
  "tool": "codex",
  "model": "gpt-5",
  "session_id": "s1",
  "project": "/path/to/project",
  "usage": {
    "input_tokens": 120,
    "output_tokens": 30,
    "cache_read_input_tokens": 20,
    "cache_creation_input_tokens": 10,
    "reasoning_tokens": 5,
    "tool_tokens": 2
  },
  "timestamp": "2026-05-03T10:00:00Z"
}
```

It also understands common vendor-specific shapes such as:

```json
{
  "usage": {
    "input_tokens_details": {
      "cached_tokens": 15
    },
    "output_tokens_details": {
      "reasoning_tokens": 12
    },
    "cache_creation": {
      "ephemeral_5m_input_tokens": 6,
      "ephemeral_1h_input_tokens": 4
    }
  }
}
```

Model and project fields are resolved from common aliases such as `model_id`, `modelId`, `model_slug`, `workspace_path`, `current_working_directory`, and `root_path`. AgentMeter also uses `session_id` context within a log file to attach later or earlier model/project metadata to usage rows.

## Cost Estimate

`Est. USD` means estimated USD cost. It is calculated offline from token counts and a built-in model price catalog.

It is not an official bill. If a model has no matching price, cost may be shown as `0`.

## Privacy

AgentMeter is local-first:

- No usage uploads.
- No API key reads.
- No account connection.
- No remote backend.

The only files read are the paths you provide or the default local agent log directories.

## Development

Run tests:

```bash
cd agentmeter
go test ./...
```

Run with sample data:

```bash
go run . serve -paths examples
```

## Releases

GitHub releases are built with GoReleaser when a `v*` tag is pushed. Release assets include macOS, Linux, and Windows archives plus `checksums.txt`.

Create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Build a local release snapshot:

```bash
make release-snapshot
```

## Status

AgentMeter is an early local tool. The current focus is reliable local parsing, clear usage summaries, and a lightweight dashboard. Future improvements may include data-source diagnostics, custom pricing overrides, session/thread views, and richer provider attribution.
