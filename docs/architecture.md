# Architecture

`content-i18n` is a standalone translation engine.

It is meant to be consumed by other repositories through config, glossary files, style packs, and content roots. It is not tied to one blog or one CMS.

## Boundary

This repo owns:
- config loading and validation
- content discovery
- translation preparation
- glossary and style-pack loading
- review and validation
- status and sync updates
- queue control
- batch orchestration
- CLI and MCP interfaces
- DeepL and Google provider adapters

This repo does not own:
- consumer-specific routing
- theme behavior
- runtime language-switch widgets
- final human editorial review
- site-specific business rules inside core packages

It is intended to stay language-agnostic and consumer-agnostic at the core layer. Consumer-specific validation belongs in adapters.

## Layers

There are three layers:

1. **Core** — business logic in `internal/core`
2. **CLI** — operator-facing commands in `cmd/content-i18n`
3. **MCP** — agent-facing workflow surface in `internal/mcp`

Rule:
- core is the source of truth
- CLI and MCP call core
- interface layers should not reimplement business rules

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

## Built-in consumer starters

The `init` command currently scaffolds two starter types:
- `hugo`
- `generic-markdown`

These are starter templates, not the full boundary of the core engine.

## Core design rules

### Fidelity first
The tool translates language only.

It preserves:
- heading hierarchy and order
- section order and paragraph coverage
- list and table shape
- examples and references
- code blocks and technical inline code
- URLs and identifiers
- the article's reasoning and tone

### Thin interfaces
CLI and MCP should only:
- parse input
- call core
- shape output

### Config-driven behavior
Consumer repos provide:
- `content-i18n.yaml`
- source and target roots
- glossary files
- style packs

Consumer policy should stay in config, not in generic core logic.

### Hash-based staleness
Freshness is based on source content hash, not file modification time.

### Official completion path
A file is complete only when:
1. review or validation passes
2. no leftover source-language prose remains where the target should be translated
3. `sync-status` succeeds

## MCP design

Public MCP tools:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site` (consumer-site validation; current built-in adapter is Hugo-oriented)

Read-only MCP resources may also expose config and content for inspection. Site validation is adapter-level behavior, not a limitation of the core translation engine.

Why keep the public tool surface narrow:
- less overlap
- better auto tool choice
- easier maintenance
- clearer agent behavior

## Queue and batch model

There are three progression modes:

1. **Single-file loop**
   - prepare
   - translate
   - review
   - sync

2. **Queue-driven rollout**
   - inspect queue
   - process the next candidate deterministically

3. **Batch orchestration**
   - run `translate-batch`
   - let the tool manage many queued files in one workflow

## Provider model

- `deepl` and `google`: provider-backed translation paths
- `auto`: tries DeepL first, then Google
- `ai-harness`: external AI writes targets; `content-i18n` handles preparation, review, queue control, and sync around those files

## Maintainability rules

The codebase should stay:
- simple
- explicit
- low-duplication
- easy to extend

Important rules:
- shared structure analysis should live in one place
- frontmatter write paths must not silently swallow errors
- MCP definitions, registration, and handlers stay split
- dead code should be removed instead of kept for later
