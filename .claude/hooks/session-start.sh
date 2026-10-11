#!/bin/bash
# SessionStart hook — provision this repo's dependencies so a session can
# actually run the workflows the skills prescribe.
#
# Three jobs:
#   1. Install `bd` (Beads), which every grooming and orchestration skill in
#      this repo assumes is on PATH. Without it those skills cannot be
#      exercised at all, which is the definition of not dogfooding.
#   2. Keep the vendored shared references current. The skills are committed
#      symlinks under .claude/skills/ (each -> plugin/skills/<name>), so the
#      session runs the working tree directly — no materialization step, and
#      no global copy can shadow it.
#   3. Install the `strata` CLI from its latest release.
#
# Idempotent and non-interactive. Safe to re-run — and sometimes it must be
# re-run by hand, because this hook is not always invoked:
#
#   Claude Code registers hooks from the PROJECT ROOT's .claude/settings.json
#   only. A session that attaches this repo alongside another clones both as
#   siblings under a shared parent, that parent becomes the project root, and
#   it has no .claude/ — so this hook is never registered. Nothing here is
#   broken when that happens, and merging to the default branch does not fix
#   it. See docs/session-bootstrap.md; CLAUDE.md carries the manual fallback,
#   and an environment setup script is the durable fix.
#
# To run it by hand:
#   CLAUDE_PROJECT_DIR=/path/to/specs bash .claude/hooks/session-start.sh
set -euo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
cd "$PROJECT_DIR"

# Only provision in the remote/cloud environment. A local machine has its own
# toolchain and its own bd; reinstalling there would be surprising.
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

log() { echo "[session-start] $*"; }

# bd resolves its global config path from $HOME. Observed once in this
# environment: it created a LITERAL `./~/.config/bd/` inside the repo instead,
# which then shows up in every `git status` and eventually gets committed.
# The trigger was not reproducible, so defend both ways — pin HOME, and sweep
# the artifact if it appears anyway.
export HOME="${HOME:-/root}"
sweep_literal_tilde() {
  if [ -d "$PROJECT_DIR/~" ]; then
    rm -rf "$PROJECT_DIR/~"
    log "removed a literal ./~ directory (bd wrote its global config there)"
  fi
}
trap sweep_literal_tilde EXIT

# --- 1. bd (Beads) -----------------------------------------------------------
GOBIN_DIR="$(go env GOPATH 2>/dev/null)/bin"

# bd warns on a group/world-readable beads dir. Quiet it up front, before any
# bd call, so the warning does not appear on every invocation and train the
# agent to ignore bd's stderr.
[ -d .beads ] && chmod 700 .beads 2>/dev/null || true

if ! command -v bd >/dev/null 2>&1 && [ ! -x /usr/local/bin/bd ]; then
  # Beads embeds Dolt, whose regex support is cgo-bound to ICU. Without these
  # headers any source build fails at `unicode/uregex.h: No such file`.
  if [ ! -f /usr/include/unicode/uregex.h ]; then
    log "installing libicu-dev (ICU headers, required by the Dolt dependency)"
    apt-get install -y libicu-dev >/dev/null 2>&1 \
      || log "WARNING: libicu-dev install failed; a source build will likely fail"
  fi
  # Use the official installer rather than calling `go install` directly: it
  # prefers a prebuilt release binary and falls back to a source build only
  # when the release lookup is unavailable — which it is behind this egress
  # proxy, where api.github.com and the releases pages return 403 while
  # raw.githubusercontent.com and proxy.golang.org do not. Delegating the
  # choice means this gets faster for free if release access is ever restored,
  # with no edit here.
  log "installing bd (prebuilt if reachable, else a source build)"
  curl -sSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh 2>/dev/null \
    | bash >/dev/null 2>&1 \
    || log "WARNING: bd installer failed"
fi

# Symlink onto the default PATH. A PATH edit alone is not enough: subagents and
# hook subprocesses do not reliably inherit this session's environment, and a
# skill shelling out to `bd` from one of those would fail with "command not
# found" — which reads as a skill defect rather than a missing dependency.
if [ -x "$GOBIN_DIR/bd" ] && [ ! -e /usr/local/bin/bd ]; then
  ln -sf "$GOBIN_DIR/bd" /usr/local/bin/bd
fi
export PATH="$GOBIN_DIR:$PATH"
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export PATH=\"$GOBIN_DIR:\$PATH\"" >> "$CLAUDE_ENV_FILE"
fi

if command -v bd >/dev/null 2>&1; then
  log "bd ready: $(bd --version 2>/dev/null | grep -o 'version [0-9.]*' | head -1)"
else
  log "WARNING: bd unavailable — grooming and orchestration skills will not run"
fi

# --- 2. Beads telemetry and database ----------------------------------------
if command -v bd >/dev/null 2>&1; then
  # bd stores data in Dolt, which phones home to eventsapi.dolthub.com. Killed
  # here rather than making every session re-do it.
  if command -v dolt >/dev/null 2>&1 \
     && ! dolt config --global --list 2>/dev/null | grep -q 'metrics.disabled = true'; then
    dolt config --global --add metrics.disabled true >/dev/null 2>&1 || true
  fi
  export BD_DISABLE_TELEMETRY=1
  [ -n "${CLAUDE_ENV_FILE:-}" ] && echo 'export BD_DISABLE_TELEMETRY=1' >> "$CLAUDE_ENV_FILE"

  # --- 2a. Leftovers from the refs/dolt/data tracker --------------------------
  #
  # A container reused from before the move may still carry the GitHub proxy
  # bypass and PAT helper the old refs/dolt/data tracker needed. Left in place
  # they keep routing this repository's pushes around the proxy with a PAT that
  # lacks `workflow` scope, so remove them. The tracker now needs no credential
  # beyond the one that pushes code.
  _origin="$(git remote get-url origin 2>/dev/null || true)"
  _origin="${_origin%.git}"
  case "$_origin" in
    https://*)
      for _k in "http.${_origin}.proxy" "http.${_origin}.git.proxy" \
                "credential.${_origin}.helper" "credential.${_origin}.git.helper"; do
        git config --global --unset-all "$_k" 2>/dev/null || true
      done ;;
  esac

  # bd must never add a Dolt remote on its own: one added without the
  # `beads/data` ref points at refs/dolt/data. scripts/beads-setup.sh writes it.
  export BD_NO_REMOTE_ADOPT=1
  [ -n "${CLAUDE_ENV_FILE:-}" ] && echo 'export BD_NO_REMOTE_ADOPT=1' >> "$CLAUDE_ENV_FILE"

  # --- 2b. A database from an earlier tracker ---------------------------------
  #
  # Must run BEFORE anything reads the tracker. If `dolt_database` names a
  # database not on disk, bd CREATES an empty one under that name and reports
  # a healthy, empty tracker — measured on bd 1.2.2: `bd stats` printed
  # "Total Issues: 0", exit 0, no error. A container reused from before the
  # tracker moved holds exactly that: the old `beads` database and no `specs`.
  # Move the whole directory aside — never delete it, it may hold unpushed
  # beads — so the setup script below clones the named database fresh.
  if [ -d .beads/embeddeddolt ]; then
    _want="$(python3 -c 'import json,io;print(json.load(io.open(".beads/metadata.json",encoding="utf-8")).get("dolt_database",""))' 2>/dev/null)" || _want=""
    if [ -n "$_want" ] && [ ! -d ".beads/embeddeddolt/$_want" ]; then
      _stale=".beads/embeddeddolt.stale.$$"
      mv .beads/embeddeddolt "$_stale" \
        && log "moved aside a database from an earlier tracker (no '$_want' in it): $_stale"
    fi
  fi

  # --- 2c. The database -------------------------------------------------------
  # scripts/beads-setup.sh clones the tracker from the `beads/data` branch on a
  # fresh clone, points the Dolt remote at that branch, and pulls otherwise.
  # It verifies by issue count and exits non-zero on an empty tracker.
  BB_DOLT_FATAL=""
  if _setup_out="$(scripts/beads-setup.sh 2>&1)"; then
    printf '%s\n' "$_setup_out" | sed 's/^beads-setup: /[session-start] /'
  else
    BB_DOLT_FATAL="scripts/beads-setup.sh failed: $(printf '%s' "$_setup_out" | tr '\n' ' ' | cut -c1-400)"
  fi

  # A database whose configured prefix disagrees with the ids it holds mints a
  # SECOND id namespace silently, because the prefix lives in the database and
  # travels with the remote. Warn only: `bd rename-prefix --repair` rewrites
  # ids to random hashes when it misdetects the prefix, which is worse than the
  # stray namespace it would fix.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads/embeddeddolt ]; then
    # Read the prefix positively: under `pipefail` a transient `bd config list`
    # failure is indistinguishable from "the prefix is wrong".
    _prefix="$(bd config list 2>/dev/null | sed -n 's/^[[:space:]]*issue_prefix = //p' | head -1)" || _prefix=""
    if [ -z "$_prefix" ]; then
      log "note: could not read the tracker prefix — skipping the namespace guard"
    elif [ "$_prefix" != "specs" ]; then
      log "WARNING: tracker prefix is '$_prefix', not 'specs' — new beads will mint a second namespace"
    fi
  fi
fi

# --- 3. Vendored references --------------------------------------------------
# The skills are committed symlinks under .claude/skills/ -> plugin/skills/,
# so nothing needs materializing here; only their vendored shared references
# are regenerated in case a source changed.
if [ -x scripts/sync-shared-references.sh ]; then
  scripts/sync-shared-references.sh >/dev/null 2>&1 || log "WARNING: reference sync failed"
fi

# --- 4. strata CLI -----------------------------------------------------------
# Every model this repository ships is generated from schemas/*.strata, and
# the `strata` skill's method is compiling what it writes, so the compiler is
# provisioned for the same reason bd is. scripts/install-strata-cli.sh
# installs the latest release binary and is a no-op when it is current.
if [ "${STRATA_CLI_SKIP:-}" != "1" ]; then
  if _strata_out="$(scripts/install-strata-cli.sh 2>&1)"; then
    log "strata CLI ready: $(strata --version 2>/dev/null || echo installed)"
  else
    BB_STRATA_FATAL="${_strata_out##*$'\n'}"
  fi
fi

# --- 5. Fail loudly if the tracker is unreachable ----------------------------
# Deliberately last: everything else still gets provisioned, so the failure is
# reported against a fully set-up session rather than aborting halfway. But it
# IS a failure — this repository's issue tracker lives on the Dolt remote, and
# a session that cannot write to it will lose any bead it closes.
if [ -n "${BB_STRATA_FATAL:-}" ] && [ -z "${STRATA_CLI_SKIP:-}" ]; then
  log ""
  log "ERROR: no strata compiler in this session — $BB_STRATA_FATAL"
  log ""
  log "  Every model this repository ships is generated from schemas/*.strata"
  log "  by scripts/generate-schemas.sh, which needs this compiler. A session"
  log "  without it can still EDIT a .strata file and cannot compile one, and"
  log "  that is how schemas/feature.strata came to be authored, reviewed and"
  log "  committed having never been transpiled."
  log ""
  log "  This is fatal rather than a warning for that reason. To start a"
  log "  session that deliberately does no schema work, set STRATA_CLI_SKIP=1."
  log ""
  log "  Otherwise the environment must provide STRATA_ACCESS_TOKEN — a"
  log "  fine-grained PAT with read access to strata-lang/strata, set at"
  log "  container creation. See docs/session-bootstrap.md."
  log ""
  exit 1
fi

if [ -n "${BB_DOLT_FATAL:-}" ]; then
  log ""
  log "ERROR: the beads tracker is not usable in this session — $BB_DOLT_FATAL"
  log ""
  log "  This repository's tracker is the Dolt database on the beads/data"
  log "  branch of origin. There is no JSONL fallback by design. Do not"
  log "  create or close beads in this session: the work would be lost when"
  log "  the container is reclaimed."
  log ""
  log "  Check that origin carries the branch:"
  log "    git ls-remote origin 'refs/heads/beads/*'"
  log "  then re-run ./scripts/beads-setup.sh."
  log ""
  log "  See docs/session-bootstrap.md."
  log ""
  exit 1
fi

log "ready."
