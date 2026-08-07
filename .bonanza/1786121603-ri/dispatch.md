# Dispatch plan — reference-integrity

Branch: `claude/reference-integrity-blueprint-groom-xc7d30` (both repos)
Standing authorization: auto-chain waves until dry. Explicit `decision-required`
beads still stop.

Worktree-parallel is UNAVAILABLE for every bead in this run: `.beads/embeddeddolt/`
is gitignored, so a `git worktree` receives no Dolt DB and `bd` either fails or
silently re-inits an empty one inside it. Everything runs in the main checkout.

| Wave | Bead | Class | Bucket | Tier | Model | Push | Reason |
|------|------|-------|--------|------|-------|------|--------|
| 0 | tix | isolated | sequential | standard | sonnet | per-bead | prerequisite; adopts strata HEAD, makes the suite green |
| 1 | j7f | isolated | sequential | standard | sonnet | per-bead | script guard; blocks MODEL |
| 2 | 9kb.1 | isolated | sequential | advanced | opus | per-bead | MODEL — foundation seven items consume; 3 generated artifacts must agree |
| 3 | 9kb.2 | isolated | shared-batch #1 | advanced | opus | defer | XPT — extension-point interface |
| 3 | 9kb.3 | isolated | shared-batch #1 | advanced | opus | defer | REX — new cli/core/references pkg |
| 4 | 9kb.4 | isolated | shared-batch #2 | advanced | opus | defer | RES — corpus.go build + warm-open paths |
| 4 | 9kb.5 | isolated | shared-batch #2 | standard→advanced | opus | defer | PREG — batched at the bucket's highest tier |
| 5 | 9kb.6 | shared | shared-batch #3 | standard | sonnet | defer | CHK — c4_changes [Corpus, Findings] |
| 5 | 9kb.7 | shared | shared-batch #3 | standard | sonnet | defer | LIFE — same [Corpus, Findings]; MUST NOT parallelize with .6 |
| 5 | 9kb.9 | isolated | shared-batch #3 | standard | sonnet | defer | contract doc; same tier, rides the batch |
| 6 | 9kb.8 | isolated | sequential | standard | sonnet | per-bead | CMD — one cobra subcommand |
| 7 | 9kb.10 | isolated | sequential | advanced | opus | per-bead | DoD — 24 ACs audited, blueprints reconciled, epic closes |

Tier spread as groomed: 5 advanced / 6 standard / 0 light. Grooming declined to
manufacture a `light` assignment; `9kb.1` is flagged as the over-routing suspect
to check if the run ends with zero escalations.

## Gate profile (AGENTS.md "## bead-bonanza gates")

- format: `cd cli && gofmt -l .` — must print nothing
- lint: `scripts/check-skills.sh` + `cd cli && go vet ./...`
- build: `cd cli && go build ./...`
- test: `scripts/check-skills.sh --strict && scripts/sync-shared-references.sh --check && scripts/sync-strata-docs.sh --check` + `cd cli && go test ./...`
- regen: `scripts/generate-schemas.sh`
- custom: `scripts/generate-schemas.sh --check`

## Ledger — diagnosed once, passed into every prompt as fact

1. `strata` on PATH is strata-lang/strata HEAD (`ce18d66`), installed this
   session. `strata --version` reports `0.7.0` for every build and cannot
   identify it; read `~/.cache/strata-cli/built-commit`. Do not rebuild.
2. `scripts/check-skills.sh --strict` is KNOWN RED, pre-existing: exits 1 on a
   warning (bead-bonanza SKILL.md is 859 lines vs an "under 500" guideline).
   Baseline `0 error(s), 1 warning(s)`. Bar is no NEW findings, not exit 0.
3. `TestWarmQueriesAndChecksMeetTheirBudgets` was already unreliable here before
   any change — 6 pass / 4 fail over 10 runs on committed artifacts. Bead `tix`
   makes it non-gating. The slowness is the deliberate cost of the generated
   `UnmarshalJSON`, which is REQUIRED for correctness and must never be removed.
4. `.beads/issues.jsonl` churns on every `bd` command. Expected, additive, and
   `bb-close-bead.sh` stages it. Never commit it alone, never gitignore it.
