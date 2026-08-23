#!/usr/bin/env bash
#
# bb-status.sh — verify a run's acceptance signals from state, not testimony.
#
# USAGE
#   bb-status.sh bead <bead-id> [--worktree <path>] [--pre-merge]
#   bb-status.sh wave [--branch <name>] [--remote <name>]
#   bb-status.sh run <epic-id>
#   bb-status.sh slot [<epic-id>]
#   bb-status.sh harvest <bead-id> [<bead-id>...]
#   bb-status.sh --self-test
#
# EXIT CODES — the caller acts on these, so they distinguish three outcomes:
#   0  all clear
#   1  REPAIRABLE failure. Signals 1, 2 and 5: close the bead, or add the
#      missing `retro:` comment from the returned summary. State the repair in
#      the report.
#   2  STOP THE WAVE. Signals 3 and 4: the merge or the acceptance never
#      happened. Not the orchestrator's to repair inline — leave the worktree
#      standing, since it is the in-flight signal, and surface it.
#
# --pre-merge DEFERS SIGNALS 3 AND 4. A wave checks what a subagent returned
# BEFORE the orchestrator merges the branch and removes the worktree, so both
# signals would fail by construction: the branch cannot be an ancestor of HEAD
# yet, and the worktree is supposed to still be there. The wave passes
# --pre-merge; the orchestrator re-runs without it after merging, which is what
# actually verifies 3 and 4.
#   3  invocation error.
#
# WHO RUNS IT
#   The ORCHESTRATOR, after each bead's acceptance check and at every wave
#   boundary; a recovering session, as its first act; a woken durable trigger.
#
#   **A subagent never runs it.** Inside its own worktree the signals read
#   differently by construction — signal 4 asks whether the worktree is gone,
#   and the subagent is standing in it.
#
# WHAT IT DOES NOT DO
#   It does not check the Phase 4 report's prose. Whether a report NAMES a
#   failing signal is a property of text this script cannot read; what it can
#   do is make the underlying facts cheap to re-derive, which is `wave` mode.
#   It never repairs anything: every mode is read-only.
#
# Every line prints the verified state AND the command that established it, so
# a reader can re-run any single check without reconstructing it.

set -uo pipefail

RC_OK=0; RC_REPAIR=1; RC_STOP=2; RC_USAGE=3

die() { echo "bb-status: $*" >&2; exit "$RC_USAGE"; }
say() { printf '  %s\n' "$*"; }
# state, verdict, and the command that established it — one line, three parts.
line() { printf '  %-4s %-34s %s\n' "$1" "$2" "$3"; }

worst=$RC_OK
note() { # note <rc>
  case "$1" in
    "$RC_STOP")   worst=$RC_STOP ;;
    "$RC_REPAIR") [ "$worst" = "$RC_STOP" ] || worst=$RC_REPAIR ;;
  esac
}

# ------------------------------------------------------------------ bead mode
mode_bead() {
  id=""; wt=""; premerge=no
  [ $# -ge 1 ] || die "usage: bb-status.sh bead <bead-id> [--worktree <path>] [--pre-merge]"
  id="$1"; shift
  while [ $# -gt 0 ]; do
    case "$1" in
      --worktree)  [ $# -ge 2 ] || die "--worktree needs a path"; wt="$2"; shift 2 ;;
      --pre-merge) premerge=yes; shift ;;
      *)           die "unknown argument: $1" ;;
    esac
  done

  echo "bead $id"

  # 1. closed in the tracker.
  #
  # Read the status field, never `bd show | grep -q`: under `pipefail`, grep -q
  # exits at the first match and closes the pipe, `bd show` dies on SIGPIPE,
  # and the pipeline reports failure for a bead that is in fact closed. Same
  # class of trap as reading `$?` through a pager.
  c="bd show $id --json | jq -r .[0].status"
  _st="$(bd show "$id" --json 2>/dev/null | jq -r '(if type=="array" then .[0] else . end).status // empty' 2>/dev/null)"
  if [ "$_st" = closed ]; then
    line OK "1 tracker: closed" "$c"
  else
    line FAIL "1 tracker: ${_st:-unreadable}" "$c"; note $RC_REPAIR
  fi

  # 2. a commit whose subject carries the id
  c="git log --oneline --all --grep=$id -1"
  sha="$(git log --oneline --all --grep="$id" -1 --format=%h 2>/dev/null)"
  if [ -n "$sha" ]; then
    line OK "2 commit: $sha" "$c"
  else
    line FAIL "2 commit: none found" "$c"; note $RC_REPAIR
  fi

  # 3. merged into the current branch. Only meaningful once 2 found a sha, and
  # only AFTER the merge has happened. A wave verifies what a subagent returned
  # before the orchestrator merges, so it passes --pre-merge and signals 3 and 4
  # are deferred to the orchestrator's own check after the merge. Without that
  # split every bead fails signal 3 by construction — the branch is not an
  # ancestor of HEAD yet, and cannot be.
  if [ "$premerge" = yes ]; then
    line SKIP "3 merged: deferred" "(--pre-merge; the merge has not happened yet)"
  elif [ -n "$sha" ]; then
    c="git merge-base --is-ancestor $sha HEAD"
    if git merge-base --is-ancestor "$sha" HEAD 2>/dev/null; then
      line OK "3 merged into HEAD" "$c"
    else
      line STOP "3 NOT merged into HEAD" "$c"; note $RC_STOP
    fi
  else
    line SKIP "3 merged: no sha to test" "(signal 2 found nothing)"
  fi

  # 4. the worktree is gone. The load-bearing signal — state, not testimony.
  git worktree prune 2>/dev/null
  c="git worktree prune && git worktree list"
  if [ -n "$wt" ]; then
    if git worktree list 2>/dev/null | grep -qF -- "$wt"; then
      if [ "$premerge" = yes ]; then
        line SKIP "4 worktree: deferred" "(--pre-merge; removal follows the merge)"
      else
        line STOP "4 worktree SURVIVES: $wt" "$c"; note $RC_STOP
      fi
    else
      line OK "4 worktree: gone" "$c"
    fi
  else
    line SKIP "4 worktree: path not given" "$c"
  fi

  # 5. findings recorded
  mode_harvest_one "$id"
}

# --------------------------------------------------------------- harvest mode
#
# `bd show <id> --json` reports comment_count 0 for a bead that HAS comments,
# so a harvest keyed on it fails every bead in the run at once. The export
# pipeline is the one that reads correctly.
harvest_count() {
  # No `head` on the end: it would close the pipe early and, under pipefail,
  # turn a correct answer into a failure.
  bd export 2>/dev/null \
    | jq -r --arg id "$1" 'select(.id==$id) | [.comments[]?|select(.text|startswith("retro:"))] | length' 2>/dev/null \
    | tr -d "[:space:]"
}

mode_harvest_one() {
  _id="$1"
  c="bd export | jq '…select(.text|startswith(\"retro:\"))…'"
  n="$(harvest_count "$_id")"
  [ -n "$n" ] || n=0
  if [ "$n" -gt 0 ] 2>/dev/null; then
    line OK "5 findings: $n retro: comment(s)" "$c"
  else
    # Zero is not automatically wrong — a clean bead may say so explicitly —
    # but it is never silently acceptable, so it is surfaced as repairable.
    line FAIL "5 findings: none recorded" "$c"; note $RC_REPAIR
  fi
}

mode_harvest() {
  [ $# -ge 1 ] || die "usage: bb-status.sh harvest <bead-id> [<bead-id>...]"
  command -v jq >/dev/null 2>&1 || die "harvest needs jq"
  total=0
  for _b in "$@"; do
    echo "bead $_b"
    mode_harvest_one "$_b"
    n="$(harvest_count "$_b")"; [ -n "$n" ] || n=0
    total=$((total + n))
  done
  echo
  # The assertion that matters at run level: a whole run harvesting zero
  # findings is the signature of a broken pipeline, not of a flawless run.
  if [ "$total" -eq 0 ]; then
    line FAIL "harvest total: 0 across $# bead(s)" "a zero-total harvest is a broken pipeline, not a clean run"
    note $RC_REPAIR
  else
    line OK "harvest total: $total" "summed over $# bead(s)"
  fi
}

# ------------------------------------------------------------------ wave mode
mode_wave() {
  branch=""; remote=""
  while [ $# -gt 0 ]; do
    case "$1" in
      --branch) [ $# -ge 2 ] || die "--branch needs a name"; branch="$2"; shift 2 ;;
      --remote) [ $# -ge 2 ] || die "--remote needs a name"; remote="$2"; shift 2 ;;
      *) die "unknown argument: $1" ;;
    esac
  done
  [ -n "$branch" ] || branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
  [ -n "$remote" ] || remote="$(git config --get "branch.$branch.remote" 2>/dev/null || echo origin)"

  echo "wave: branch $branch, remote $remote"

  c="git status --porcelain"
  if [ -z "$(git status --porcelain 2>/dev/null)" ]; then
    line OK "tree: clean" "$c"
  else
    line FAIL "tree: DIRTY" "$c"; note $RC_REPAIR
  fi

  # The landing check. An orchestrator that dies between committing and pushing
  # leaves every other signal reading "finished": no worktrees, clean tree,
  # beads closed. Only this comparison separates a landed wave from an unlanded
  # one.
  c="git rev-list --count $remote/$branch..$branch"
  if git rev-parse --verify -q "$remote/$branch" >/dev/null 2>&1; then
    ahead="$(git rev-list --count "$remote/$branch..$branch" 2>/dev/null)"
    [ -n "$ahead" ] || ahead=0
    if [ "$ahead" -eq 0 ]; then
      line OK "landed: 0 unpushed commits" "$c"
    else
      line FAIL "landed: $ahead commit(s) UNPUSHED" "$c"; note $RC_REPAIR
    fi
  else
    line FAIL "landed: no $remote/$branch to compare" "$c"; note $RC_REPAIR
  fi

  git worktree prune 2>/dev/null
  c="git worktree prune && git worktree list"
  n="$(git worktree list 2>/dev/null | tail -n +2 | grep -c . || true)"
  [ -n "$n" ] || n=0
  if [ "$n" -eq 0 ]; then
    line OK "in flight: 0 worktrees" "$c"
  else
    # Not a failure by itself: a surviving worktree is the in-flight signal.
    line INFO "in flight: $n worktree(s)" "$c"
    git worktree list 2>/dev/null | tail -n +2 | sed 's/^/      /'
  fi

  c="bd merge-slot check"
  if out="$(bd merge-slot check 2>/dev/null)"; then
    say "merge slot: $(printf '%s' "$out" | tr '\n' ' ' | sed 's/  */ /g')"
    say "            ($c)"
  fi
}

# ------------------------------------------------------------------ self-test
mode_self_test() {
  fails=0
  t() { # t <name> <expected> <actual>
    if [ "$2" = "$3" ]; then printf '  ok   %s\n' "$1"
    else printf '  FAIL %s: expected %s, got %s\n' "$1" "$2" "$3"; fails=$((fails+1)); fi
  }
  worst=$RC_OK; note $RC_REPAIR; t "repair raises worst"      "1" "$worst"
  worst=$RC_OK; note $RC_STOP;   t "stop raises worst"        "2" "$worst"
  worst=$RC_STOP; note $RC_REPAIR; t "repair cannot lower stop" "2" "$worst"
  worst=$RC_REPAIR; note $RC_STOP; t "stop overrides repair"  "2" "$worst"

  # Invocation surface. These fail before any tracker read, so the self-test
  # stays hermetic — the modes' behaviour against real beads is not covered
  # here and is exercised by using them.
  _self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
  "$_self" run >/dev/null 2>&1;        t "run with no epic exits 3"  "3" "$?"
  "$_self" run a b >/dev/null 2>&1;    t "run with two args exits 3" "3" "$?"
  "$_self" nosuchmode >/dev/null 2>&1; t "unknown mode exits 3"      "3" "$?"
  "$_self" >/dev/null 2>&1;            t "no mode exits 3"           "3" "$?"
  echo
  [ "$fails" -eq 0 ] && { echo "RESULT: PASS"; return 0; }
  echo "RESULT: FAIL"; return 1
}

# --------------------------------------------------------------- slot mode
#
# `bd merge-slot release` does not check the holder: any agent can release
# anyone's slot, and releasing an unheld slot succeeds silently. So the
# recovery path for a slot held by a dead agent and the way to corrupt a live
# merge are the same command — and nothing distinguishes dead from live.
#
# This reports the holder ALONGSIDE the evidence for and against that holder
# still being alive, and recommends a release only when every signal agrees.
# It never releases anything: the script is read-only, so a wrong reading here
# costs a second look rather than someone else's merge.
mode_slot() {
  [ $# -le 1 ] || die "usage: bb-status.sh slot [<epic-id>]"
  _epic="${1:-}"
  _slot="$(bd merge-slot check 2>&1)"
  echo "merge slot"

  case "$_slot" in
    *"not found"*)
      line OK "no slot bead — nothing holds a merge" "bd merge-slot check"
      return ;;
    *"available"*|*"Available"*)
      line OK "slot is FREE" "bd merge-slot check"
      return ;;
  esac

  _holder="$(printf '%s\n' "$_slot" | sed -n 's/.*Holder: *//p' | head -1)"
  _id="$(printf '%s\n' "$_slot" | sed -n 's/.*held: *\([^ ]*\).*/\1/p' | head -1)"
  [ -n "$_id" ] || _id="$(bd_prefix_slot)"
  line OK "slot HELD by: ${_holder:-unknown}" "bd merge-slot check"

  # Evidence 1 — how long. A slot acquired minutes ago is a merge in progress;
  # one acquired hours ago is a question.
  _held_since="$(bd show "$_id" --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: raise SystemExit
d = d[0] if isinstance(d, list) and d else d
print((d or {}).get("updated_at") or "")' 2>/dev/null)"
  [ -n "$_held_since" ] && line OK "  held since: $_held_since" "bd show $_id --json | jq -r .updated_at"

  # Evidence 2 — is any agent standing at all?
  _wt="$(git worktree list 2>/dev/null | grep -c "bb-worktrees\|\.claude/worktrees" || true)"
  line OK "  standing worktrees: ${_wt:-0}" "git worktree list"

  # Evidence 3 — is work still landing?
  _last="$(git log -1 --format=%cr 2>/dev/null)"
  line OK "  last commit: ${_last:-unknown}" "git log -1 --format=%cr"

  # Evidence 4 — the strongest, when a run epic is known: an agent that says
  # it is running. This is the dispatch record, not an inference.
  _live=""
  if [ -n "$_epic" ]; then
    _live="$(bd list --parent "$_epic" --metadata-field dispatch_state=running --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: d=[]
print(" ".join(x.get("id","") for x in d) if isinstance(d,list) else "")' 2>/dev/null)"
    if [ -n "$_live" ]; then
      line OK "  dispatches reporting running: $_live" \
           "bd list --parent $_epic --metadata-field dispatch_state=running"
    else
      line OK "  dispatches reporting running: none" \
           "bd list --parent $_epic --metadata-field dispatch_state=running"
    fi
  else
    line OK "  dispatch records: not checked (pass the run epic to include them)" \
         "bb-status.sh slot <epic-id>"
  fi

  # The verdict. Release is recommended only when NOTHING argues for the
  # holder. Any single signal in the holder's favour is enough to withhold it.
  if [ "${_wt:-0}" -gt 0 ] 2>/dev/null; then
    line FAIL "do NOT release — $_wt worktree(s) standing; an agent is holding work" \
         "git worktree list"; note $RC_STOP
  elif [ -n "$_live" ]; then
    line FAIL "do NOT release — a dispatch reports running: $_live" \
         "bd list --parent $_epic --metadata-field dispatch_state=running"; note $RC_STOP
  else
    line OK "no worktree and no running dispatch backs ${_holder:-the holder}" ""
    line OK "RELEASE IS DEFENSIBLE: bd merge-slot release" "bd merge-slot release"
    line OK "  (this script never releases it — read the evidence above first)" ""
  fi
}

bd_prefix_slot() {
  _p="$(bd show --json 2>/dev/null | head -0; bd list -n 1 --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: raise SystemExit
i=(d[0] if isinstance(d,list) and d else {}).get("id","")
print(i.rsplit("-",1)[0] if "-" in i else "")' 2>/dev/null)"
  [ -n "$_p" ] && printf '%s-merge-slot' "$_p"
}

# ---------------------------------------------------------------- run mode
#
# The run-level Definition of Done. `bd epic status` counts DIRECT children,
# so on a nested run epic it answers about sub-epics rather than beads — it
# read 0/2 with four leaves closed. `bd list --parent` is direct-children-only
# too, unlike `bd ready --parent` which is transitive. Neither gives a
# transitive count, so this walks the tree.
mode_run() {
  [ $# -eq 1 ] || die "usage: bb-status.sh run <epic-id>"
  _epic="$1"
  command -v python3 >/dev/null 2>&1 || die "run mode needs python3"
  _out="$(python3 - "$_epic" <<'PY'
import json, subprocess, sys

epic = sys.argv[1]

# Child types that are run bookkeeping rather than work.
SYNTHETIC = {"event", "molecule"}

def show(i):
    r = subprocess.run(["bd", "show", i, "--json"], capture_output=True, text=True)
    try:
        d = json.loads(r.stdout)
    except Exception:
        return None
    d = d[0] if isinstance(d, list) and d else d
    return d if isinstance(d, dict) and not d.get("error") else None

def children(i):
    r = subprocess.run(["bd", "list", "--parent", i, "--status",
                        "open,in_progress,blocked,deferred,closed", "--json"],
                       capture_output=True, text=True)
    try:
        d = json.loads(r.stdout)
    except Exception:
        return []
    return d if isinstance(d, list) else []

root = show(epic)
if root is None:
    print("ERR|no such bead"); raise SystemExit
if root.get("issue_type") != "epic":
    print("ERR|%s is type '%s', not an epic" % (epic, root.get("issue_type"))); raise SystemExit

md = root.get("metadata") or {}
if isinstance(md, str):
    try: md = json.loads(md)
    except Exception: md = {}

seen, leaves, nested = set(), [], False
stack = [epic]
while stack:
    cur = stack.pop()
    for c in children(cur):
        cid = c.get("id")
        if not cid or cid in seen:
            continue
        seen.add(cid)
        kind = c.get("issue_type")
        if kind == "epic":
            nested = True
            stack.append(cid)
        elif kind in SYNTHETIC:
            # `bd set-state` writes an `event` child on every health change and
            # a run writes several, so counting them as leaves would report a
            # run with no work as having beads in scope. Measured: an epic with
            # two health events and zero beads read "2 leaf bead(s)".
            continue
        else:
            leaves.append(c)

by = {}
for c in leaves:
    by.setdefault(c.get("status") or "?", []).append(c.get("id"))

admitted = md.get("admitted") or []
excluded = md.get("excluded") or []
closed = set(by.get("closed", []))
outstanding = [i for i in admitted if i not in closed] if admitted else []

print("NESTED|%s" % ("yes" if nested else "no"))
print("TOTAL|%d" % len(leaves))
for st in sorted(by):
    print("STATUS|%s|%d|%s" % (st, len(by[st]), " ".join(sorted(by[st]))))
print("ADMITTED|%d" % len(admitted))
print("EXCLUDED|%d|%s" % (len(excluded),
      "; ".join("%s (%s)" % (e.get("id"), e.get("reason", "no reason recorded"))
                for e in excluded if isinstance(e, dict))))
print("OUTSTANDING|%d|%s" % (len(outstanding), " ".join(outstanding)))
PY
)" || die "run mode: could not walk $_epic"

  case "$_out" in
    ERR\|*) die "run mode: ${_out#ERR|}" ;;
  esac

  _nested="$(printf '%s\n' "$_out" | sed -n 's/^NESTED|//p')"
  _total="$(printf '%s\n' "$_out" | sed -n 's/^TOTAL|//p')"
  _adm="$(printf '%s\n' "$_out" | sed -n 's/^ADMITTED|//p')"

  echo "run $_epic"
  if [ "$_nested" = "yes" ]; then
    line OK "epic is NESTED — count is walked, not \`bd epic status\`" \
         "bb-status.sh run $_epic"
  fi
  line OK "$_total leaf bead(s) in scope" "bd ready --parent $_epic --exclude-type=epic"
  # Ids only while the list is readable; past that the count plus the command
  # that re-derives it is more useful than a wall of ids.
  printf '%s\n' "$_out" | sed -n 's/^STATUS|//p' | while IFS='|' read -r _st _n _ids; do
    if [ "$_n" -le 8 ] 2>/dev/null; then
      line OK "  $_st ($_n): $_ids" "bb-status.sh run $_epic"
    else
      line OK "  $_st: $_n bead(s)" "bb-status.sh run $_epic"
    fi
  done

  _exc_n="$(printf '%s\n' "$_out" | sed -n 's/^EXCLUDED|\([0-9]*\)|.*/\1/p')"
  _exc="$(printf '%s\n' "$_out" | sed -n 's/^EXCLUDED|[0-9]*|//p')"
  if [ "${_exc_n:-0}" -gt 0 ] 2>/dev/null; then
    # Not a failure — a deliberately narrowed scope. But a run that reads as
    # complete while beads sat out is exactly the misreading this prevents.
    line OK "$_exc_n bead(s) EXCLUDED at admission: $_exc" "manifest .excluded"
  fi

  # Health is the one thing neither git nor bead status can express: an agent
  # that is stuck looks exactly like one that is slow. The label is the
  # orchestrator's claim; the worktree count and the last commit time are the
  # evidence for or against it, so a reader can disagree with a stale claim.
  _health="$(bd show "$_epic" --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: raise SystemExit
d = d[0] if isinstance(d, list) and d else d
labs = (d or {}).get("labels") or []
print(next((l.split(":",1)[1] for l in labs if l.startswith("health:")), ""))' 2>/dev/null)"
  _wt="$(git worktree list 2>/dev/null | grep -c "bb-worktrees\|\.claude/worktrees" || true)"
  _last="$(git log -1 --format=%cr 2>/dev/null)"
  if [ -n "$_health" ]; then
    line OK "health: $_health (worktrees: ${_wt:-0}, last commit: ${_last:-unknown})" \
         "bd show $_epic --json | jq -r '.labels[]|select(startswith(\"health:\"))'"
  else
    line FAIL "health: UNSET — a watchdog cannot tell stalled from slow" \
         "bd set-state $_epic health=healthy --reason '<why>'"; note $RC_REPAIR
  fi

  # --- readiness gate, as checks rather than as an assertion -------------
  # Two things about `bd lint`: it exits 0 even when it reports warnings, so
  # the FINDINGS are the verdict and the status is not; and it has no scope
  # flag, so it reports the whole tracker. Blaming a run for the repository's
  # other 30 beads is a false signal, so intersect with this run's queue.
  _lint="$(bd lint 2>&1 | python3 -c 'import re,subprocess,sys,json
flagged=set()
cur=None
for ln in sys.stdin:
    m=re.match(r"^(\S+) \[", ln)
    if m: cur=m.group(1)
    elif "⚠" in ln and cur: flagged.add(cur)
epic=sys.argv[1]
r=subprocess.run(["bd","ready","--parent",epic,"--exclude-type=epic","--json"],
                 capture_output=True,text=True)
try: q={x.get("id") for x in json.loads(r.stdout)}
except Exception: q=set()
print(" ".join(sorted(flagged & q)))' "$_epic" 2>/dev/null)"
  if [ -n "$_lint" ]; then
    line FAIL "bd lint flags beads in this run: $_lint" "bd lint"; note $RC_REPAIR
  else
    line OK "bd lint: nothing flagged in this run's queue" "bd lint"
  fi

  _human="$(bd human list 2>&1 | grep -c '^ *[a-z].*-' || true)"
  case "$(bd human list 2>&1)" in
    *"No human-needed"*) line OK "human queue: empty" "bd human list" ;;
    *) line FAIL "human queue: $_human item(s) awaiting a decision" "bd human list"
       note $RC_REPAIR ;;
  esac

  _notier="$(bd ready --parent "$_epic" --exclude-type=epic --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: d=[]
out=[]
for x in (d if isinstance(d,list) else []):
    md=x.get("metadata") or {}
    if isinstance(md,str):
        try: md=json.loads(md)
        except Exception: md={}
    if not md.get("tier"): out.append(x.get("id"))
print(" ".join(out))' 2>/dev/null)"
  if [ -n "$_notier" ]; then
    line FAIL "no tier on: $_notier — dispatch cannot resolve a model" \
         "bd ready --parent $_epic --exclude-type=epic --json"; note $RC_REPAIR
  else
    line OK "every ready bead carries a tier" "bd ready --parent $_epic --exclude-type=epic"
  fi

  # --- dispatched versus returned ----------------------------------------
  _running="$(bd list --parent "$_epic" --metadata-field dispatch_state=running --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: d=[]
print(" ".join(x.get("id","") for x in d) if isinstance(d,list) else "")' 2>/dev/null)"
  if [ -n "$_running" ]; then
    line OK "dispatches still running: $_running" \
         "bd list --parent $_epic --metadata-field dispatch_state=running"
  else
    line OK "dispatches still running: none" \
         "bd list --parent $_epic --metadata-field dispatch_state=running"
  fi

  _out_n="$(printf '%s\n' "$_out" | sed -n 's/^OUTSTANDING|\([0-9]*\)|.*/\1/p')"
  _out_ids="$(printf '%s\n' "$_out" | sed -n 's/^OUTSTANDING|[0-9]*|//p')"
  if [ "${_adm:-0}" -eq 0 ] 2>/dev/null; then
    line FAIL "manifest records no admitted set — DoD cannot be checked" \
         "bd show $_epic --json | jq '.[0].metadata.admitted'"; note $RC_REPAIR
  elif [ "${_out_n:-0}" -gt 0 ] 2>/dev/null; then
    line FAIL "$_out_n admitted bead(s) not closed: $_out_ids" "manifest .admitted"
    note $RC_STOP
  else
    line OK "every admitted bead is closed" "manifest .admitted"
  fi

  # --- what this script CANNOT verify ------------------------------------
  # Naming them is the point. An agent that reads a clean run report and
  # infers the whole checklist passed is the failure this list prevents.
  echo "  --  not verifiable here (agentic; check them yourself):"
  echo "      - decisions surfaced to the user before dispatch, not during"
  echo "      - implicit decisions recorded on the beads that made them"
  echo "      - tier escalations reported with their reason"
  echo "      - the final report's prose: totals, follow-ups, stashes, pushes"
}

[ $# -ge 1 ] || die "usage: bb-status.sh bead|wave|run|slot|harvest|--self-test"
sub="$1"; shift
case "$sub" in
  bead)        mode_bead "$@" ;;
  wave)        mode_wave "$@" ;;
  run)         mode_run "$@" ;;
  slot)        mode_slot "$@" ;;
  harvest)     mode_harvest "$@" ;;
  --self-test) mode_self_test; exit $? ;;
  *)           die "unknown mode: $sub" ;;
esac
exit "$worst"
