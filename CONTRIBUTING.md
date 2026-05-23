# Contributing

[中文](CONTRIBUTING.zh-CN.md)

Thanks for helping improve AgentMeter.

## Before You Submit

Run the relevant checks:

```bash
make test
make build
```

For a full local check:

```bash
make validate
```

If Make is not available, run:

```bash
go test ./...
go build .
```

## Change Guidelines

- Parser changes need tests with small sanitized fixtures.
- Web UI changes should support both English and Chinese labels.
- CLI output changes should update tests and README examples when needed.
- Release workflow changes should keep `.goreleaser.yml` and `.github/workflows/release.yml` aligned.
- Keep the tool offline-first. Do not add network upload, hidden telemetry, or account connections.

## Privacy Rules

Do not include real user data in issues, tests, docs, or commits:

- API keys
- prompts
- private file paths
- full local logs
- account identifiers

Use tiny synthetic fixtures instead.

## Pull Requests

Please include:

- What changed
- How it was tested
- Any known risks or limitations
- Whether README or release notes need updates
