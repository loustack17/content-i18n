# Architecture

`content-i18n` is a standalone AI translation tool.

Consumers provide:

- `content-i18n.yaml`
- glossary
- style pack
- source content
- target content path

Core packages stay generic. Site-specific rules live in adapters and consumer config.
