# Run log

One row per swarm run. The record named in the last column carries that run's
results, every finding it captured verbatim, and what it could not verify.

| date | run epic | target | admitted / closed | findings | unrouted | filed | record |
|---|---|---|---|---|---|---|---|
| 2026-09-04 | `agent-plugins-d7w5` | `agent-plugins-f6tl` — spec corpus structure extraction | 8 / 8 — 4 work orders, then +4 review findings admitted by `AMENDMENT 3` | 45 (26 project, 12 skill, 7 `proof`) | 0 | 16 of 19 candidates (12 → `.sf/outbox/swarm/`, 4 beads: `3oiv`, `yqjh`, `nduj`, `yois`); 2 declined; 1 held (A1, pending root-cause investigation on `yztv`) | [2026-09-04-d7w5-spec-corpus-structure-extraction.md](2026-09-04-d7w5-spec-corpus-structure-extraction.md) |
| 2026-09-05 | `agent-plugins-59ic` | `agent-plugins-rc14` — spec corpus anchor references | 9 / 9 — 0 excluded, 0 moved, 2 admissions at the gate | 42 (29 project, 1 skill, 1 split, 10 `proof`) | 1 | 8 of 9 candidates (5 → `.sf/outbox/swarm/`, 3 beads: `flvl`, `x59j`, `e1zx`); 0 declined; 1 (C) folded into an existing outbox issue as recurrence 1 | [2026-09-05-59ic-spec-corpus-anchor-references.md](2026-09-05-59ic-spec-corpus-anchor-references.md) |
| 2026-09-07 | `agent-plugins-w85b` | `agent-plugins-kgir` — corpus listing (`spec list`) | 10 / 10 — 6 admitted at the gate, +4 review findings admitted across two loops; 0 excluded, 0 moved | 39 (23 project, 0 skill, 11 `proof`, 4 `environment`) | 1 | 0 of 7 candidates — **pending**, put to a human, no answer recorded; 7 further clusters deduped into existing homes; 3 recurrence bumps on existing `.sf/outbox/swarm/` issues | [2026-09-07-w85b-corpus-listing-spec-list.md](2026-09-07-w85b-corpus-listing-spec-list.md) |
| 2026-10-07 | `specs-sra` | `specs-c4pre` + `specs-f6tl` — C4 prerequisites with structure record blocks | 10 / 10 — 3 at the gate, +7 review findings across two loops; 1 excluded (`specs-f6tl.7`, deferred); review budget exhausted with `specs-sra.8`, `.9` left open | 16 (6 project, 1 skill, 8 `proof`, 0 `environment`) | 1 | 6 of 6 candidates (4 → `outbox/swarm/`, 2 beads: `xoe`, `30x`); 0 declined; 1 dedupe (`specs-84yu`). Dolt remote diverged mid-run; reconciled by pull+push after the retro | [2026-10-07-sra-c4-prerequisites-record-blocks.md](2026-10-07-sra-c4-prerequisites-record-blocks.md) |
| 2026-10-07 | `specs-c4s1` | C4 diagrams stage 1 — Mermaid C4 parser and `c4` rule identifiers | 12 / 12 — 2 admitted at the gate (after the base moved to PR #47's head), +9 review findings across two loops, +1 coordinator gofmt fix; 0 excluded, 0 moved; 14 held out to backlog | 29 (7 project, 5 skill, 8 `proof`, 0 `environment`) | 9 | 11 of 11 candidates (7 → `outbox/swarm/`, 2 beads: `qlr`, `77f`; `3kx` folded to P2; 1 recurrence fold on `84yu`, `7jxt`, `a484`); 0 declined | [2026-10-07-c4s1-c4-diagrams-parser.md](2026-10-07-c4s1-c4-diagrams-parser.md) |
| 2026-10-07 | `specs-c4s2` | C4 diagrams stage 2 — element resolution and `c4` checks | 16 / 16 — 8 at the gate, +8 review findings admitted by pass 1 (7 fixed, `.15` closed by design); pass 2 admitted 0; 0 excluded, 0 moved; 12 held out to backlog | 40 (19 project, 7 skill, 13 `proof`, 1 `environment`) | 0 | 0 of 8 candidates — **pending**, put to a human by the coordinator (2 new → `outbox/swarm/`, 4 folds into existing `outbox/swarm/` issues, 2 beads) | [2026-10-07-c4s2-c4-diagrams-checks.md](2026-10-07-c4s2-c4-diagrams-checks.md) |
| 2026-10-07 | `specs-c4model` | `feature-spec-c4` — C4 model and the `specs c4` command | 12 / 12 — 8 at the gate, +3 review findings admitted by pass 1, +1 by pass 2 (`.12`, a regression from `.10`); pass 3 admitted 0; 0 excluded, 0 moved; 11 held out to backlog | 23 (10 project, 1 skill, 8 `proof`, 0 `environment`) | 4 | 1 of 7 candidates (1 bead: `pxc`); 6 declined (2 new and 2 folds → `outbox/swarm/`, 2 recurrence comments on `84yu`/`xoe`) | [2026-10-07-c4model-c4-model-specs-c4.md](2026-10-07-c4model-c4-model-specs-c4.md) |
| 2026-10-08 | `specs-xqm` | Spec CLI QA fixes before the first release | 34 / 34 — 26 at the gate (+`.1` closed before the run), +2 run-introduced amendments (`hlok`, `.25`), +5 review findings across two loops (`.26`–`.30`, all fixed); pass 3 admitted 0; `specs-8i2` excluded by standing decision; 16 review beads held out to backlog | 97 (54 project, 7 skill, 34 `proof`, 2 `environment`) | 4 | 4 of 9 candidates filed (F–I: comments on `84yu`/`7jxt`, `specs-9x4n`, `97ib`+`o5kz` → dup of `1sz`, `59ic.8` closed + `specs-4ejq`); A–E (swarm outbox) declined | [2026-10-08-xqm-spec-cli-qa-fixes.md](2026-10-08-xqm-spec-cli-qa-fixes.md) |

A record whose "admitted / closed" cell names work still in flight is
provisional: its results are revised when that run truly ends, and its findings,
routing and filing decision are not. No row is provisional today.

The 2026-09-07 row is **pending** in the filed column: its seven candidates were
put and no answer has been recorded. Its `retro_filed` / `retro_declined`
counters read 0 for that reason, not because the candidates were declined.

A "pending" count in the filed column means candidates were put to a human and
the answer is not yet recorded. The record names them, and that run's
`retro_filed` / `retro_declined` / `retro_routed` still carry the figures from
before they were put.
