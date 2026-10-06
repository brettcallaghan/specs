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
#   3. Install the `strata` CLI from its latest release.
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

  # --- 2a. The DoltHub credential ---------------------------------------------
  #
  # The tracker lives on DoltHub (sync.remote in .beads/config.yaml). DoltHub
  # authenticates with a Dolt credential — an ed25519 key in JWK form under
  # ~/.dolt/creds/, selected by `user.creds` in ~/.dolt/config_global.json —
  # which the embedded Dolt inside bd reads. The environment supplies it as
  # DOLT_CREDS_JWK, the contents of the .jwk file `dolt creds new` writes.
  #
  # Installed here without the dolt CLI, which is not on the image. The file
  # must be named by the key id, which Dolt derives as base32 (alphabet
  # 0-9a-v, unpadded) of SHA-512/224 over the key's `d` member — Dolt stores
  # the PUBLIC key in `d` (libraries/doltcore/creds: PubKeyToKID,
  # JWKCredsDeserialize). A wrong name is silent until the first push.
  #
  # There is deliberately NO JSONL fallback: a session that cannot reach the
  # tracker is broken and says so.
  BB_DOLT_FATAL=""
  if [ -z "${DOLT_CREDS_JWK:-}" ]; then
    BB_DOLT_FATAL="DOLT_CREDS_JWK is not set in this container."
  elif ! DOLT_CREDS_JWK="$DOLT_CREDS_JWK" python3 -I - <<'PY' 2>/dev/null
import base64, hashlib, json, os
jwk = json.loads(os.environ["DOLT_CREDS_JWK"])
pub = base64.urlsafe_b64decode(jwk["d"] + "=" * (-len(jwk["d"]) % 4))
kid = base64.b32hexencode(hashlib.new("sha512_224", pub).digest()).decode().lower().rstrip("=")
root = os.path.join(os.path.expanduser("~"), ".dolt")
os.makedirs(os.path.join(root, "creds"), mode=0o700, exist_ok=True)
path = os.path.join(root, "creds", kid + ".jwk")
fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
with os.fdopen(fd, "w") as f:
    json.dump(jwk, f)
cfg_path = os.path.join(root, "config_global.json")
try:
    cfg = json.load(open(cfg_path))
except (OSError, ValueError):
    cfg = {}
cfg["user.creds"] = kid
cfg["metrics.disabled"] = "true"
json.dump(cfg, open(cfg_path, "w"))
PY
  then
    BB_DOLT_FATAL="DOLT_CREDS_JWK is set but could not be installed — it must be the JSON content of a .jwk file from \`dolt creds new\`."
  else
    log "DoltHub credential installed"
  fi

  # Never let a push fall back to adopting the git origin as a Dolt remote:
  # the tracker's old home there is retired.
  export BD_NO_REMOTE_ADOPT=1
  [ -n "${CLAUDE_ENV_FILE:-}" ] && echo 'export BD_NO_REMOTE_ADOPT=1' >> "$CLAUDE_ENV_FILE"

  # A container reused from before the move may still carry the GitHub proxy
  # bypass and PAT helper the old refs/dolt/data tracker needed. Left in place
  # they keep routing this repository's pushes around the proxy with a PAT that
  # lacks `workflow` scope, so remove them.
  _origin="$(git remote get-url origin 2>/dev/null || true)"
  _origin="${_origin%.git}"
  case "$_origin" in
    https://*)
      for _k in "http.${_origin}.proxy" "http.${_origin}.git.proxy" \
                "credential.${_origin}.helper" "credential.${_origin}.git.helper"; do
        git config --global --unset-all "$_k" 2>/dev/null || true
      done ;;
  esac

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
      log "tracker is empty ($_n issues) — bootstrapping from DoltHub"
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
        log "beads DB bootstrapped from DoltHub ($_n issues)"
      else
        BB_DOLT_FATAL="bootstrap left the tracker empty ($_n issues); it may have exited 0 anyway. Output: $(printf '%s' "$_bootstrap_out" | tr '\n' ' ' | cut -c1-300)"
      fi
    else
      # A container can be reused across sessions, so a non-empty tracker is
      # not necessarily a CURRENT one. Pull rather than assume; a no-op pull
      # costs ~2s.
      if timeout 300 bd dolt pull >/dev/null 2>&1; then
        _n="$(_count)"; [ -n "$_n" ] || _n=0
        log "tracker pulled from DoltHub ($_n issues)"
      else
        log "WARNING: \`bd dolt pull\` failed — the tracker may be behind DoltHub"
      fi
    fi
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

  # Prefer a checkout the session already has: if swarm is attached as a
  # source, the CLI should match the code in front of the user.
  swarm_src=""
  for candidate in "${SWARM_SRC_DIR:-}" "$PROJECT_DIR/../swarm" "$HOME/swarm" \
                   "$SWARM_CLONE"; do
    [ -n "$candidate" ] || continue
    [ -f "$candidate/cmd/swarm/main.go" ] || continue
    swarm_src="$(cd "$candidate" && pwd)"
    break
  done

  # brettcallaghan/swarm is private. AGENT_PLUGINS_ACCESS_TOKEN, a GitHub PAT
  # the environment provides, reads it
  # (SWARM_ACCESS_TOKEN overrides, for an environment that scopes them apart).
  # Handed to git through a credential helper that reads the variable at call
  # time, never written into the remote URL: a URL carrying the token persists
  # in .git/config and leaks into `git remote -v`.
  #
  # Scoped to this clone via `git clone -c`, never global — a global github.com
  # helper would also answer for this repository's own pushes.
  #
  # No proxy bypass: the clone succeeds through the egress proxy with this
  # token. Do not add one — a github.com bypass strips the proxy's brokered
  # credential from every other github.com remote in the container.
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
    # THIS repository is the public brettcallaghan marketplace (specs); the
    # swarm checkout is the brettcallaghan-private marketplace (swarm), whose
    # swarm plugin depends on specs@brettcallaghan across marketplaces. Both
    # install from working trees — no clone, no token. specs first, since
    # swarm declares it as a dependency.
    if ! command -v claude >/dev/null 2>&1; then
      log "WARNING: claude not on PATH — skipping the swarm plugin install"
    else
      _plugin_ok=1
      for _m in "$PROJECT_DIR" "$swarm_src"; do
        claude plugin marketplace add "$_m" --scope user >>"$SWARM_LOG" 2>&1 \
          || { _plugin_ok=0; log "WARNING: could not add the marketplace at $_m (see $SWARM_LOG)"; }
      done
      for _p in specs@brettcallaghan swarm@brettcallaghan-private; do
        claude plugin install "$_p" --scope user --yes >>"$SWARM_LOG" 2>&1 \
          || { _plugin_ok=0; log "WARNING: could not install plugin $_p (see $SWARM_LOG)"; }
      done
      [ "$_plugin_ok" = "1" ] && log "specs@brettcallaghan and swarm@brettcallaghan-private installed at user scope"
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
  log "  This repository's tracker is the Dolt database on DoltHub"
  log "  (bcallaghan/beads). There is no JSONL fallback by design. Do not"
  log "  create or close beads in this session: the work would be lost when"
  log "  the container is reclaimed."
  log ""
  log "  The environment must provide BOTH:"
  log "    1. DOLT_CREDS_JWK — the contents of a DoltHub credential's .jwk"
  log "       file, set at container creation (claude.ai/code -> the"
  log "       environment's variables)."
  log "    2. Network access to doltremoteapi.dolthub.com and DoltHub's chunk"
  log "       storage hosts (see docs/session-bootstrap.md for the list)."
  log ""
  log "  See docs/session-bootstrap.md."
  log ""
  exit 1
fi

log "ready."
