# content-i18n

## Boundary

Owns:
- config
- discovery
- work packets
- glossary/style enforcement
- validation
- status/sync
- queue
- batch orchestration
- CLI
- MCP
- provider adapters

Does not own:
- consumer routing/theme/runtime widgets
- consumer-specific article logic
- final human publication review

## Core rules

- Fidelity-first: translate language only.
- Preserve: heading order, section order, paragraph coverage, lists, tables, examples, references, code blocks, inline code, URLs, argument flow, style class.
- Do not: summarize, compress, restructure, add facts/examples/commentary, change article genre.
- Completion requires: validation pass + CJK clean + `sync-status` success.
- Never fake completion with manual status edits.
- `internal/core` = business-logic source of truth.
- CLI/MCP = thin wrappers over core.
- Hash-based staleness only. No mtime authority.
- Preserve unknown frontmatter fields on write paths.
- Do not silently swallow frontmatter parse/encode errors on official write paths.

## Public MCP surface

Keep only:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site`

## Workflow

Single file:
1. prepare
2. translate
3. review
4. fix until `ready_to_sync=true`
5. sync-status

Batch:
- queue = `batch-status` / `next`
- orchestration = `translate-batch`
- per file: prepare → translate → review → validate → CJK check → sync-status

Provider modes:
- `deepl` / `google` = provider-backed
- `ai-harness` = external AI writes target, repo handles prepare/review/sync/orchestration
- `auto` = DeepL then Google fallback

## MCP design

- Keep MCP split: server / defs / register / handlers / response helpers.
- Prefer task-oriented tools, not overlapping low-level tools.
- Queue/batch features should reduce agent freelancing.

## Verify

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/content-i18n
```

MCP changes:

```bash
npx -y @modelcontextprotocol/inspector -- go run ./cmd/content-i18n mcp --config examples/generic-markdown/content-i18n.yaml
```
