# content-i18n

Standalone AI-assisted i18n harness for Markdown and static-site content.

Fidelity-first translation: only the language changes. Structure, content coverage, argument flow, emphasis, and style class are preserved.

## Quick start

```bash
go build ./cmd/content-i18n
./content-i18n status --config content-i18n.yaml
```

## Commands

```bash
content-i18n status --config content-i18n.yaml
content-i18n list --config content-i18n.yaml
content-i18n plan --config content-i18n.yaml --file <source.md> --to <lang>
content-i18n prepare --config content-i18n.yaml --file <source.md> --to <lang>
content-i18n review --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n repair-plan --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n next --config content-i18n.yaml [--group DevOps]
content-i18n batch-status --config content-i18n.yaml [--group DevOps]
content-i18n translate-batch --config content-i18n.yaml --provider deepl [--group DevOps] [--limit 10] [--stop-on-fail] [--continue-on-error] [--dry-run]
content-i18n apply-work --config content-i18n.yaml --slug <slug> [--dry-run] [--force]
content-i18n validate-content --config content-i18n.yaml --file <target.md> [--source <source.md>]
content-i18n validate-site --config content-i18n.yaml
content-i18n mcp --config content-i18n.yaml
```

## AI agent workflow

An AI agent can translate a post with a single instruction:

> Use content-i18n to translate `<source>` fidelity-first and keep fixing until review passes.

The tool provides all harness context, review criteria, and repair guidance. No large custom prompt needed.
It also provides deterministic batch queue control so agents can keep moving file-by-file without human reselection.

### MCP workflow (preferred)

1. Call `content_i18n_prepare_translation(source, language)` — returns source, prompt, glossary, style, context, fingerprint, target_path
2. Translate using the returned context
3. Call `content_i18n_review_translation(source, target)` — returns passed, ready_to_sync, word ratio, severity-tagged issues
4. Fix issues and re-review until pass
5. Use `content_i18n_translation_queue` for deterministic batch progression (includes next candidate)
6. Call `content_i18n_translate_batch` for full orchestration, or `content_i18n_sync_status` for direct-target workflows
7. Write to target path or use CLI `apply-work`

### CLI workflow

```bash
# Step 1: prepare — get source + context + fingerprint
content-i18n prepare --file content/posts/source.md --to en

# Step 2: translate (agent writes target.md)

# Step 3: review — check fidelity
content-i18n review --file work/slug/target.md --source content/posts/source.md

# Step 4: repair if needed
content-i18n repair-plan --file work/slug/target.md --source content/posts/source.md

# Step 5: for batches, ask for the next file deterministically
content-i18n next --group DevOps
content-i18n batch-status --group DevOps

# Step 6: apply when review passes
content-i18n apply-work --slug <slug>
```

### Batch translation (one-command orchestration)

For translating many files without manual per-file looping:

```bash
# DeepL/Google: autonomous — translate via API, review, validate, sync
content-i18n translate-batch --provider deepl --group DevOps

# AI harness: batch review/sync of pre-filled targets (agent writes target.md first)
content-i18n translate-batch --provider ai-harness --group DevOps

# Dry run: see what would be processed
content-i18n translate-batch --provider deepl --dry-run

# Stop on first failure
content-i18n translate-batch --provider deepl --stop-on-fail

# Continue processing after failures
content-i18n translate-batch --provider deepl --continue-on-error

# Limit to N files
content-i18n translate-batch --provider deepl --limit 5
```

Batch pipeline per file: prepare → translate (provider API or pre-filled target) → review → repair (if --continue-on-error) → validate → CJK check → sync-status.

Never marks complete unless validate-content passes, CJK is clean, and sync-status succeeds.

Provider modes:
- **deepl/google**: Calls provider API to translate, writes target with full frontmatter preserved, reviews, validates, syncs.
- **ai-harness**: Prepares work packets, reviews/validates/syncs pre-filled `target.md` files. Reports unfilled targets as "pending". Does not call AI — agent writes targets externally first.
- **auto**: Tries DeepL first, falls back to Google.

MCP equivalent: `content_i18n_translate_batch(provider="deepl", group="DevOps")`

## Fidelity-first contract

content-i18n is a translation harness, not an editorial rewriting tool.

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

## Architecture

See [docs/architecture.md](docs/architecture.md) for package structure.
See [docs/agent-workflow.md](docs/agent-workflow.md) for AI agent usage details.

## Verification

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/content-i18n
```
