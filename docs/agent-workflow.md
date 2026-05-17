# Agent workflow

`content-i18n` is built so an AI agent can use a short instruction and rely on the tool for context, validation, queue control, and completion rules.

## Repo role in the workflow

`content-i18n` is the translation engine.

The consumer repo provides:
- source content
- target roots
- glossary
- style pack
- project config

The agent should treat `content-i18n` as the workflow authority.

## Tiny prompt model

A compliant agent should be able to work from a short instruction like:

> Use content-i18n to translate `<source>` fidelity-first and keep fixing until review passes.

For batch mode:

> Use content-i18n to translate the queued DevOps batch fidelity-first and keep going until the queue is drained or a real blocker appears.

## Public MCP workflow

The public MCP surface is intentionally narrow:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site`

## Single-file workflow

### MCP-first

| Step | Tool | Purpose |
|------|------|---------|
| 1 | `content_i18n_prepare_translation` | Get source, prompt, glossary, style, context, fingerprint, target path |
| 2 | agent translates | Produce fidelity-first target content |
| 3 | `content_i18n_review_translation` | Check whether translation is structurally/content correct |
| 4 | repeat step 2–3 | Fix until review is good |
| 5 | `content_i18n_sync_status` | Mark translation complete officially |

### CLI-first

```bash
content-i18n prepare --file content/posts/source.md --to en
content-i18n review --file work/slug/target.md --source content/posts/source.md
content-i18n sync-status --file content/en/posts/example.md --source content/zh-TW/posts/example.md
```

## What prepare returns

`content_i18n_prepare_translation` returns:
- `source`
- `prompt`
- `glossary`
- `style`
- `context`
- `fingerprint`
- `slug`
- `target_path`

This should be enough to translate without unrelated repo exploration.

## What review returns

`content_i18n_review_translation` returns:
- `passed`
- `ready_to_sync`
- `source_words`
- `target_words`
- `word_ratio`
- `issues[]`

`ready_to_sync=true` means:
- structural/content validation passed
- no error-severity issues remain
- the file is ready for official completion sync

Then call:
- `content_i18n_sync_status`

## Queue-driven workflow

Use queue mode when you want deterministic next-file progression without picking files manually.

Tool:
- `content_i18n_translation_queue`

Returns:
- total
- completed
- stale
- missing
- next candidate

CLI equivalents:
```bash
content-i18n batch-status --group DevOps
content-i18n next --group DevOps
```

## Batch orchestration workflow

Use batch orchestration when you want one command/tool call to process many queued files.

### CLI

```bash
content-i18n translate-batch --provider deepl --group DevOps
content-i18n translate-batch --provider google --group DevOps --stop-on-fail
content-i18n translate-batch --provider ai-harness --group DevOps --continue-on-error
```

### MCP

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

### Batch behavior

Per file, orchestration does:
- prepare
- translate (provider API or pre-filled target)
- review
- repair if configured
- validate
- CJK check
- sync-status

A file is never counted complete unless:
- validation passes
- CJK is clean
- sync-status succeeds

## Provider modes

- **deepl / google**: provider-backed translation path
- **ai-harness**: external AI writes target first; content-i18n handles prepare/review/sync/orchestration
- **auto**: DeepL first, then Google fallback

## Fidelity rules for agents

The agent must not:
- summarize
- compress
- restructure
- merge or split sections casually
- add new facts or examples
- remove caution/troubleshooting detail
- change the article into a different genre

The target should be the same article in another language.

## When to use which workflow

| Scenario | Best path |
|----------|-----------|
| One post, manual review | prepare → review → sync |
| One post, AI-assisted | MCP prepare/review/sync |
| Many posts, deterministic manual progression | translation_queue |
| Many posts, provider-backed orchestration | translate-batch |
| Final site-level check | validate_site |
