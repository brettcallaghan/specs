#!/usr/bin/env bash
# pre-flight-check.sh — run bead-bonanza's Phase 0 as a measurement, not as prose.
#
# Contract
#   Measures what Phase 0 needs and reports it as a cited fact, not
#   remembered prose. stderr: human-readable report. stdout: shell-evalable
#   RUN_DIR=, SCRIPTS=, BB_* facts — eval it. Exit 0 proceed, 1 a blocking
#   condition named in the report, 2 bad invocation.
#
# Usage
#   pre-flight-check.sh <run-id> [--self-test]
#
#   pre-flight-check.sh 1786345424-bbh                  # probe and report
#   pre-flight-check.sh 1786345424-bbh   # probe, and capture the gate baseline
#   pre-flight-check.sh --self-test                     # portability guards only
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

upper_key() { printf '%s' "$1" | tr '[:lower:]' '[:upper:]' | tr -cs '[:alnum:]' '_' | sed 's/_$//'; }

SELF="$(abs_dir "$0")/$(basename "$0")"
SCRIPTS="$(abs_dir "$0")"

# ------------------------------------------------------------------ reporting

REPORT_FAIL=0
BLOCKERS=""

say()  { printf '  %s\n' "$*" >&2; }
head_() { printf '\n%s\n' "$*" >&2; }
emit() { printf '%s=%s\n' "$1" "$2"; }
emitq() { printf '%s=%s\n' "$1" "$(printf '%s' "$2" | sed "s/'/'\\\\''/g; s/^/'/; s/$/'/")"; }
block() { BLOCKERS="$BLOCKERS
  - $*"; REPORT_FAIL=1; }

# ================================================================== self-test
#
# Every probe above is exercised against a scratch fixture, so the AGENTS.md
# targeted-test entry for plugin/skills/bead-bonanza/scripts/** has something to
# run. It must exit 0 on Linux and on macOS, and it must never touch the real
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

  # 1. Bad invocation exits 2, and --self-test does not require a run id.
  "$SELF" >/dev/null 2>&1; st "no args exits 2" 2 "$?"
  "$SELF" a b c d e >/dev/null 2>&1; st "too many args exits 2" 2 "$?"

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

  # 7. AGENTS.md block detection, both directions.
  printf '%s\n' '# x' '## bead-bonanza gates' 'body' '## model lineup' 'body' > "$fixture/AGENTS.md"
  st "lineup block detected" present "$(has_block "$fixture/AGENTS.md" '## model lineup')"
  printf '%s\n' '# x' > "$fixture/EMPTY.md"
  st "absent block reported" absent "$(has_block "$fixture/EMPTY.md" '## model lineup')"

  # 8. The RESULT-line reader. Only stdout carries verdicts, and a gate whose
  #    own output contains the word FAIL must not be read as one.
  rl='RESULT skills-strict fail 1'
  st "gate id read"      "skills-strict" "$(printf '%s' "$rl" | awk '{print $2}')"
  st "gate verdict read" "fail"          "$(printf '%s' "$rl" | awk '{print $3}')"
  st "gate key derived"  "BB_GATE_SKILLS_STRICT" "BB_GATE_$(upper_key 'skills-strict')"

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

  echo
  if [ "$st_fail" -eq 0 ]; then echo "RESULT: PASS"; return 0; fi
  echo "RESULT: FAIL"; return 1
}

has_block() {
  if grep -q "^$2\$" "$1" 2>/dev/null; then printf 'present'; else printf 'absent'; fi
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

RUN_ID=""
[ $# -ge 1 ] || die "usage: pre-flight-check.sh <run-id> [--self-test]"
[ $# -le 2 ] || die "usage: pre-flight-check.sh <run-id> [--self-test]"
for arg in "$@"; do
  case "$arg" in
    --self-test)     self_test; exit $? ;;
    -*)              die "unknown option '$arg'" ;;
    *)               [ -z "$RUN_ID" ] || die "only one run-id may be given"
                     RUN_ID="$arg" ;;
  esac
done
[ -n "$RUN_ID" ] || die "a run-id is required (or pass --self-test)"
case "$RUN_ID" in
  */*|*' '*) die "run-id must be a single path segment, got '$RUN_ID'" ;;
esac

# ================================================================ 1. the repo

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  die "not inside a git work tree"
fi
REPO="$(git rev-parse --show-toplevel)"
cd "$REPO" || die "cannot cd to repo root $REPO"

head_ "pre-flight-check — run $RUN_ID"
say "repo: $REPO"

# Prune stale worktree registrations before anything downstream trusts
# `git worktree list`. A reclaimed container deletes a worktree's directory
# without deregistering it, and a stale entry would read as in-flight forever.
git worktree prune 2>/dev/null || true
say "worktree registrations pruned"

# ================================================= 2. the project's gates

# The project declares its gates in an executable `checks` script, never in
# prose this script parses. The declaration lives in `sf.toml` at the
# repository root (ADR-022), which is a config file with a named key rather
# than prose with a shape. Resolution order, first hit wins:
#
#   $BB_CHECKS       explicit override
#   sf.toml          [checkpoints] checks = "<repo-root-relative path>"
#   scripts/checks   the conventional location, when sf.toml declares nothing
head_ "Project gates"
# One key out of a config file, not a command out of a sentence: the value is
# whatever is quoted after `checks =`, and a missing key yields empty.
_declared="$(sed -n 's/^[[:space:]]*checks[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' sf.toml 2>/dev/null | head -1)"
CHECKS=""
for _c in "${BB_CHECKS:-}" "$_declared" scripts/checks; do
  [ -n "$_c" ] || continue
  if [ -x "$_c" ]; then CHECKS="$_c"; break; fi
  if [ -f "$_c" ]; then
    block "$_c exists but is not executable — chmod +x it; a non-executable checks script is the same as none"
  fi
done
GATE_IDS=""
if [ -z "$CHECKS" ]; then
  block "no checks script found (looked at \$BB_CHECKS, sf.toml's [checkpoints] checks key, scripts/checks) — read references/project-setup.md and write one WITH the user; do not invent gates"
else
  say "checks script: $CHECKS"
  GATE_IDS="$("$CHECKS" list 2>/dev/null | tr '\n' ' ')"
  if [ -z "$GATE_IDS" ]; then
    block "$CHECKS list printed nothing — a checks script must name its gates"
  else
    say "gates: $GATE_IDS"
  fi
fi

LINEUP_BLOCK="$(has_block AGENTS.md '## model lineup')"
say "## model lineup:      $LINEUP_BLOCK"
if [ "$LINEUP_BLOCK" = absent ]; then
  block "AGENTS.md has no '## model lineup' block — read references/project-setup.md and draft it WITH the user; do not invent one"
fi

# Echo the lineup so the orchestrator does not have to open AGENTS.md at all —
# that read was Phase 0 step 1, and this replaces it.
echo_block() { awk -v h="$1" '$0==h{f=1} /^## /{if(f && $0!=h)exit} f' AGENTS.md 2>/dev/null; }
printf '\n' >&2
echo_block '## model lineup' | sed 's/^/    /' >&2


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

head_ "Run directory"
RUN_ROOT="${TMPDIR:-/tmp}/claude/bead-bonanza"
RUN_DIR="$RUN_ROOT/$RUN_ID"
if mkdir -p "$RUN_DIR" 2>/dev/null && touch "$RUN_DIR/.ok" 2>/dev/null; then
  say "run dir: $RUN_DIR"
  # -n is load-bearing: without it the second run's link is created INSIDE the
  # first run's directory instead of replacing the pointer.
  ln -sfn "$RUN_DIR" "$RUN_ROOT/current" 2>/dev/null && say "current -> $RUN_ID"
else
  block "run dir is not writable: $RUN_DIR"
  say "run dir: UNWRITABLE $RUN_DIR"
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
     block "jq is not on PATH — the close script, the retro dump and the bucketing need it"; fi

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

# worktree.baseRef — hard gate, blocking. Every bead runs in its own worktree
# (universal, not opt-in — see SKILL.md, "Every bead runs in its own
# worktree"), so an unset baseRef silently branches EVERY worktree from the
# DEFAULT branch instead of session HEAD. Measured, because this lives in
# user-scope settings that no repository owns.
USER_SETTINGS="$HOME/.claude/settings.json"
if [ "$JQ" = ok ] && [ -f "$USER_SETTINGS" ]; then
  WORKTREE_BASEREF="$(jq -r '.worktree.baseRef // "unset"' "$USER_SETTINGS" 2>/dev/null)"
  [ -n "$WORKTREE_BASEREF" ] || WORKTREE_BASEREF=unset
else
  WORKTREE_BASEREF=unknown
fi
say "worktree.baseRef: $WORKTREE_BASEREF"
if [ "$WORKTREE_BASEREF" != head ]; then
  say "  ^ NOT 'head'. isolation:\"worktree\" will branch from the default branch,"
  say "    not session HEAD, and every worktree wave loses the sprint branch's commits."
  block "worktree.baseRef is '$WORKTREE_BASEREF', not 'head' — every worktree dispatched this run will branch from the default branch instead of session HEAD, silently dropping every commit the sprint branch carries, isolated beads included. Fix: set .worktree.baseRef = \"head\" in $USER_SETTINGS (user scope; takes effect immediately, no restart), then re-run pre-flight-check.sh before dispatching."
fi


# =============================================================== 7. PR CI

head_ "CI"
# Glob each extension separately and test each file. A single `ls` over both
# globs exits nonzero whenever one of them matches nothing, which silently
# reported "no CI" in a repo that has only .yml workflows.
PR_CI=no
for wf in .github/workflows/*.yml .github/workflows/*.yaml; do
  [ -f "$wf" ] || continue
  if grep -q 'pull_request' "$wf" 2>/dev/null; then PR_CI=yes; break; fi
done
if [ "$PR_CI" = yes ]; then
  say "pull_request-triggered workflows found — a draft PR opened now gets every"
  say "wave the full CI matrix. Ask the user once."
else
  say "no pull_request trigger found — do not ask about a draft PR."
fi

# ======================================================== 8. gate baseline

head_ "Gate baseline"
GATE_BASELINE=skipped
GATE_KEYS=""
GATE_EMIT=""
if [ -n "$CHECKS" ]; then
  GATE_BASELINE=captured
  # One invocation runs every gate the project declared, in the order it
  # declared them. Parse only stdout: the contract puts RESULT lines there and
  # everything human on stderr, so a gate that prints "FAIL" in its own output
  # cannot be mistaken for a verdict.
  _results="$("$CHECKS" run 2>/dev/null)"
  while IFS= read -r line; do
    case "$line" in RESULT\ *) ;; *) continue ;; esac
    # RESULT <gate> <verdict> <exit>
    gate="$(printf '%s' "$line" | awk '{print $2}')"
    verdict="$(printf '%s' "$line" | awk '{print $3}')"
    [ -n "$gate" ] || continue
    key="BB_GATE_$(upper_key "$gate")"
    say "$gate: $verdict"
    GATE_EMIT="$GATE_EMIT
$key=$verdict"
    GATE_KEYS="$GATE_KEYS $key"
  done <<EOF
$_results
EOF
  if [ -z "$GATE_KEYS" ]; then
    say "$CHECKS run emitted no RESULT lines — treating the baseline as uncaptured"
    GATE_BASELINE=skipped
  else
    say "the bar for every bead is 'no findings beyond this baseline', not 'exit 0'."
    say "a gate with no recorded baseline is expected to exit 0."
  fi

else
  say "not captured — no checks script resolved, so every gate is expected to"
  say "exit 0 and a pre-existing failure will be misread as a regression."
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
  block "refs/dolt/data is not writable (probed: $DOLT_WRITE_PROBE — push failed). Every bead closure this run makes dies with the container when it is reclaimed; the tracker has no JSONL fallback. Fix write access (AGENT_PLUGINS_ACCESS_TOKEN and unproxied github.com egress — see docs/session-bootstrap.md), then re-run pre-flight-check.sh before Phase 1."
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
emitq RUN_DIR "$RUN_DIR"
emitq SCRIPTS "$SCRIPTS"
emitq BB_CHECKS "$CHECKS"
emit  BB_LINEUP_BLOCK "$LINEUP_BLOCK"
emitq BB_SUBAGENT_MODEL_OVERRIDE "$SUBAGENT_MODEL_OVERRIDE"
emit  BB_PR_CI "$PR_CI"
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
emit  BB_DOLT "$DOLT"
emit  BB_DOLT_TELEMETRY "$DOLT_TELEMETRY"
emit  BB_DOLT_WRITE "$DOLT_WRITE"
emit  BB_DENY_ENTRY "$DENY_ENTRY"
emit  BB_SETTINGS_LOCAL_ALLOW "$SETTINGS_LOCAL_ALLOW"
emitq BB_WORKTREE_BASEREF "$WORKTREE_BASEREF"
emit  BB_GATE_BASELINE "$GATE_BASELINE"
[ -n "$GATE_EMIT" ] && printf '%s\n' "$GATE_EMIT" | sed '/^$/d'

exit "$REPORT_FAIL"
