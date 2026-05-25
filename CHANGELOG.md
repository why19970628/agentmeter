# Changelog

All notable changes to AgentMeter are documented here.

This project follows semantic versioning for release tags.

## Unreleased

- Standardized the Go module path for `go install github.com/why19970628/agentmeter@latest`.
- Added `agentmeter doctor` for local data-source diagnostics.
- Updated CI and release workflows to run `make validate`.
- Added a macOS/Linux shell installer and Homebrew cask release configuration.
- Added Web dark/light theme switching.
- Added a global readable/raw token display toggle, defaulting to readable counts.
- Moved Web display controls into the page header and added static asset versioning to avoid stale UI scripts.
- Added MIT license.
- Added dashboard screenshot and fuller installation instructions.
- Added common CLI examples for tool, model, tool + model, path selection, and Web startup.
- Included README assets, changelog, and license files in release archives.

## v0.1.2 - 2026-05-24

- Added open-source governance docs, including contributing guides, security policy, issue templates, and PR template.
- Added repository automation for CI, release builds, and dependency updates.
- Added README cover images and editable SVG sources.

## v0.1.1 - 2026-05-24

- Added GoReleaser-based release assets for macOS, Linux, and Windows.
- Added release workflow and local release snapshot support.
- Improved project documentation for releases and usage.

## v0.1.0 - 2026-05-24

- Added the initial offline-first CLI and Web dashboard.
- Added local log scanning for Codex, Claude Code, Cursor, Gemini CLI, and generic JSON/JSONL records.
- Added token summaries, model details, project ranking, request trends, CSV export, and English/Chinese UI switching.
