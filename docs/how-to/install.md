# Install the plugin and the CLI

The Specs plugin holds the skills. The `specs` CLI checks and queries the
documents the skills write. The plugin installs the CLI for you, so most
projects need only the first section.

## Install the plugin

The `brettcallaghan` marketplace at
[brettcallaghan/agent-plugins](https://github.com/brettcallaghan/agent-plugins)
serves the latest 0.x release of the plugin. In Claude Code, run:

```
/plugin marketplace add brettcallaghan/agent-plugins
/plugin install specs@brettcallaghan
```

When a session starts, the plugin puts the CLI release that matches the
plugin's version first on the session's `PATH`. It downloads and
checksum-verifies that release the first time, and keeps each version it
downloads, so a later update downloads only once.

- **On your own machine:** a different version of `specs` that's already
  installed stays as it is. It's shadowed for the session only.
- **In a Claude Code cloud session:** the plugin's version also replaces
  `/usr/local/bin/specs`.

If the plugin can't install the CLI, the session's context gets one line that
begins `specs plugin:`, names the cause, and says how to fix it.

## Use your own `specs` instead

To keep the plugin from installing the CLI, set `SPECS_CLI_PROVISION=off`.
The skills then run whichever `specs` is on your `PATH`, whatever its
version. If no `specs` is available, the skills check documents by hand and
say so. A manual check is slower, and a reviewer can't rerun it.

Install the CLI with Go:

```bash
go install github.com/brettcallaghan/specs/cli/specs@latest
```

Or download a prebuilt binary from a `cli/vX.Y.Z` release. Each release has
binaries for Linux, macOS, and Windows on amd64 and arm64. The binaries are
statically linked, so they need no Go toolchain, runtime, or other software.

## Configure the spec corpus

The CLI and the skills read [`specs.toml`](../reference/configuration.md) to
find your spec root. The CLI looks for the nearest `specs.toml` in the
working directory or a directory above it, up to the repository root. To name
a file instead, pass `--config <file>`.

Create `specs.toml` at the root of your repository:

```toml
[workspace]
name = "my-project"
root = "specs"

[profile]
name = "specs"
```

Without a `specs.toml`, a CLI command exits with status 3, and its message
says where it searched and suggests `--config`.

## Install the skills from a checkout

If you're developing the skills, you can install them into
`~/.claude/skills` from a checkout of this repository instead of through the
plugin system:

```bash
scripts/install-spec-skill.sh
```
