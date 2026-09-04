#!/bin/bash
# SessionStart hook — provision this repo's dependencies so a session can
# actually run the workflows the skills prescribe.
#
# Four jobs:
#   1. Install `bd` (Beads), which every grooming and orchestration skill in
#      this repo assumes is on PATH. Without it those skills cannot be
#      exercised at all, which is the definition of not dogfooding.
#   2. Keep the vendored shared references current. The skills are committed
#      symlinks under .claude/skills/ (each -> plugin/skills/<name>), so the
#      session runs the working tree directly — no materialization step, and
#      no global copy can shadow it.
#   3. Build the `strata` CLI.
#   4. Install the `swarm` plugin at user scope and build the `swarm` CLI. Its
#      skills are the sprint orchestration this repo grooms beads for, and each
#      of them shells out to that CLI.
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
  # bd stores data in Dolt, which phones home to eventsapi.dolthub.com. Killed
  # here rather than making every session re-do it.
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
  #      time. The token is never written to disk. It is registered under a
  #      URL-scoped key, never the bare `credential.helper` — see below.
  #   2. Proxy bypass, scoped to THIS repository's origin. The egress proxy
  #      REPLACES the Authorization header with a brokered GitHub App
  #      installation token, which is scoped to refs/heads and returns 403 on
  #      refs/dolt/*. Proof it is the proxy and not GitHub: a push with a
  #      deliberately INVALID token succeeded, which is only possible if the
  #      proxy substituted its own credential. Direct egress to github.com is
  #      permitted, so bypassing the proxy lets the real PAT through.
  #
  #      SCOPE IT THE SAME WAY AS THE HELPER — never `http.https://github.com/`.
  #      A host-wide bypass reaches every OTHER github.com remote in the
  #      container while the helper beside it does not, so those remotes lose
  #      the proxy's brokered credential and gain nothing in its place. In a
  #      session with a second repository attached this breaks both directions:
  #      a push to the other repo dies with "could not read Username for
  #      'https://github.com'", and a push to THIS one starts using the PAT,
  #      which lacks `workflow` scope and is refused on any commit touching
  #      .github/workflows/. Both were measured in a session holding
  #      agent-plugins and swarm.
  #
  # The token name is this repository's environment contract, not any skill's:
  # a skill needing a durable tracker requires only "write access to
  # refs/dolt/data" and never names a credential. See docs/session-bootstrap.md.
  BB_DOLT_FATAL=""
  BB_STRATA_FATAL="${BB_STRATA_FATAL:-}"
  # The helper body is guarded exactly like STRATA_CRED_HELPER further down.
  # `[ "$1" = get ]` because git invokes a helper for get, store and erase, and
  # only get wants an answer; the token test because an unset token would
  # otherwise emit `password=` and authenticate as the empty string instead of
  # declining. One printf rather than two echos: echo with a backslash in the
  # value is implementation-defined in sh.
  DOLT_CRED_HELPER='!f() { [ "$1" = get ] || exit 0; [ -n "${AGENT_PLUGINS_ACCESS_TOKEN:-}" ] || exit 0; printf "username=x-access-token\npassword=%s\n" "$AGENT_PLUGINS_ACCESS_TOKEN"; }; f'

  if [ -z "${AGENT_PLUGINS_ACCESS_TOKEN:-}" ]; then
    BB_DOLT_FATAL="AGENT_PLUGINS_ACCESS_TOKEN is not set in this container."
  else
    # Scoped to origin's URL, NEVER the bare `credential.helper` key.
    # credential.helper is multi-valued and git tries global values before
    # local ones, so a bare global helper wins over any per-clone helper and
    # hands this repository's PAT to a clone of an unrelated private repo —
    # which GitHub rejects with a 403 "Write access to repository not granted",
    # a write-shaped message for a read that was never permitted.
    #
    # Registered under both the suffixless and the `.git` form, because git
    # matches the config key's path against the request's path EXACTLY and only
    # the matching form answers — measured against a clone whose origin carries
    # `.git`: with only the suffixless key registered the push died with "could
    # not read Username for 'https://github.com'", and with only the `.git` key
    # it succeeded. Whether origin carries the suffix varies by how the clone
    # was made, so register both rather than guess. The bare key is unset, not
    # merely shadowed — leaving it in place would keep answering for every
    # host, which is the defect.
    _origin="$(git remote get-url origin 2>/dev/null || true)"
    _origin="${_origin%.git}"
    case "$_origin" in
      https://*) _origin_scoped=1 ;;
      *)         _origin_scoped=0 ;;
    esac

    if [ "$_origin_scoped" = 0 ]; then
      BB_DOLT_FATAL="origin is '${_origin:-<none>}', not an https:// URL — an https credential helper cannot be scoped to it, and a bare one would capture every other host."
    else
      git config --global --unset-all credential.helper 2>/dev/null || true
      git config --global "credential.${_origin}.helper" "$DOLT_CRED_HELPER"
      git config --global "credential.${_origin}.git.helper" "$DOLT_CRED_HELPER"

      # Both URL forms, for the reason the helper registers both: git matches
      # an http.<url>.* key's path against the request's path, and only the
      # matching form answers.
      #
      # The host-wide key is UNSET rather than merely shadowed. An earlier
      # revision of this hook set it, so a container whose git config it
      # already touched still carries it — and a bypass left in place keeps
      # answering for every github.com remote, which is the defect.
      git config --global --unset-all 'http.https://github.com/.proxy' 2>/dev/null || true
      git config --global "http.${_origin}.proxy" ''
      git config --global "http.${_origin}.git.proxy" ''

      # Verify rather than assume — a mis-scoped helper breaks nothing in the
      # session that writes it and everything in the next one. Pushing HEAD
      # (already on the remote in any normal clone) to a scratch ref under
      # refs/dolt/ transfers no objects and exercises the exact permission that
      # matters, then removes it.
      _probe="refs/dolt/write-probe-$$"
      if git push -q origin "HEAD:$_probe" >/dev/null 2>&1; then
        git push -q origin --delete "$_probe" >/dev/null 2>&1 || \
          log "note: left a stray probe ref $_probe on origin — delete it by hand"
        log "verified write access to refs/dolt/* (probe pushed and removed)"
      else
        BB_DOLT_FATAL="a probe push to refs/dolt/* was rejected."
      fi
    fi
  fi

  # --- 2b. The database name --------------------------------------------------
  #
  # Must run BEFORE anything reads the tracker, because reading it is what
  # destroys the evidence. If `dolt_database` names a database not on disk, bd
  # CREATES an empty one under that name and reports a healthy, empty tracker:
  # measured on bd 1.2.2 against a database holding 318 issues, `bd stats`
  # printed "Total Issues: 0" and `bd ready` printed "No open issues", both
  # exit 0, no error. Ordered after the count below, the sequence measured was
  # worse than the bug: the count read 0, the hook moved the REAL database
  # aside as stale, re-cloned under the wrong name, and the guard then found
  # the directory present and stayed silent.
  #
  # Fatal rather than a warning — an empty tracker that reports itself healthy
  # is the thing this hook exists to prevent. Skipped when there is no database
  # yet, which is the ordinary fresh-clone case handled below.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads/embeddeddolt ]; then
    _want="$(python3 -c 'import json,io;print(json.load(io.open(".beads/metadata.json",encoding="utf-8")).get("dolt_database",""))' 2>/dev/null)" || _want=""
    if [ -n "$_want" ] && [ ! -d ".beads/embeddeddolt/$_want" ]; then
      BB_DOLT_FATAL="metadata.json names database '$_want' but it is absent from .beads/embeddeddolt/ (present: $(ls .beads/embeddeddolt/ 2>/dev/null | tr '\n' ' ')). bd would report an empty tracker rather than an error."
    fi
  fi

  # --- 2c. The database -------------------------------------------------------
  # The database is gitignored, so a fresh clone has none and every bd command
  # fails with "no beads database found".
  #
  # Verify by ISSUE COUNT, never by directory existence, and never by exit code.
  # A stale but genuinely initialised `.beads/embeddeddolt/` can survive into a
  # new container, and bootstrap cannot clone over one: measured on bd 1.2.2, it
  # exits 1 with "can't create database beads; database exists". A guard that
  # skips bootstrap because that directory exists therefore skips it exactly
  # when it is needed, which is why the stale database is moved aside below
  # rather than treated as proof of a tracker.
  #
  # The count is the outcome test. bd 1.2.2 fails loudly on the ways bootstrap
  # can go wrong — including a workspace-identity check comparing metadata.json's
  # project_id against the database's — so this is no longer catching a silent
  # empty tracker at this step. It stays because it tests the thing that
  # matters (issues are present) rather than a proxy for it, and because the
  # database-name mismatch further down still is silent.
  if [ -z "$BB_DOLT_FATAL" ] && [ -d .beads ]; then
    # `|| true` inside, not outside: `bd stats` EXITS NON-ZERO when there is no
    # database, which under `set -euo pipefail` makes `_n="$(_count)"` kill the
    # script — silently, with no message and exit 1, before bootstrap can run.
    # That is precisely the fresh-clone case this block exists to handle, so
    # without this the hook can only ever succeed where it was not needed.
    _count() { { bd stats 2>/dev/null || true; } | sed -n 's/^[[:space:]]*Total Issues:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -1; }
    _n="$(_count)"; [ -n "$_n" ] || _n=0

    if [ "$_n" -eq 0 ]; then
      log "tracker is empty ($_n issues) — bootstrapping from origin refs/dolt/data"
      # A stale empty database short-circuits bootstrap's own probe. Move it
      # aside rather than deleting it: if the remote turns out to be the empty
      # one, the copy is the only remaining evidence.
      if [ -d .beads/embeddeddolt ]; then
        _stale=".beads/embeddeddolt.stale.$$"
        mv .beads/embeddeddolt "$_stale" \
          && log "moved a stale empty database aside: $_stale"
      fi
      _bootstrap_out="$(BD_NON_INTERACTIVE=1 bd bootstrap --yes 2>&1)"
      _n="$(_count)"; [ -n "$_n" ] || _n=0
      if [ "$_n" -gt 0 ]; then
        log "beads DB bootstrapped from origin refs/dolt/data ($_n issues)"
      else
        BB_DOLT_FATAL="bootstrap left the tracker empty ($_n issues); it may have exited 0 anyway. Output: $(printf '%s' "$_bootstrap_out" | tr '\n' ' ' | cut -c1-300)"
      fi
    else
      # A container can be reused across sessions, so a non-empty tracker is
      # not necessarily a CURRENT one. Pull rather than assume; a no-op pull
      # costs ~2s.
      if timeout 300 bd dolt pull >/dev/null 2>&1; then
        _n="$(_count)"; [ -n "$_n" ] || _n=0
        log "tracker pulled from origin refs/dolt/data ($_n issues)"
      else
        log "WARNING: \`bd dolt pull\` failed — the tracker may be behind origin"
      fi
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
    # `-c credential.helper=` FIRST — defence in depth, and cheap. The Dolt
    # block's helper is scoped to this repository's own remote and so does not
    # answer for strata-lang/strata, but credential.helper is multi-valued and
    # global values are tried before local ones: any bare global helper this
    # environment (or a stale ~/.gitconfig) still carries would answer with the
    # wrong identity, and GitHub reports that as 403 "Write access to
    # repository not granted" — a write-shaped message for a read that was
    # never permitted. The empty value resets the list so only the strata
    # helper answers.
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

# --- 5. swarm plugin and CLI -------------------------------------------------
# The `swarm` plugin (skills `swarm`, `sprint-*`, `swarm-dispatch`) orchestrates
# a sprint over this repository's bead backlog, and every one of those skills
# shells out to the `swarm` CLI. Neither half is useful alone, so both are
# provisioned here: the plugin at user scope, the CLI on PATH.
#
# Non-fatal throughout. A session with no swarm can still do everything the
# rest of this hook provisions.
if [ "${SWARM_SKIP:-}" != "1" ]; then
  SWARM_CACHE="$HOME/.cache/swarm-cli"
  SWARM_CLONE="$HOME/.cache/swarm-src"
  SWARM_BIN=/usr/local/bin/swarm
  SWARM_STAMP="$SWARM_CACHE/built-commit"
  SWARM_LOG="$SWARM_CACHE/build.log"
  mkdir -p "$SWARM_CACHE"

  # Prefer a checkout the session already has, for the same reason strata does:
  # if swarm is attached as a source, the CLI should match the code in front of
  # the user.
  swarm_src=""
  for candidate in "${SWARM_SRC_DIR:-}" "$PROJECT_DIR/../swarm" "$HOME/swarm" \
                   "$SWARM_CLONE"; do
    [ -n "$candidate" ] || continue
    [ -f "$candidate/cmd/swarm/main.go" ] || continue
    swarm_src="$(cd "$candidate" && pwd)"
    break
  done

  # brettcallaghan/swarm is private. The token that already exists in this
  # environment for the Dolt remote — AGENT_PLUGINS_ACCESS_TOKEN — reads it
  # (SWARM_ACCESS_TOKEN overrides, for an environment that scopes them apart).
  # Handed to git through a credential helper that reads the variable at call
  # time, never written into the remote URL: a URL carrying the token persists
  # in .git/config and leaks into `git remote -v`.
  #
  # Scoped to this clone via `git clone -c`, never global — a global github.com
  # helper would also answer for this repository's own pushes.
  #
  # No proxy bypass: the clone succeeds through the egress proxy with this
  # token. Do not add one — the bypass this repo configures is scoped to
  # origin's URL on purpose (see AGENTS.md).
  SWARM_CRED_HELPER='!f() { [ "$1" = get ] || exit 0; t="${SWARM_ACCESS_TOKEN:-${AGENT_PLUGINS_ACCESS_TOKEN:-}}"; [ -n "$t" ] || exit 0; printf "username=x-access-token\npassword=%s\n" "$t"; }; f'

  if [ -z "$swarm_src" ]; then
    if [ -z "${SWARM_ACCESS_TOKEN:-${AGENT_PLUGINS_ACCESS_TOKEN:-}}" ]; then
      log "WARNING: no swarm checkout and neither SWARM_ACCESS_TOKEN nor AGENT_PLUGINS_ACCESS_TOKEN is set — brettcallaghan/swarm is private, so the clone would fail. Point SWARM_SRC_DIR at a checkout, or set SWARM_SKIP=1."
    fi
    log "cloning brettcallaghan/swarm into $SWARM_CLONE"
    # `-c credential.helper=` first resets the multi-valued list, so a global
    # helper carrying the wrong identity cannot answer ahead of this one —
    # GitHub reports that case as a 403 about write access, for a read.
    if git clone --filter=blob:none --quiet \
         -c credential.helper= -c "credential.helper=$SWARM_CRED_HELPER" \
         https://github.com/brettcallaghan/swarm.git \
         "$SWARM_CLONE" >>"$SWARM_LOG" 2>&1 \
       || git clone --quiet \
            -c credential.helper= -c "credential.helper=$SWARM_CRED_HELPER" \
            https://github.com/brettcallaghan/swarm.git \
            "$SWARM_CLONE" >>"$SWARM_LOG" 2>&1; then
      swarm_src="$SWARM_CLONE"
    else
      log "WARNING: could not clone swarm (see $SWARM_LOG) — plugin and CLI unavailable"
    fi
  fi

  if [ -n "$swarm_src" ]; then
    # Reassert the helper on the clone this hook owns, so a clone made by an
    # earlier version of this script can still fetch. Never on an attached
    # checkout: that one belongs to the user.
    if [ "$swarm_src" = "$SWARM_CLONE" ] \
       && git -C "$swarm_src" rev-parse --git-dir >/dev/null 2>&1; then
      git -C "$swarm_src" config --replace-all credential.helper "" 2>/dev/null || true
      git -C "$swarm_src" config --add credential.helper "$SWARM_CRED_HELPER" 2>/dev/null || true
    fi

    # Fast-forward only, and only a clean default branch. A session may be
    # mid-change in an attached swarm checkout; a hook that resets someone's
    # branch is data loss, not convenience.
    _swarm_fetch_unverified=0
    if ! git -C "$swarm_src" rev-parse --git-dir >/dev/null 2>&1; then
      log "swarm: $swarm_src is not a git checkout — building it as-is"
    else
      _sdefault="$(git -C "$swarm_src" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed 's|^origin/||' || true)"
      _sdefault="${_sdefault:-main}"
      _sbranch="$(git -C "$swarm_src" rev-parse --abbrev-ref HEAD 2>/dev/null || echo '')"
      if [ "$_sbranch" != "$_sdefault" ]; then
        log "swarm: checkout is on '$_sbranch', not '$_sdefault' — building it as-is"
      elif [ -n "$(git -C "$swarm_src" status --porcelain 2>/dev/null)" ]; then
        log "swarm: checkout has uncommitted changes — building it as-is"
      else
        git -C "$swarm_src" fetch --quiet origin "$_sdefault" >>"$SWARM_LOG" 2>&1 \
          && git -C "$swarm_src" merge --ff-only --quiet "origin/$_sdefault" >>"$SWARM_LOG" 2>&1 \
          || { _swarm_fetch_unverified=1; log "WARNING: swarm: fetch/fast-forward against upstream failed — using the checkout as-is; freshness relative to upstream is unverified this session (see $SWARM_LOG)"; }
      fi
    fi

    # --- the plugin ---------------------------------------------------------
    # Both installs are idempotent: `marketplace add` on an already-declared
    # marketplace and `install` of an already-installed plugin both report the
    # existing state and exit 0. `--yes` is required because stdout here is not
    # a TTY.
    #
    # sf first. swarm declares sf@spec-driven-dev as a plugin dependency, and
    # without it `claude plugin install swarm@swarm` reports "1 dependency
    # still unresolved". THIS repository is the spec-driven-dev marketplace, so
    # sf installs from the working tree — no clone, no token, and the same
    # files the committed .claude/skills/ symlinks already point at.
    if ! command -v claude >/dev/null 2>&1; then
      log "WARNING: claude not on PATH — skipping the swarm plugin install"
    else
      _plugin_ok=1
      for _m in "$PROJECT_DIR" "$swarm_src"; do
        claude plugin marketplace add "$_m" --scope user >>"$SWARM_LOG" 2>&1 \
          || { _plugin_ok=0; log "WARNING: could not add the marketplace at $_m (see $SWARM_LOG)"; }
      done
      for _p in sf@spec-driven-dev swarm@swarm; do
        claude plugin install "$_p" --scope user --yes >>"$SWARM_LOG" 2>&1 \
          || { _plugin_ok=0; log "WARNING: could not install plugin $_p (see $SWARM_LOG)"; }
      done
      [ "$_plugin_ok" = "1" ] && log "swarm plugin installed at user scope (with sf@spec-driven-dev, its declared dependency)"
    fi

    # --- the CLI ------------------------------------------------------------
    # An empty _shead means the source is not a git checkout, which rebuilds
    # every session: there is no commit to compare, and a placeholder stamp
    # would match itself forever and pin the binary.
    _shead="$(git -C "$swarm_src" rev-parse HEAD 2>/dev/null || true)"
    if ! command -v go >/dev/null 2>&1; then
      log "WARNING: go not on PATH — cannot build the swarm CLI"
    elif [ -n "$_shead" ] && [ -x "$SWARM_BIN" ] \
         && [ "$(cat "$SWARM_STAMP" 2>/dev/null)" = "$_shead" ]; then
      if [ "$_swarm_fetch_unverified" = "1" ]; then
        # The stamp matching local HEAD proves the binary matches disk, not
        # that disk is caught up with upstream.
        log "WARNING: swarm CLI stamp matches local HEAD (${_shead:0:8}), but this session could not verify that HEAD against upstream (fetch failed above) — the installed binary may be stale. Re-run the hook once github.com is reachable, or inspect $SWARM_LOG."
      else
        log "swarm CLI current: ${_shead:0:8}"
      fi
    else
      log "building swarm CLI from $swarm_src${_shead:+ @ ${_shead:0:8}}"
      _swarm_tmp="$SWARM_CACHE/swarm.new"
      if (cd "$swarm_src" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
            -o "$_swarm_tmp" ./cmd/swarm) >>"$SWARM_LOG" 2>&1 \
         && install -m 0755 "$_swarm_tmp" "$SWARM_BIN" 2>>"$SWARM_LOG"; then
        rm -f "$_swarm_tmp"
        # Prove the binary runs before stamping. `--help` rather than a bare
        # invocation: `swarm` with no arguments exits 3 by design.
        if "$SWARM_BIN" --help >/dev/null 2>&1; then
          if [ -n "$_shead" ]; then printf '%s' "$_shead" > "$SWARM_STAMP"; else rm -f "$SWARM_STAMP"; fi
          log "swarm CLI ready: $SWARM_BIN${_shead:+ (${_shead:0:8})}"
        else
          rm -f "$SWARM_STAMP"
          log "WARNING: $SWARM_BIN was installed but does not run — see $SWARM_LOG"
        fi
      else
        rm -f "$_swarm_tmp"
        log "WARNING: swarm CLI build failed — see $SWARM_LOG"
      fi
    fi
  fi
fi

# --- 6. Fail loudly if the tracker is unreachable ----------------------------
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
  log "ERROR: the beads tracker is not usable in this session — $BB_DOLT_FATAL"
  log ""
  log "  This repository's tracker is the Dolt remote (refs/dolt/data on"
  log "  origin). There is no JSONL fallback by design. Do not create or close"
  log "  beads in this session: the work would be lost when the container is"
  log "  reclaimed."
  log ""
  log "  If the cause above is a credentials or push failure, the environment"
  log "  must provide BOTH:"
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
