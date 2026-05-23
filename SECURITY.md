# Security Policy

[中文](SECURITY.zh-CN.md)

## Supported Versions

Security fixes target the latest released version and the `main` branch.

## Reporting a Vulnerability

Please avoid public disclosure before maintainers have time to investigate.

Use GitHub private vulnerability reporting when available, or open an issue with a minimal sanitized reproduction.

## Data and Privacy Boundaries

AgentMeter is offline-first. Security-sensitive changes must preserve these rules:

- Do not upload local logs, prompts, API keys, token usage, or project paths by default.
- Do not read, print, hash, fingerprint, or persist raw API keys.
- Do not add hidden telemetry, background sync, or account connections.
- Prefer small synthetic fixtures over real user logs.
- Remove private paths, prompts, and personal data from bug reports.

## Report Contents

A useful report includes:

- affected version or commit
- operating system
- reproduction steps
- expected behavior
- actual behavior
- impact

Please remove secrets and personal data before sharing logs or fixtures.
