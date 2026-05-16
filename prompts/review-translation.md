Review this translation against the source for fidelity-first compliance.

FIDELITY CHECKS (fail if any violated):
- code blocks unchanged (byte-for-byte match)
- inline code unchanged
- heading count and hierarchy match source exactly
- list count and nesting match source exactly
- table count and dimensions match source exactly
- paragraph count per section matches source
- all URLs preserved exactly
- all examples and code samples present
- all references, citations, and footnotes preserved
- argument flow matches source (introduction → problem → analysis → solution → conclusion)
- no sections added, removed, merged, or split
- no editorial commentary, opinions, or summaries added
- style class matches source (debugging walkthrough, tutorial, post-mortem, etc.)

QUALITY CHECKS:
- glossary terms applied when source terms present
- no invented facts
- no missing troubleshooting steps
- clear DevOps/Platform wording
- no awkward literal Chinese phrasing
- title uses natural English capitalization

Return:
1. pass or fail
2. exact issues with section reference
3. corrected Markdown if fail
