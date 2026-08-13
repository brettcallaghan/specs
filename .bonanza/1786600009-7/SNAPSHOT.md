# bead-bonanza run snapshot — `1786600009-7`

This directory is a **copy**, committed so the run survives a container
reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.

## The run

| Field | Value |
|---|---|
| Run id | `1786600009-7` |
| As of | `2026-08-13T06:05:25Z` |
| Live run directory | `/tmp/claude/bead-bonanza/1786600009-7` |
| Repo | `/home/user/agent-plugins` |
| Run branch | `claude/triage-rfc-self-report-anchors-7vlncg` |
| Default branch | `main` |
| HEAD when taken | `f6986f1` agent-plugins-61f.7 P2 bead-bonanza: record the run's effort setting in the dispatch manifest, and carry it into routed RFCs (triage T1) |
| Files copied | 44 |
| Beads with artifacts | 5 |

Resume the run from the branch, not from this file: `git log --oneline main..claude/triage-rfc-self-report-anchors-7vlncg`
lists every bead that has landed, one close commit each.

## Beads with artifacts in this run

| Bead | Last progress phase | Recorded sha | Progress | Handoff |
|---|---|---|---|---|
| `agent-plugins-61f.1` | `committed` | `350485d` | yes | yes |
| `agent-plugins-61f.2` | `committed` | `c329e9f` | yes | yes |
| `agent-plugins-61f.3` | `committed` | `a04004e` | yes | yes |
| `agent-plugins-61f.6` | `committed` | `189bdb9` | yes | yes |
| `agent-plugins-61f.7` | `committed` | `f6986f1` | yes | yes |

The phase column is a **hint, never the authority** — the ledger stops when
the agent writing it stops, so a trail ending before `committed` does not
mean the bead is unfinished. Resolve any status question with
`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.

## Other files copied

- `agent-plugins-61f.1.build.out`
- `agent-plugins-61f.1.close.out`
- `agent-plugins-61f.1.gofmt.out`
- `agent-plugins-61f.1.lint.out`
- `agent-plugins-61f.2.build.out`
- `agent-plugins-61f.2.close.out`
- `agent-plugins-61f.2.gofmt.out`
- `agent-plugins-61f.2.lint.out`
- `agent-plugins-61f.2.sync-shared.out`
- `agent-plugins-61f.3.build.out`
- `agent-plugins-61f.3.close.out`
- `agent-plugins-61f.3.gofmt.out`
- `agent-plugins-61f.3.lint.out`
- `agent-plugins-61f.6.build.out`
- `agent-plugins-61f.6.close.out`
- `agent-plugins-61f.6.gofmt.out`
- `agent-plugins-61f.6.lint.out`
- `agent-plugins-61f.6.sync-shared.out`
- `agent-plugins-61f.7.build.out`
- `agent-plugins-61f.7.close.out`
- `agent-plugins-61f.7.gofmt.out`
- `agent-plugins-61f.7.lint.out`
- `baseline.check-skills-strict.out`
- `baseline.sync-shared.out`
- `baseline.sync-strata.out`
- `dispatch.md`
- `final.build.out`
- `final.check-skills-strict.out`
- `final.gofmt.out`
- `final.lint.out`
- `final.sync-shared.out`
- `final.sync-strata.out`
- `trailers.txt`
- `.ok`

`dispatch.md`, where present, is the plan: the mandate, the enumerated bead
list, the standing authorization, the wave table, the gate profile and the
amendments log. Read it in full and do not re-derive it.

## How to read this snapshot

**The live run directory is authoritative; this copy never is.** Before
reading anything here — for recovery, a retro, or a DoD audit — check
whether the live directory still exists and use it when it does:

    ls -A /tmp/claude/bead-bonanza/1786600009-7

Read this copy only WHEN that directory is gone, and then state this
snapshot's as-of time (`2026-08-13T06:05:25Z`) in whatever you conclude from it.

**It preserves operational state, not findings.** Findings are `retro:`
comments on the beads themselves, durable in the tracker independently of
this directory. Nothing here is the sole copy of anything a later reader
needs; it exists to make an interrupted run resumable.

**It is deleted in Phase 4, and that is not a loss.** It is removed as its
own commit, after the sprint report is written and committed:

    git rm -r .bonanza/1786600009-7/

The commits that added this directory stay in the branch history, so
nothing is destroyed by the deletion:

    git log --diff-filter=A -- '.bonanza/1786600009-7/'

finds them, and `git show <sha>:<path>` reads any file back out.
