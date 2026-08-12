# bead-bonanza run snapshot — `1786562615-17`

This directory is a **copy**, committed so the run survives a container
reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.

## The run

| Field | Value |
|---|---|
| Run id | `1786562615-17` |
| As of | `2026-08-12T20:12:26Z` |
| Live run directory | `/tmp/claude/bead-bonanza/1786562615-17` |
| Repo | `/home/user/agent-plugins` |
| Run branch | `claude/bead-bonanza-refinement-261qvk` |
| Default branch | `main` |
| HEAD when taken | `59983d8` agent-plugins-viy P2 bead-bonanza restates the findings-harvest pipeline instead of vendoring it |
| Files copied | 23 |
| Beads with artifacts | 10 |

Resume the run from the branch, not from this file: `git log --oneline main..claude/bead-bonanza-refinement-261qvk`
lists every bead that has landed, one close commit each.

## Beads with artifacts in this run

| Bead | Last progress phase | Recorded sha | Progress | Handoff |
|---|---|---|---|---|
| `agent-plugins-0z0` | `committed` | `7405577` | yes | yes |
| `agent-plugins-8mv` | `committed` | `e0fd2a5` | yes | yes |
| `agent-plugins-9p4` | `committed` | `9f79570` | yes | yes |
| `agent-plugins-b13` | `committed` | `f892cc7` | yes | yes |
| `agent-plugins-d6x` | `committed` | `d1e446d` | yes | yes |
| `agent-plugins-d70` | `committed` | `b1f5c1f` | yes | yes |
| `agent-plugins-q1h` | `committed` | `c1ed842` | yes | yes |
| `agent-plugins-td1` | `committed` | `daae2d3` | yes | yes |
| `agent-plugins-txq` | `committed` | `baf0dd5` | yes | yes |
| `agent-plugins-viy` | `committed` | `59983d8` | yes | yes |

The phase column is a **hint, never the authority** — the ledger stops when
the agent writing it stops, so a trail ending before `committed` does not
mean the bead is unfinished. Resolve any status question with
`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.

## Other files copied

- `dispatch.md`
- `trailers.txt`
- `.ok`

`dispatch.md`, where present, is the plan: the mandate, the enumerated bead
list, the standing authorization, the wave table, the gate profile and the
amendments log. Read it in full and do not re-derive it.

## How to read this snapshot

**The live run directory is authoritative; this copy never is.** Before
reading anything here — for recovery, a retro, or a DoD audit — check
whether the live directory still exists and use it when it does:

    ls -A /tmp/claude/bead-bonanza/1786562615-17

Read this copy only WHEN that directory is gone, and then state this
snapshot's as-of time (`2026-08-12T20:12:26Z`) in whatever you conclude from it.

**It preserves operational state, not findings.** Findings are `retro:`
comments on the beads themselves, durable in the tracker independently of
this directory. Nothing here is the sole copy of anything a later reader
needs; it exists to make an interrupted run resumable.

**It is deleted in Phase 4, and that is not a loss.** It is removed as its
own commit, after the sprint report is written and committed:

    git rm -r .bonanza/1786562615-17/

The commits that added this directory stay in the branch history, so
nothing is destroyed by the deletion:

    git log --diff-filter=A -- '.bonanza/1786562615-17/'

finds them, and `git show <sha>:<path>` reads any file back out.
