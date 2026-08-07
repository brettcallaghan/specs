# Decisions — reference-integrity run

## Explicit

### agent-plugins-j7f — generate-schemas.sh has no strata version guard

**Problem.** `scripts/generate-schemas.sh` runs whatever `strata` is on PATH and
reports success regardless of what it emitted. Measured this run, THREE distinct
strata builds give three different results against the same sources:

| strata commit | result |
|---|---|
| `44c1955` (session-start hook installs this) | schema destroyed — every document type emits `"properties": {}`, 2202 lines -> 1326. Script reports SUCCESS. |
| `ce18d66` (HEAD; AGENTS.md says build this) | hard error: `/go/optionalStyle: unknown property` — #296 removed `optionalStyle` AND `generateValidation` |
| `840f9ba` | correct — byte-identical to committed artifacts |

So the AGENTS.md ledger entry is now WRONG: "building from the strata checkout
at HEAD produces output byte-identical to what is committed" held on 2026-08-04
and does not hold today. There is both a floor AND a ceiling.

**Options.**

| # | Option | Blast radius | Trade-off |
|---|---|---|---|
| 1 | Version floor naming commit 8997153 | small, in-script | Guards ONLY the old-binary case. Would NOT have caught today's failure at all — `ce18d66` is above the floor and still wrong. Version strings are useless here: `44c1955` and `840f9ba` both report `strata-cli 0.7.0`. |
| 2 | Post-generation sanity assertion: every document-type definition must carry a non-empty `required` list | small, in-script | Version-independent; catches this SHAPE of regression rather than one build. Would have caught `44c1955`. Would NOT have caught `ce18d66` — but that one already fails loudly on its own, which is the acceptable outcome. |
| 3 | Pin the strata version the session-start hook installs | hook + docs | Fixes provisioning, not the script. A hand-run `generate-schemas.sh` with a stray binary is still unguarded. Complements 2; not a substitute. |
| 4 | Wire `generate-schemas.sh --check` into CI | CI workflow | Needs a Rust toolchain in CI — the reason it is deliberately absent today. Option 2 removes most of the need. |

**Recommendation: 2 + 3.** Option 2 is the actual guard and satisfies the bead's
acceptance criterion verbatim ("fails loudly rather than emitting a schema whose
document types have no properties"). Option 3 pins provisioning to `840f9ba` so
a fresh container starts correct instead of starting broken. Option 1 is
dominated by 2 — it is more brittle and catches strictly less. Option 4 is out
of this bead's scope and already has its own bead (`agent-plugins-9ko` covers
CI-gate expansion).

**Status:** [ ] pending
