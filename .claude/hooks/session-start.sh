#!/bin/bash
# SessionStart hook — provision this repo's dependencies so a session can
# actually run the workflows the skills prescribe.
#
# Two jobs:
#   1. Install `bd` (Beads), which every grooming and orchestration skill in
#      this repo assumes is on PATH. Without it those skills cannot be
#      exercised at all, which is the definition of not dogfooding.
#   2. Keep the vendored shared references current and build the `strata`
#      CLI. The skills themselves are committed symlinks under .claude/skills/
#      (each -> plugin/skills/<name>), so the session runs the working tree
#      directly — no materialization step, and no global copy can shadow it.
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
#   CLAUDE_PROJECT_DIR=/path/to/agent-plugins bash .claude/hooks/session-start.sh
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

  # --- 2a. Write access to refs/dolt/data -----------------------------------
  #
  # The tracker's source of truth is the Dolt remote (`refs/dolt/data` on
  # origin), not a committed JSONL export. That makes WRITE access to that ref
  # a hard precondition for this repository, not a nice-to-have: a session that
  # can read but not write can close a bead and lose it on container reclaim.
  #
  # There is deliberately NO JSONL fallback. Carrying both models means two
  # code paths, a rule for which one wins, and every reader having to ask which
  # mode it is in — more complexity than either model alone, and it preserves
  # exactly the machinery agent-plugins-92q.31 retired. An environment that
  # cannot reach the tracker is broken and says so.
  #
  # Two settings make it work, and both are global git config rather than
  # per-command environment juggling, because `bd dolt push` shells out to git
  # from wherever the agent happens to run it:
  #
  #   1. A credential helper that reads the PAT from the environment at call
  #      time. The token is never written to disk.
  #   2. Proxy bypass for github.com. The egress proxy REPLACES the
  #      Authorization header with a brokered GitHub App installation token,
  #      which is scoped to refs/heads and returns 403 on refs/dolt/*. This was
  #      misdiagnosed for a full session as "GitHub refuses refs/dolt/* writes";
  #      it is not. Proof: a push with a deliberately INVALID token succeeded,
  #      which is only possible if the proxy substituted its own credential.
  #      Direct egress to github.com is permitted, so bypassing the proxy lets
  #      the real PAT through.
  #
  # The token name is this repository's environment contract, not the skill's:
  # bead-bonanza requires only "write access to refs/dolt/data" and never names
  # a credential. See docs/session-bootstrap.md.
  BB_DOLT_FATAL=""
  BB_STRATA_FATAL="${BB_STRATA_FATAL:-}"
  if [ -z "${AGENT_PLUGINS_ACCESS_TOKEN:-}" ]; then
    BB_DOLT_FATAL="AGENT_PLUGINS_ACCESS_TOKEN is not set in this container."
  else
    git config --global credential.helper \
      '!f() { echo username=x-access-token; echo "password=$AGENT_PLUGINS_ACCESS_TOKEN"; }; f'
    git config --global 'http.https://github.com/.proxy' ''

    # Verify rather than assume. Pushing HEAD (already on the remote in any
    # normal clone) to a scratch ref under refs/dolt/ transfers no objects and
    # exercises the exact permission that matters, then removes it.
    _probe="refs/dolt/write-probe-$$"
    if git push -q origin "HEAD:$_probe" >/dev/null 2>&1; then
      git push -q origin --delete "$_probe" >/dev/null 2>&1 || \
        log "note: left a stray probe ref $_probe on origin — delete it by hand"
      log "verified write access to refs/dolt/* (probe pushed and removed)"
    else
      BB_DOLT_FATAL="a probe push to refs/dolt/* was rejected."
    fi
  fi

  # --- 2b. The database -------------------------------------------------------
  # Gitignored, so a fresh container has none and every bd command fails with
  # "no beads database found". Test for the directory rather than running bd:
  # an unrelated bd failure should not trigger a re-bootstrap.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads ] \
     && [ ! -d .beads/embeddeddolt ] && [ ! -d .beads/dolt ]; then
    log "no beads database in this clone — bootstrapping from the Dolt remote"
    _bootstrap_out="$(bd bootstrap 2>&1)"

    # Verify the POST-CONDITION, never the exit code. `bd bootstrap` exits 0
    # after doing nothing: its "database already exists" probe matches the
    # .beads DIRECTORY, which is committed here, so a clone that produced no
    # database still reports success. Trusting the status is how every fresh
    # session silently came up with an empty tracker.
    if [ -d .beads/embeddeddolt ] || [ -d .beads/dolt ]; then
      log "beads DB bootstrapped from origin refs/dolt/data"
    else
      BB_DOLT_FATAL="bd bootstrap produced no database (it may have exited 0 anyway). Output: $(printf '%s' "$_bootstrap_out" | tr '\n' ' ' | cut -c1-300)"
    fi
  fi

  # The database name must match the one on the remote, or bd silently creates
  # an EMPTY database under the configured name and reports no issues. Measured:
  # .beads/metadata.json said `spec_driven_dev` while the remote carries `beads`,
  # and `bd stats` created the empty one rather than reporting the mismatch.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads/embeddeddolt ] && command -v jq >/dev/null 2>&1; then
    _want="$(jq -r '.dolt_database // empty' .beads/metadata.json 2>/dev/null)"
    if [ -n "$_want" ] && [ ! -d ".beads/embeddeddolt/$_want" ]; then
      log "WARNING: metadata.json names database '$_want' but it is absent from .beads/embeddeddolt/ — present: $(ls .beads/embeddeddolt/ 2>/dev/null | tr '\n' ' ')"
    fi
  fi

  # A database whose configured prefix disagrees with the ids it holds mints a
  # SECOND id namespace silently — measured: a bootstrapped clone carrying 146
  # `agent-plugins-*` issues minted `spec-driven-dev-w0n`, because the prefix
  # lives in the database config and travels with the remote. Repaired at the
  # source and pushed, so this is a regression guard rather than a fix-up: it
  # should never fire. `--repair` rewrites no ids when they are already correct.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads/embeddeddolt ]; then
    # Read the prefix positively rather than testing for its absence: under
    # `pipefail` a transient `bd config list` failure is indistinguishable from
    # "the prefix is wrong", and acting on that ran a repair that then errored
    # with "new prefix is the same as current prefix" — two alarming warnings
    # about a database that was correct all along. Only act on a value actually
    # read.
    _prefix="$(bd config list 2>/dev/null | sed -n 's/^[[:space:]]*issue_prefix = //p' | head -1)" || _prefix=""
    if [ -z "$_prefix" ]; then
      log "note: could not read the tracker prefix — skipping the namespace guard"
    elif [ "$_prefix" != "agent-plugins" ]; then
      log "WARNING: tracker prefix is '$_prefix', not 'agent-plugins' — repairing"
      bd rename-prefix agent-plugins --repair >/dev/null 2>&1 \
        && log "repaired tracker prefix -> agent-plugins" \
        || log "WARNING: prefix repair failed — new beads may mint a second namespace"
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

  # strata-lang/strata is private, so an unauthenticated clone fails with
  # "could not read Username for 'https://github.com'" — and the session's own
  # git credentials do not cover it: the egress proxy injects credentials
  # scoped to the repositories the session attached, not to strata. A
  # fine-grained PAT with read access to that repo, supplied as
  # STRATA_ACCESS_TOKEN in the environment's variables, is what closes the gap.
  #
  # The token is handed to git through a credential helper rather than being
  # written into the remote URL. A URL carrying the token is persisted in the
  # clone's .git/config, where it survives the session on disk and leaks into
  # any `git remote -v` output; the helper reads $STRATA_ACCESS_TOKEN at call
  # time, so only the shell snippet is stored. Passed via `git clone -c`, it
  # authenticates the clone AND persists into the clone's local config, so the
  # fetch/fast-forward below reuses it with no further work.
  #
  # Scoped to this clone's local config, never global: a global github.com
  # helper would also answer for this repository's own pushes, and a PAT
  # scoped to strata would be rejected there.
  STRATA_CRED_HELPER='!f() { [ "$1" = get ] || exit 0; [ -n "${STRATA_ACCESS_TOKEN:-}" ] || exit 0; printf "username=x-access-token\npassword=%s\n" "$STRATA_ACCESS_TOKEN"; }; f'

  if [ -z "$strata_src" ]; then
    if [ -z "${STRATA_ACCESS_TOKEN:-}" ]; then
      log "WARNING: no strata checkout and STRATA_ACCESS_TOKEN is unset — strata-lang/strata is private, so the clone would fail. Set it in the environment's variables (fine-grained PAT, read access to strata-lang/strata), or point STRATA_SRC_DIR at a checkout."
    fi
    log "cloning strata-lang/strata into $STRATA_CLONE"
    # Blobless rather than shallow: `scripts/sync-strata-docs.sh` stamps each
    # vendored doc with the commit that last touched it, which a --depth 1
    # clone cannot answer. Falls back to a full clone if the git proxy in this
    # environment refuses partial-clone filters.
    # `-c credential.helper=` FIRST, and it is load-bearing. credential.helper
    # is multi-valued and git tries helpers in listed order; the Dolt block
    # above installs an unguarded GLOBAL helper that answers every request for
    # every host with $AGENT_PLUGINS_ACCESS_TOKEN. That PAT has no access to
    # strata-lang/strata, so without this reset the clone authenticates as the
    # wrong identity and GitHub answers 403 "Write access to repository not
    # granted" — a write-shaped message for a read that was never permitted.
    # The empty value resets the list so only the strata helper answers.
    if git clone --filter=blob:none --quiet \
         -c credential.helper= -c "credential.helper=$STRATA_CRED_HELPER" \
         https://github.com/strata-lang/strata.git \
         "$STRATA_CLONE" >>"$STRATA_CACHE/build.log" 2>&1 \
       || git clone --quiet \
            -c credential.helper= -c "credential.helper=$STRATA_CRED_HELPER" \
            https://github.com/strata-lang/strata.git \
            "$STRATA_CLONE" >>"$STRATA_CACHE/build.log" 2>&1; then
      strata_src="$STRATA_CLONE"
    else
      BB_STRATA_FATAL="the clone failed (see $STRATA_CACHE/build.log)."
      log "ERROR: could not clone strata (see $STRATA_CACHE/build.log) — CLI unavailable"
    fi
  fi

  if [ -n "$strata_src" ]; then
    # Reassert the credential helper on the clone this hook owns, so a clone
    # made by an earlier version of this script — or one whose helper snippet
    # has since changed — can still fetch. Only ever on $STRATA_CLONE: an
    # attached checkout belongs to the user and has its own credentials, and
    # a hook that edits someone's .git/config is the same class of surprise
    # as a hook that rebases their branch.
    if [ "$strata_src" = "$STRATA_CLONE" ] \
       && git -C "$strata_src" rev-parse --git-dir >/dev/null 2>&1; then
      # Two values, in this order: the empty one resets the list inherited
      # from global config, then the strata helper answers. A lone local
      # helper would still lose, because global is tried first.
      git -C "$strata_src" config --replace-all credential.helper "" 2>/dev/null || true
      git -C "$strata_src" config --add credential.helper "$STRATA_CRED_HELPER" 2>/dev/null || true
    fi

    # Fast-forward only, and only when there is nothing to lose. A session may
    # be mid-change in an attached strata checkout — this hook runs in that
    # repo too — and a hook that resets or rebases someone's branch is data
    # loss, not convenience. Anything other than a clean default branch gets
    # built exactly as it sits.
    #
    # _fetch_unverified tracks whether we actually confirmed the checkout is
    # caught up with upstream this session. A "stamp matches local HEAD"
    # match below only proves the binary matches what's on disk — it says
    # nothing about whether that disk state is current if the fetch that was
    # supposed to update it failed. Left unguarded, a persistent fetch
    # failure (e.g. a flaky egress proxy) plus a stamp that happens to match
    # the unmoved local HEAD reports "strata CLI current" indefinitely while
    # actually installing whatever was on disk when the fetch first started
    # failing (see agent-plugins-j7f). Only the explicit fetch/merge failure
    # below sets this — the "not a git checkout" and "as-is" (non-default
    # branch / dirty tree) cases are deliberate, not failures, and are
    # already self-explanatory in their own log lines.
    _fetch_unverified=0
    if ! git -C "$strata_src" rev-parse --git-dir >/dev/null 2>&1; then
      # STRATA_SRC_DIR pointed at a source tree with no git in it. Nothing to
      # update and nothing to stamp, so it rebuilds every session.
      log "strata: $strata_src is not a git checkout — building it as-is"
    else
      # `|| true` is load-bearing: this script runs under `set -o pipefail`,
      # and a clone with no origin/HEAD ref (git does not always create one)
      # makes symbolic-ref fail, which would take the whole hook down
      # mid-pipeline.
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
          || { _fetch_unverified=1; log "WARNING: strata: fetch/fast-forward against upstream failed — building the checkout as-is; freshness relative to upstream is unverified this session (see $STRATA_CACHE/build.log)"; }
      fi
    fi

    # An empty _head means the source is not a git checkout — STRATA_SRC_DIR
    # pointing at an extracted tarball or a vendored tree does that. Such a
    # tree gets rebuilt every session: there is no commit to compare, and
    # stamping a placeholder would match itself forever and pin the binary to
    # whatever happened to be built first.
    _head="$(git -C "$strata_src" rev-parse HEAD 2>/dev/null || true)"
    if [ -n "$_head" ] && [ -x "$STRATA_BIN" ] \
       && [ "$(cat "$STRATA_STAMP" 2>/dev/null)" = "$_head" ]; then
      if [ "$_fetch_unverified" = "1" ]; then
        # Do NOT report plain success here: the stamp matching local HEAD
        # only proves the binary matches disk, not that disk is caught up —
        # this is exactly the "stale stamp reports success" failure mode
        # this branch exists to close.
        log "WARNING: strata CLI stamp matches local HEAD (${_head:0:8}), but this session could not verify that HEAD against upstream (fetch failed above) — the installed binary may be stale. Re-run the hook once network access to github.com is restored, or inspect $STRATA_CACHE/build.log."
      else
        log "strata CLI current: $("$STRATA_BIN" --version 2>/dev/null || echo installed) (${_head:0:8})"
      fi
    else
      log "building strata CLI from $strata_src${_head:+ @ ${_head:0:8}} (~2 min cold${_head:+, cached after})"
      # -p strata-cli, not the workspace: this needs the binary, not the WASM
      # crate, and skipping it saves both time and disk.
      if ! command -v cargo >/dev/null 2>&1; then
        log "WARNING: cargo not on PATH — cannot build the strata CLI"
      elif cargo build --release -p strata-cli \
           --manifest-path "$strata_src/Cargo.toml" >>"$STRATA_CACHE/build.log" 2>&1; then
        # Install a copy rather than a symlink into target/: `cargo clean` in
        # the strata checkout is a routine thing to do, and it would leave a
        # dangling /usr/local/bin/strata that reads as a broken install.
        if install -m 0755 "$strata_src/target/release/strata" "$STRATA_BIN" 2>>"$STRATA_CACHE/build.log"; then
          # Only stamp a real commit — see the note above. A stale stamp is
          # worse than none: it silently pins the binary.
          if [ -n "$_head" ]; then printf '%s' "$_head" > "$STRATA_STAMP"; else rm -f "$STRATA_STAMP"; fi
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
  log "  Otherwise the environment must provide ONE of:"
  log "    1. STRATA_ACCESS_TOKEN — a fine-grained PAT with read access to"
  log "       strata-lang/strata, set at container creation. See"
  log "       docs/session-bootstrap.md."
  log "    2. STRATA_SRC_DIR — a path to a checkout already on disk."
  log "    3. strata-lang/strata attached to the session as a source."
  log ""
  exit 1
fi

if [ -n "${BB_DOLT_FATAL:-}" ]; then
  log ""
  log "ERROR: no write access to the beads tracker — $BB_DOLT_FATAL"
  log ""
  log "  This repository's tracker is the Dolt remote (refs/dolt/data on"
  log "  origin). There is no JSONL fallback by design. Do not create or close"
  log "  beads in this session: the work would be lost when the container is"
  log "  reclaimed."
  log ""
  log "  The environment must provide BOTH:"
  log "    1. AGENT_PLUGINS_ACCESS_TOKEN — a GitHub PAT with push access to"
  log "       this repository, set at container creation (claude.ai/code ->"
  log "       the environment's variables)."
  log "    2. Direct, unproxied egress to github.com. The egress proxy"
  log "       substitutes a brokered token scoped to refs/heads, which 403s on"
  log "       refs/dolt/*."
  log ""
  log "  See docs/session-bootstrap.md."
  log ""
  exit 1
fi

log "ready."
