#!/usr/bin/env bash
# bb-close-bead.sh — the per-bead close ceremony: atomic close, verify against
# the database, stage code, one commit, push the tracker.
#
# Why this exists
#   Every closed bead must land as exactly one commit, and the closure must be
#   durable before the container goes away. The ordering constraints, easy to
#   state and easy to reorder under pressure:
#
#     1. close BEFORE staging, so the closure is inside the implementation
#        commit rather than trailing it as a second commit;
#     2. the close must be VERIFIED against the database before anything is
#        staged — a `bd update` that silently did not take would otherwise
#        produce a commit claiming a closure that never happened;
#     3. stage the code, and only the code;
#     4. one commit, no --amend of anything pushed, no --no-verify.
#
#   Then the tracker is pushed (step 5), which is what makes the closure
#   durable.
#
#   The same ceremony appears in three subagent prompt templates and has to be
#   replayed by hand during crash recovery. One script means one behavior.
#
# Usage
#   bb-close-bead.sh [--trailers-file <path>] [--trailers-from <epic>] \
#                    <bead-id> <priority> <title> [path ...]
#   bb-close-bead.sh --self-test
#
#   --trailers-file <path>  optional; must precede <bead-id>. Its contents
#       are appended to the commit message body after a blank line (e.g.
#       Co-Authored-By, Claude-Session trailers a project's convention
#       requires). Omitting the flag leaves behavior identical to before it
#       existed — no separate `git commit --amend` is needed when it is
#       passed, since the trailers land in the one commit this script makes.
#
#   Paths are the code files this bead changed. WHEN omitted, everything
#   currently modified is staged.
#
#   Running inside the bead's own worktree — the normal case — OMIT the
#   paths. Everything dirty in that tree
#   belongs to this bead by construction, so a path list can only be wrong:
#   too narrow and the commit is incomplete, and it has to be maintained by
#   hand as the implementation grows.
#
#   Pass explicit paths only where the tree is genuinely shared: the
#   orchestrator committing its own work on the sprint branch, and the
#   main-loop fallback (no Agent tool), where every bead is implemented in
#   the shared checkout.
#
#   The close does NOT touch the bead's description: grooming wrote it, and
#   closing must not rewrite it. Findings belong in `retro:`-prefixed
#   `bd comment`s on the bead.
#
#
# Exit status
#   0  bead closed, committed, tracker pushed (short sha on stdout)
#   1  a step failed — nothing is committed; the message says what to fix.
#   2  bad invocation
#
#   One warning is NOT an error: the commit landed but `bd dolt push` failed.
#   The bead IS closed and committed; the closure is in the local database
#   only. Push it before the container is reclaimed — never re-run
#   `bd update --status closed`.

set -uo pipefail

die()  { echo "bb-close-bead: $*" >&2; exit 2; }
fail() { echo "bb-close-bead: $*" >&2; exit 1; }

# ---------------------------------------------------------------- self-test
# The invocation surface ONLY. Every case below fails before the script
# touches the tracker, git, or the filesystem, so the self-test needs no
# database, no repository and no network — which is what makes it runnable
# anywhere, including as a declared check.
#
# What it deliberately does not cover: the close ceremony itself. Steps 1-5
# mutate a tracker and a git history, and a self-test that did so would either
# need a scratch repo per run or leave litter in a real one. Those steps are
# exercised by every bead that closes.
self_test() {
  fails=0
  t() { # t <name> <expected-rc> <args...>
    n="$1"; e="$2"; shift 2
    "$0" "$@" >/dev/null 2>&1; r=$?
    if [ "$r" = "$e" ]; then printf '  ok   %s\n' "$n"
    else printf '  FAIL %s (expected %s, got %s)\n' "$n" "$e" "$r"; fails=$((fails+1)); fi
  }
  # 2 = invocation error. Each of these is a distinct branch in the parser or
  # the validation block below it.
  t "no arguments"                       2
  t "too few positionals"                2 some-id 2
  t "unknown option"                     2 --bogus some-id 2 title
  t "--trailers-file with no value"      2 --trailers-file
  t "--trailers-from with no value"      2 --trailers-from
  t "both trailer sources at once"       2 --trailers-file /dev/null \
                                           --trailers-from some-epic some-id 2 title
  t "empty bead id"                      2 "" 2 title
  t "empty title"                        2 some-id 2 ""
  t "priority is not a digit"            2 some-id x title
  t "priority out of range"              2 some-id 5 title
  t "priority is multi-digit"            2 some-id 12 title
  t "--trailers-file names nothing"      2 --trailers-file /nonexistent/trailers some-id 2 title
  [ "$fails" = 0 ] || return 1
  return 0
}

if [ "${1-}" = "--self-test" ]; then
  [ $# -eq 1 ] || die "--self-test takes no other arguments"
  echo "bb-close-bead --self-test"; self_test; exit $?
fi

trailers_file=""
trailers_epic=""
while [ $# -gt 0 ]; do
  case "$1" in
    --trailers-file)
      [ $# -ge 2 ] || die "--trailers-file requires a path argument"
      trailers_file="$2"
      shift 2
      ;;
    --trailers-file=*)
      trailers_file="${1#--trailers-file=}"
      shift ;;
    --trailers-from)
      [ $# -ge 2 ] || die "--trailers-from requires an epic id"
      trailers_epic="$2"
      shift 2 ;;
    --trailers-from=*)
      trailers_epic="${1#--trailers-from=}"
      shift
      ;;
    --)
      shift
      break
      ;;
    --*)
      die "unknown option: $1"
      ;;
    *)
      break
      ;;
  esac
done

[ $# -ge 3 ] || die "usage: bb-close-bead.sh [--trailers-file <path>] [--trailers-from <epic>] <bead-id> <priority> <title> [path ...]"
[ -z "$trailers_file" ] || [ -z "$trailers_epic" ] || die "pass --trailers-file or --trailers-from, not both"

# The run's trailers live in the run epic's manifest, so a bead can fetch them
# from the tracker rather than from a directory that may not outlive the run.
if [ -n "$trailers_epic" ]; then
  trailers_file="$(mktemp "${TMPDIR:-/tmp}/bb-trailers.XXXXXX")" || die "cannot create a temp file for trailers"
  # shellcheck disable=SC2064
  trap "rm -f '$trailers_file'" EXIT
  bd show "$trailers_epic" --json 2>/dev/null | python3 -c 'import json,sys
try:
    d = json.load(sys.stdin)
except Exception:
    raise SystemExit(1)
d = d[0] if isinstance(d, list) and d else d
if not isinstance(d, dict) or d.get("error"):
    raise SystemExit(1)
md = d.get("metadata") or {}
if isinstance(md, str):
    try: md = json.loads(md)
    except Exception: md = {}
t = (md.get("trailers") or "").strip()
if not t:
    raise SystemExit(2)
sys.stdout.write(t + "\n")' > "$trailers_file" || {
    _st=$?
    [ "$_st" = 2 ] && die "$trailers_epic carries no \`trailers\` in its manifest — omit --trailers-from when the run resolved none"
    die "cannot read the manifest on $trailers_epic"
  }
fi
id="$1"; priority="$2"; title="$3"; shift 3

if [ -n "$trailers_file" ]; then
  [ -f "$trailers_file" ] || die "--trailers-file: no such file: $trailers_file"
fi

[ -n "$id" ]    || die "bead-id must not be empty"
[ -n "$title" ] || die "title must not be empty"
case "$priority" in ''|*[!0-4]*|??*) die "priority must be a single digit 0-4, got '$priority'" ;; esac
command -v bd >/dev/null 2>&1 || fail "\`bd\` not on PATH"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not inside a git work tree"

branch="$(git rev-parse --abbrev-ref HEAD)"
case "$branch" in
  main|master) fail "refusing to commit directly to '$branch' — switch to a working branch" ;;
esac

# Cleanup runs on every exit path — success, fail()'s exit 1, die()'s exit 2 —
# so the commit-message temp file (created only when --trailers-file is
# given, step 4 below) never survives a failed `git commit`.
msgfile=""
cleanup_msgfile() { [ -n "$msgfile" ] && rm -f "$msgfile"; }
trap cleanup_msgfile EXIT

# 1. Atomic close — status only. `bd close` would be a second write and a
#    second commit; `--body-file` would overwrite the groomed description with
#    a hand-retyped copy, which is how descriptions used to lose content.
bd update "$id" --status closed \
  || fail "\`bd update $id --status closed\` failed — bead not closed, nothing staged"

# 2. Verify the close actually took. The DATABASE is the authority; there is
#    no committed ledger file to cross-check against any more.
#
#    The tracker is origin refs/dolt/data, .beads/*.jsonl is untracked, and a
#    closure becomes durable when it is pushed — step 5 below.
db_status() {
  if command -v jq >/dev/null 2>&1; then
    bd show "$1" --json 2>/dev/null | jq -r '.[0].status // empty'
  else
    bd show "$1" --json 2>/dev/null | grep -o '"status": *"[a-z_]*"' | head -1 | cut -d'"' -f4
  fi
}

observed="$(db_status "$id")"
if [ "$observed" != "closed" ]; then
  fail "the close did not take: bd reports $id status '${observed:-unknown}' after \`bd update --status closed\`.
  Nothing staged, nothing committed. Fix the tracker write, then re-run this script."
fi

# 3. Stage code only. Everything uncommitted in the worktree belongs to this
#    bead — inside a batch's shared tree too, since the previous bead committed
#    before this one started — so no paths need naming; the Usage note above
#    gives the two cases that still pass them.
#
#    The exclusion names the JSONL exports alone, not all of .beads/. Excluding
#    the whole directory silently drops tracked tracker config — config.yaml,
#    .gitignore, the bd hooks — from the commit of any bead whose scope
#    includes it, and the close still reports success.
if [ $# -gt 0 ]; then
  git add -- "$@" || fail "git add failed for: $*"
else
  git add -A -- ':!.beads/*.jsonl' || fail "git add failed"
fi

git diff --staged --quiet && fail "nothing staged — no changes to commit for $id.
  If the bead needed no code change, this is a SKIP, not a close:
  run bb-verify-skip.sh and take the skipped path instead."

# 4. One commit. No --no-verify (hooks are a gate), no --amend (the commit may
#    already be pushed; a missed step becomes a follow-up commit instead).
#    WHEN --trailers-file was given, its contents are appended to the message
#    body after a blank line, so the trailers land in this one commit and no
#    caller ever needs to amend it back in.
subject="$id P$priority $title"
if [ -n "$trailers_file" ]; then
  msgfile="$(mktemp "${TMPDIR:-/tmp}/bb-commit-msg.XXXXXX")" \
    || fail "cannot create commit-message temp file under ${TMPDIR:-/tmp}"
  { printf '%s\n\n' "$subject"; cat "$trailers_file"; } > "$msgfile"
  git commit -q -F "$msgfile" \
    || fail "git commit failed — hooks may have rejected it. Fix and re-run; the bead is already closed, so do not re-run bd update."
else
  git commit -q -m "$subject" \
    || fail "git commit failed — hooks may have rejected it. Fix and re-run; the bead is already closed, so do not re-run bd update."
fi

sha="$(git rev-parse --short HEAD)"
# 5. Push the tracker. PER-CLOSE, not per-batch. A bead's closure is the unit of durable work, and
#    deferring to the end of a batch loses every closure since the last push when
#    a container is reclaimed mid-batch, which is the loss this epic exists to
#    prevent. The cost is one push per bead, seconds.
#
#    Skipped quietly where no Dolt remote is configured: this script ships with
#    a project-agnostic skill, and a project tracking beads another way must
#    still be able to close one.
if bd dolt remote list 2>/dev/null | grep -q .; then
  bd dolt push >/dev/null 2>&1 \
    || echo "bb-close-bead: WARNING — commit landed but \`bd dolt push\` failed. The closure is in the local database only; push it before the container is reclaimed." >&2
fi

echo "$sha"
