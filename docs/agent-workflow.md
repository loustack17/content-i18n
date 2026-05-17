# Agent workflow

`content-i18n` is designed so an AI agent can work from a short instruction and rely on the tool for context, validation, queue state, and completion rules.

## Working model

The agent should treat `content-i18n` as the workflow authority for machine-checked translation work.

That authority covers:
- preparation context
- review output
- queue state
- completion rules

It does not replace final human editorial sign-off.

## First-run setup

Before an agent can use a consumer repo cleanly, someone should run `init` in that consumer repo:

```bash
content-i18n init --type hugo --output ./content-i18n.yaml
```

Current built-in starter types:
- `hugo`
- `generic-markdown`

This creates the starter config, glossary, and style files that the later workflow depends on.

## Tiny prompts

Single file:

> Use content-i18n to translate `<source>` fidelity-first and keep fixing until review passes.

Batch:

> Use content-i18n to translate the queued batch fidelity-first and keep going until the queue is drained or a real blocker appears.

## Public MCP tool surface

Main tools used in agent workflows:
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`

Other public tools:
- `content_i18n_status`
- `content_i18n_validate_site` (consumer-site validation; current built-in adapter is Hugo-oriented)

## Single-file workflow

### MCP-first

| Step | Tool | Purpose |
|------|------|---------|
| 1 | `content_i18n_prepare_translation` | Get source, prompt, glossary, style, context, fingerprint, and target path |
| 2 | agent translates | Write the target file at the returned path |
| 3 | `content_i18n_review_translation` | Check structural and content fidelity |
| 4 | repeat 2–3 | Fix until `ready_to_sync=true` |
| 5 | `content_i18n_sync_status` | Mark completion officially |

Completion rule for a single file:
- review passed
- no leftover source-language prose remains where the target should be translated
- `sync-status` succeeds

### CLI-first

```bash
content-i18n prepare --file content/zh-TW/posts/example.md --to en
content-i18n review --file content/en/posts/example.md --source content/zh-TW/posts/example.md
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

`fingerprint` is the source-content fingerprint used for stale detection and safe completion sync.

## What review returns

`content_i18n_review_translation` returns:
- `passed`
- `ready_to_sync`
- `source_words`
- `target_words`
- `word_ratio`
- `issues[]`

`ready_to_sync=true` means the review passed with no remaining issues and the file can be synced.

## Queue workflow

Use queue mode when you want deterministic next-file progression without picking files manually.

Tool:
- `content_i18n_translation_queue`

Main fields:
- `total`
- `completed`
- `stale`
- `missing`
- `next`

CLI equivalents:

```bash
content-i18n batch-status --group DevOps
content-i18n next --group DevOps
```

## Batch workflow

Use batch mode when you want one workflow to process many queued files.

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

Per file, batch orchestration does:
- prepare
- translate or load the externally written target
- review
- validate
- CJK check
- sync-status

A file is never counted complete unless all completion gates pass.

## Provider modes

- `deepl` and `google`: provider-backed translation
- `auto`: tries DeepL first, then Google
- `ai-harness`: external AI writes the target first; `content-i18n` runs the workflow around it and does not generate the translation itself

## Rules for agents

Do not:
- summarize
- compress
- restructure
- add facts or examples
- hide troubleshooting detail
- change the article into another genre

The target should be the same article in another language.

## Picking the workflow

| Scenario | Main tool or flow | Notes |
|----------|-------------------|-------|
| One file, manual review | prepare → review → sync | Best when a human is editing the target |
| One file, AI-assisted | MCP prepare/review/sync | Best when an agent writes the target |
| Many files, deterministic progression | `translation_queue` | Pick the next file without manual reselection |
| Many files, provider-backed or AI-assisted batch work | `translate-batch` | Use provider translation or externally written targets |
| Final site-level check | `validate_site` | Consumer-site validation; current built-in adapter is Hugo-oriented |
