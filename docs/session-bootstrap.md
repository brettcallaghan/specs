# Session bootstrap

How this repository's dependencies get installed in a Claude Code cloud
session, why the SessionStart hook alone is not enough, and what to
paste where.

This is an operations note, not a specification — it carries no
frontmatter and is not a corpus member.

## The problem

`.claude/settings.json` registers `.claude/hooks/session-start.sh` as a
SessionStart hook. That hook installs `bd`, installs the `strata` CLI,
installs the `swarm` plugin and builds the `swarm` CLI, and regenerates
the vendored shared references. (The skills are committed
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

The symptom is no `bd`, `strata` or `swarm` on PATH and no `swarm`
plugin; this repository's own skills are committed symlinks and are
present either way. The hook script is fine —
run by hand it completes. It is simply never invoked.

`strata` has the same exposure through its own
`scripts/setup-cloud-env.sh`.

## The tracker needs nothing from the environment

This repository's issue tracker is a Dolt database (prefix `specs-`) stored on
the ordinary branch `beads/data` of this repository's git remote.
`.beads/config.yaml`'s `sync.remote` names the remote; the branch is written
into each clone's Dolt config by `scripts/beads-setup.sh`, because `bd` has no
config key for a ref. The SessionStart hook runs that script and **exits 1**
when it leaves the tracker empty: `.beads/*.jsonl` is untracked and is **not**
a fallback, and a session that cannot sync the tracker can close a bead and
lose it when the container is reclaimed.

No variable, secret or network rule is needed. A Claude Code cloud session's
GitHub credential is accepted on `refs/heads/*` (and refused on `refs/dolt/*`,
Dolt's default, which is why the data is on a branch), and the git host is
already reachable. A local developer runs `./scripts/beads-setup.sh` once
after cloning and uses whatever git credential pushes their code.

### The pre-move history

Until 2026-10-06 the tracker lived on this repository's own git remote under
`refs/dolt/data`, with the `agent-plugins-` prefix. Only open work moved to
the new database. The full tracker as it stood at the move — all 690 beads,
closed ones included, with their comments and dependencies — is committed as
one `bd export` file under [`docs/archive/`](archive/). Ids of the form
`agent-plugins-…` anywhere in this repository resolve there, not in the live
tracker.

### Not needed here any more: `worktree.baseRef`

`{ "worktree": { "baseRef": "head" } }` must be in effect. Unset, a git
worktree branches from the repository's **default branch** rather than session
HEAD, so every bead in a run silently loses the commits the sprint branch
already carries. Every bead runs in a worktree, so this affects every run, not
just parallel ones.

**This repository commits it to `.claude/settings.json`**, which every clone and
every fresh container inherits, so the setup script below no longer has to write
it and a session that skips the script is not exposed. Claude Code resolves the
key highest-precedence-first — `.claude/settings.local.json`, then
`.claude/settings.json`, then `~/.claude/settings.json` — so the committed value
holds unless someone sets the local file. Pre-flight reads all three and reports
which one answered.

**It is no longer this repository's concern.** `bead-bonanza` and its worktree
dispatch moved to [brettcallaghan/swarm](https://github.com/brettcallaghan/swarm),
which commits the same key for the same reason; nothing left here runs a bead in
a worktree. The committed value stays because a `swarm` run against THIS
repository still needs it in effect, and because it costs nothing. The
`$HOME/.claude/settings.json` step in the script below is belt-and-braces rather
than the only source, and is safe to drop.

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
# Provision bd for sessions in this environment.
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

# Worktree base ref: without this a git worktree branches from the default
# branch instead of session HEAD, and every bead in a bead-bonanza run loses
# the commits its sprint branch already carries. User-scope, so it belongs here.
mkdir -p "$HOME/.claude"
if [ ! -s "$HOME/.claude/settings.json" ]; then
  printf '{"worktree":{"baseRef":"head"}}\n' > "$HOME/.claude/settings.json"
elif command -v python3 >/dev/null 2>&1; then
  python3 -c 'import json,os; p=os.path.expanduser("~/.claude/settings.json"); d=json.load(open(p)); d.setdefault("worktree",{})["baseRef"]="head"; json.dump(d,open(p,"w"),indent=2)' 2>/dev/null || true
fi

exit 0
```

The script does not install `strata`: it runs before this repository is
cloned, and `scripts/install-strata-cli.sh` — which the hook runs, and which
the `CLAUDE.md` fallback names — installs the release binary in seconds. That
script reads `STRATA_ACCESS_TOKEN`, a fine-grained PAT with read access to
`strata-lang/strata`, so set it as an environment variable on the
environment. It goes away when `strata-lang/strata` becomes public —
`agent-plugins-6gn` tracks the removal.

### Trade-offs

- **It duplicates logic** that also lives in
  `.claude/hooks/session-start.sh`, so the two can drift. Keep them
  different in scope: the setup script provisions only `bd`;
  the hook does the repository-specific work (hydrating the beads database
  from the `beads/data` branch, regenerating vendored references) that cannot run before the repository is cloned.
- **The five-minute budget is real.** The `bd` source build takes a few
  minutes cold.
- **The cache expires** after roughly seven days, and rebuilds when the
  script or the network allowlist changes. A session that lands on a
  cold cache waits for the script.
- **It is per-environment, not per-repository.** It cannot be committed
  here, so it needs re-pasting into any new environment.

## The fallback that lives in this repository

`CLAUDE.md` opens with a bootstrap check telling the agent to run the
hook by hand when `bd` or `strata` is missing. CLAUDE.md *is* read from
a nested repository even when its hooks are not, so this covers the
multi-repo case without any environment configuration. It costs a few
minutes at the start of the session that needs it, and it depends on the
agent acting on the instruction rather than on the harness enforcing it.

## Checking which path ran

```bash
command -v bd strata swarm    # all three present -> provisioning worked
claude plugin list            # specs@brettcallaghan and swarm@brettcallaghan-private, user scope
cat ~/.cache/strata-cli/installed-release   # the strata release tag installed
cat ~/.cache/swarm-cli/build.log    # the hook's swarm clone, plugin and build output
```

`swarm` and `strata` are provisioned by the hook only — the setup script
above installs neither, so a session where the hook did not run has neither
even when `bd` is present from the cache. The clone reads
`SWARM_ACCESS_TOKEN`, falling back to `AGENT_PLUGINS_ACCESS_TOKEN`, which
that same PAT covers; `SWARM_SRC_DIR` points the build at a checkout
already on disk, and `SWARM_SKIP=1` skips the step.

A swarm build log ending in `could not read Username for 'https://github.com'`
means the clone was unauthenticated — neither token is set, or the PAT no
longer reads `brettcallaghan/swarm`. Re-running the hook by hand with the
variable set recovers it.

A build log ending in **`remote: Write access to repository not granted`**
and a **403** means something different, and the message is misleading twice
over: it names *write* access for a read-only clone, and it looks like a
proxy or a scope problem. It is neither. `credential.helper` is multi-valued
and git tries **global** values before local ones, so any bare global helper
answers before the clone's own. A PAT that cannot read the cloned repository
therefore authenticates the clone, and GitHub refuses. The remedy is
`-c credential.helper=` before the clone's helper, which resets the inherited
list; the hook's swarm clone does this. Measured on a clone of
`strata-lang/strata`: the identical clone returns 403 without the reset and
succeeds with it.

The hook registers no github.com credential helper of its own, and removes the
URL-scoped helper and proxy bypass that the old `refs/dolt/data` tracker
needed from any container that still carries them. The reset stays as defence
in depth against a bare helper arriving from somewhere else.

`strata` is the signal that the hook ran, and `bd` that the hook or the
setup script did; the skills under `.claude/skills/` are committed symlinks
and are present either way.
