---
title: "Resolve the run's commit trailers onto the manifest, or stop passing --trailers-from"
description: Nothing in a run sets manifest.trailers, so bb-close-bead.sh --trailers-from exits 2 on every close and no commit in the run's history names the bead it landed.
slug: issue-a-run-resolves-no-trailers-so-no-commit-carries-one
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-05T00:00:00Z

# Evidence — what this issue rests on
findings: []                         # none: recorded by the run's coordinator
                                     # outside the `retro:` finding channel —
                                     # see
                                     # issue-coordinator-tooling-findings-are-invisible-to-the-harvest
recurrence: 0
report: docs/retros/2026-09-05-59ic-spec-corpus-anchor-references.md

# Provenance — who and what produced this
run: agent-plugins-59ic
session: https://claude.ai/code/session_0185zfzFm9maHZjPX4FQqhbZ
branch: sprint/agent-plugins-59ic
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<derived, not observed — manifest.models holds only the tier→model map (standard→sonnet, advanced→opus) and nothing records what served each dispatch>"
---

## What is wrong, and what it costs

`bb-close-bead.sh --trailers-from <run-epic>` reads the run's commit trailers
out of the run epic's manifest and **exits 2 when that key is unset**. Nothing
in a run sets it. No step in scope, plan or dispatch resolves `trailers` onto
the manifest, so the flag the close ceremony is told to pass can only fail,
and every close falls back to a commit with no trailers at all.

The cost is that **a run's commit history does not name the beads it landed.**
Nine beads closed on the run that hit this and not one of the nine commits
carries a `Bead:` trailer, so the link from a commit back to its work order
exists only in the subject line's id text and in the tracker. Where the
harness or the project requires attribution trailers — co-authorship, a
session link — those reach nothing either: the run is exactly the surface that
was supposed to carry them into every commit it makes. Nobody notices at the
time, because the fallback is silent from the caller's point of view: the bead
still closes and the commit still lands.

## Evidence

- **The key is unset.** Measured on the run epic:

  ```
  $ swarm manifest get --run agent-plugins-59ic --key trailers
  trailers:
    ok: false
    value: null
    issues:
      - "trailers is not set. Set it with: swarm manifest set --run agent-plugins-59ic --key trailers --value <value>"
  ```

- **That is exactly what the script exits 2 on.** `bb-close-bead.sh`'s
  `--trailers-from` block (opening `if [ -n "$trailers_epic" ]`, line 150 as a
  hint) pipes `bd show "$trailers_epic" --json` into a Python filter ending:

  ```python
  t = (md.get("trailers") or "").strip()
  if not t:
      raise SystemExit(2)
  ```

  and the shell handler beside it reads:

  ```sh
  [ "$_st" = 2 ] && die "$trailers_epic carries no \`trailers\` in its manifest — omit --trailers-from when the run resolved none"
  ```

  The script's own message names the remedy the run never takes.

- **The consequence, measured on the finished tree.**
  `git log --format='%H%n%b%n---' 225751e..a852925 | grep -c 'Bead:'` returns
  **0** over the run's nine commits, and `git log -1 --format='%B' a852925` is
  the subject line alone with an empty body.

- **`rfc_path` is not the only manifest key a run wants and cannot have** — a
  separate observation from the same run, recorded here because it bears on
  where `trailers` should be resolved: `swarm manifest set --run
  agent-plugins-59ic --key rfc_path --value '"…"'` is refused with `manifest:
  refusing the write, and storing none of it — "rfc_path" is not a declared
  manifest key`. `trailers` **is** declared; it is simply never written.

## Proposed change

The run's opening step resolves `trailers` onto the manifest, so that
`--trailers-from` either works or is never passed. Concretely: after scope
takes the authorization gate, the coordinator writes `trailers` — the
attribution block the harness requires where it states one, and an
explicitly-empty value where it does not — with `swarm manifest set --run
<epic> --key trailers --value <value>`.

Done is: after the scope step, `swarm manifest get --run <epic> --key
trailers` returns `ok: true`; every close ceremony's commit body carries the
resolved trailers; and `git log --format='%b' <base>..<head> | grep -c 'Bead:'`
equals the number of beads the run closed. Where a run legitimately resolves
none, the close ceremony is told not to pass `--trailers-from` rather than
being told to pass a flag that always fails.
