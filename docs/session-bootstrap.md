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

## The cloud environment contract: write access to `refs/dolt/data`

**This section applies to Claude's cloud environments only.** A local
developer already has git credentials for this repository, so `bd dolt push`
works with no token and no proxy bypass; skip to [the setup
script](#the-fix-an-environment-setup-script). In a cloud container neither is
true, and neither can be fixed from inside a running session — read this first.

This repository's issue tracker is a Dolt database synced to `refs/dolt/data`
on the GitHub remote. `.beads/*.jsonl` is untracked and is **not** a fallback:
there deliberately isn't one, because carrying both models means two code
paths and a rule for which wins. A session that can read but not write the
tracker can close a bead and lose it when the container is reclaimed, so
`.claude/hooks/session-start.sh` **verifies** write access with a real probe
push and **exits 1** when it cannot.

A cloud environment must supply two things:

1. **`AGENT_PLUGINS_ACCESS_TOKEN`** — a GitHub PAT with push access to this
   repository. Set it at [claude.ai/code](https://claude.ai/code) on the
   environment's variables. **Environment variables are read at container
   creation**, so adding one cannot rescue a session already running; that
   session has to be restarted.

2. **Direct, unproxied egress to `github.com`.** The egress proxy replaces the
   `Authorization` header with a brokered GitHub App installation token scoped
   to `refs/heads`, which returns 403 on `refs/dolt/*` — the 403 is the proxy,
   **not** GitHub refusing `refs/dolt/*` writes. The hook handles this by
   setting the proxy to empty **for this repository's own origin URL** —
   `http.<origin>.proxy`, never the host-wide `http.https://github.com/.proxy`,
   which would reach every other github.com remote in the container while the
   credential helper beside it does not — and installing a credential helper
   scoped the same way, which reads the PAT from the environment at call time
   so the token is never written to disk.

   One consequence to know about: the bypass is what routes this remote to the
   PAT, and that PAT has no `workflow` scope, so a push touching
   `.github/workflows/` is refused. Git config cannot scope a bypass by
   refspec, so it cannot apply to `refs/dolt/*` alone. Push such a change with
   `git -c "http.<origin>.proxy=$HTTPS_PROXY" push`, which restores the proxy's
   brokered token for that one command, or grant the PAT the scope. If an environment blocks direct egress,
   the hook will fail its probe and say so.

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
  the hook does the repository-specific work (bootstrapping the beads
  database from the Dolt remote and verifying write access to it,
  regenerating vendored references) that cannot run before the
  repository is cloned.
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
claude plugin list            # sf@spec-driven-dev and swarm@swarm, user scope
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

The hook's own Dolt helper is registered under a URL-scoped key
(`credential.https://github.com/<owner>/<repo>.helper`, and the same key with
a `.git` suffix), never the bare one, so it no longer answers for other repos.
The reset stays as defence in depth against a bare helper arriving from
somewhere else.

`strata` is the signal that the hook ran, and `bd` that the hook or the
setup script did; the skills under `.claude/skills/` are committed symlinks
and are present either way.
