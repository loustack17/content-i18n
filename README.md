# content-i18n

Repo-local i18n harness for Markdown and static-site content.

## Scope

`content-i18n` manages content discovery, translation work packets, glossary/style packs, provider fallback, MCP tools, and validation.

It does not own site routing, theme behavior, or runtime translation widgets.

## Planned commands

```bash
content-i18n status --config content-i18n.yaml
content-i18n list --config content-i18n.yaml
content-i18n plan --config content-i18n.yaml --file path/to/source.md --to en
content-i18n translate --config content-i18n.yaml --file path/to/source.md --to en
content-i18n validate --config content-i18n.yaml --file path/to/target.md
content-i18n validate-site --config content-i18n.yaml
content-i18n mcp --config content-i18n.yaml
```
