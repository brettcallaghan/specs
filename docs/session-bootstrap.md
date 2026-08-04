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

exit 0
```

### Trade-offs

- **It duplicates logic** that also lives in
  `.claude/hooks/session-start.sh`. The two can drift. They are kept
  deliberately different in scope: the setup script provisions only the
  two binaries, which is VM provisioning and exactly what a setup script
  is for; the hook additionally does repository-specific work
  (reconstructing the beads database from `.beads/issues.jsonl`,
  regenerating vendored references) that cannot run before the
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
