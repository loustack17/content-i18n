# Architecture

`content-i18n` is a standalone fidelity-first translation engine.

It is designed to be consumed by other repos through config, content files, glossary files, and style packs. It is not tied to one blog or one site structure.

## What the repo owns

This repo owns:
- config loading and validation
- content discovery
- translation work packet generation
- glossary and style-pack loading
- fidelity validation
- status sidecar updates
- queue control
- batch orchestration
- MCP server and CLI
- translation provider adapters

This repo does not own:
- consumer-specific routing policy
- theme behavior
- runtime translation plugins
- final editorial review
- site-specific business logic inside core packages

## Main layers

`content-i18n` has three operating layers:

1. **Core** — product logic in `internal/core`
2. **CLI** — operator-facing commands in `cmd/content-i18n`
3. **MCP** — agent-facing tool surface in `internal/mcp`

Design rule:
- core is the business-logic source of truth
- CLI and MCP are thin wrappers over core

## Package layout

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

## Core design principles

### 1. Fidelity-first translation
Only the language changes.

The system preserves:
- heading hierarchy and order
- section order
- paragraph coverage
- lists and nesting
- tables
- examples and references
- code blocks
- inline code
- URLs
- argument flow
- style class

### 2. Thin interface layers
CLI and MCP should not reimplement product logic.

They should:
- parse input
- call core
- shape output

### 3. Config-driven consumer behavior
Consumer repos provide:
- `content-i18n.yaml`
- glossary
- style pack
- source files
- target roots

Site-specific rules belong in config and adapters, not in generic core logic.

### 4. Hash-based staleness
Translation freshness is based on source content hash, not modification time.

### 5. Official completion path
A translation is only complete when:
- review/validation passes
- CJK check is clean
- sync-status succeeds

## MCP design

The MCP surface is intentionally narrow and task-oriented.

Public MCP tools:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site`

Why narrow:
- better auto tool choice
- less overlap
- easier maintenance
- clearer agent behavior

Low-level internal helpers may still exist in core/CLI, but the MCP public surface should stay small and powerful.

## Batch model

There are three progression modes:

1. **Single-file manual/agent loop**
   - prepare
   - translate
   - review
   - sync

2. **Queue-driven rollout**
   - inspect queue
   - process next candidate deterministically

3. **Batch orchestration**
   - `translate-batch`
   - one command/tool call handles many queued files

## Provider model

- **DeepL / Google**: provider-backed translation path
- **ai-harness**: external AI writes target; content-i18n handles prepare/review/sync/orchestration around it
- **auto**: fallback path between providers

## Maintainability rules

The codebase should stay:
- simple
- explicit
- low-duplication
- easy to extend

Important implementation rules:
- shared structure analysis should live in one place
- frontmatter write paths must not silently swallow errors
- MCP definitions, registration, and handlers stay split
- dead code should be removed, not preserved “just in case”
