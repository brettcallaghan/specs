#!/usr/bin/env bash
# bb-backup-run.sh — commit a self-describing snapshot of the live run
# directory to .bonanza/<run-id>/ on the run's branch.
#
# Why this exists
#   Per-bead work is durable the moment it is a commit on a branch, and the
#   tracker is durable the moment it is pushed to the Dolt remote. The RUN
#   DIRECTORY is neither: dispatch.md, the per-bead .progress.jsonl and the
#   per-bead .handoff.md live under ${TMPDIR:-/tmp} and die with the container,
#   taking with them everything a fresh orchestrator would use to resume an
#   interrupted run. The recover-sprint skill names the committed snapshot as
#   the route that survives a reclaim; this script is what puts it there.
#
#   Committed, not gitignored. A gitignored copy sits in the working tree,
#   which is reclaimed with the container exactly like /tmp — it would move the
#   artifacts without making them durable. Only a commit that reaches the
#   remote survives. The cost is small: a full run directory measured 152 KB
#   across 17 files.
#
# Cadence — per bead, not per run
#   Run it after EVERY bead's acceptance check, for the same reason
#   bb-close-bead.sh pushes the tracker per close rather than per wave: a
#   container reclaimed mid-wave must not lose the run record. A copy taken
#   once mid-run and never refreshed is indistinguishable from a complete one;
#   the copy that motivated this script went stale within 30 minutes. Hooking
#   it to the acceptance check is what stops the two drifting.
#
#   Run it from the run's own checkout on the run's branch — NOT from a bead's
#   worktree. The snapshot is the orchestrator's record of the whole run, and
#   committing it inside a bead's worktree would either pollute that bead's one
#   commit or strand the snapshot on a branch that is deleted after merge.
#
# The live directory is authoritative; a snapshot never is
#   Any procedure that reads the backup — recovery, a retro, a DoD audit —
#   first checks whether the live run directory still exists and uses it when
#   it does. Read the backup only WHEN $RUN_DIR is gone, and then state the
#   snapshot's as-of time in whatever you conclude from it. This script obeys
#   the same rule from the writing side: every fact in SNAPSHOT.md is measured
#   from the live directory, and nothing is ever read back out of the copy.
#
# Delete it at the end of Phase 4
#   `git rm -r .bonanza/<run-id>/` as its own commit, after the sprint report
#   is written and committed — never before. The snapshot must not reach the PR
#   that carries the run's actual work. Deleting it loses nothing: the commits
#   that added it stay in the branch history, reachable with
#   `git log --diff-filter=A -- '.bonanza/<run-id>/'` and `git show`.
#
# Usage
#   bb-backup-run.sh <run-dir> [<dest>]
#
#   bb-backup-run.sh "$RUN_DIR"                      # -> <repo>/.bonanza/<run-id>/
#   bb-backup-run.sh "$RUN_DIR" .bonanza/1786492547-bbh
#
#   <run-dir> may be the `current` symlink; it is resolved to its physical path
#   so the snapshot is named for the run and not for the pointer. <dest>
#   defaults to <repo-root>/.bonanza/<run-id> and is interpreted relative to
#   the current directory when given.
#
#   Idempotent: re-run it as many times per run as there are beads. Each run
#   replaces the copy wholesale, so a file deleted from the live directory
#   disappears from the snapshot instead of lingering. WHEN no RUN-DIRECTORY
#   file changed since the last snapshot, no commit is made and the committed
#   SNAPSHOT.md keeps its original as-of time — which is the correct reading:
#   as-of names when that content was observed, not when the script last ran.
#   (SNAPSHOT.md itself is excluded from that decision on purpose. It records
#   the current HEAD and a fresh timestamp, so including it would make every
#   invocation a change and every bead an empty commit.)
#
#   It does NOT push. The snapshot rides the run's branch, and branch pushes
#   are the orchestrator's call (a run may be in defer-push mode). A snapshot
#   that is committed but never pushed dies with the container like everything
#   else — push the branch.
#
# Portability — must run on macOS (bash 3.2, BSD userland) and on Linux.
#   No associative arrays, no mapfile, no ${var^^}, no `stat -c`, `sed -i`,
#   `grep -P`, `readlink -f`.
#
# Exit status
#   0  snapshot committed (short sha on stdout), or already current
#      (`unchanged` on stdout)
#   1  a step failed — the message says what to fix
#   2  bad invocation

set -uo pipefail

die()  { echo "bb-backup-run: $*" >&2; exit 2; }
fail() { echo "bb-backup-run: $*" >&2; exit 1; }
say()  { printf '  %s\n' "$*" >&2; }

# ------------------------------------------------------------------ arguments

while [ $# -gt 0 ]; do
  case "$1" in
    --) shift; break ;;
    --*) die "unknown option: $1" ;;
    *) break ;;
  esac
done

[ $# -ge 1 ] || die "usage: bb-backup-run.sh <run-dir> [<dest>]"
[ $# -le 2 ] || die "usage: bb-backup-run.sh <run-dir> [<dest>]"

run_dir="$1"
dest_arg="${2-}"

[ -n "$run_dir" ] || die "run-dir must not be empty"
[ -d "$run_dir" ] || fail "no such run directory: $run_dir
  The live directory is the only source for a snapshot. WHEN it is already
  gone, there is nothing to back up — read the last committed snapshot under
  .bonanza/ instead, and state its as-of time in whatever you conclude."

# Resolve to the physical path so `current` snapshots as the run it points at.
src="$(cd "$run_dir" 2>/dev/null && pwd -P)" || fail "cannot resolve run directory: $run_dir"
run_id="$(basename "$src")"
case "$run_id" in
  ''|.|..|/) die "cannot derive a run id from '$run_dir'" ;;
esac

# --------------------------------------------------------------------- repo

git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git work tree"
repo="$(git rev-parse --show-toplevel)" || fail "cannot resolve the repo root"

branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
case "$branch" in
  main|master) fail "refusing to commit a run snapshot to '$branch' — a run happens on its own branch" ;;
esac

if [ -n "$dest_arg" ]; then
  case "$dest_arg" in
    /*) dest="$dest_arg" ;;
    *)  dest="$(pwd -P)/$dest_arg" ;;
  esac
else
  dest="$repo/.bonanza/$run_id"
fi

# Normalize before the guards below compare paths, or `.` resolves to the
# literal string "<repo>/." and slips past the repo-root check as a path that
# merely differs textually. `..` is rejected outright rather than resolved: a
# `..` under a not-yet-created parent cannot be resolved by `cd`, and a
# textual containment check on such a path is exactly the check that can be
# fooled into copying outside the repo.
case "/$dest/" in
  *"/../"*) die "destination must not contain '..': $dest" ;;
esac
if [ -d "$dest" ]; then
  dest="$(cd "$dest" 2>/dev/null && pwd -P)" || fail "cannot resolve destination"
elif [ -d "$(dirname "$dest")" ]; then
  dest="$(cd "$(dirname "$dest")" 2>/dev/null && pwd -P)/$(basename "$dest")"
fi
# Neither branch taken means the parent does not exist yet — the normal case
# for the FIRST snapshot of a run, when .bonanza/ has never been created. The
# path is already absolute and free of `..`, and `mkdir -p` creates the parent
# below, after the guards have had their say.

# The destination has to be inside the repo, or the "commit it" half of this
# script has nothing to commit.
case "$dest/" in
  "$repo"/*) ;;
  *) fail "destination is outside the repo: $dest (repo: $repo)" ;;
esac
[ "$dest" != "$repo" ] || fail "destination must not be the repo root"

# Gitignored is the one configuration that silently defeats this script: the
# copy lands in the working tree, is never committed, and is reclaimed with the
# container exactly like /tmp. Fail loudly rather than produce a snapshot that
# does not survive.
#
# `--no-index` is load-bearing: without it `git check-ignore` stays silent for a
# path that is already tracked, so the guard would pass on every run after the
# first — exactly the runs where an ignore rule added mid-sprint needs catching.
if git check-ignore -q --no-index "$dest" 2>/dev/null; then
  fail "the destination is gitignored: $dest
  A gitignored snapshot is not durable — it is reclaimed with the container
  like the /tmp directory it copies. Remove the ignore rule (.bonanza/ is meant
  to be committed while the run is live, and deleted in Phase 4)."
fi

# --------------------------------------------------------------------- copy
#
# Wholesale replace, so the snapshot cannot accumulate files the live directory
# has since dropped. Guarded: refuse to delete a directory that is not already
# one of ours, so a mistyped <dest> cannot take a source tree with it.
if [ -e "$dest" ]; then
  [ -d "$dest" ] || fail "destination exists and is not a directory: $dest"
  if [ ! -f "$dest/SNAPSHOT.md" ] && [ -n "$(ls -A "$dest" 2>/dev/null)" ]; then
    fail "destination is non-empty and is not a previous snapshot (no SNAPSHOT.md): $dest
  Refusing to replace it. Point <dest> at .bonanza/<run-id>/ or an empty directory."
  fi
  rm -rf "$dest" || fail "cannot clear the previous snapshot at $dest"
fi

mkdir -p "$(dirname "$dest")" || fail "cannot create $(dirname "$dest")"
cp -Rf "$src" "$dest" || fail "copy failed: $src -> $dest"

# ------------------------------------------------------- describe the snapshot
#
# EVERY fact below is measured from $src, the LIVE directory — never read back
# out of $dest. The copy is never authoritative, including for this script.

as_of="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
head_sha="$(git rev-parse --short HEAD 2>/dev/null)"
head_subject="$(git log -1 --pretty=%s 2>/dev/null)"
default_branch="$(git symbolic-ref --short refs/remotes/origin/HEAD 2>/dev/null | sed 's#^origin/##')"
[ -n "$default_branch" ] || default_branch="main"
file_count="$(find "$src" -type f 2>/dev/null | wc -l | tr -d ' ')"
rel_dest="${dest#"$repo"/}"

# The bead list: every id with a progress ledger or a handoff file in the live
# directory. bash 3.2 — no associative arrays, so collect ids into a sorted
# temp file and de-duplicate with `sort -u`.
ids="${TMPDIR:-/tmp}/bb-backup-ids.$$"
: > "$ids"
for f in "$src"/*.progress.jsonl "$src"/*.handoff.md "$src"/*.retro.md; do
  [ -f "$f" ] || continue
  b="$(basename "$f")"
  b="${b%.progress.jsonl}"; b="${b%.handoff.md}"; b="${b%.retro.md}"
  printf '%s\n' "$b" >> "$ids"
done
bead_ids="$(sort -u "$ids" 2>/dev/null)"
rm -f "$ids"

bead_count=0
bead_rows=""
if [ -n "$bead_ids" ]; then
  while IFS= read -r b; do
    [ -n "$b" ] || continue
    bead_count=$(( bead_count + 1 ))
    phase="—"; extra=""
    if [ -f "$src/$b.progress.jsonl" ]; then
      last="$(tail -n 1 "$src/$b.progress.jsonl" 2>/dev/null)"
      p="$(printf '%s' "$last" | sed -n 's/.*"phase":"\([^"]*\)".*/\1/p')"
      [ -n "$p" ] && phase="\`$p\`"
      s="$(printf '%s' "$last" | sed -n 's/.*"sha":"\([^"]*\)".*/\1/p')"
      [ -n "$s" ] && extra="\`$s\`"
    fi
    ho="no"; [ -f "$src/$b.handoff.md" ] && ho="yes"
    pr="no"; [ -f "$src/$b.progress.jsonl" ] && pr="yes"
    bead_rows="$bead_rows| \`$b\` | $phase | $extra | $pr | $ho |
"
  done <<EOF
$bead_ids
EOF
fi

# Everything in the run directory that is not a per-bead artifact — dispatch.md
# and the trailers file being the ones a resuming orchestrator needs by name.
# The second glob catches dotfiles (pre-flight's `.ok` marker), which the first
# does not — a snapshot that copies a file without listing it is not
# self-describing, and the file count would not add up.
other=""
for f in "$src"/* "$src"/.[!.]*; do
  [ -f "$f" ] || continue
  n="$(basename "$f")"
  case "$n" in
    *.progress.jsonl|*.handoff.md|*.retro.md) continue ;;
  esac
  other="$other- \`$n\`
"
done
[ -n "$other" ] || other="_(none)_
"

{
  printf '# bead-bonanza run snapshot — `%s`\n\n' "$run_id"
  printf 'This directory is a **copy**, committed so the run survives a container\n'
  printf 'reclaim. It is written by `scripts/bb-backup-run.sh` after every bead.\n\n'
  printf '## The run\n\n'
  printf '| Field | Value |\n|---|---|\n'
  printf '| Run id | `%s` |\n' "$run_id"
  printf '| As of | `%s` |\n' "$as_of"
  printf '| Live run directory | `%s` |\n' "$src"
  printf '| Repo | `%s` |\n' "$repo"
  printf '| Run branch | `%s` |\n' "$branch"
  printf '| Default branch | `%s` |\n' "$default_branch"
  printf '| HEAD when taken | `%s` %s |\n' "${head_sha:-unknown}" "$head_subject"
  printf '| Files copied | %s |\n' "$file_count"
  printf '| Beads with artifacts | %s |\n\n' "$bead_count"
  printf 'Resume the run from the branch, not from this file: `git log --oneline %s..%s`\n' "$default_branch" "$branch"
  printf 'lists every bead that has landed, one close commit each.\n\n'
  printf '## Beads with artifacts in this run\n\n'
  if [ "$bead_count" -gt 0 ]; then
    printf '| Bead | Last progress phase | Recorded sha | Progress | Handoff |\n'
    printf '|---|---|---|---|---|\n'
    printf '%s' "$bead_rows"
    printf '\n'
    printf 'The phase column is a **hint, never the authority** — the ledger stops when\n'
    printf 'the agent writing it stops, so a trail ending before `committed` does not\n'
    printf 'mean the bead is unfinished. Resolve any status question with\n'
    printf '`git worktree list`, then `bd show <id>`, then `git log --grep=<id>`.\n\n'
  else
    printf '_No per-bead artifacts were present when this snapshot was taken._\n\n'
  fi
  printf '## Other files copied\n\n'
  printf '%s\n' "$other"
  printf '`dispatch.md`, where present, is the plan: the mandate, the enumerated bead\n'
  printf 'list, the standing authorization, the wave table, the gate profile and the\n'
  printf 'amendments log. Read it in full and do not re-derive it.\n\n'
  printf '## How to read this snapshot\n\n'
  printf '**The live run directory is authoritative; this copy never is.** Before\n'
  printf 'reading anything here — for recovery, a retro, or a DoD audit — check\n'
  printf 'whether the live directory still exists and use it when it does:\n\n'
  printf '    ls -A %s\n\n' "$src"
  printf 'Read this copy only WHEN that directory is gone, and then state this\n'
  printf 'snapshot'"'"'s as-of time (`%s`) in whatever you conclude from it.\n\n' "$as_of"
  printf '**It preserves operational state, not findings.** Findings are `retro:`\n'
  printf 'comments on the beads themselves, durable in the tracker independently of\n'
  printf 'this directory. Nothing here is the sole copy of anything a later reader\n'
  printf 'needs; it exists to make an interrupted run resumable.\n\n'
  printf '**It is deleted in Phase 4, and that is not a loss.** It is removed as its\n'
  printf 'own commit, after the sprint report is written and committed:\n\n'
  printf '    git rm -r %s/\n\n' "$rel_dest"
  printf 'The commits that added this directory stay in the branch history, so\n'
  printf 'nothing is destroyed by the deletion:\n\n'
  printf '    git log --diff-filter=A -- '"'"'%s/'"'"'\n\n' "$rel_dest"
  printf 'finds them, and `git show <sha>:<path>` reads any file back out.\n'
} > "$dest/SNAPSHOT.md" || fail "cannot write $dest/SNAPSHOT.md"

# -------------------------------------------------------------------- commit
#
# Stage and commit ONLY the snapshot path. A pathspec commit cannot swallow
# unrelated staged work in the orchestrator's checkout, which matters because
# this runs between beads rather than on a quiescent tree.
git add -A -- "$dest" || fail "git add failed for $dest"

# The change decision reads the COPIED RUN FILES only. SNAPSHOT.md is excluded
# because it carries the current HEAD and a fresh timestamp, so counting it
# would make every invocation a change — and at one invocation per bead that is
# an empty commit per bead for the life of the run.
changed="$(git status --porcelain -- "$dest" 2>/dev/null | grep -v '/SNAPSHOT\.md$')" || true
if [ -z "$changed" ] && git cat-file -e "HEAD:$rel_dest/SNAPSHOT.md" 2>/dev/null; then
  # Put the committed SNAPSHOT.md back: its as-of time names when this exact
  # content was observed, which is still true, and rewriting it would commit a
  # new timestamp over an unchanged snapshot.
  git checkout HEAD -- "$dest/SNAPSHOT.md" 2>/dev/null \
    || fail "cannot restore the committed $rel_dest/SNAPSHOT.md"
  say "run $run_id: snapshot already current at $rel_dest ($file_count files, $bead_count beads)"
  echo "unchanged"
  exit 0
fi

git commit -q -m "bonanza: snapshot run $run_id ($bead_count beads, as of $as_of)" -- "$dest" \
  || fail "git commit failed for $dest — hooks may have rejected it. Fix and re-run; the copy is already on disk."

sha="$(git rev-parse --short HEAD)"
say "run $run_id: snapshot committed $sha -> $rel_dest ($file_count files, $bead_count beads)"
say "not pushed — the snapshot rides the run branch; push it with the branch."
echo "$sha"
