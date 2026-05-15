# AGENTS.md — content-i18n

`content-i18n` is a standalone AI-assisted i18n harness for Markdown/static-site content.

## Project Boundary

This repo owns:

- content discovery
- translation work packets
- validation
- glossary/style rules
- CLI
- MCP wrappers
- provider fallback

This repo does not own:

- Hugo routing
- theme behaviour
- runtime translation widgets
- LouStackBase-specific URLs

Keep core logic generic. Hugo is an adapter/consumer, not the product boundary.

## Current Priority

V1 order:

1. Core + CLI
2. MCP wrappers after CLI commands work locally
3. DeepL/Google provider fallback Post-V1

V1 commands:

```bash
content-i18n status
content-i18n list
content-i18n plan --file <source.md> --to en
content-i18n apply-work --slug <slug> --dry-run
content-i18n apply-work --slug <slug>
content-i18n validate-content --file <target.md>
content-i18n validate-site
```

Post-V1:

```bash
content-i18n translate --file <source.md> --to en --provider deepl|google
```

## Engineering Rules

- Keep `internal/core` independent from CLI and MCP.
- CLI and MCP must call the same core functions.
- Prefer hash-based stale detection via `.content-i18n/status.json`; do not use mtime as authority.
- Only validated write paths should update status metadata.
- `apply-work --dry-run` must show intended changes before write.
- Provider credentials must come from environment variables, never config files.
- Runtime `work/` packets are generated local state and must not be committed.

## Verification

Run before reporting done:

```bash
gofmt -w cmd internal
go test ./...
go run ./cmd/content-i18n status --config examples/generic-markdown/content-i18n.yaml
```
