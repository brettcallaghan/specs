# Fresh-session handoff prompt

Paste the block below into a new session on branch
`claude/bead-bonanza-improvements-j6jxjv`.

---

We are mid-sprint on the `bead-bonanza` hardening epic. Do not start work
until you have read the context below — a previous session closed 24 beads
under assumptions that have since been retired, and dispatching against a
stale brief is the specific failure mode this sprint keeps hitting.

**Read these three, in order:**

1. `bd show agent-plugins-92q` — the epic. Its comments carry the Dolt-sync
   investigation handoff, the direction-change record, and the sprint-plan
   pointer. This is the authoritative entry point.
2. `.bonanza/1786345424-bbh/dispatch.md` — the sprint plan. Sprint 1's manifest,
   gate baseline and environmental ledger are at the top; the section headed
   **"Sprint 2 plan — re-scope, 2026-08-11"** at the bottom is what you are
   executing. It answers which of the 24 closed beads need revisiting and how
   the re-scope reshapes the 15 open ones.
3. `bd show agent-plugins-92q.31` and `bd show agent-plugins-92q.30` — the two
   beads that are next.

**Where we are.** Waves 1–3 of sprint 1 closed 24 beads and are pushed to draft
PR #18. A review then retired two assumptions the shipped machinery was written
around: that subagents share one working tree, and that the tracker travels as
a committed `.beads/issues.jsonl`. `92q.30` (one worktree per bead) and
`92q.31` (Dolt remote as source of truth) replace both. Further content waves
are **paused** — `92q.7` and `8ef` are explicitly frozen, because roughly half
the text `92q.7` was going to extract is deleted by the re-scope.

**Start at stage 0: `92q.31`'s credential probe.** It is a measurement, not a
design question, and it gates everything downstream. Established already —
do not re-derive it: reads work with the brokered credential
(`git fetch '+refs/dolt/*:refs/dolt/*'`, `bd bootstrap` returns 546 chunks /
111 issues), writes return **403 on `addTableFiles`** tested three ways, and
the egress proxy is innocent (19 failures logged, zero github.com entries).
The probe re-runs the write with a real PAT. Two outcomes, both already
planned for:

- writes work → run the migration on `92q.31`, and close `92q.27` as superseded
- writes still 403 → take `92q.31`'s recorded fallback, and **promote** `92q.27`
  — without Dolt transport the JSONL file stays the transport, and the merge
  driver becomes load-bearing the moment worktrees make concurrent branches
  normal

**Two hazards that will bite if you skip them.** The remote is **31 beads
behind** — never `bd bootstrap` before pushing, or a fresh container silently
discards the whole epic. And `92q.30` must be implemented in the orchestrator's
own loop, not dispatched to a worktree subagent: it re-plumbs the mechanism
that would have dispatched it.

Stop and report at each stage boundary, as in sprint 1. Before the container
ends, run `bd export > .beads/issues.jsonl`, commit it, and snapshot the run
directory into `.bonanza/<run-id>/` — the live directory is under `/tmp` and
does not survive.
