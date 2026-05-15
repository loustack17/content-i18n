# content-i18n

Repo-local i18n harness for Markdown and static-site content.

## Status

Current accepted progress:

- Tasks 1–12 complete
- Core + CLI complete
- Validation engine complete for content integrity and Hugo site URL policy
- Next target: Task 13 MCP wrappers

## Scope

`content-i18n` manages:

- content discovery
- translation work packets
- validation
- glossary/style packs
- provider fallback
- MCP wrappers

It does not own site routing, theme behavior, or runtime translation widgets.

## Current commands

```bash
content-i18n status --config content-i18n.yaml
content-i18n list --config content-i18n.yaml
content-i18n plan --config content-i18n.yaml --file path/to/source.md --to en
content-i18n apply-work --config content-i18n.yaml --slug my-post --dry-run
content-i18n apply-work --config content-i18n.yaml --slug my-post
content-i18n validate-content --config content-i18n.yaml --file path/to/target.md
content-i18n validate-site --config content-i18n.yaml
content-i18n mcp --config content-i18n.yaml
```

## Example smoke commands

Generic Markdown validation:

```bash
go run ./cmd/content-i18n validate-content \
  --config examples/generic-markdown/content-i18n.yaml \
  --file examples/generic-markdown/docs/en/test.md
```

Hugo URL policy validation:

```bash
go run ./cmd/content-i18n validate-site \
  --config examples/hugo/content-i18n.yaml
```

## Verification

```bash
gofmt -w cmd internal
go test ./...
go build ./cmd/content-i18n
```
