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

1. Call `content_i18n_prepare_translation(source, language)` — returns source, prompt, glossary, style, context, fingerprint
2. Translate using the returned context
3. Call `content_i18n_review_translation(source, target)` — returns pass/fail, word ratio, severity-tagged issues
4. If review fails, fix issues and call `content_i18n_repair_translation(slug, content)` — validates before writing
5. Repeat review until pass
6. Use `content_i18n_next_translation` / `content_i18n_translation_queue` for deterministic batch progression when translating many files
7. Call `content_i18n_apply_work` or write directly to target path

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
