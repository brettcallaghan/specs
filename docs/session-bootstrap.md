# Session bootstrap

How this repository's dependencies get installed in a Claude Code cloud
session, why the SessionStart hook alone is not enough, and what to
paste where.

This is an operations note, not a specification — it carries no
frontmatter and is not a corpus member.

## The problem

`.claude/settings.json` registers `.claude/hooks/session-start.sh` as a
SessionStart hook. That hook installs `bd`, builds the `strata` CLI, and
regenerates the vendored shared references. (The skills are committed
symlinks under `.claude/skills/` → `plugin/skills/`, so they need no
materialization and are unaffected by whether the hook runs.)

Claude Code loads hooks from the **project root** only. There is no
mechanism for a multi-root workspace to register hooks per repository
(see [hooks](https://code.claude.com/docs/en/hooks) — settings resolve
from managed policy, `~/.claude/settings.json`, and the project root's
`.claude/settings.json`).

So the hook runs in a session that attaches **this repository alone**,
and does not run in a session that attaches it **alongside another**. In
the multi-repo case both repositories are cloned as siblings:

```
/home/user/            <- project root; no .claude/, so no hooks
  agent-plugins/       <- .claude/settings.json here is never read
  strata/              <- same
```

Observed 2026-08-01 in a session with `agent-plugins` and `strata` both
attached: no `bd` on PATH and no `strata` binary. (The skills are
committed symlinks, so they were present regardless — only the two
binaries were missing.) The hook script itself is fine — run by hand it
completes successfully. It was simply never invoked.

`strata` has the same exposure through its own
`scripts/setup-cloud-env.sh`.

## The environment contract: write access to `refs/dolt/data`

Read this before the setup script — it is the one requirement that cannot be
satisfied from inside a running session.

This repository's issue tracker is a Dolt database synced to `refs/dolt/data`
on the GitHub remote. `.beads/*.jsonl` is untracked and is **not** a fallback:
there deliberately isn't one, because carrying both models means two code
paths and a rule for which wins. A session that can read but not write the
tracker can close a bead and lose it when the container is reclaimed, so
`.claude/hooks/session-start.sh` **verifies** write access with a real probe
push and **exits 1** when it cannot.

The environment must supply two things:

1. **`AGENT_PLUGINS_ACCESS_TOKEN`** — a GitHub PAT with push access to this
   repository. Set it at [claude.ai/code](https://claude.ai/code) on the
   environment's variables. **Environment variables are read at container
   creation**, so adding one cannot rescue a session already running; that
   session has to be restarted.

2. **Direct, unproxied egress to `github.com`.** The egress proxy replaces the
   `Authorization` header with a brokered GitHub App installation token scoped
   to `refs/heads`, which returns 403 on `refs/dolt/*`. The hook handles this
   by setting `http.https://github.com/.proxy` to empty and installing a
   credential helper that reads the PAT from the environment at call time, so
   the token is never written to disk. If an environment blocks direct egress,
   the hook will fail its probe and say so.

This cost a full session to diagnose, so the correction is worth keeping: the
403 was **not** GitHub refusing `refs/dolt/*` writes. The measurement that
settled it — a push carrying a **deliberately invalid** token succeeded, which
is only possible if something upstream substituted a working credential.

### Also user-scope: `worktree.baseRef`

`~/.claude/settings.json` must carry `{ "worktree": { "baseRef": "head" } }`.
Unset, a git worktree branches from the repository's **default branch** rather
than session HEAD, so every bead in a run silently loses the commits the sprint
branch already carries. Now that every bead runs in a worktree
(`agent-plugins-92q.30`), this affects every run, not just parallel ones.

It is user-scope global state that no repository owns, and **a fresh container
does not have it** — measured 2026-08-11, where pre-flight reported
`BB_WORKTREE_BASEREF='unset'`. Set it in the environment setup script below, or
expect to set it by hand at the start of each session. Pre-flight reports it,
and the skill treats a wrong value as stop-and-fix at worktree dispatch.

The `bead-bonanza` skill never names a credential. Its requirement is the
capability, "write access to `refs/dolt/data`"; the token and the proxy bypass
are this repository's environment configuration.

## The fix: an environment setup script

A [setup script](https://code.claude.com/docs/en/cloud-environments#setup-scripts)
runs before Claude Code launches, as root, independent of the project
root. It is snapshotted into the
[environment cache](https://code.claude.com/docs/en/cloud-environments#environment-caching),
so it runs once and later sessions start with the results already on
disk.

Configure it at [claude.ai/code](https://claude.ai/code): open the
environment selector above the message box, hover the environment, and
use the settings icon. Paste into **Setup script**.

```bash
#!/bin/bash
# Provision bd and the strata CLI for sessions in this environment.
# Runs before Claude Code launches, so it does not depend on which
# repository is the project root. Cached after the first run.
#
# Every step is best-effort: a non-zero exit fails the whole session.

export HOME="${HOME:-/root}"
export DEBIAN_FRONTEND=noninteractive

# bd embeds Dolt, whose regex support is cgo-bound to ICU. Without these
# headers the source build fails at `unicode/uregex.h: No such file`.
apt-get update -qq >/dev/null 2>&1 || true
apt-get install -y -qq libicu-dev >/dev/null 2>&1 || true

# Release-binary lookup is unavailable behind the egress proxy
# (api.github.com returns 403), so this falls back to a source build.
go install github.com/steveyegge/beads/cmd/bd@latest >/dev/null 2>&1 || true
[ -x "$(go env GOPATH)/bin/bd" ] \
  && install -m 0755 "$(go env GOPATH)/bin/bd" /usr/local/bin/bd 2>/dev/null || true

# strata publishes no reachable release binary either; build from source.
if [ ! -x /usr/local/bin/strata ]; then
  git clone --filter=blob:none --quiet \
    https://github.com/strata-lang/strata.git /opt/strata-src >/dev/null 2>&1 || true
  if [ -f /opt/strata-src/crates/strata-cli/Cargo.toml ]; then
    ( cd /opt/strata-src && cargo build --release -p strata-cli ) >/dev/null 2>&1 || true
    [ -x /opt/strata-src/target/release/strata ] \
      && install -m 0755 /opt/strata-src/target/release/strata /usr/local/bin/strata || true
  fi
fi

# Worktree base ref: without this a git worktree branches from the default
# branch instead of session HEAD, and every bead in a bead-bonanza run loses
# the commits its sprint branch already carries. User-scope global state that
# no repository owns, so it belongs here.
mkdir -p "$HOME/.claude"
if [ ! -s "$HOME/.claude/settings.json" ]; then
  printf '{"worktree":{"baseRef":"head"}}\n' > "$HOME/.claude/settings.json"
elif command -v python3 >/dev/null 2>&1; then
  python3 -c 'import json,os; p=os.path.expanduser("~/.claude/settings.json"); d=json.load(open(p)); d.setdefault("worktree",{})["baseRef"]="head"; json.dump(d,open(p,"w"),indent=2)' 2>/dev/null || true
fi

exit 0
```

### Trade-offs

- **It duplicates logic** that also lives in
  `.claude/hooks/session-start.sh`. The two can drift. They are kept
  deliberately different in scope: the setup script provisions only the
  two binaries, which is VM provisioning and exactly what a setup script
  is for; the hook additionally does repository-specific work
  (bootstrapping the beads database from the Dolt remote and verifying
  write access to it, regenerating vendored references) that cannot run before the
  repository is cloned.
- **The five-minute budget is real.** A cold `cargo build --release` of
  `strata-cli` took roughly two minutes in this environment, and the
  `bd` source build a few more. If the script starts timing out, drop
  the strata half and leave it to the hook, since the strata skill is
  the only thing that needs it.
- **The cache expires** after roughly seven days, and rebuilds when the
  script or the network allowlist changes. A session that lands on a
  cold cache waits for the script.
- **It is per-environment, not per-repository.** It cannot be committed
  here, so it needs re-pasting into any new environment. That is the
  cost of the layer that actually runs.

## The fallback that lives in this repository

`CLAUDE.md` opens with a bootstrap check telling the agent to run the
hook by hand when `bd` or `strata` is missing. CLAUDE.md *is* read from
a nested repository even when its hooks are not, so this covers the
multi-repo case without any environment configuration.

It is a fallback rather than the fix: it costs a few minutes at the
start of the session that needs it, and it depends on the agent acting
on the instruction rather than on the harness enforcing it.

## Checking which path ran

```bash
command -v bd strata          # both present -> provisioning worked
cat ~/.cache/strata-cli/build.log   # the hook's strata build output
```

`bd` and `strata` are the only signal that the hook (or the setup
script) ran; the skills under `.claude/skills/` are committed symlinks
and are present either way.
