#!/usr/bin/env bash
# Background monitor: one stdout line per bead state change, for the whole
# session. Claude Code starts it when the plugin declares it and delivers each
# stdout line as a notification, so nothing has to arm a watch at the right
# moment.
#
# WHAT IT TAILS, AND WHY IT GLOBS RATHER THAN BEING TOLD.
# The run's event log is `${TMPDIR:-/tmp}/bb-events-<run-epic>.log`, written by
# the `.beads/hooks/on_update` hook and named by pre-flight-check.sh. A monitor
# is spawned by the harness, not by the orchestrator's shell, so the
# `BB_EVENT_LOG` that pre-flight emits for that shell to `eval` never reaches
# this process -- `tail -F "$BB_EVENT_LOG"` here would follow an empty path.
# The epic id is not knowable at start either: this process outlives any one
# run and starts before the first one is scoped.
#
# So it watches the DIRECTORY and follows whatever appears. A run that starts
# an hour from now is picked up without anyone restarting anything, which is
# the always-on property that makes the per-wave ceremony unnecessary.
#
# STDOUT IS THE PRODUCT. Every line printed here becomes a notification, so
# this prints event lines and nothing else -- no banners, no status, no
# "watching N files". Diagnostics go to stderr.
#
# Environment:
#   BB_EVENT_DIR    directory to watch      (default ${TMPDIR:-/tmp})
#   BB_EVENT_GLOB   filename pattern        (default bb-events-*.log)
#   BB_EVENT_POLL   seconds between reglobs (default 5)
set -uo pipefail

DIR="${BB_EVENT_DIR:-${TMPDIR:-/tmp}}"
GLOB="${BB_EVENT_GLOB:-bb-events-*.log}"
POLL="${BB_EVENT_POLL:-5}"
case "$POLL" in ''|*[!0-9.]*) POLL=5 ;; esac

PIDS=""
cleanup() { for _p in $PIDS; do kill "$_p" 2>/dev/null; done; }
# INT/TERM must EXIT, not just clean up. A trap handler that returns normally
# resumes the loop, so `trap cleanup TERM` alone makes this process ignore the
# signal the harness stops a monitor with -- it detaches and runs forever.
trap cleanup EXIT
trap 'cleanup; exit 143' INT TERM

# One `tail` per log, started once and never restarted. The start offset is
# the whole subtlety, and it differs by WHEN the log appeared:
#
#   already there when the monitor started -> `-n 0`. It is a backlog, not
#     news. A monitor restarting mid-run must not replay the run so far as
#     fresh notifications.
#   appeared afterwards -> `-n +1`, from byte one. Everything in it was
#     written while this monitor was running, so nothing in it is old -- and
#     a poll interval's worth of events lands in a log BEFORE the reglob
#     notices the file, so `-n 0` here would silently drop a run's first
#     seconds.
#
# `-q` suppresses `==> file <==` headers: each would be a notification saying
# nothing, and the bead id in the line already carries the project. `-F`
# follows by name and retries, so a log truncated or recreated between runs
# keeps working.
mode_watch() {
  seen="" first=1
  while :; do
    for f in "$DIR"/$GLOB; do
      [ -f "$f" ] || continue
      case "$seen" in *"|$f|"*) continue ;; esac
      seen="$seen|$f|"
      if [ "$first" = 1 ]; then tail -n 0 -q -F "$f" 2>/dev/null &
      else                      tail -n +1 -q -F "$f" 2>/dev/null &
      fi
      PIDS="$PIDS $!"
    done
    first=0
    sleep "$POLL"
  done
}

# ------------------------------------------------------------------ self-test
# No database, no repository, no network: temp directory, files appended by
# hand. What it proves is that the monitor picks up a log created AFTER it
# started, and a second one after that -- the property that lets it replace an
# armed-per-wave watch.
mode_self_test() {
  t=0 f=0
  ok()  { t=$((t+1)); printf '  ok   %s\n' "$1"; }
  bad() { t=$((t+1)); f=$((f+1)); printf '  FAIL %s\n' "$1"; }
  is()  { if [ "$2" = "$3" ]; then ok "$1"; else bad "$1 (expected [$3], got [$2])"; fi; }

  tmp="$(mktemp -d)"
  mon=""
  stop() { [ -n "$mon" ] && { pkill -P "$mon" 2>/dev/null; kill "$mon" 2>/dev/null; }; mon=""; sleep 0.4; }
  trap 'stop; rm -rf "$tmp"' EXIT
  self="$(cd "$(dirname "$0")" && pwd)/$(basename "$0")"
  out="$tmp/out"; : > "$out"
  start() { BB_EVENT_DIR="$tmp" BB_EVENT_POLL=0.2 "$self" > "$out" 2>/dev/null & mon=$!; sleep 1; }

  echo "bb-events-monitor --self-test"

  start
  is "no output before any log exists" "$(cat "$out")" ""

  # A log that did not exist when the monitor started. The append happens
  # BEFORE the reglob notices the file -- which is the real ordering, and the
  # one a naive `-n 0` drops.
  printf 'T1\tzz-aaa\tupdate\tclosed\n' >> "$tmp/bb-events-zz-run1.log"
  sleep 1.5
  is "reports a log created after it started, from byte one" \
     "$(cat "$out")" "T1	zz-aaa	update	closed"

  # A second run's log, later still, with no restart.
  printf 'T2\tzz-bbb\tcreate\topen\n' >> "$tmp/bb-events-zz-run2.log"
  sleep 1.5
  is "picks up a second log without a restart" "$(grep -c . "$out")" "2"
  is "no ==> file <== headers on stdout" "$(grep -c '==>' "$out")" "0"

  # Appends to the first log still flow once a second is being followed.
  printf 'T3\tzz-ccc\tupdate\tin_progress\n' >> "$tmp/bb-events-zz-run1.log"
  sleep 1.5
  is "still following the first log" "$(grep -c 'zz-ccc' "$out")" "1"

  # Restart with both logs already on disk: a backlog, not news.
  stop
  : > "$out"
  start
  sleep 0.5
  is "a restart does not replay the backlog" "$(cat "$out")" ""
  printf 'T4\tzz-ddd\tupdate\tclosed\n' >> "$tmp/bb-events-zz-run1.log"
  sleep 1.5
  is "but does report the next event after a restart" "$(grep -c 'zz-ddd' "$out")" "1"

  # SIGTERM must actually stop it. A trap handler that cleans up and RETURNS
  # resumes the watch loop, and the monitor detaches and runs forever.
  m="$mon"
  kill "$m" 2>/dev/null
  sleep 1
  is "SIGTERM stops the monitor" "$(kill -0 "$m" 2>/dev/null && echo alive || echo gone)" "gone"
  is "and takes its tails with it" "$(pgrep -P "$m" 2>/dev/null | wc -l)" "0"
  mon=""

  stop

  echo
  if [ "$f" -eq 0 ]; then echo "RESULT: PASS ($t assertions)"; return 0; fi
  echo "RESULT: FAIL ($f of $t)"; return 1
}

case "${1:-}" in
  --self-test) mode_self_test ;;
  "")          mode_watch ;;
  *)           echo "usage: $(basename "$0") [--self-test]" >&2; exit 2 ;;
esac
