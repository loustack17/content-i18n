# content-i18n

Standalone fidelity-first translation harness for Markdown and static-site content.

`content-i18n` translates language only. It preserves structure, content coverage, argument flow, examples, references, code blocks, inline code, URLs, and style class from the source post.

## What this repo is

This repo is the translation engine, not the consumer content repo.

It owns:
- config loading and content discovery
- work packet generation
- glossary and style-pack enforcement
- fidelity validation
- direct-target completion sync
- batch queue control
- batch orchestration
- MCP server and CLI
- provider adapters (DeepL / Google)

It does not own:
- Hugo routing or theme behavior
- consumer-specific article content
- runtime language-switch plugins
- editorial rewriting
- final human publication review

## Product shape

`content-i18n` supports three operating layers:

1. **Core** — reusable business logic in `internal/core`
2. **CLI** — operator-facing commands in `cmd/content-i18n`
3. **MCP** — agent-facing tool surface in `internal/mcp`

The MCP surface is intentionally narrow and task-oriented.

## Public MCP tool surface

`content-i18n` exposes 7 MCP tools:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site`

These tools are meant to cover the full agent workflow without exposing noisy low-level internals.

## Repo layout

```text
cmd/content-i18n/          CLI entrypoint
internal/config/           config schema + validation
internal/content/          content discovery + status sidecar I/O
internal/core/             main product logic
internal/frontmatter/      frontmatter parsing + metadata injection
internal/mcp/              MCP server, tool/resource definitions, handlers
internal/providers/deepl/  DeepL provider
internal/providers/google/ Google provider
internal/structure/        shared structure analysis helpers
internal/translator/       code/inline protection helpers
internal/validator/        fidelity/content validation
examples/                  runnable example configs
schemas/                   JSON schema for config
```

## Quick start

```bash
go build ./cmd/content-i18n
./content-i18n status --config content-i18n.yaml
```

## CLI commands

```bash
content-i18n status --config content-i18n.yaml
content-i18n list --config content-i18n.yaml
content-i18n plan --config content-i18n.yaml --file <source.md> --to <lang>
content-i18n prepare --config content-i18n.yaml --file <source.md> --to <lang>
content-i18n review --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n repair-plan --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n next --config content-i18n.yaml [--group DevOps]
content-i18n batch-status --config content-i18n.yaml [--group DevOps]
content-i18n sync-status --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n translate-batch --config content-i18n.yaml --provider deepl [--group DevOps] [--limit 10] [--stop-on-fail] [--continue-on-error] [--dry-run]
content-i18n apply-work --config content-i18n.yaml --slug <slug> [--dry-run] [--force]
content-i18n validate-content --config content-i18n.yaml --file <target.md> [--source <source.md>]
content-i18n validate-site --config content-i18n.yaml
content-i18n mcp --config content-i18n.yaml
```

## Main workflows

### 1. Single-file AI workflow

MCP-first flow:
1. `content_i18n_prepare_translation`
2. agent translates
3. `content_i18n_review_translation`
4. fix until `ready_to_sync=true`
5. `content_i18n_sync_status`

CLI-first flow:
```bash
content-i18n prepare --file content/posts/source.md --to en
content-i18n review --file work/slug/target.md --source content/posts/source.md
content-i18n sync-status --file content/en/posts/example.md --source content/zh-TW/posts/example.md
```

### 2. Queue-driven rollout

```bash
content-i18n batch-status --group DevOps
content-i18n next --group DevOps
```

Use when you want deterministic file-by-file progression.

### 3. One-command batch orchestration

```bash
content-i18n translate-batch --provider deepl --group DevOps
content-i18n translate-batch --provider google --group DevOps --stop-on-fail
content-i18n translate-batch --provider ai-harness --group DevOps --continue-on-error
```

Pipeline per file:
- prepare
- translate (provider API or pre-filled target)
- review
- repair if configured
- validate
- CJK check
- sync-status

A file is never counted complete unless validation passes, CJK is clean, and sync-status succeeds.

## Provider modes

- **deepl/google**: provider-backed translation, then review/validate/sync
- **ai-harness**: batch review/sync for pre-filled targets; does not call an LLM itself
- **auto**: tries DeepL first, falls back to Google

## Fidelity-first contract

`content-i18n` is a translation harness, not an editorial rewriting tool.

| Preserved | Translated |
|-----------|------------|
| Heading hierarchy and order | Prose within each element |
| Paragraph count per section | Link text (without changing meaning) |
| List count and nesting | Frontmatter title/description/keywords |
| Table dimensions | Glossary terms applied when applicable |
| Code blocks (byte-for-byte) | |
| Inline code | |
| URLs | |
| Examples and references | |
| Argument flow | |
| Style class | |

## Docs

- [docs/architecture.md](docs/architecture.md)
- [docs/agent-workflow.md](docs/agent-workflow.md)

## Verification

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/content-i18n
```
