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
  # bd stores data in Dolt, which phones home to eventsapi.dolthub.com. The
  # bead-bonanza skill tells its own users to kill this; do it here too rather
  # than making every session re-do it.
  if command -v dolt >/dev/null 2>&1 \
     && ! dolt config --global --list 2>/dev/null | grep -q 'metrics.disabled = true'; then
    dolt config --global --add metrics.disabled true >/dev/null 2>&1 || true
  fi
  export BD_DISABLE_TELEMETRY=1
  [ -n "${CLAUDE_ENV_FILE:-}" ] && echo 'export BD_DISABLE_TELEMETRY=1' >> "$CLAUDE_ENV_FILE"

  # The Dolt database is gitignored — only .beads/issues.jsonl is committed —
  # so a fresh container has no database and every bd command fails with "no
  # beads database found". Test for the database directory rather than running
  # bd: an unrelated bd failure should not trigger a re-init.
  if [ -d .beads ] && [ ! -d .beads/embeddeddolt ] && [ ! -d .beads/dolt ]; then
    log "no beads database in this clone — reconstructing"
    # `bd init` scaffolds agent config files AND git-commits them. That
    # scaffolding is already committed in this repo, so on a fresh container
    # the only thing genuinely needed is the (gitignored) Dolt database.
    # Record HEAD so any commit init makes can be unwound: a session hook must
    # never leave a commit the user did not ask for.
    _head_before="$(git rev-parse HEAD 2>/dev/null || echo none)"
    if BD_NON_INTERACTIVE=1 bd init --non-interactive >/dev/null 2>&1; then
      _head_after="$(git rev-parse HEAD 2>/dev/null || echo none)"
      if [ "$_head_before" != "$_head_after" ] && [ "$_head_before" != none ]; then
        git reset --soft "$_head_before" 2>/dev/null || true
        git restore --staged . 2>/dev/null || true
        log "unwound the commit bd init made (scaffolding is already in the repo)"
      fi
      # .beads/issues.jsonl is the durable export and the only thing version
      # control carries. Restore it into the fresh database when present.
      if [ -s .beads/issues.jsonl ]; then
        BD_NON_INTERACTIVE=1 bd import .beads/issues.jsonl >/dev/null 2>&1 \
          && log "beads DB reconstructed from .beads/issues.jsonl" \
          || log "WARNING: bd import failed — DB is empty but usable"
      else
        log "beads DB created (no .beads/issues.jsonl to restore)"
      fi
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

# --- 4. strata CLI -----------------------------------------------------------
# The `strata` skill teaches an agent to write .strata schemas, and its entire
# method is `strata build --target ast` — the compiler is the only reliable
# validator of the language. Without the binary that skill degrades to
# guessing, so it gets provisioned here for the same reason bd does.
#
# Built from the strata repo at HEAD rather than a tagged release: "latest
# code" is what this is for, and the published binaries lag main.
#
# Runs last because it is the long pole (~2 min cold, no-op warm). If the hook
# is ever killed on a timeout, everything above it has already landed.
if [ "${STRATA_CLI_SKIP:-}" != "1" ]; then
  STRATA_CACHE="$HOME/.cache/strata-cli"
  STRATA_CLONE="$HOME/.cache/strata-src"
  STRATA_BIN=/usr/local/bin/strata
  STRATA_STAMP="$STRATA_CACHE/built-commit"
  mkdir -p "$STRATA_CACHE"

  # cargo lives in ~/.cargo/bin, which a hook subprocess may not have on PATH.
  if [ -d "$HOME/.cargo/bin" ]; then
    export PATH="$HOME/.cargo/bin:$PATH"
  fi

  # Prefer a checkout the session already has: when strata is attached as a
  # source, that clone is the one the user is actually working in, and
  # building it means the CLI matches the code in front of them. The private
  # clone is the fallback for sessions where strata is not attached.
  strata_src=""
  for candidate in "${STRATA_SRC_DIR:-}" "$PROJECT_DIR/../strata" "$HOME/strata" \
                   "$STRATA_CLONE"; do
    [ -n "$candidate" ] || continue
    [ -f "$candidate/crates/strata-cli/Cargo.toml" ] || continue
    strata_src="$(cd "$candidate" && pwd)"
    break
  done

  if [ -z "$strata_src" ]; then
    log "cloning strata-lang/strata into $STRATA_CLONE"
    # Blobless rather than shallow: `scripts/sync-strata-docs.sh` stamps each
    # vendored doc with the commit that last touched it, which a --depth 1
    # clone cannot answer. Falls back to a full clone if the git proxy in this
    # environment refuses partial-clone filters.
    if git clone --filter=blob:none --quiet https://github.com/strata-lang/strata.git \
         "$STRATA_CLONE" >>"$STRATA_CACHE/build.log" 2>&1 \
       || git clone --quiet https://github.com/strata-lang/strata.git \
            "$STRATA_CLONE" >>"$STRATA_CACHE/build.log" 2>&1; then
      strata_src="$STRATA_CLONE"
    else
      log "WARNING: could not clone strata (see $STRATA_CACHE/build.log) — CLI unavailable"
    fi
  fi

  if [ -n "$strata_src" ]; then
    # Fast-forward only, and only when there is nothing to lose. A session may
    # be mid-change in an attached strata checkout — this hook runs in that
    # repo too — and a hook that resets or rebases someone's branch is data
    # loss, not convenience. Anything other than a clean default branch gets
    # built exactly as it sits.
    # `|| true` is load-bearing: this script runs under `set -o pipefail`, and
    # a clone with no origin/HEAD ref (git does not always create one) makes
    # symbolic-ref fail, which would take the whole hook down mid-pipeline.
    _default="$(git -C "$strata_src" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||' || true)"
    _default="${_default:-main}"
    _branch="$(git -C "$strata_src" rev-parse --abbrev-ref HEAD 2>/dev/null || echo '')"
    if [ "$_branch" != "$_default" ]; then
      log "strata: checkout is on '$_branch', not '$_default' — building it as-is"
    elif [ -n "$(git -C "$strata_src" status --porcelain 2>/dev/null)" ]; then
      log "strata: checkout has uncommitted changes — building it as-is"
    else
      git -C "$strata_src" fetch --quiet origin "$_default" >>"$STRATA_CACHE/build.log" 2>&1 \
        && git -C "$strata_src" merge --ff-only --quiet "origin/$_default" >>"$STRATA_CACHE/build.log" 2>&1 \
        || log "strata: fetch/fast-forward failed — building the checkout as-is"
    fi

    _head="$(git -C "$strata_src" rev-parse HEAD 2>/dev/null || echo unknown)"
    if [ -x "$STRATA_BIN" ] && [ "$(cat "$STRATA_STAMP" 2>/dev/null)" = "$_head" ]; then
      log "strata CLI current: $("$STRATA_BIN" --version 2>/dev/null || echo installed) (${_head:0:8})"
    else
      log "building strata CLI from $strata_src @ ${_head:0:8} (~2 min cold, cached after)"
      # -p strata-cli, not the workspace: this needs the binary, not the WASM
      # crate, and skipping it saves both time and disk.
      if cargo build --release -p strata-cli \
           --manifest-path "$strata_src/Cargo.toml" >>"$STRATA_CACHE/build.log" 2>&1; then
        # Install a copy rather than a symlink into target/: `cargo clean` in
        # the strata checkout is a routine thing to do, and it would leave a
        # dangling /usr/local/bin/strata that reads as a broken install.
        if install -m 0755 "$strata_src/target/release/strata" "$STRATA_BIN" 2>>"$STRATA_CACHE/build.log"; then
          printf '%s' "$_head" > "$STRATA_STAMP"
          log "strata CLI ready: $("$STRATA_BIN" --version 2>/dev/null || echo installed)"
        else
          log "WARNING: could not install to $STRATA_BIN — see $STRATA_CACHE/build.log"
        fi
      else
        log "WARNING: strata build failed — see $STRATA_CACHE/build.log"
      fi
    fi

    # The skill's references are vendored copies of strata's own docs. Report
    # drift; never write, because a hook that edits tracked files turns every
    # session start into an unexplained diff.
    if [ -x scripts/sync-strata-docs.sh ]; then
      STRATA_SRC_DIR="$strata_src" scripts/sync-strata-docs.sh --check >/dev/null 2>&1 \
        || log "note: vendored strata docs are behind — run scripts/sync-strata-docs.sh"
    fi
  fi
fi

log "ready."
