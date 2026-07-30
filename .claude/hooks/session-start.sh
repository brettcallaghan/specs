#!/bin/bash
# SessionStart hook — provision this repo's dependencies so a session can
# actually run the workflows the skills prescribe.
#
# Two jobs:
#   1. Install `bd` (Beads), which every grooming and orchestration skill in
#      this repo assumes is on PATH. Without it those skills cannot be
#      exercised at all, which is the definition of not dogfooding.
#   2. Materialize the working-tree skills into .claude/skills/ so the
#      session runs the plugin it is editing rather than a published copy.
#
# Idempotent and non-interactive. Safe to re-run.
set -euo pipefail

PROJECT_DIR="${CLAUDE_PROJECT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
cd "$PROJECT_DIR"

# Only provision in the remote/cloud environment. A local machine has its own
# toolchain and its own bd; reinstalling there would be surprising.
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

log() { echo "[session-start] $*"; }

# --- 1. bd (Beads) -----------------------------------------------------------
# Pinned rather than @latest: an unpinned toolchain means a session can fail
# for a reason unrelated to anything in this repo, and the skills' `bd` flag
# probing already handles version drift deliberately.
BD_VERSION="v1.1.2"
export GOPATH="${GOPATH:-/root/go}"
export GOBIN="${GOBIN:-$GOPATH/bin}"
export PATH="$GOBIN:$PATH"

# bd warns on a group/world-readable beads dir. Quiet it up front, before any
# bd call, so the warning does not appear on every invocation and train the
# agent to ignore bd's stderr.
[ -d .beads ] && chmod 700 .beads 2>/dev/null || true

if command -v bd >/dev/null 2>&1; then
  log "bd already present: $(bd --version 2>&1 | grep -o 'bd version.*' | head -1)"
else
  # Beads embeds Dolt, whose regex support is cgo bound to ICU. Without these
  # headers the build fails at `unicode/uregex.h: No such file or directory`.
  if [ ! -f /usr/include/unicode/uregex.h ]; then
    log "installing libicu-dev (ICU headers, required by the Dolt dependency)"
    apt-get install -y libicu-dev >/dev/null 2>&1 \
      || log "WARNING: libicu-dev install failed; the bd build will likely fail"
  fi
  log "building bd $BD_VERSION from source (a few minutes on a cold container)"
  if go install "github.com/steveyegge/beads/cmd/bd@$BD_VERSION" 2>&1 | tail -3; then
    log "bd installed: $(bd --version 2>&1 | grep -o 'bd version.*' | head -1)"
  else
    log "WARNING: bd install failed — grooming and orchestration skills will not run"
  fi
fi

# Persist PATH so later tool calls in this session find bd.
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export PATH=\"$GOBIN:\$PATH\"" >> "$CLAUDE_ENV_FILE"
fi

# --- 2. Beads telemetry and database ----------------------------------------
if command -v bd >/dev/null 2>&1; then
  # bd stores data in Dolt, which phones home to eventsapi.dolthub.com. The
  # bead-bonanza skill tells its own users to kill this; do it here too rather
  # than making every session re-do it.
  if command -v dolt >/dev/null 2>&1 \
     && ! dolt config --global --list 2>/dev/null | grep -q 'metrics.disabled = true'; then
    dolt config --global --add metrics.disabled true >/dev/null 2>&1 || true
  fi
  export BD_DISABLE_TELEMETRY=1
  [ -n "${CLAUDE_ENV_FILE:-}" ] && echo 'export BD_DISABLE_TELEMETRY=1' >> "$CLAUDE_ENV_FILE"

  # A fresh container has .beads/ metadata but no Dolt database — the database
  # directory is gitignored, so it never survives a clone. Without this, every
  # bd command fails with "no beads database found" and the grooming and
  # orchestration skills cannot run at all.
  if ! bd list >/dev/null 2>&1; then
    log "no beads database in this clone — initializing"
    # `bd init` scaffolds agent config files AND git-commits them. That
    # scaffolding is already committed in this repo, so on a fresh container
    # the only thing genuinely needed is the (gitignored) Dolt database.
    # Record HEAD so any commit init makes can be unwound: a session hook must
    # never leave a commit the user did not ask for.
    _head_before="$(git rev-parse HEAD 2>/dev/null || echo none)"
    if bd init --quiet >/dev/null 2>&1; then
      _head_after="$(git rev-parse HEAD 2>/dev/null || echo none)"
      if [ "$_head_before" != "$_head_after" ] && [ "$_head_before" != none ]; then
        git reset --soft "$_head_before" 2>/dev/null || true
        git restore --staged . 2>/dev/null || true
        log "unwound the commit bd init made (scaffolding is already in the repo)"
      fi
      # .beads/issues.jsonl is the durable export and the only thing version
      # control carries. Restore it into the fresh database when present.
      if [ -s .beads/issues.jsonl ]; then
        bd import .beads/issues.jsonl >/dev/null 2>&1 \
          && log "restored beads from .beads/issues.jsonl" \
          || log "WARNING: bd import of .beads/issues.jsonl failed"
      fi
      log "beads database ready"
    else
      log "WARNING: bd init failed — bd-dependent skills will not run"
    fi
  fi
fi

# --- 3. Dogfood the plugin ---------------------------------------------------
if [ -x scripts/sync-shared-references.sh ]; then
  scripts/sync-shared-references.sh >/dev/null 2>&1 || log "WARNING: reference sync failed"
fi
if [ -x scripts/dogfood-install.sh ]; then
  log "materializing working-tree skills into .claude/skills/"
  scripts/dogfood-install.sh 2>&1 | sed 's/^/[session-start]   /'
fi

log "ready."
