# Run log

One row per swarm run. The record named in the last column carries that run's
results, every finding it captured verbatim, and what it could not verify.

| date | run epic | target | admitted / closed | findings | unrouted | filed | record |
|---|---|---|---|---|---|---|---|
| 2026-09-04 | `agent-plugins-d7w5` | `agent-plugins-f6tl` — spec corpus structure extraction | 4 / 4, then **+4 admitted by AMENDMENT 3 and still in flight** | 28 (16 project, 11 skill, 1 `proof`) | 0 | 13 of 15 candidates (10 → `.sf/outbox/swarm/`, 3 beads: `3oiv`, `yqjh`, `nduj`); 2 declined | [2026-09-04-d7w5-spec-corpus-structure-extraction.md](2026-09-04-d7w5-spec-corpus-structure-extraction.md) |

A record whose "admitted / closed" cell names work still in flight is
provisional: its results are revised when that run truly ends, and its findings,
routing and filing decision are not.
