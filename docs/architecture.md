# Architecture

`content-i18n` is a standalone AI translation tool.

Consumers provide:

- `content-i18n.yaml`
- glossary
- style pack
- source content
- target content path

Core packages stay generic. Site-specific rules live in adapters and consumer config.

## Package structure

```
cmd/content-i18n/          CLI entry point
internal/
  config/                  YAML config loading and validation
  content/                 File discovery, status tracking, hash-based staleness
  core/                    Business logic: status, list, plan, apply, validate, translate
  frontmatter/             YAML frontmatter parsing and metadata injection
  translator/              Token protection (code blocks, inline code, URLs), section splitting
  validator/               Content integrity + structural + tone validation
  providers/               Translation providers (DeepL, Google)
  mcp/                     MCP server with high-level and low-level tools
  adapters/                Site-specific adapters (Hugo, generic-markdown)
prompts/                   Translation and review prompts
examples/                  Example configs and style packs
```

## Design principles

1. **Fidelity-first**: Translation preserves structure, coverage, argument flow, and style class. Only language changes.
2. **Core isolation**: `internal/core` is independent from CLI and MCP. Both call the same core functions.
3. **Hash-based staleness**: `.content-i18n/status.json` tracks source hashes, not mtime.
4. **Env-only credentials**: Provider credentials come from environment variables, never config files.
5. **Local-only work packets**: `work/` directory is generated local state, not committed.

## AI agent design

content-i18n is designed for short-prompt AI use. The tool provides:

- **prepare**: returns source, prompt, glossary, style, context, fingerprint in one call
- **review**: returns structured issues with severity (error/warning), word ratio, actionable fixes
- **repair**: validates before writing, returns pass/fail with issue list

An agent needs only: "Use content-i18n to translate `<source>` fidelity-first and keep fixing until review passes."

See [docs/agent-workflow.md](agent-workflow.md) for the full agent workflow.
