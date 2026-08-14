# agent-plugins-92u handoff

Deleted `scripts/validate-corpus.py`. Added `rfcs/specs.toml` (a second spec
root, `root = "."`, workspace `agent-plugins-rfcs`) so `spec check --config
rfcs/specs.toml` covers the RFC queue with the document-local rule set:
33 documents checked, no findings, exit 0. Ignored the new `rfcs/specs.json`
index artifact in `.gitignore`. Dropped the dead script from `AGENTS.md`
(custom-gate block and the `checks.yml` description), `schemas/README.md` and
`ADR-016`'s two Consequences mentions. Filed `agent-plugins-y09` for the
configurable `corpus/empty` floor covering the retired `MIN_CHECKED = 20`
guard; it does not block this bead.

Gates: `gofmt -l` clean; `check-skills.sh` PASS; `--strict` exits 1 on the
one pre-existing `bead-bonanza/SKILL.md` size warning, 0 errors — the measured
baseline, unchanged; `sync-shared-references.sh --check` and
`sync-strata-docs.sh --check` both exit 0; `spec check --strict` over `specs/`
exits 0 with 50 documents. `cli/` and `schemas/*.strata` untouched, so the Go
and schema-generation gates do not apply.

Caveat: `.github/workflows/checks.yml` is NOT in this commit — the container's PAT lacks the `workflow` scope — so the intended edit is written up as a fragment at `/tmp/claude/bead-bonanza/1786686364-15/agent-plugins-92u.workflow-fragment.md` for the orchestrator to apply through the GitHub App (three edits: delete the pip-install and `validate-corpus.py` steps from the `checks` job, append the RFC-queue `spec check` step to the `corpus` job where the binary is built, and refresh the header comment); until it lands, `checks.yml` still invokes the deleted script and CI will fail, and AGENTS.md's `## bead-bonanza gates` block no longer lists the retired `python3 scripts/validate-corpus.py` custom gate — that gate is retired by this bead, not skipped.
