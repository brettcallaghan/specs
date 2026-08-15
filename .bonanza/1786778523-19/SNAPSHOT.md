# bead-bonanza run snapshot — `1786778523-19`

This directory is a **copy**, committed so the run survives a container
reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.

## The run

| Field | Value |
|---|---|
| Run id | `1786778523-19` |
| As of | `2026-08-15T17:55:19Z` |
| Live run directory | `/tmp/claude/bead-bonanza/1786778523-19` |
| Repo | `/home/user/agent-plugins` |
| Run branch | `claude/gherkin-adoption-strategy-fikj4i` |
| Default branch | `main` |
| HEAD when taken | `8a53856` Align plugin.json description with marketplace.json (Gherkin and EARS) |
| Files copied | 52 |
| Beads with artifacts | 18 |

Resume the run from the branch, not from this file: `git log --oneline main..claude/gherkin-adoption-strategy-fikj4i`
lists every bead that has landed, one close commit each.

## Beads with artifacts in this run

| Bead | Last progress phase | Recorded sha | Progress | Handoff |
|---|---|---|---|---|
| `agent-plugins-0i6` | `retro-written` |  | yes | yes |
| `agent-plugins-4vz` | `pushed` |  | yes | yes |
| `agent-plugins-a7u` | `pushed` |  | yes | yes |
| `agent-plugins-krb` | `committed` | `83c575d` | yes | yes |
| `agent-plugins-nu2` | `committed` | `ff839a8` | yes | yes |
| `agent-plugins-r19` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.1` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.10` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.11` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.12` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.2` | `committed` | `a7a8df2` | yes | yes |
| `agent-plugins-wvd.3` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.4` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.5` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.6` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.7` | `committed` | `b41c9db` | yes | yes |
| `agent-plugins-wvd.8` | `pushed` |  | yes | yes |
| `agent-plugins-wvd.9` | `pushed` |  | yes | yes |

The phase column is a **hint, never the authority** — the ledger stops when
the agent writing it stops, so a trail ending before `committed` does not
mean the bead is unfinished. Resolve any status question with
`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.

## Other files copied

- `check-skills.log`
- `dispatch.md`
- `generate-schemas-check.log`
- `go-test.log`
- `spec`
- `sync-shared-references.log`
- `sync-strata-docs.log`
- `trailers.txt`
- `.ok`

`dispatch.md`, where present, is the plan: the mandate, the enumerated bead
list, the standing authorization, the wave table, the gate profile and the
amendments log. Read it in full and do not re-derive it.

## How to read this snapshot

**The live run directory is authoritative; this copy never is.** Before
reading anything here — for recovery, a retro, or a DoD audit — check
whether the live directory still exists and use it when it does:

    ls -A /tmp/claude/bead-bonanza/1786778523-19

Read this copy only WHEN that directory is gone, and then state this
snapshot's as-of time (`2026-08-15T17:55:19Z`) in whatever you conclude from it.

**It preserves operational state, not findings.** Findings are `retro:`
comments on the beads themselves, durable in the tracker independently of
this directory. Nothing here is the sole copy of anything a later reader
needs; it exists to make an interrupted run resumable.

**It is deleted in Phase 4, and that is not a loss.** It is removed as its
own commit, after the sprint report is written and committed:

    git rm -r .bonanza/1786778523-19/

The commits that added this directory stay in the branch history, so
nothing is destroyed by the deletion:

    git log --diff-filter=A -- '.bonanza/1786778523-19/'

finds them, and `git show <sha>:<path>` reads any file back out.
