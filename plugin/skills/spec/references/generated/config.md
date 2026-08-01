# Config 

## Table of Contents

- [SpecsConfig](#specsconfig)
- [WorkspaceConfig](#workspaceconfig)
- [ProfileConfig](#profileconfig)
- [ConventionsConfig](#conventionsconfig)

## SpecsConfig

The `specs.toml` file an adopting repository authors at its root.

This is the public half of the configuration contract: the surface teams
outside the CLI's own team write, and therefore the surface whose
compatibility promise costs something to keep. It is deliberately small.
A repository says which repository it is, where its documents live,
which profile is active, and a short list of house conventions —
and nothing else.

In particular there is no way to define a document type, register a
block tag, or supply a schema from here. Those belong to the profile,
and holding them back is what keeps the promise cheap: a key that was
never public cannot be broken by changing it. Promotion runs one way and
is additive, so a repository that never wrote a key it could not write
is unaffected by a later version that lets it.

Undeclared keys are a defect rather than a courtesy. A misspelled key is
far more common than a key from the future, and silently ignoring one
would let a repository believe it had configured something it had not.

| Property | Type | Documentation |
|----------|------|---------------|
| workspace | [WorkspaceConfig](#workspaceconfig) | Identifies this repository and the directory its documents live in. |
| profile | [ProfileConfig](#profileconfig) | Selects the document vocabulary to check against. |
| conventions | [ConventionsConfig](#conventionsconfig)? | House conventions. Every key resolves from the active profile when absent, so the whole table may be omitted. |


## WorkspaceConfig

Identifies a repository's document collection.

| Property | Type | Documentation |
|----------|------|---------------|
| name | String | The name this repository's documents are attributed to.<br><br>It exists so that indexes emitted by several repositories can be merged into one architecture graph without two documents that both declare `adr-001` colliding. |
| root | String | The directory the CLI walks to discover the corpus.<br><br>Resolved relative to the directory containing `specs.toml`, so a configuration file is portable with its repository and a check run in CI resolves the same corpus as a check run locally. |


## ProfileConfig

Selects the active document vocabulary.

| Property | Type | Documentation |
|----------|------|---------------|
| name | String | The name of a registered profile.<br><br>Exactly one profile applies per invocation. A value naming no registered profile is reported as a finding listing the names that are registered, rather than as a bare startup failure. |


## ConventionsConfig

House conventions an adopting repository may set, and the profile
supplies when it does not.

| Property | Type | Documentation |
|----------|------|---------------|
| line_width | UInt? | The maximum width of a prose line, measured in Unicode codepoints rather than bytes, so an em dash counts as one character and not as its three UTF-8 bytes.<br><br>Lines inside fenced code blocks, table rows, and lines containing a URL are exempt. Defaults to the active profile's registered value. |


