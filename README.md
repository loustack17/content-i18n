# content-i18n

`content-i18n` is a standalone translation tool for language conversion across content formats and publishing systems.

It is built for fidelity-first translation. The target should be the same document in another language, not a cleaned-up rewrite.

## What this tool guarantees

`content-i18n` changes language only.

It preserves:
- heading order and section order
- paragraph coverage
- lists, tables, examples, and references
- code blocks, technical inline code, URLs, commands, identifiers, and error strings
- the article's overall reasoning and tone

A translation is complete only when:
1. review or validation passes
2. no leftover source-language prose remains where the target should be translated
3. `sync-status` succeeds

## What this tool will not do

It does not:
- own consumer routing or theme behavior
- replace final human editorial review
- rewrite the article into a different structure or genre
- change code blocks, technical literals, URLs, or references

## Repo shape

The repo has three layers:
- `internal/core`: source of truth for business logic
- `cmd/content-i18n`: CLI for operators
- `internal/mcp`: MCP surface for AI agents

CLI and MCP stay thin. Core owns the workflow rules.

## Before you start

You need:
- Go installed
- a `content-i18n.yaml` config file for your consumer repo
- source and target content roots defined in that config

Example configs live under:
- `examples/hugo/content-i18n.yaml`
- `examples/generic-markdown/content-i18n.yaml`

Run commands from the `content-i18n` repo root.

## Quick start

Scaffold a starter config in your consumer repo:

```bash
go build -o content-i18n ./cmd/content-i18n
./content-i18n init --type hugo --output ./content-i18n.yaml
```

That creates:
- `content-i18n.yaml`
- `.content-i18n/glossary.yaml`
- `.content-i18n/style/technical-english.yaml`

Then run a basic status check:

```bash
./content-i18n status --config ./content-i18n.yaml
```

You can also skip the build step:

```bash
go run ./cmd/content-i18n init --type hugo --output ./content-i18n.yaml
go run ./cmd/content-i18n status --config ./content-i18n.yaml
```

## Main CLI commands

```bash
content-i18n init --type hugo --output ./content-i18n.yaml
content-i18n status --config content-i18n.yaml
content-i18n prepare --config content-i18n.yaml --file <source.md> --to <lang>
content-i18n review --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n sync-status --config content-i18n.yaml --file <target.md> --source <source.md>
content-i18n batch-status --config content-i18n.yaml [--group DevOps]
content-i18n next --config content-i18n.yaml [--group DevOps]
content-i18n translate-batch --config content-i18n.yaml --provider deepl [--group DevOps]
content-i18n validate-site --config content-i18n.yaml
content-i18n mcp --config content-i18n.yaml
```

## Main workflows

### Single file

MCP-first:
1. `content_i18n_prepare_translation`
2. translate the target file at the returned `target_path`
3. `content_i18n_review_translation`
4. fix until `ready_to_sync=true`
5. `content_i18n_sync_status`

CLI-first example using language-root content paths:

```bash
content-i18n prepare --file content/zh-TW/posts/example.md --to en
content-i18n review --file content/en/posts/example.md --source content/zh-TW/posts/example.md
content-i18n sync-status --file content/en/posts/example.md --source content/zh-TW/posts/example.md
```

### Queue-driven rollout

```bash
content-i18n batch-status --group DevOps
content-i18n next --group DevOps
```

Use this when you want deterministic file-by-file progression.

### Batch orchestration

```bash
content-i18n translate-batch --provider deepl --group DevOps
content-i18n translate-batch --provider google --group DevOps --stop-on-fail
content-i18n translate-batch --provider ai-harness --group DevOps --continue-on-error
```

Per file, the batch flow is:
- prepare
- translate or load the externally written target
- review
- validate
- CJK check
- sync-status

## Provider behavior

- `deepl`: provider-backed translation, then review and sync
- `google`: provider-backed translation, then review and sync
- `auto`: tries DeepL first, then Google
- `ai-harness`: does not generate the translation itself; an external AI writes the target, and `content-i18n` handles prepare, review, queue control, and sync

## MCP tools and resources

Public MCP tools:
- `content_i18n_status`
- `content_i18n_prepare_translation`
- `content_i18n_review_translation`
- `content_i18n_sync_status`
- `content_i18n_translation_queue`
- `content_i18n_translate_batch`
- `content_i18n_validate_site` (consumer-site validation; current built-in adapter is Hugo-oriented)

Read-only MCP resources are also available for config and content inspection.

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

## Consumer example

A multilingual content repo can use language roots like:

```text
content/
  en/
    posts/
  zh-TW/
    posts/
```

`content-i18n` works on the content files and config. It does not own consumer routing or theme behavior.

## Markdown example

Source:

```markdown
## 問題起點

先確認 Cloud Run runtime contract 是否真的符合預期。
```

Target:

```markdown
## Starting Point

First confirm whether the Cloud Run runtime contract really matches the expected behavior.
```

The structure stays the same. Only the language changes.

## Validation guarantees

`content-i18n` checks:
- structure and section order
- code block and technical inline literal preservation
- URL preservation
- glossary and tone rules from config
- stale vs completed state through source hashes

Typical review failures include:
- changed heading order
- modified code blocks
- altered technical inline literals
- changed URLs

It does not replace final human publication review.

## Common failures

- missing or wrong `--config` path
- source and target files do not match the configured language roots
- review fails because structure drifted from the source
- `sync-status` fails because review has not passed yet

## Docs

- [docs/architecture.md](docs/architecture.md)
- [docs/agent-workflow.md](docs/agent-workflow.md)
- [docs/content-i18n.md](docs/content-i18n.md)

## Verify

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/content-i18n
```
