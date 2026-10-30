# Configuration

The `specs` CLI and the skills read their configuration from `specs.toml`.
For the full Contract, including how the CLI validates the file, see
[Contract: Specs Configuration](../../specs/contracts/specs-configuration-contract.md).

## Where the CLI looks

The CLI uses the nearest `specs.toml` in the working directory or a directory
above it. The search stops at the repository root, or at the filesystem root
outside a repository. `--config <file>` names a file and turns the search
off.

When it finds no `specs.toml`, the CLI exits with status 3 and doesn't guess
a spec root.

## Keys

| Key | Required | Description |
|:---|:---|:---|
| `[workspace].name` | Yes | Identifies this repository's documents when indexes from several repositories are merged. |
| `[workspace].root` | Yes | The spec root that the CLI walks, relative to the directory that holds `specs.toml`. |
| `[profile].name` | Yes | The document vocabulary to activate. The Specs plugin uses `specs`. |
| `[conventions].line_width` | No | The line width, in Unicode code points. When it's absent, the active profile supplies the value. |

The CLI reports any other key as a finding, so a misspelled key doesn't go
unnoticed.

## Example

This repository's own `specs.toml`:

```toml
[workspace]
name = "agent-plugins"
root = "specs"

[profile]
name = "specs"

[conventions]
line_width = 80
```
