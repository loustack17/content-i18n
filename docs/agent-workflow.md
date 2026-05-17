# Agent workflow

content-i18n is designed so an AI agent can translate content with a short instruction. The tool itself provides all harness context, review criteria, and repair guidance.

## Tiny prompt

### MCP-first

> Use content-i18n to translate `<source>` fidelity-first and keep fixing until review passes.

### CLI-first

> Use `content-i18n prepare`, translate the output, then `content-i18n review` and fix until it passes.

That's it. No large custom operator prompt. The tool carries everything the agent needs.

## Expected tool call sequence

### MCP

| Step | Tool | Purpose |
|------|------|---------|
| 1 | `content_i18n_prepare_translation` | Get source, prompt, glossary, style, context, fingerprint, target_path |
| 2 | (agent translates) | Use returned context to produce fidelity-first translation |
| 3 | `content_i18n_review_translation` | Validate against source structure and content |
| 4 | Repeat 3 until review passes | Iterative repair loop |
| 5 | `content_i18n_sync_status` | Mark direct-target translation complete (optional) |

For batch workflows:
| Step | Tool | Purpose |
|------|------|---------|
| 1 | `content_i18n_translation_queue` | See queue status and next candidate |
| 2 | `content_i18n_translate_batch` | Full orchestration: prepare → translate → review → validate → sync |

### CLI

| Step | Command | Purpose |
|------|---------|---------|
| 1 | `content-i18n prepare --file <source> --to <lang>` | Get source + context + fingerprint |
| 2 | (agent translates) | Write target.md |
| 3 | `content-i18n review --file <target> --source <source>` | Validate translation |
| 4 | `content-i18n repair-plan --file <target> --source <source>` | Pre-write validation |
| 5 | Repeat 3-4 until review passes | Iterative repair loop |
| 6 | `content-i18n apply-work --slug <slug>` | Deploy to target path |

## Required success criteria

Review passes when:

- Heading hierarchy and order match source
- Paragraph count per section matches source
- List count and nesting match source
- Table dimensions match source
- Code blocks preserved byte-for-byte
- Inline code preserved
- URLs preserved
- Blockquotes preserved
- Word ratio is reasonable (not <50% of source)
- No tone/style violations

## What prepare returns

A single `content_i18n_prepare_translation` call returns:

- `source` — full source markdown
- `prompt` — translation prompt from config
- `glossary` — glossary terms (if configured)
- `style` — style pack rules (if configured)
- `context` — structure fingerprint, heading order, preserved URLs, 7-point self-check
- `fingerprint` — H2/H3/H4 counts, list counts, table rows, paragraphs, blockquotes, code blocks
- `slug` — work packet identifier
- `target_path` — where to write the translation

## What review returns

A single `content_i18n_review_translation` call returns:

- `passed` — boolean
- `ready_to_sync` — true when passed=true AND no error-severity issues (structure/code/URL intact). When true, call `content_i18n_sync_status` to mark complete.
- `source_words` / `target_words` / `word_ratio` — coverage check
- `issues[]` — array of:
  - `severity` — "error" (structure/code/URL) or "warning" (tone/style)
  - `field` — which check failed
  - `section` — where in the document
  - `message` — what went wrong
  - `suggested_fix` — how to fix it

## Fidelity-first, not editorial rewriting

content-i18n enforces that the translated output is the same article in another language. The agent must not:

- Summarize or compress sections
- Merge or split paragraphs
- Reorder sections
- Add facts, examples, or commentary not in the source
- Remove caveats, troubleshooting steps, or lessons learned
- Change the genre (debugging walkthrough → tutorial, etc.)

## Self-sufficient work packets

Once `prepare` is called, the work packet directory contains:

- `source.md` — source content
- `prompt.md` — translation prompt
- `glossary.md` — glossary (if configured)
- `style.md` — style pack (if configured)
- `context.md` — structure fingerprint + self-check checklist
- `meta.json` — metadata with structure hash and fingerprint
- `target.md` — translation output (agent writes here)

The agent needs no external context beyond the work packet to complete the translation.

## Batch translation orchestration

For translating many files without manual per-file looping, use `translate-batch` (CLI) or `content_i18n_translate_batch` (MCP).

### When to use batch vs manual flow

| Scenario | Approach |
|----------|----------|
| Single file or small edits | Manual prepare → review → repair → apply |
| Many files, DeepL/Google provider | `translate-batch --provider deepl` (fully autonomous) |
| Many files, AI agent workflow | `translate-batch --provider ai-harness` (processes pre-filled targets) |
| Need control over error handling | `--stop-on-fail` or `--continue-on-error` |
| Want to preview before running | `--dry-run` |

### Batch pipeline

For each queued file:

1. **Prepare** — generate work packet with source, prompt, glossary, style, context, fingerprint
2. **Translate** — call provider API (DeepL/Google) or check for pre-filled target (ai-harness)
3. **Review** — validate against source using `TranslateReview` (config-driven glossary/tone/style checks)
4. **Repair** — if review fails and `--continue-on-error`, attempt repair
5. **Validate** — CJK character check on target body
6. **Sync-status** — update official status store only if all checks pass

### Provider modes

- **`deepl` / `google`**: Calls provider API to translate body text, writes target with full frontmatter preserved (all source fields + provider metadata), reviews, validates, syncs.
- **`ai-harness`**: Prepares work packets, then reviews/validates/syncs pre-filled `target.md` files. Does not call AI — the agent must write targets externally first. Reports unfilled targets as "pending".
- **`auto`**: Tries DeepL first, falls back to Google.

### Success guarantees

Batch never marks a file complete unless:
- `validate-content --source` passes (structure, code, URLs, glossary, tone)
- CJK character check is clean (no source-language characters in target)
- `sync-status` succeeds (path validation, status store update)

### MCP batch tool

```json
{
  "tool": "content_i18n_translate_batch",
  "arguments": {
    "provider": "deepl",
    "group": "DevOps",
    "limit": 10,
    "stop_on_fail": false,
    "continue_on_error": true
  }
}
```

Returns:
```json
{
  "total": 15,
  "completed": [{"source_path": "...", "target_path": "...", "language": "en", "status": "completed"}],
  "failed": [{"source_path": "...", "language": "en", "status": "failed", "error": "..."}],
  "remaining": [{"source_path": "...", "language": "en", "status": "remaining"}]
}
```
