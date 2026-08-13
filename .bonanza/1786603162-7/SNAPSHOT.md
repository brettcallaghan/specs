# bead-bonanza run snapshot — `1786603162-7`

This directory is a **copy**, committed so the run survives a container
reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.

## The run

| Field | Value |
|---|---|
| Run id | `1786603162-7` |
| As of | `2026-08-13T07:10:06Z` |
| Live run directory | `/tmp/claude/bead-bonanza/1786603162-7` |
| Repo | `/home/user/agent-plugins` |
| Run branch | `claude/validate-corpus-spec-cli-0t0cxo` |
| Default branch | `main` |
| HEAD when taken | `2256efb` agent-plugins-qz9.3 P2 spec check reports what it checked: document count, exclusions, and failures grouped by file |
| Files copied | 18 |
| Beads with artifacts | 7 |

Resume the run from the branch, not from this file: `git log --oneline main..claude/validate-corpus-spec-cli-0t0cxo`
lists every bead that has landed, one close commit each.

## Beads with artifacts in this run

| Bead | Last progress phase | Recorded sha | Progress | Handoff |
|---|---|---|---|---|
| `agent-plugins-3xx` | `committed` | `a85b2dc` | yes | yes |
| `agent-plugins-j5n` | `committed` | `0ecad82` | yes | yes |
| `agent-plugins-k9j` | `committed` | `cfe3360` | yes | yes |
| `agent-plugins-lwu` | `committed` | `332dea1` | yes | yes |
| `agent-plugins-qz9.2` | `committed` | `04cb850` | yes | yes |
| `agent-plugins-qz9.3` | `pushed` |  | yes | yes |
| `agent-plugins-ziv` | `committed` | `c0e9997` | yes | yes |

The phase column is a **hint, never the authority** — the ledger stops when
the agent writing it stops, so a trail ending before `committed` does not
mean the bead is unfinished. Resolve any status question with
`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.

## Other files copied

- `baseline-strict.out`
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

    ls -A /tmp/claude/bead-bonanza/1786603162-7

Read this copy only WHEN that directory is gone, and then state this
snapshot's as-of time (`2026-08-13T07:10:06Z`) in whatever you conclude from it.

**It preserves operational state, not findings.** Findings are `retro:`
comments on the beads themselves, durable in the tracker independently of
this directory. Nothing here is the sole copy of anything a later reader
needs; it exists to make an interrupted run resumable.

**It is deleted in Phase 4, and that is not a loss.** It is removed as its
own commit, after the sprint report is written and committed:

    git rm -r .bonanza/1786603162-7/

The commits that added this directory stay in the branch history, so
nothing is destroyed by the deletion:

    git log --diff-filter=A -- '.bonanza/1786603162-7/'

finds them, and `git show <sha>:<path>` reads any file back out.
