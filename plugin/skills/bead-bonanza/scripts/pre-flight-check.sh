#!/usr/bin/env bash
# pre-flight-check.sh — run bead-bonanza's Phase 0 as a measurement, not as prose.
#
# Contract
#   Measures what Phase 0 needs and reports it as a cited fact, not
#   remembered prose. stderr: human-readable report. stdout: shell-evalable
#   SCRIPTS= and BB_* facts — eval it. Exit 0 proceed, 1 a blocking
#   condition named in the report, 2 bad invocation.
#
#   There is NO run directory. The run's record — manifest, baseline,
#   amendments — lives on the run epic, which outlives a reclaimed container.
#
# Where a new fact goes — the rule, because there are two destinations
#   A fact about the RUN is durable: write it to the epic's manifest AND emit
#   it. A resumed session reads the manifest and has nothing else.
#   A fact about the CONTAINER is ephemeral: emit it only. In the manifest it
#   would let a run resumed in a DIFFERENT container read a stale measurement
#   of an environment that no longer exists — worse than absent, because it
#   reads as current.
#
#   Manifest + emitted:  checks_ids, checks_baseline, concurrency_cap,
#                        the model lineup
#   Emitted only:        SCRIPTS, BB_NO_AGENT, BB_TREE_CLEAN,
#                        BB_PUSH_NEEDS_REFSPEC, BB_GIT_CONFIG_WRITABLE,
#                        BB_DOLT_*, BB_BRANCH / REMOTE / UPSTREAM,
#                        BB_WORKTREE_BASEREF, BB_WORKTREE_BASEREF_SRC,
#                        BB_EVENT_LOG
#   Audited against this rule when it was written; every fact is on the side
#   the rule puts it.
#
# Usage
#   pre-flight-check.sh [<epic-id>] [--json] [--self-test]
#
#   pre-flight-check.sh                    # environment only — run this FIRST
#   pre-flight-check.sh agent-plugins-bir9 # + run dir and checks baseline
#   pre-flight-check.sh --json             # the same facts as ONE JSON object
#   pre-flight-check.sh --self-test        # portability guards only
#
#   --json is an ADDITIONAL mode, not a replacement. The default is
#   shell-evalable because the caller runs `eval "$(pre-flight-check.sh)"`, and
#   a shell cannot eval JSON. Both render from the same emit calls, so they
#   cannot carry different keys — asserted in --self-test.
#
#   The run id IS the run epic's bead id. Two modes, because admission has to
#   write to the tracker to resolve that epic and must not do so before the
#   environment pass has shown a closure can be pushed.
#
#   The script never writes user-global state (~/.claude/settings.json, the
#   global Dolt config). Where Dolt telemetry suppression or the
#   eventsapi.dolthub.com deny entry is absent, it says so, notes that any
#   related permission prompt is safe to deny, and prints the exact command
#   to run yourself if you want it suppressed.
#
# Portability — this must run on macOS (bash 3.2, BSD userland) and on Linux.
#   No associative arrays, no mapfile, no ${var^^}, no `declare -A`.
#   No `stat -c`, `sed -i`, `grep -P`, `readlink -f`, `timeout` — see the
#   portability lint in --self-test, which fails the script if one appears
#   without a probed fallback.

set -uo pipefail

die() { echo "pre-flight-check: $*" >&2; exit 2; }

# ---------------------------------------------------------------- portability

# Absolute directory of a path, without readlink -f (absent on older macOS).
abs_dir() { ( cd "$(dirname "$1")" 2>/dev/null && pwd -P ); }

# mtime of a file. GNU stat and BSD stat disagree on both the flag and the
# format string, so probe once and bind a function rather than branching at
# every call site.
if stat -f '%m' . >/dev/null 2>&1; then   # portability-lint-ok (the probe itself)
  bb_mtime() { stat -f '%m' "$1"; }          # portability-lint-ok (BSD probe)
  BB_STAT_FLAVOUR=bsd
else
  bb_mtime() { stat -c '%Y' "$1"; }          # portability-lint-ok (GNU branch)
  BB_STAT_FLAVOUR=gnu
fi


SELF="$(abs_dir "$0")/$(basename "$0")"
SCRIPTS="$(abs_dir "$0")"

# swarm_capability <exit-status> <stderr>  ->  ok | stale:unknown-command | stale:other
#
# A pure classifier so the self-test can drive it with synthetic (status,
# stderr) pairs and needs no `swarm` binary at all — old, missing, or built
# with `status` all read the same to a caller that never runs this.
swarm_capability() {
  case "$1:$2" in
    0:*) printf 'ok' ;;
    *:*[Uu]nknown*[Cc]ommand*) printf 'stale:unknown-command' ;;
    *) printf 'stale:other' ;;
  esac
}

# ------------------------------------------------------------------ reporting

REPORT_FAIL=0
BLOCKERS=""

say()  { printf '  %s\n' "$*" >&2; }
head_() { printf '\n%s\n' "$*" >&2; }
# Both output modes render from the SAME emit call, so they cannot carry
# different keys. Nothing prints until _flush; JSON_OUT picks the buffer.
_SHELL=""
_JSON=""
_json_add() { # _json_add <key> <value>
  _e=$(printf '%s' "$2" | sed 's/\\/\\\\/g; s/"/\\"/g')
  _JSON="${_JSON}  \"$1\": \"${_e}\",
"
}
emit() { _SHELL="${_SHELL}$1=$2
"; _json_add "$1" "$2"; }
emitq() { _SHELL="${_SHELL}$1=$(printf '%s' "$2" | sed "s/'/'\\\\''/g; s/^/'/; s/$/'/")
"; _json_add "$1" "$2"; }
_flush() {
  if [ "${JSON_OUT:-no}" = yes ]; then
    printf '{\n%s\n}\n' "$(printf '%s' "$_JSON" | sed '$ s/,$//')"
  else
    printf '%s' "$_SHELL"
  fi
}
block() { BLOCKERS="$BLOCKERS
  - $*"; REPORT_FAIL=1; }

# ------------------------------------------------------- the manifest, merged
#
# Every manifest write is a read-merge-write, and this one FAILS rather than
# falling back to an empty object on an unreadable manifest. Two reasons, and
# the first is the one that bites today:
#
#   A read error must not be reported as a write success. Swallowing it means
#   the caller prints "written: ..." having merged onto whatever it guessed.
#
#   `bd update --metadata` merges at the top level on bd 1.2.2 (measured:
#   writing `{"delta":9}` over `{"alpha":1,"beta":"two"}` leaves all three) —
#   so an empty base is survivable HERE. It is not survivable by contract:
#   a nested value IS replaced wholesale (`{"a":1,"b":2}` overwritten with
#   `{"c":3}` yields `{"c":3}`), and merge-on-top-level is undocumented
#   behaviour this script should not stake the admission record on.
#
# manifest_merge <epic> <key=value>...   value prefixes:
#   @path   the file's contents, as a string
#   #digits a JSON number
#   %words  whitespace-split into an array of strings
#   (none)  the literal string
manifest_merge() {
  python3 - "$@" <<'PY'
import json, subprocess, sys

eid, pairs = sys.argv[1], sys.argv[2:]

def fail(msg):
    sys.stderr.write("manifest_merge: %s\n" % msg)
    sys.exit(1)

r = subprocess.run(["bd", "show", eid, "--json"], capture_output=True, text=True)
if r.returncode != 0:
    fail("bd show %s exited %d: %s" % (eid, r.returncode, r.stderr.strip()))
try:
    d = json.loads(r.stdout)
except Exception as e:
    fail("bd show %s returned unparseable JSON (%s)" % (eid, e))
d = d[0] if isinstance(d, list) and d else d
if not isinstance(d, dict):
    fail("bd show %s returned no issue" % eid)
md = d.get("metadata") or {}
try:
    cur = json.loads(md) if isinstance(md, str) else (md or {})
except Exception as e:
    fail("%s's metadata is not JSON (%s)" % (eid, e))
if not isinstance(cur, dict):
    fail("%s's metadata is not an object" % eid)

for p in pairs:
    k, sep, v = p.partition("=")
    if not sep:
        fail("argument %r is not key=value" % p)
    if v.startswith("@"):
        try:
            with open(v[1:]) as fh:
                cur[k] = fh.read()
        except OSError as e:
            fail("cannot read %s for key %s (%s)" % (v[1:], k, e))
    elif v.startswith("#"):
        try:
            cur[k] = int(v[1:])
        except ValueError:
            fail("value for %s is not a number: %r" % (k, v[1:]))
    elif v.startswith("%"):
        cur[k] = v[1:].split()
    else:
        cur[k] = v

w = subprocess.run(["bd", "update", eid, "--metadata", json.dumps(cur)],
                   capture_output=True, text=True)
if w.returncode != 0:
    fail("bd update %s exited %d: %s" % (eid, w.returncode, w.stderr.strip()))
PY
}

# ================================================================== self-test
#
# Every probe above is exercised against a scratch fixture, so the
# targeted-test entry for plugin/skills/bead-bonanza/scripts/** has something
# to run. It must exit 0 on Linux and on macOS, and it must never touch the real
# $HOME or the real repository.

st_fail=0
st() {
  # st <name> <expected> <actual>
  if [ "$2" = "$3" ]; then
    printf '  ok   %s\n' "$1"
  else
    printf '  FAIL %s\n    expected: %s\n    actual:   %s\n' "$1" "$2" "$3"
    st_fail=1
  fi
}

self_test() {
  echo "pre-flight-check --self-test  (bash ${BASH_VERSION:-?}, stat flavour: $BB_STAT_FLAVOUR)"
  fixture="$(mktemp -d "${TMPDIR:-/tmp}/bb-preflight-selftest.XXXXXX")" || {
    echo "  FAIL cannot create fixture dir"; return 1; }
  trap 'rm -rf "$fixture"' EXIT

  # 1. Bad invocation exits 2. No args is NOT bad invocation — it is the
  #    environment-only pass, which admission runs before it has an epic id,
  #    so it may exit 0 or 1 but must never exit 2.
  "$SELF" >/dev/null 2>&1
  case "$?" in 2) st "no args is a valid mode" "not 2" "2" ;;
               *) st "no args is a valid mode" ok ok ;; esac
  "$SELF" --nope >/dev/null 2>&1; st "unknown option exits 2" 2 "$?"
  "$SELF" a b c d e >/dev/null 2>&1; st "too many args exits 2" 2 "$?"
  "$SELF" one two >/dev/null 2>&1; st "two epic ids exit 2" 2 "$?"

  # 2. mtime wrapper agrees with the platform: an integer, close to now.
  : > "$fixture/probe"
  m="$(bb_mtime "$fixture/probe" 2>/dev/null)"
  case "$m" in
    ''|*[!0-9]*) st "mtime is an integer" "integer" "$m" ;;
    *) now="$(date +%s)"; delta=$(( now - m )); [ "$delta" -lt 0 ] && delta=$(( -delta ))
       if [ "$delta" -le 300 ]; then st "mtime near now" ok ok
       else st "mtime near now" "<=300s skew" "${delta}s"; fi ;;
  esac

  # 3. abs_dir resolves without readlink -f, and finds this script's own dir.
  st "abs_dir finds SCRIPTS" "$SCRIPTS" "$(abs_dir "$0")"
  mkdir -p "$fixture/a/b"
  st "abs_dir is physical" "$(cd "$fixture/a/b" && pwd -P)" "$(abs_dir "$fixture/a/b/x")"

  # 4. The braced TMPDIR default. A bare $TMPDIR yields /claude/... on any host
  #    where TMPDIR is unset — an absolute path nobody can write.
  st "TMPDIR default" "/tmp/claude/bead-bonanza/r1" \
     "$(TMPDIR= ; unset TMPDIR; printf '%s' "${TMPDIR:-/tmp}/claude/bead-bonanza/r1")"

  # 5. ln -sfn does not nest. Without -n the second link lands INSIDE the first
  #    run's directory instead of replacing the pointer.
  mkdir -p "$fixture/run1" "$fixture/run2"
  ln -sfn "$fixture/run1" "$fixture/current"
  ln -sfn "$fixture/run2" "$fixture/current"
  st "ln -sfn repoints" "$fixture/run2" "$(readlink "$fixture/current")"
  st "ln -sfn does not nest" "" "$(ls "$fixture/run1" 2>/dev/null)"

  # 6. deny_entry_present is read-only: it reports state, never writes.
  if command -v jq >/dev/null 2>&1; then
    st "deny entry absent on missing file" absent "$(deny_entry_present "$fixture/nope.json")"
    printf '%s\n' '{"permissions":{"deny":["WebFetch(domain:example.com)"],"allow":["Bash(ls)"]}}' > "$fixture/s2.json"
    before="$(cat "$fixture/s2.json")"
    st "deny entry absent when not listed" absent "$(deny_entry_present "$fixture/s2.json")"
    st "deny_entry_present does not write" "$before" "$(cat "$fixture/s2.json")"
    printf '%s\n' '{"permissions":{"deny":["WebFetch(domain:eventsapi.dolthub.com)"]}}' > "$fixture/s3.json"
    st "deny entry present when listed" present "$(deny_entry_present "$fixture/s3.json")"
  else
    printf '  skip deny-entry tests (no jq)\n'
  fi

  printf '%s\n' '# x' > "$fixture/EMPTY.md"

  # 7. swarm_capability — the classifier, driven with synthetic (status,
  #    stderr) pairs so this needs no `swarm` binary at all.
  st "swarm capability, a working binary" "ok" "$(swarm_capability 0 '')"
  st "swarm capability, a binary predating status" "stale:unknown-command" \
     "$(swarm_capability 3 'Error: unknown command "status" for "swarm"')"
  st "swarm capability, an unrelated failure" "stale:other" \
     "$(swarm_capability 1 'panic: something else entirely')"
  st "swarm capability, exit 0 wins even with matching text" "ok" \
     "$(swarm_capability 0 'unknown command mentioned in passing')"

  # 9. Portability lint over this script's own source. Every GNU-only or
  #    bash-4-only construct must carry the `portability-lint-ok` marker,
  #    meaning it sits behind a probe. Whole-line comments are excluded — the
  #    header documents these constructs by name and must be free to do so.
  code="$fixture/self.code"
  grep -v '^[[:space:]]*#' "$SELF" | grep -v 'portability-lint-ok' > "$code"
  for pat in 'declare -A' 'mapfile' 'readlink -f' 'stat -c' 'stat -f' 'grep -P' 'sed -i' 'timeout '; do
    hits="$(grep -n -- "$pat" "$code" | grep -v 'for pat in')"
    st "no unguarded '$pat'" "" "$hits"
  done
  # ${var^^} is bash 4 only; match it without writing it literally.
  hits="$(grep -n '\${[A-Za-z_][A-Za-z_0-9]*\^\^}' "$code")"
  st "no bash-4 case expansion" "" "$hits"

  # --json renders from the same emit calls as the eval block, so the two
  # cannot drift apart in KEYS. Assert that on a synthetic emit set rather
  # than by running the probe: this self-test must stay free of git and
  # network I/O, which is what lets CI and a per-bead check run it.
  _SHELL=""; _JSON=""
  emitq SELFTEST_A "a value with 'quotes' and spaces"
  emit  SELFTEST_B plain
  emitq SELFTEST_C 'back\slash and "double"'
  _sh_keys="$(printf '%s' "$_SHELL" | sed -n 's/^\([A-Za-z_][A-Za-z_0-9]*\)=.*/\1/p' | tr '\n' ' ')"
  JSON_OUT=yes; _json_text="$(_flush)"; JSON_OUT=no
  _js_keys="$(printf '%s' "$_json_text" | sed -n 's/^  "\([^"]*\)".*/\1/p' | tr '\n' ' ')"
  st "--json carries the same keys as the eval block" "$_sh_keys" "$_js_keys"

  if command -v python3 >/dev/null 2>&1; then
    if printf '%s' "$_json_text" | python3 -c 'import json,sys; json.load(sys.stdin)' 2>/dev/null
    then st "--json is valid JSON" ok ok
    else st "--json is valid JSON" ok "PARSE FAILED"; fi
  else
    st "--json is valid JSON" skipped skipped
  fi
  _SHELL=""; _JSON=""

  echo
  if [ "$st_fail" -eq 0 ]; then echo "RESULT: PASS"; return 0; fi
  echo "RESULT: FAIL"; return 1
}

# Report whether the dolthub telemetry deny entry is present in a settings
# file. Read-only — pre-flight detects and advises, it does not write
# user-global state.
deny_entry_present() {
  _s="$1"
  _e='WebFetch(domain:eventsapi.dolthub.com)'
  command -v jq >/dev/null 2>&1 || { printf 'unknown'; return 1; }
  [ -f "$_s" ] || { printf 'absent'; return 0; }
  if jq -e --arg e "$_e" '(.permissions.deny // []) | index($e)' "$_s" >/dev/null 2>&1; then
    printf 'present'; return 0
  fi
  printf 'absent'; return 0
}

# ================================================================== arguments

# The run id IS the run epic's bead id. It is optional, and the two modes
# exist because of an ordering constraint: admission has to CREATE or adopt
# into that epic, which is a tracker write, and a tracker write must not
# happen before section 9 has shown that closures can be pushed.
#
#   pre-flight-check.sh                 environment only — run this FIRST
#   pre-flight-check.sh <epic-id>       everything, once admission has an epic
#
# Without an epic id there is no run directory and no checks baseline, and the
# report says so rather than leaving their absence to be inferred.
EPIC=""
JSON_OUT=no
[ $# -le 3 ] || die "usage: pre-flight-check.sh [<epic-id>] [--json] [--self-test]"
for arg in "$@"; do
  case "$arg" in
    --self-test)     self_test; exit $? ;;
    --json)          JSON_OUT=yes ;;
    -*)              die "unknown option '$arg'" ;;
    *)               [ -z "$EPIC" ] || die "only one epic id may be given"
                     EPIC="$arg" ;;
  esac
done
case "$EPIC" in
  */*|*' '*) die "an epic id is a single bead id, got '$EPIC'" ;;
esac

# ================================================================ 1. the repo

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  die "not inside a git work tree"
fi
REPO="$(git rev-parse --show-toplevel)"
cd "$REPO" || die "cannot cd to repo root $REPO"

head_ "pre-flight-check${EPIC:+ — run $EPIC}"
say "repo: $REPO"

# Prune stale worktree registrations before anything downstream trusts
# `git worktree list`. A reclaimed container deletes a worktree's directory
# without deregistering it, and a stale entry would read as in-flight forever.
git worktree prune 2>/dev/null || true
say "worktree registrations pruned"

# ================================================ 2. the project's checks

# The project declares its checks in `sf.toml` at the repository root — an
# array of `[[checks]]` tables carrying an `id` and a `command` each, never in
# prose this script parses. `bb-checks.py` reads that array and runs it; the
# only knob is `$BB_SF_TOML`, which moves the config file.
head_ "Project checks"
CHECKS_RUNNER="$SCRIPTS/bb-checks.py"
CHECKS_IDS=""
if [ ! -x "$CHECKS_RUNNER" ]; then
  block "$CHECKS_RUNNER is missing or not executable — it ships with this skill, so a broken checkout is the likely cause"
else
  # The runner validates the whole declaration before it would run anything,
  # so asking it for the ids is also the syntax check. Its message names the
  # fault precisely; repeating that here in weaker words would only obscure it.
  _ids_out="$("$CHECKS_RUNNER" --ids 2>&1)"; _ids_st=$?
  if [ "$_ids_st" -ne 0 ]; then
    block "the project declares no runnable checks — $_ids_out — read references/project-setup.md and declare them WITH the user; do not invent checks"
  else
    CHECKS_IDS="$(printf '%s\n' "$_ids_out" | tr '\n' ' ' | sed 's/  *$//')"
    say "checks declared: $CHECKS_IDS"
  fi
fi

# `[models]` and `[run]` are flat key reads, so they need no TOML parser:
# sf_get <section> <key> prints the quoted value, or nothing. Section-scoped,
# because `standard` under [models] and `standard` anywhere else are different
# keys. `[[checks]]` is an array of tables and is NOT readable this way — that
# is bb-checks.py's job, and why it exists.
sf_get() {
  awk -v want="[$1]" -v key="$2" '
    /^[[:space:]]*\[/ { sec=$0; gsub(/[[:space:]]/,"",sec); next }
    sec==want {
      line=$0
      sub(/^[[:space:]]*/,"",line)
      if (index(line, key) == 1) {
        rest=substr(line, length(key)+1)
        sub(/^[[:space:]]*=[[:space:]]*/,"",rest)
        if (match(rest, /^"[^"]*"/)) {
          v=substr(rest, RSTART+1, RLENGTH-2)
          print v; exit
        }
      }
    }
  ' "${BB_SF_TOML:-sf.toml}" 2>/dev/null
}

# The run epic. Verified rather than assumed: everything downstream reads the
# queue from it, and a typo'd or non-epic id would produce an empty run that
# looks like a finished one.
EPIC_KIND=""; EPIC_QUEUE=""; EVENT_LOG=""
if [ -n "$EPIC" ]; then
  # `bd show` on an unknown id prints {"error": ...} and exits 1, so a
  # non-empty stdout is not proof the bead exists. Read the type, and report
  # "missing" separately from "wrong type" — they need different fixes.
  _ej="$(bd show "$EPIC" --json 2>/dev/null)"
  EPIC_KIND="$(printf '%s' "$_ej" | python3 -c 'import json,sys
try:
    d = json.load(sys.stdin)
except Exception:
    print("!unreadable"); raise SystemExit
d = d[0] if isinstance(d, list) and d else d
if not isinstance(d, dict) or d.get("error") or not d.get("id"):
    print("!missing")
else:
    print(d.get("issue_type", "") or "!untyped")' 2>/dev/null)"
  [ -n "$EPIC_KIND" ] || EPIC_KIND="!unreadable"
  if [ "$EPIC_KIND" = "!missing" ] || [ "$EPIC_KIND" = "!unreadable" ]; then
    block "no bead $EPIC in the tracker — the run id IS the run epic's id, so admission must create or name one that exists"
  else
    if [ "$EPIC_KIND" != "epic" ]; then
      block "$EPIC is type '$EPIC_KIND', not an epic — a run is scoped by an epic; admission creates one when the user gave beads"
    else
      # --exclude-type=epic is mandatory, not tidiness: --parent is TRANSITIVE
      # and returns intermediate epics too, so without it a run would dispatch
      # its own sub-epics as if they were work.
      EPIC_QUEUE="$(bd ready --parent "$EPIC" --exclude-type=epic --json 2>/dev/null | python3 -c 'import json,sys
try: d=json.load(sys.stdin)
except Exception: d=[]
print(len(d))' 2>/dev/null)"
      [ -n "$EPIC_QUEUE" ] || EPIC_QUEUE=0
      say "run epic: $EPIC — $EPIC_QUEUE bead(s) ready in scope"
      [ "$EPIC_QUEUE" != "0" ] || say "NOTE: nothing ready. Either every bead is closed or blocked, or admission put nothing in."
    fi
  fi
fi
unset _ej

# The tier lineup is a routing fact, so it is read as configuration rather than
# as prose. Each tier resolves independently: a project may declare only the
# tiers it uses, and an undeclared tier is reported empty rather than guessed.
MODEL_LIGHT="$(sf_get models light)"
MODEL_STANDARD="$(sf_get models standard)"
MODEL_ADVANCED="$(sf_get models advanced)"
MODEL_FRONTIER="$(sf_get models frontier)"
if [ -n "$MODEL_LIGHT$MODEL_STANDARD$MODEL_ADVANCED$MODEL_FRONTIER" ]; then
  LINEUP=declared
  say "models: light=${MODEL_LIGHT:--} standard=${MODEL_STANDARD:--} advanced=${MODEL_ADVANCED:--} frontier=${MODEL_FRONTIER:--}"
else
  # Not a blocker, and deliberately not a guess. Dispatching without a `model`
  # param makes every subagent inherit the session model, which is a stateable
  # cost profile; inventing a lineup is not.
  LINEUP=undeclared
  say "no [models] lineup in sf.toml — dispatch will omit the model param and every"
  say "subagent inherits the session model. Say so once at run start."
fi

# Project-level run defaults. Anything a project genuinely owns rather than
# decides per run.
PUSH_MODE_DEFAULT="$(sf_get run push_mode)"
[ -n "$PUSH_MODE_DEFAULT" ] || PUSH_MODE_DEFAULT=unset


# =================================================== 3. beads telemetry + deny

head_ "Beads telemetry"
if command -v dolt >/dev/null 2>&1; then
  DOLT=present
  if dolt config --global --list 2>/dev/null | grep -q '^metrics.disabled *= *true'; then
    DOLT_TELEMETRY=disabled
    say "dolt present; metrics.disabled already true"
  else
    DOLT_TELEMETRY=not-suppressed
    say "dolt present; telemetry is not suppressed globally."
    say "a network permission prompt for Dolt's telemetry endpoint may appear"
    say "during this run — it is safe to deny. To suppress it permanently,"
    say "run this yourself:"
    say "  dolt config --global --add metrics.disabled true"
  fi
else
  DOLT=absent
  DOLT_TELEMETRY=na
  say "no dolt CLI — normal for a bd that embeds Dolt as a library. Nothing to configure."
fi

DENY_ENTRY="$(deny_entry_present "$HOME/.claude/settings.json")"
case "$DENY_ENTRY" in
  present) say "deny entry for eventsapi.dolthub.com: already present in ~/.claude/settings.json" ;;
  absent)  say "deny entry for eventsapi.dolthub.com: not present."
           say "a WebFetch permission prompt for that domain may appear during this"
           say "run — it is safe to deny. To suppress it permanently, add this yourself"
           say "to ~/.claude/settings.json under permissions.deny:"
           say '  "WebFetch(domain:eventsapi.dolthub.com)"' ;;
  *)       DENY_ENTRY=unknown
           say "could not read ~/.claude/settings.json to check for a deny entry (no jq)." ;;
esac

# ======================================================== 4. run dir + scripts

head_ "Run record"
if [ -z "$EPIC" ]; then
  say "no epic id given — environment-only pass; no manifest, no checks baseline."
  say "Re-run as \`pre-flight-check.sh <epic-id>\` once admission has resolved one."
else
  # There is no run directory. The manifest lives in the run epic's metadata
  # and the amendments in its comments, so the run's record is as durable as
  # the tracker and survives a reclaimed container, which a /tmp path does not.
  say "run record: metadata on $EPIC (no run directory — the tracker IS the record)"

  # The event log is the opposite kind of thing: a live STREAM for the
  # Monitor to tail, not a record. It is ephemeral by design — `swarm status`
  # is the durable view, measured from the beads and from git.
  EVENT_LOG="${TMPDIR:-/tmp}/bb-events-$EPIC.log"
  if : >> "$EVENT_LOG" 2>/dev/null; then
    say "event log: $EVENT_LOG"
    say "export BB_EVENT_LOG before installing .beads/hooks/on_update — the hook"
    say "appends to it, and an unset variable makes every write fail silently."
  else
    EVENT_LOG=""
    say "event log: UNWRITABLE — progress streaming will be unavailable; use"
    say "\`swarm status --run $EPIC\` instead."
  fi
fi
say "scripts:  $SCRIPTS"

# ========================================================== 5. push mechanics
#
# Measured, not asserted. SKILL.md used to carry "the sandbox blocks .git/config
# writes so --set-upstream never persists" as a standing claim; it is true in
# some containers and false in others, and an agent cannot tell which from prose.

head_ "Push mechanics (measured)"
BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null)"
REMOTE="$(git remote | head -1)"
[ -n "$REMOTE" ] || REMOTE=none
if [ -w .git/config ]; then GIT_CONFIG_WRITABLE=yes; else GIT_CONFIG_WRITABLE=no; fi
UPSTREAM="$(git rev-parse --abbrev-ref '@{u}' 2>/dev/null)"
[ -n "$UPSTREAM" ] || UPSTREAM=none
if [ "$UPSTREAM" = none ]; then PUSH_NEEDS_REFSPEC=yes; else PUSH_NEEDS_REFSPEC=no; fi

DEFAULT_BRANCH="$(git symbolic-ref --short "refs/remotes/$REMOTE/HEAD" 2>/dev/null | sed "s#^$REMOTE/##")"
if [ -z "$DEFAULT_BRANCH" ]; then
  for cand in main master; do
    if git show-ref --verify --quiet "refs/remotes/$REMOTE/$cand"; then DEFAULT_BRANCH="$cand"; break; fi
  done
fi
[ -n "$DEFAULT_BRANCH" ] || DEFAULT_BRANCH=unknown

say "branch:            $BRANCH"
say "remote:            $REMOTE"
say ".git/config write: $GIT_CONFIG_WRITABLE"
say "upstream:          $UPSTREAM"
say "default branch:    $DEFAULT_BRANCH"
if [ "$PUSH_NEEDS_REFSPEC" = yes ]; then
  say "=> push with an explicit refspec: git push $REMOTE $BRANCH"
else
  say "=> upstream is configured; a bare 'git push' resolves. Verify landings"
  say "   with 'git rev-parse $UPSTREAM' after fetch, not 'git status -sb'."
fi
if [ "$DEFAULT_BRANCH" != unknown ] && [ "$BRANCH" = "$DEFAULT_BRANCH" ]; then
  block "HEAD is the default branch ($BRANCH) — branch before running beads"
fi

# ============================================================= 6. environment

head_ "Environment"
if command -v bd >/dev/null 2>&1; then
  BD=ok
  BD_DB="$(bd where 2>/dev/null | sed -n 's/^ *database: *//p' | head -1)"
  [ -n "$BD_DB" ] || BD_DB=none
  say "bd:  $(command -v bd)"
  say "     database: $BD_DB"
  [ "$BD_DB" = none ] && block "bd is installed but 'bd where' resolves no database"
else
  BD=missing; BD_DB=none
  say "bd:  MISSING"
  block "bd is not on PATH — every bead operation in this run needs it"
fi

if command -v jq >/dev/null 2>&1; then JQ=ok; say "jq:  $(command -v jq)"
else JQ=missing; say "jq:  MISSING"
     block "jq is not on PATH — the close script, the retro dump and the batching need it"; fi

# `swarm manifest` and `swarm status` are required steps, not optional ones,
# so absence and staleness are both blocking. The probe is behavioural
# rather than a version string: `swarm` carries no version string yet, and
# even if it did, a binary built before `swarm status` existed would still
# report SOME version and pass a version check while still lacking the
# subcommand this run is about to invoke. Measured: a binary predating that
# command exits 3 ("unknown command") on `swarm status --help`, which is
# exactly the failure a version number would not catch.
if command -v swarm >/dev/null 2>&1; then
  # Captured into a variable and matched with `case`, never piped through
  # `grep -q`: a small writer can still lose the SIGPIPE race against a
  # reader that closes early, which makes a piped check intermittent (see
  # AGENTS.md, "Pipes hide exit statuses").
  SWARM_HELP_ERR="$(swarm status --help 2>&1 >/dev/null)"; SWARM_HELP_RC=$?
  SWARM_CAP="$(swarm_capability "$SWARM_HELP_RC" "$SWARM_HELP_ERR")"
  case "$SWARM_CAP" in
    ok)
      SWARM=ok; say "swarm: $(command -v swarm)" ;;
    stale:unknown-command)
      SWARM=stale
      say "swarm: $(command -v swarm) (STALE — no 'status' subcommand)"
      block "swarm on PATH predates 'swarm status' — rebuild it: (cd cli && go build -o /usr/local/bin/swarm ./swarm)" ;;
    *)
      SWARM=stale
      say "swarm: $(command -v swarm) (STALE — 'swarm status --help' failed: $SWARM_HELP_ERR)"
      block "swarm on PATH does not run 'swarm status --help' — rebuild it: (cd cli && go build -o /usr/local/bin/swarm ./swarm)" ;;
  esac
else
  SWARM=missing
  say "swarm: MISSING"
  block "swarm is not on PATH — swarm manifest and swarm status are required steps: (cd cli && go build -o /usr/local/bin/swarm ./swarm)"
fi

# `tomllib` is stdlib from 3.11, so probing the import settles both the
# interpreter and the version in one call.
if python3 -c 'import tomllib' >/dev/null 2>&1; then
  PY=ok; say "python3: $(command -v python3) (tomllib present)"
else
  PY=missing; say "python3: MISSING or older than 3.11"
  block "python3 3.11+ is not on PATH — bb-checks.py reads the project's [[checks]] declaration with tomllib"
fi

if [ -z "$(git status --porcelain 2>/dev/null)" ]; then
  TREE_CLEAN=yes; say "working tree: clean"
else
  TREE_CLEAN=no; say "working tree: DIRTY"
  git status --porcelain | sed 's/^/    /' >&2
  block "the working tree is dirty at t=0 — every later skip check and pathspec guard is poisoned by it"
fi

if [ -n "${CLAUDE_CODE_SUBAGENT_MODEL:-}" ]; then
  SUBAGENT_MODEL_OVERRIDE="${CLAUDE_CODE_SUBAGENT_MODEL}"
  say "CLAUDE_CODE_SUBAGENT_MODEL: $SUBAGENT_MODEL_OVERRIDE"
  say "  ^ this OVERRIDES the per-invocation model, so every tier in the dispatch"
  say "    plan is silently ignored. Unset it or record the plan as advisory."
else
  SUBAGENT_MODEL_OVERRIDE=unset
  say "CLAUDE_CODE_SUBAGENT_MODEL: unset (tier assignments will be honoured)"
fi

SETTINGS_LOCAL=.claude/settings.local.json
if [ -f "$SETTINGS_LOCAL" ] && [ "$JQ" = ok ]; then
  SETTINGS_LOCAL_ALLOW="$(jq -r '(.permissions.allow // []) | length' "$SETTINGS_LOCAL" 2>/dev/null)"
  [ -n "$SETTINGS_LOCAL_ALLOW" ] || SETTINGS_LOCAL_ALLOW=0
  say "$SETTINGS_LOCAL: $SETTINGS_LOCAL_ALLOW allow entries"
  jq -r '(.permissions.allow // [])[]' "$SETTINGS_LOCAL" 2>/dev/null | sed 's/^/    /' >&2
else
  SETTINGS_LOCAL_ALLOW=absent
  say "$SETTINGS_LOCAL: absent — expect a permission prompt per distinct command"
fi

# worktree.baseRef — hard gate, blocking. Every dispatch runs in its own worktree
# (universal, not opt-in — see SKILL.md, "Every bead runs in its own
# worktree"), so a baseRef that is not 'head' silently branches EVERY worktree
# from the DEFAULT branch instead of session HEAD.
#
# Resolved across Claude Code's documented settings precedence, HIGHEST level
# first, taking the first file that DEFINES the key:
#   1. .claude/settings.local.json  (project local)
#   2. .claude/settings.json        (shared project)
#   3. $HOME/.claude/settings.json  (user)
# A higher level OVERRIDES a lower one, so "any file says head" is the wrong
# test — a project-local 'fresh' beats a user-scope 'head' and must still block.
#
# A project-scope value IS honoured: a subagent dispatched with
# isolation:"worktree", from a session whose HEAD carried two commits absent
# from the default branch, came up on session HEAD carrying both, with
# worktree.baseRef set only in the repository's own .claude/settings.json and
# no user-scope file present at all.
USER_SETTINGS="$HOME/.claude/settings.json"
PROJECT_SETTINGS="$REPO/.claude/settings.json"
PROJECT_SETTINGS_LOCAL="$REPO/.claude/settings.local.json"

# Echo the key's value from one settings file, or nothing when that file does
# not define it. An absent file, an unreadable one and malformed JSON all read
# as "does not define it", so that level simply does not participate.
_baseref_in() {
  [ -f "$1" ] || return 0
  jq -r 'if (.worktree|type) == "object" and (.worktree|has("baseRef"))
         then (.worktree.baseRef|tostring) else empty end' "$1" 2>/dev/null
}

WORKTREE_BASEREF=unset
WORKTREE_BASEREF_SRC=none
if [ "$JQ" != ok ]; then
  WORKTREE_BASEREF=unknown
  WORKTREE_BASEREF_SRC='jq missing'
else
  for _s in "$PROJECT_SETTINGS_LOCAL" "$PROJECT_SETTINGS" "$USER_SETTINGS"; do
    _v="$(_baseref_in "$_s")"
    if [ -n "$_v" ]; then
      WORKTREE_BASEREF="$_v"
      WORKTREE_BASEREF_SRC="$_s"
      break
    fi
  done
fi
if [ "$WORKTREE_BASEREF_SRC" = none ]; then
  say "worktree.baseRef: $WORKTREE_BASEREF (no settings file on the precedence chain defines it)"
else
  say "worktree.baseRef: $WORKTREE_BASEREF (from $WORKTREE_BASEREF_SRC)"
fi
if [ "$WORKTREE_BASEREF" != head ]; then
  say "  ^ NOT 'head'. isolation:\"worktree\" will branch from the default branch,"
  say "    not session HEAD, and every worktree loses the sprint branch's commits."
  block "worktree.baseRef is '$WORKTREE_BASEREF', not 'head' — every worktree dispatched this run will branch from the default branch instead of session HEAD, silently dropping every commit the sprint branch carries, isolated beads included. Fix: set .worktree.baseRef = \"head\" in $PROJECT_SETTINGS (committed, so every clone and every fresh container inherits it) or in $USER_SETTINGS (this machine only). Either takes effect immediately, no restart. A value at a HIGHER precedence level overrides a lower one, so also clear any baseRef in $PROJECT_SETTINGS_LOCAL. Then re-run pre-flight-check.sh before dispatching."
fi


# =============================================================== 7. PR CI

# ============================================== 7b. the Agent tool assertion
#
# Whether the session has an Agent/Task tool is a harness property no script
# can see, so this is an ASSERTION the agent makes, not a probe. It is carried
# as a fact anyway: every other Phase 0 finding reaches the dispatch manifest,
# and one that does not is one a resumed session cannot re-read.
head_ "Delegation"
if [ "${BB_ASSUME_NO_AGENT:-0}" = 1 ]; then
  NO_AGENT=yes
  block "BB_ASSUME_NO_AGENT=1 — no Agent tool, so nothing can be delegated: not grooming, not planning, not the beads. Do not start a run."
else
  NO_AGENT=no
  say "assuming an Agent tool is available (set BB_ASSUME_NO_AGENT=1 if not)."
fi

# ======================================================== 8. checks baseline

head_ "Checks baseline"
CHECKS_BASELINE=skipped
if [ -z "$EPIC" ]; then
  say "not captured — no epic id was given, so there is nowhere to record it."
  say "The baseline is per-run, not per-repo: capture it on the epic-scoped pass."
elif [ -n "$CHECKS_IDS" ]; then
  # The baseline is what the project's checks PRINT on the untouched tree, plus
  # the status they exit with. Nothing is parsed: a bead compares its own run's
  # output against this text, so a line present in both was already failing and
  # a line only in the bead's is the bead's.
  # Two scratch files, deleted before this script returns. They are the
  # runner's report and the assembled text on its way into the manifest, not
  # the run's record — the record is the epic.
  _tsv="$(mktemp "${TMPDIR:-/tmp}/bb-baseline.XXXXXX")" || die "cannot create a temp file"
  _raw="$(mktemp "${TMPDIR:-/tmp}/bb-baseline-md.XXXXXX")" || die "cannot create a temp file"
  _out="$("$CHECKS_RUNNER" --report "$_tsv" 2>&1)"; _st=$?
  CHECKS_BASELINE=captured
  # One row per check, read from the runner's --report file rather than from
  # its output. The `=== <id>` delimiters share a stream with bytes the runner
  # does not control, so a tool that prints one forges a verdict — measured: a
  # check echoing `=== other FAILED (exit 9)` made an in-band reader report a
  # passing check as failed.
  #
  # A reading aid over the raw block below, never a substitute for it:
  # attribution is still a bead comparing text against text.
  _verdicts="$(awk -F'\t' '
    { printf "| `%s` | %s |\n", $1, ($2 == 0 ? "passed" : "**FAILED** (exit " $2 ")") }
  ' "$_tsv" 2>/dev/null)"
  {
    printf '# Checks baseline\n\n'
    printf 'What the project'"'"'s declared checks printed on the untouched tree,\n'
    printf 'before any bead ran, and the status they exited with.\n\n'
    printf '| check | on the untouched tree |\n|---|---|\n%s\n\n' "$_verdicts"
    if [ "$_st" -eq 0 ]; then
      printf 'Exit status: **0 — everything passed.** Every bead is expected to leave it\n'
      printf 'that way; any failure you see is yours.\n\n'
    else
      printf 'Exit status: **%s — something was ALREADY failing.** It is not yours to\n' "$_st"
      printf 'fix. The bar for your bead is "no findings beyond what is below", not\n'
      printf '"exit 0": run the checks, and compare your output against this text.\n\n'
    fi
    printf '```\n%s\n```\n' "$_out"
  } > "$_raw" 2>/dev/null
  if [ "$_st" -eq 0 ]; then
    say "baseline: everything passed (exit 0)"
  else
    say "baseline: ALREADY FAILING (exit $_st) — a bead is not expected to fix it"
  fi

  # Onto the epic. `checks_ids` goes as an array, not the space-joined string:
  # a consumer that has to re-split a string is one that will split it wrong
  # the first time an id contains a space.
  if manifest_merge "$EPIC" \
       "checks_ids=%$CHECKS_IDS" \
       "checks_baseline_status=#$_st" \
       "checks_baseline=@$_raw"
  then
    say "written: manifest on $EPIC (\`bd show $EPIC --json | jq -r '.[0].metadata.checks_baseline'\`)"
    say "INLINE it into every subagent prompt — a baseline that stops at the"
    say "planner is one the agent running the checks never sees."
  else
    block "could not write the checks baseline into $EPIC's manifest"
  fi
  rm -f "$_tsv" "$_raw"
else
  say "not captured — no checks were declared, so every check is expected to"
  say "pass and a pre-existing failure will be misread as a regression."
fi

# ================================================== 8b. concurrency cap

# How many subagents the harness will run at once, and therefore how large a
# dev pool a plan may size (N_dev <= cap - 1: the coordinator holds a slot).
#
# A RUN fact, not a container fact, even though its source is a session
# environment variable: a run resumed on a differently-configured session has
# to be able to read what the PLAN assumed, not what the current container
# happens to allow. So it goes to the manifest as well as being emitted.
head_ "Concurrency"
CONCURRENCY_CAP="${CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS:-20}"
case "$CONCURRENCY_CAP" in
  ''|*[!0-9]*)
    # A string in the manifest is worse than the default: the consumer is a
    # count comparison, and it would compare a number against text.
    say "CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS='${CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS:-}' is not a number — using 20"
    CONCURRENCY_CAP=20 ;;
esac
say "concurrency_cap: $CONCURRENCY_CAP (a plan may size N_dev <= $((CONCURRENCY_CAP - 1)))"

# Merged onto the epic on ANY epic-scoped pass. Deliberately NOT inside the
# checks-baseline branch above: a project that declares no checks still plans
# its dev pool against the cap, and a key written only on the baseline path would
# be null for exactly those runs.
if [ -n "$EPIC" ]; then
  # A JSON number, matching the shape sprint-plan's run-record.md documents.
  if manifest_merge "$EPIC" "concurrency_cap=#$CONCURRENCY_CAP"
  then
    say "written: concurrency_cap on $EPIC (\`bd show $EPIC --json | jq '.[0].metadata.concurrency_cap'\`)"
  else
    block "could not write concurrency_cap into $EPIC's manifest"
  fi
else
  say "not written — no epic id was given, so there is no manifest to merge it into."
fi

# ==================================================== 9. tracker durability
#
# Can a closure actually be persisted? Where beads syncs to a Dolt remote, a
# run whose closures cannot be pushed loses them when the container is
# reclaimed — the failure this whole ceremony exists to prevent. Measured, not
# assumed, and measured HERE as well as in any SessionStart hook: a hook is
# registered from the project root only, so a session that attached this repo
# alongside another never ran one. This MUST run before the verdict block
# below — it is the fact that decides whether the run is safe to proceed.
#
# `none` means no Dolt remote is configured, which is not a fault: a project
# may track beads another way. `no` means one IS configured and cannot be
# written — that is a blocking condition for the run.
head_ "Tracker durability"
DOLT_WRITE=none
DOLT_WRITE_PROBE=""
if [ "$BD" = ok ] && bd dolt remote list 2>/dev/null | grep -q .; then
  _probe_ref="refs/dolt/preflight-probe-$$"
  DOLT_WRITE_PROBE="git push ${REMOTE:-origin} HEAD:$_probe_ref"
  if git push -q "${REMOTE:-origin}" "HEAD:$_probe_ref" >/dev/null 2>&1; then
    git push -q "${REMOTE:-origin}" --delete "$_probe_ref" >/dev/null 2>&1 || true
    DOLT_WRITE=yes
    say "refs/dolt/data write: yes (probed: $DOLT_WRITE_PROBE)"
  else
    DOLT_WRITE=no
    say "refs/dolt/data write: NO (probed: $DOLT_WRITE_PROBE)"
  fi
else
  say "refs/dolt/data write: n/a — no Dolt remote configured"
fi
if [ "$DOLT_WRITE" = no ]; then
  block "refs/dolt/data is not writable (probed: $DOLT_WRITE_PROBE — push failed). Every bead closure this run makes dies with the container when it is reclaimed; the tracker has no JSONL fallback. Fix write access — the credential and egress path are the project's to supply — then re-run pre-flight-check.sh before Phase 1."
fi

# ================================================================= 10. verdict

head_ "Verdict"
if [ "$REPORT_FAIL" -eq 0 ]; then
  say "OK — safe to proceed to Phase 1."
else
  say "BLOCKED:"
  printf '%s\n' "$BLOCKERS" | sed '/^$/d; s/^/  /' >&2
fi

# The evalable block. Values that can contain spaces are single-quoted.
emitq SCRIPTS "$SCRIPTS"
emitq BB_CHECKS_IDS "$CHECKS_IDS"
emit  BB_CONCURRENCY_CAP "$CONCURRENCY_CAP"
emit  BB_LINEUP "$LINEUP"
emitq BB_MODEL_LIGHT "$MODEL_LIGHT"
emitq BB_MODEL_STANDARD "$MODEL_STANDARD"
emitq BB_MODEL_ADVANCED "$MODEL_ADVANCED"
emitq BB_MODEL_FRONTIER "$MODEL_FRONTIER"
emitq BB_PUSH_MODE_DEFAULT "$PUSH_MODE_DEFAULT"
emitq BB_SUBAGENT_MODEL_OVERRIDE "$SUBAGENT_MODEL_OVERRIDE"
emit  BB_NO_AGENT "$NO_AGENT"
emit  BB_TREE_CLEAN "$TREE_CLEAN"
emit  BB_PUSH_NEEDS_REFSPEC "$PUSH_NEEDS_REFSPEC"
emit  BB_GIT_CONFIG_WRITABLE "$GIT_CONFIG_WRITABLE"
emitq BB_BRANCH "$BRANCH"
emitq BB_REMOTE "$REMOTE"
emitq BB_UPSTREAM "$UPSTREAM"
emitq BB_DEFAULT_BRANCH "$DEFAULT_BRANCH"
emit  BB_BD "$BD"
emitq BB_BD_DB "$BD_DB"
emit  BB_JQ "$JQ"
emit  BB_SWARM "$SWARM"
emit  BB_DOLT "$DOLT"
emit  BB_DOLT_TELEMETRY "$DOLT_TELEMETRY"
emit  BB_DOLT_WRITE "$DOLT_WRITE"
emit  BB_DENY_ENTRY "$DENY_ENTRY"
emit  BB_SETTINGS_LOCAL_ALLOW "$SETTINGS_LOCAL_ALLOW"
emitq BB_WORKTREE_BASEREF "$WORKTREE_BASEREF"
emitq BB_WORKTREE_BASEREF_SRC "$WORKTREE_BASEREF_SRC"
emit  BB_CHECKS_BASELINE "$CHECKS_BASELINE"
emitq BB_EPIC "$EPIC"
emitq BB_EVENT_LOG "$EVENT_LOG"
emitq BB_EPIC_QUEUE "$EPIC_QUEUE"

_flush

exit "$REPORT_FAIL"
