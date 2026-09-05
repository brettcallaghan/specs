# Run log

One row per swarm run. The record named in the last column carries that run's
results, every finding it captured verbatim, and what it could not verify.

| date | run epic | target | admitted / closed | findings | unrouted | filed | record |
|---|---|---|---|---|---|---|---|
| 2026-09-04 | `agent-plugins-d7w5` | `agent-plugins-f6tl` — spec corpus structure extraction | 8 / 8 — 4 work orders, then +4 review findings admitted by `AMENDMENT 3` | 45 (26 project, 12 skill, 7 `proof`) | 0 | 16 of 19 candidates (12 → `.sf/outbox/swarm/`, 4 beads: `3oiv`, `yqjh`, `nduj`, `yois`); 2 declined; 1 held (A1, pending root-cause investigation on `yztv`) | [2026-09-04-d7w5-spec-corpus-structure-extraction.md](2026-09-04-d7w5-spec-corpus-structure-extraction.md) |
| 2026-09-05 | `agent-plugins-59ic` | `agent-plugins-rc14` — spec corpus anchor references | 9 / 9 — 0 excluded, 0 moved, 2 admissions at the gate | 42 (29 project, 1 skill, 1 split, 10 `proof`) | 1 | **pending** — 9 candidates put to a human, none filed | [2026-09-05-59ic-spec-corpus-anchor-references.md](2026-09-05-59ic-spec-corpus-anchor-references.md) |

A record whose "admitted / closed" cell names work still in flight is
provisional: its results are revised when that run truly ends, and its findings,
routing and filing decision are not. No row is provisional today.

A "pending" count in the filed column means candidates were put to a human and
the answer is not yet recorded. The record names them, and that run's
`retro_filed` / `retro_declined` / `retro_routed` still carry the figures from
before they were put.
