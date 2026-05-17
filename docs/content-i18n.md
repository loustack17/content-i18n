# content-i18n in LouStackBase

This guide explains how LouStackBase uses `content-i18n` in day-to-day translation work. It is a consumer example, not the product boundary.

## Repo boundary

`content-i18n` owns:
- translation preparation
- review and validation
- queue and batch control
- completion sync

LouStackBase owns:
- Hugo site structure
- route and language policy
- article source files
- final publication decisions
- Cloudflare deployment actions

## First-time setup

In a consumer repo, start by scaffolding the config and support files:

```bash
content-i18n init --type hugo --output ./content-i18n.yaml
```

That creates:
- `content-i18n.yaml`
- `.content-i18n/glossary.yaml`
- `.content-i18n/style/technical-english.yaml`

Edit the generated paths and style settings if the consumer repo needs different roots or wording rules.

## Language-root policy

LouStackBase uses language roots:

```text
content/
  en/
    posts/
  zh-TW/
    posts/
```

Meaning:
- `en` and `zh-TW` are language roots
- `posts` is the shared section

For a translated pair, the relative path should match under both language roots.

Example:

```text
content/zh-TW/posts/DevOps/example.md
content/en/posts/DevOps/example.md
```

Do not create a third copy under `content/posts/` for new multilingual posts.

## Canonical content rule during migration

For this migration, English is the target publication language.

Mandarin source files remain the source of truth for fidelity checks. Reviewers compare the English output against the Mandarin source.

## Runtime plugin boundary

LouStackBase may still use runtime helpers such as `cmpt-translate` for site behavior.

That sits outside `content-i18n`. The translation tool should not own theme logic, route generation, runtime language switching, or Cloudflare deployment behavior. Those are consumer-side concerns.

## Typical operator flow

For one translated post:
1. prepare translation context from the Mandarin source
2. write or repair the English target
3. run review or strict source validation
4. confirm the English target has no leftover Mandarin prose
5. run `sync-status`

## Review checklist for a translated post

A file is only accepted when all of these are true:
1. review or `validate-content --source` passes
2. no leftover Mandarin prose remains in the English target
3. `sync-status` succeeds
4. no workaround hacks were used to satisfy the validator
5. code blocks and technical literals still match the source

## Batch operating rule

For rollout work:
- use queue or batch workflow from `content-i18n`
- do not manually fake completion by editing status files
- if the validator and the content contract conflict, fix `content-i18n` instead of corrupting the article
