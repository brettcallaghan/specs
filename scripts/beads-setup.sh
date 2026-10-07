#!/usr/bin/env bash
#
# beads-setup.sh — point this clone's beads tracker at the repository's data
# branch and sync it. Safe to run repeatedly; a no-op once the tracker is
# current.
#
#   ./scripts/beads-setup.sh
#
# Run it after a fresh clone, and any time `bd` reports an empty tracker. The
# SessionStart hook (.claude/hooks/session-start.sh) calls it, so a Claude Code
# cloud session gets it automatically — but nothing here is cloud-specific: no
# root, no package manager, no token. It needs `bd`, `git`, `python3` and a
# checkout whose `origin` is this repository.
#
# WHY A BRANCH. The tracker's data lives on the ordinary branch `beads/data`
# rather than in the `refs/dolt/*` namespace Dolt defaults to. A Claude Code
# cloud session's GitHub credential is refused on `refs/dolt/*` and accepted on
# `refs/heads/*`, and for everyone else a branch is written with the same
# access that writes code — so repository write access IS tracker write access,
# with no second account or credential. A fresh clone has to be told where to
# look, or `bd` reports an empty tracker rather than an error.
# `.beads/embeddeddolt/` is gitignored, so this is per-clone setup that no
# committed file can do for you.
#
# Dolt's own bookkeeping branch keeps its default name, `__dolt_remote_info__`.
# Renaming it would mean exporting DOLT_REMOTE_INFO_BRANCH in every shell that
# runs `bd`, since Dolt reads it from the environment on each call — a per-shell
# step traded for a tidier branch name. Not worth it.
#
# It cannot be a committed config value: `bd`'s storage interface is
# `AddRemote(ctx, name, url)` — no parameters — and `.beads/config.yaml` has no
# key for a ref, so nothing `bd` reads can carry it (`bd dolt remote add` takes
# only a name and a URL as of bd 1.3.1). Dolt's own CLI exposes it as
# `dolt remote add --ref`; `bd` does not pass it through. Until it does, the
# value is version-controlled here and written into Dolt's repo_state.json.

set -uo pipefail

DATA_REF="refs/heads/beads/data"
# Dolt's built-in default ref. Used ONLY as a local staging name for the
# clone in hydrate_from_data_branch; nothing is ever pushed to it.
DOLT_DEFAULT_REF="refs/dolt/data"
ORIGIN_URL="$(git remote get-url origin 2>/dev/null || true)"

cd "$(git rev-parse --show-toplevel 2>/dev/null)" || {
  echo "beads-setup: not inside a git repository" >&2; exit 1; }

command -v bd >/dev/null 2>&1 || {
  echo "beads-setup: \`bd\` is not on PATH. Install it with:" >&2
  echo "  curl -sSL https://raw.githubusercontent.com/steveyegge/beads/main/scripts/install.sh | bash" >&2
  exit 1; }
command -v python3 >/dev/null 2>&1 || {
  echo "beads-setup: python3 is required (used to edit Dolt's repo_state.json)" >&2
  exit 1; }
[[ -d .beads ]] || { echo "beads-setup: no .beads/ directory here" >&2; exit 1; }

count() {
  # `|| true` INSIDE the substitution: `bd stats` exits non-zero when there is
  # no database, which is exactly the fresh-clone case this handles.
  { bd stats 2>/dev/null || true; } \
    | sed -n 's/^[[:space:]]*Total Issues:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | head -1
}

# HYDRATION MUST NOT GO THROUGH `bd dolt pull`, AND MUST NOT USE THE DEFAULT REF.
#
# Two `bd` facts drive the shape of this. Getting either wrong yields a tracker
# that looks populated and is not the one this repository uses:
#
#   1. `bd init` CLONES from `sync.remote`, and a clone always reads Dolt's
#      DEFAULT ref, `refs/dolt/data` — never $DATA_REF, because repo_state.json
#      (the only place the ref can live) does not exist until init has finished.
#      Initialising first and setting the ref afterwards hydrates from the wrong
#      ref, and the resulting count is non-zero, so every count-based check
#      passes while the tracker serves the wrong data.
#   2. `bd dolt pull` MERGES. It cannot reconcile two unrelated Dolt histories
#      and fails with "local changes would be stomped by merge: events" — which
#      `bd dolt commit` does not clear, because nothing is actually pending.
#
# So publish $DATA_REF under Dolt's default ref name inside THIS clone's own
# .git and clone from there over file://. The clone reads the tracker branch
# while still going through the only code path that works. set_data_ref then
# points the remote back at the real origin for push/pull.
#
# `--skip-agents --skip-hooks` matter: without them `bd init` rewrites AGENTS.md
# and CLAUDE.md and adds .cursor/. It commits either way (see the reset below).
hydrate_from_data_branch() {
  local gitdir sha head_before dirty_before
  gitdir="$(git rev-parse --absolute-git-dir 2>/dev/null)" || return 1
  head_before="$(git rev-parse HEAD 2>/dev/null || true)"
  # Tracked files already modified before init. Whatever is dirty that is NOT in
  # this set was dirtied by `bd init`, and gets restored below.
  dirty_before="$(git diff --name-only HEAD 2>/dev/null | sort)"
  git fetch -q origin "+$DATA_REF:refs/remotes/origin/beads-data" 2>/dev/null || {
    echo "beads-setup: cannot fetch $DATA_REF from origin" >&2; return 1; }
  sha="$(git rev-parse refs/remotes/origin/beads-data 2>/dev/null)" || return 1
  git update-ref "$DOLT_DEFAULT_REF" "$sha" || return 1
  BD_NON_INTERACTIVE=1 bd init --non-interactive --skip-agents --skip-hooks \
    --remote "git+file://$gitdir" >/dev/null 2>&1 || return 1
  # `bd init` COMMITS to the current branch, sweeping up whatever else is dirty.
  # Undo the commit while keeping every file exactly as it is: --mixed, never
  # --hard, which would discard uncommitted work that was never bd's to touch.
  if [[ -n "$head_before" && "$(git rev-parse HEAD)" != "$head_before" ]]; then
    git reset -q --mixed "$head_before" || return 1
  fi
  # `bd init` also edits .gitignore and .beads/.gitignore in passing. Setting up
  # a tracker must not leave the working tree dirty — a later step that gates on
  # a clean tree would read those edits as the caller's own uncommitted work.
  # Restore only what init touched; anything dirty beforehand is left untouched.
  local dirty_now newly_dirty
  dirty_now="$(git diff --name-only HEAD 2>/dev/null | sort)"
  newly_dirty="$(comm -13 <(printf '%s\n' "$dirty_before") <(printf '%s\n' "$dirty_now"))"
  if [[ -n "$newly_dirty" ]]; then
    printf '%s\n' "$newly_dirty" | while IFS= read -r f; do
      [[ -n "$f" ]] && git checkout -q -- "$f" 2>/dev/null
    done
  fi
  # `bd init --remote` persists the file:// mirror as sync.remote; put the real
  # origin back so later pushes go to GitHub rather than into this clone.
  if [[ -n "$ORIGIN_URL" ]] && grep -q '^sync\.remote:' .beads/config.yaml 2>/dev/null; then
    BEADS_CFG=.beads/config.yaml BEADS_URL="$ORIGIN_URL" python3 - <<'PYEOF'
import io, os, re
p = os.environ["BEADS_CFG"]
url = os.environ["BEADS_URL"].rstrip("/")
if not url.endswith(".git"):
    url += ".git"
s = io.open(p, encoding="utf-8").read()
s = re.sub(r'^sync\.remote:.*$', 'sync.remote: "git+%s"' % url, s, count=1, flags=re.M)
io.open(p, "w", encoding="utf-8").write(s)
PYEOF
  fi
}

# An initialised-but-empty database is the case that needs FILLING, not one to
# pull into. `bd` leaves one behind as a side effect of ordinary commands and a
# reused container carries it forward, so directory existence is exactly the
# wrong signal here. `bd dolt pull` merges, and it cannot reconcile two
# unrelated Dolt histories ("no common ancestor"), so from this state the pull
# path can only fail. Discard it and take the clone path instead.
#
# Safe by construction, but ONLY on a definite zero: a database `bd stats` says
# holds 0 issues holds nothing that could be lost. An UNREADABLE count is not a
# zero — `count` returns empty when `bd stats` fails for any reason, and treating
# that as 0 would rm -rf a populated database over a transient failure, taking
# any locally-created beads not yet pushed with it. So discard only on a literal
# 0 and leave every other case to the reporting below, which warns and preserves.
if [[ -d .beads/embeddeddolt ]]; then
  _n="$(count)"
  if [[ "$_n" == 0 ]]; then
    echo "beads-setup: discarding an initialised-but-empty database and re-cloning" >&2
    rm -rf .beads/embeddeddolt
  fi
  unset _n
fi

# A fresh clone is by definition current, so it skips the pull below entirely —
# which would otherwise fail on the very database it just created.
JUST_CLONED=no
if [[ ! -d .beads/embeddeddolt ]]; then
  hydrate_from_data_branch || {
    echo "beads-setup: could not hydrate the tracker from $DATA_REF" >&2; exit 1; }
  JUST_CLONED=yes
fi

# Written only when absent or wrong. `bd` adds remotes with DOLT_REMOTE('add')
# and ignores the "already exists" error rather than overwriting, so on an
# already-configured clone this finds the value correct and does nothing.
set_data_ref() {
  local rs
  rs="$(ls -d .beads/embeddeddolt/*/.dolt/repo_state.json 2>/dev/null | head -1)"
  [[ -n "$rs" ]] || return 2
  BEADS_RS="$rs" BEADS_REF="$DATA_REF" BEADS_URL="$(git remote get-url origin 2>/dev/null)" \
    python3 - <<'PYEOF'
import json, os, sys
p = os.environ["BEADS_RS"]
url = (os.environ.get("BEADS_URL") or "").rstrip("/")
if not url.endswith(".git"):
    url += ".git"
want = {"name": "origin", "url": "git+" + url,
        "fetch_specs": ["refs/heads/*:refs/remotes/origin/*"],
        "params": {"git_ref": os.environ["BEADS_REF"]}}
d = json.load(open(p))
if d.setdefault("remotes", {}).get("origin") == want:
    sys.exit(3)          # already correct — nothing written
d["remotes"]["origin"] = want
json.dump(d, open(p, "w"), indent=2)
PYEOF
}

if set_data_ref; then
  echo "beads-setup: pointed the Dolt remote at $DATA_REF"
fi

# NEVER discard this exit status. Reporting "synced" on the strength of a
# non-zero issue count is what let a wrong-ref tracker pass for a good one:
# the count is satisfied by whatever the database already held.
pull_rc=0
pull_err=""
if [[ "$JUST_CLONED" == no ]]; then
  pull_err="$(bd dolt pull 2>&1)"; pull_rc=$?
fi

# Verify by ISSUE COUNT, never by directory existence. `bd` leaves an
# initialised-but-empty database behind as a side effect of ordinary commands,
# so "the directory exists" is true precisely in the case that needs filling.
n="$(count)"; [[ -n "$n" ]] || n=0

if [[ "$n" -eq 0 ]]; then
  echo "beads-setup: tracker is EMPTY — work recorded against it would not be shared." >&2
  [[ $pull_rc -ne 0 ]] && echo "             bd dolt pull said: $pull_err" >&2
  echo "             Check that origin carries $DATA_REF:" >&2
  echo "                 git ls-remote origin 'refs/heads/beads/*'" >&2
  exit 1
fi

if [[ $pull_rc -ne 0 ]]; then
  # Data is present, so the tracker is usable — but it is NOT freshly synced,
  # and saying so is the whole point of this branch.
  echo "beads-setup: tracker has $n issues but is NOT freshly synced — bd dolt pull failed:" >&2
  echo "             $pull_err" >&2
  echo "             To refresh: rm -rf .beads/embeddeddolt && ./scripts/beads-setup.sh" >&2
  echo "             (re-clones from $DATA_REF; locally-created beads not yet pushed" >&2
  echo "              with 'bd dolt push' would be lost, so push first)" >&2
  exit 0
fi

echo "beads-setup: tracker synced from origin $DATA_REF ($n issues)"
exit 0
