# agent-plugins-qz9.3 handoff

Caveat: ran `scripts/regenerate-cli-snapshots.sh` and reviewed the diff — it touched only `cli/testdata/cases/check-*/expected/checks.log` (8 files: check-empty-corpus, check-error, check-frontmatter-parse, check-path, check-path-reference-elsewhere, check-retired-target, check-strict-warning-only, check-unresolved-reference), each gaining a "checked N documents" line, an "excluded:" section where applicable, and per-file grouping headers on the existing finding lines; no other case's golden output changed, and no `--json` golden changed.
