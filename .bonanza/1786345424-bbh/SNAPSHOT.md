# bead-bonanza run snapshot — `1786345424-bbh`

This directory is a **copy**, committed so the run survives a container
reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.

## The run

| Field | Value |
|---|---|
| Run id | `1786345424-bbh` |
| As of | `2026-08-12T06:39:13Z` |
| Live run directory | `/tmp/claude/bead-bonanza/1786345424-bbh` |
| Repo | `/home/user/agent-plugins` |
| Run branch | `claude/bead-bonanza-sprint-2-yqpua6` |
| Default branch | `main` |
| HEAD when taken | `5283e52` agent-plugins-6xn P2 references/final-report.md does not exist, but bead Designs reference it |
| Files copied | 66 |
| Beads with artifacts | 30 |

Resume the run from the branch, not from this file: `git log --oneline main..claude/bead-bonanza-sprint-2-yqpua6`
lists every bead that has landed, one close commit each.

## Beads with artifacts in this run

| Bead | Last progress phase | Recorded sha | Progress | Handoff |
|---|---|---|---|---|
| `agent-plugins-34g` | `committed` | `fcf6a10` | yes | no |
| `agent-plugins-4sp` | `retro-written` |  | yes | yes |
| `agent-plugins-6xn` | `committed` | `18a6dfa` | yes | yes |
| `agent-plugins-7eh` | `retro-written` |  | yes | yes |
| `agent-plugins-92q.1` | `committed` | `223da6c` | yes | no |
| `agent-plugins-92q.10` | `committed` | `d8af7f2` | yes | yes |
| `agent-plugins-92q.11` | `committed` | `6c3f533` | yes | yes |
| `agent-plugins-92q.14` | `committed` | `2067740` | yes | yes |
| `agent-plugins-92q.15` | `committed` | `c9dfa4f` | yes | yes |
| `agent-plugins-92q.16` | `committed` | `32d85f8` | yes | yes |
| `agent-plugins-92q.17` | `committed` | `d16abbc` | yes | yes |
| `agent-plugins-92q.18` | `committed` | `179179e` | yes | yes |
| `agent-plugins-92q.2` | `committed` | `a66763e` | yes | no |
| `agent-plugins-92q.20` | `committed` | `d5a6de5` | yes | yes |
| `agent-plugins-92q.22` | `committed` | `9d1e0fc` | yes | yes |
| `agent-plugins-92q.23` | `committed` | `905aee4` | yes | yes |
| `agent-plugins-92q.24` | `committed` | `004ca6a` | yes | yes |
| `agent-plugins-92q.25` | `committed` | `140fba7` | yes | yes |
| `agent-plugins-92q.26` | `committed` | `b046949` | yes | yes |
| `agent-plugins-92q.27` | `implementing` |  | yes | no |
| `agent-plugins-92q.3` | `committed` | `6a77c77` | yes | no |
| `agent-plugins-92q.4` | `committed` | `4d00af8` | yes | no |
| `agent-plugins-92q.5` | `committed` | `d2f7696` | yes | yes |
| `agent-plugins-92q.6` | `committed` | `0ccc092` | yes | no |
| `agent-plugins-92q.8` | `committed` | `7948664` | yes | yes |
| `agent-plugins-92q.9` | `committed` | `faff74e` | yes | yes |
| `agent-plugins-ht3` | `committed` | `4079abd` | yes | yes |
| `agent-plugins-jh6` | `committed` | `8c54565` | yes | no |
| `agent-plugins-psw` | `committed` | `c35f056` | yes | yes |
| `agent-plugins-qqm` | `committed` | `1b97627` | yes | yes |

The phase column is a **hint, never the authority** — the ledger stops when
the agent writing it stops, so a trail ending before `committed` does not
mean the bead is unfinished. Resolve any status question with
`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.

## Other files copied

- `HANDOFF-PROMPT.md`
- `SNAPSHOT.md`
- `desc-before.sha`
- `dispatch.md`
- `msg.txt`
- `trailers.txt`
- `.ok`

`dispatch.md`, where present, is the plan: the mandate, the enumerated bead
list, the standing authorization, the wave table, the gate profile and the
amendments log. Read it in full and do not re-derive it.

## How to read this snapshot

**The live run directory is authoritative; this copy never is.** Before
reading anything here — for recovery, a retro, or a DoD audit — check
whether the live directory still exists and use it when it does:

    ls -A /tmp/claude/bead-bonanza/1786345424-bbh

Read this copy only WHEN that directory is gone, and then state this
snapshot's as-of time (`2026-08-12T06:39:13Z`) in whatever you conclude from it.

**It preserves operational state, not findings.** Findings are `retro:`
comments on the beads themselves, durable in the tracker independently of
this directory. Nothing here is the sole copy of anything a later reader
needs; it exists to make an interrupted run resumable.

**It is deleted in Phase 4, and that is not a loss.** It is removed as its
own commit, after the sprint report is written and committed:

    git rm -r .bonanza/1786345424-bbh/

The commits that added this directory stay in the branch history, so
nothing is destroyed by the deletion:

    git log --diff-filter=A -- '.bonanza/1786345424-bbh/'

finds them, and `git show <sha>:<path>` reads any file back out.
