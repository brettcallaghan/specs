---
title: "Spec CLI"
entity_name: Spec CLI
description: The command-line binary that indexes a specification corpus
  and reports conformance findings, wrapping the Document Core and
  profile modules in a command surface
slug: container-spec-cli
type: container-blueprint
status: approved
contracts_consumed: []
modules:
  - "module-document-core"
  - "module-spec-profile"
---

# Spec CLI

## Overview

The Spec CLI is the single deployable artifact that implements this
framework's specification tooling. It serves several FRDs, of which
[[frd-spec-index]] is one, and this blueprint describes the container as a
whole rather than that one feature. It walks a specification corpus, builds
a machine-readable index of every document's metadata, structure, and
relationships, and reports deterministic conformance findings against the
rules the active profile registers.

The container itself is deliberately thin. Every capability lives in one of
two composed modules: [[module-document-core]], which knows about Markdown
documents with frontmatter that reference each other, and
[[module-spec-profile]], which supplies this framework's document
vocabulary. The container contributes only the command surface: argument
parsing, configuration loading, output rendering, and exit codes. The split
exists so that the same capabilities can later be wrapped in an HTTP API or
an MCP server without first extracting them from a CLI.

- **Responsibility:** present the indexing and conformance capabilities of
  its composed modules as a command-line interface, and own the
  process-level concerns those modules deliberately don't: argument
  parsing, configuration discovery, output format, and exit status.
- **Runtime:** a single statically linked binary, invoked per command. It
  has no daemon, no persistent process, and no network listener.

## Tech stack

These are the technologies the command surface itself adds.
[[module-document-core]] and [[module-spec-profile]] list those of the
composed modules, and `cli/go.mod` is the full list.

- **Language:** Go, built as one statically linked binary. [[adr-015]]
- **Command tree:** `spf13/cobra`, which provides commands, flags, and help
  output.
- **Configuration:** `BurntSushi/toml` reads `specs.toml`, and
  `santhosh-tekuri/jsonschema` validates it against the schema generated
  from `schemas/config.strata`. [[adr-016]]
- **Base revisions:** the `git` executable, which only
  `criterion --changed` requires on `PATH`. [[adr-031]]

## Container context

```mermaid
C4Container
    title Container Context — Spec CLI
    Person(author, "Spec Author", "Writes and reviews specification documents")
    System_Ext(agent, "Coding Agent", "Authors and validates documents in a session")
    System_Ext(ci, "CI Pipeline", "Runs conformance checks on push and pull request")
    System_Ext(catalogue, "Architecture Catalogue", "Central service aggregating indexes across repositories")
    Container(cli, "Spec CLI", "Go binary", "Indexes a spec corpus and reports conformance findings")
    ContainerDb(corpus, "Spec Corpus", "Markdown files on disk", "The specification documents and the configuration file")
    Rel(author, cli, "Runs checks and queries", "shell")
    Rel(agent, cli, "Queries the corpus and checks documents", "subprocess, JSON")
    Rel(ci, cli, "Runs conformance checks", "subprocess, exit status")
    Rel(catalogue, cli, "Consumes the emitted index", "JSON artifact")
    Rel(cli, corpus, "Reads documents and configuration", "filesystem")
```

**Spec Author → Spec CLI.** A human runs the CLI directly to check a
document before review, or to ask for the next identifier. The surface is
the command tree and its `--help` output.

**Coding Agent → Spec CLI.** An agent invokes the CLI as a subprocess and
parses its JSON output instead of reading documents itself. The latency
requirements in FEAT-SPEC-009 exist for this interaction: the agent asks
one question and acts on the answer.

**CI Pipeline → Spec CLI.** An adopting repository runs the CLI in its
build and fails on a nonzero exit status. The pipeline invokes the CLI as a
subprocess, with no runtime installed beyond the binary itself.

**Architecture Catalogue → Spec CLI.** A central service consumes the
emitted index to compose an organization-wide architecture graph. The CLI
doesn't push. It emits an artifact that a pipeline uploads, so the CLI has
no network dependency and no knowledge of the catalogue. The index schema
that governs this exchange isn't a Contract document yet (see Open
questions).

**Spec CLI → Spec Corpus.** The CLI reads Markdown documents and one
configuration file, and modifies no document. Read-only operation is a
property of this feature, not an implementation detail: [[frd-spec-index]]
places document modification out of scope. `links` reaffirms this rather
than narrowing it. Its rewritten document goes only to standard output, and
it creates, modifies, and deletes no file (AC-SPECMD-005.3). That includes
the derived index cache that, as the Data section below describes, `check`,
`refs`, `slug-exists`, and `next-id` write on a cold cache. `links` opens
that cache warm without persisting it, for exactly this reason.

## Components

```mermaid
C4Component
    title Components — Spec CLI
    Container_Boundary(cli, "Spec CLI") {
        Component(cmd, "Command Surface", "Go, cobra", "Command tree, flag parsing, output rendering, exit codes")
        Component(core, "Document Core", "Go module", "Corpus indexing and conformance for markdown documents")
        Component(profile, "Specs Profile", "Go module", "This framework's document vocabulary and rules")
    }
    Rel(cmd, profile, "Activates and configures")
    Rel(cmd, core, "Invokes indexing, checks and queries")
    Rel(profile, core, "Registers types, schemas and rules into")
```

### Command Surface

```component
entity_name: Command Surface
slug: component-command-surface
description: Command tree, flag parsing, configuration discovery, output rendering and exit codes
```

(Go, `cobra`) — Command tree, flag parsing, configuration discovery, output
rendering, and exit codes. This is the only component the container owns
outright. It resolves `specs.toml`, activates the profile, dispatches to
Document Core, and turns findings into either human-readable text or the
JSON that AC-SPEC-006.5 requires. It holds no domain logic. A check that
lives here instead of in a module is a layering defect, because it would be
unavailable to a future API or MCP wrapper.

**Built (configuration resolution)** — `cli/cmd`. `Load` reads `specs.toml`
and checks it against the generated
`cli/core/schema/specs-config.schema.json`. It reports every defect in one
run, each as an offending key with its span (AC-SPEC-008.4) and never
defaulted (AC-SPEC-008.5), through the same `finding.Report` that a corpus
check produces its findings through. It resolves `[workspace].root`
relative to the configuration file's own directory (AC-SPEC-008.1). On a
clean run, it returns what the rest of the container needs: the spec root,
workspace identifier, activated profile, and resolved conventions
(AC-SPEC-008.2). An absent `specs.toml` is reported as
`*cmd.ErrConfigNotFound` rather than as a finding, per
[[contract-specs-configuration]]'s "no file to carry a span into"
(AC-SPEC-008.3), and the run exits 3.

Every command reaches `Load` through `loadConfig`. Given `--config`, it
reads the file that flag names and searches for no other (AC-SPEC-008.8).
Without `--config`, it runs `discoverConfig`, which searches the working
directory and then each directory above it. The nearest `specs.toml` wins
(AC-SPEC-008.7). The search stops after the repository root: the nearest
directory holding a `.git` entry, which is a directory in a clone and a
file in a worktree or submodule. Outside any repository, the search runs to
the filesystem root (AC-SPEC-008.9). Discovery finds the root by looking
for `.git` rather than by running `git`, so it costs no subprocess and
works where `git` isn't installed. [[adr-031]]'s `git` dependency belongs
to change detection alone. A failed search names every directory it
checked, nearest first, and `--config` as the remedy (AC-SPEC-008.3). The
discovered path is handed on to `corpusOptions`, so the paths the file
declares resolve beside it and every reported path is relative to its
directory (AC-SPEC-008.1). Discovery runs inside a command, never before
one, so help and invocation errors need no `specs.toml`.

`Registry` is the profile registry that [[contract-specs-configuration]]
names as the customization seam. It maps a `[profile].name` value to a
`corpus.Profile`. `BuiltinRegistry` is the one place in this container that
imports `cli/profile`, to register `specs` as the built-in profile. A
`[profile].name` that names no entry in the registry is reported as a
finding listing the registered names, not as a bare lookup failure.

Once `[profile].name` resolves to a registered profile, `Load` also runs
`corpus.ValidateRegistration` against it. A registration is compiled in and
has no position of its own, so each defect it reports becomes a finding
positioned at `[profile].name`, the one place in `specs.toml` that selected
it. The finding names the offending key and, where the defect belongs to
one document type rather than to the profile's lifecycle registration, that
type (AC-SPECREF-004.3). As with every other `specs.toml` defect, a
registration defect is never defaulted around: no `Config` is resolved
while one is reported.

**Built (link-map reading)** — also `cli/cmd`, beside `Load` and reading
the same way. `LoadLinkMap` takes a YAML file that the caller names and
resolves it to a `LinkMap` (AC-SPECMD-005.4). The file is a mapping whose
keys are corpus document slugs and whose value for each slug is a record
carrying a `path` string. `LoadLinkMap` ignores any property other than
`path`, so the file can carry a caller's own per-document data that the CLI
never interprets (AC-SPECMD-005.6). It decodes with `goccy/go-yaml`
([[adr-015]]), so it accepts a map written in JSON unchanged, because JSON
is valid YAML (AC-SPECMD-005.7).

The link map is file input that the caller supplies, not repository
configuration, so unlike `specs.toml` it has no not-found contract of its
own. A file that can't be read or parsed, or a record that declares no
`path`, is reported as a finding that names the file and the fault. No map
is resolved at all, not even partially, which is the same discipline `Load`
applies to `specs.toml` (AC-SPECMD-005.8). A file that parses but has the
wrong shape (a top-level list or scalar, an entry that isn't a record, or a
`path` that isn't a string) is reported once per offending entry. Each
finding names what was found and states the shape a link map holds, so the
message itself says how to fix it (AC-SPECMD-005.13). A slug repeated as a
top-level key is a parse fault: `goccy/go-yaml` itself rejects it as a
duplicate mapping key, instead of letting one record silently win.
`LoadLinkMap` doesn't interpret, resolve, or normalize a `path` value. It
reads and returns the value exactly as written.

**Built (command tree)** — also `cli/cmd`, via `cobra` ([[adr-015]]).
`Execute` builds the command tree (`index`, `check`, `slug-exists`,
`next-id`, `refs`, `links`, `feature`, `criterion`, `headings`, `list`,
`c4`). It converts each command's flags into calls against Document Core
(`corpus.Build`, `corpus.Open`, `corpus.Load`, `corpus.Corpus.Check`/`CheckFor`,
`corpus.Corpus.Slug`, `corpus.Corpus.InboundReferences`/`OutboundReferences`,
`corpus.Corpus.Feature`/`Features`/`FeaturesByCode`,
`corpus.Corpus.Criterion`/`CriteriaOf`/`CriteriaByTag`,
`corpus.Corpus.DocumentSummaries`, `corpus.Corpus.C4Model`,
`numbering.Next`, `references.Extract`/`Resolve`,
`mdlinks.Resolve`/`Substitute`, `gherkin.Render`), and converts each result
back into standard output and an exit status. No file here decides what
makes a document valid or scores a check. Each of those calls is the whole
of what a subcommand does with its input.

**Built (exit statuses)** — `cli/cmd/exit.go`, per
[[cli-exit-status-standards]]. Every run exits with one of five statuses
and no other:

- `0`: success.
- `1`: a gate failed (the corpus author acts).
- `2`: an invocation error, such as an unknown command or flag, a missing
  or surplus argument, a value outside a registered set, or an operand the
  command can't act on (the caller acts).
- `3`: an environment error, such as no `specs.toml`, a `specs.toml`
  defect, or an output path that can't be written (the repository's
  maintainer acts).
- `4`: an internal error (this tool's maintainer acts).

A failing gate isn't an error: it sets the status and returns its findings
as the answer. A run ending `2`, `3`, or `4` returns a classified error
instead (`invocationError`, `environmentError`, or `internalError`). An
error that cobra raises before any command runs is `2`, an unclassified
`*fs.PathError` is `3`, and any other error nobody classified is `4`. A
missing classification therefore surfaces as the bug it is, not as a
plausible status. Document Core returns its own sentinel errors.
Classifying them is this component's job, because a future API or MCP
wrapper would map the same sentinels to its own codes.

Standard output is buffered and reaches the process only when the run ends
`0` or `1`. A run ending `2`, `3`, or `4` therefore writes nothing there,
not even half an answer written before it failed. It writes its reason to
standard error instead, beginning with the offending value or condition and
with no `Error: ` prefix. A panic in any command is recovered, and the run
exits `4` with the panic's value and stack on standard error. An
unrecovered Go panic would exit `2` and read as "you called me wrong".
Where a run meets more than one class, the highest wins (`4`, `3`, `2`,
`1`). The findings for a `specs.toml` defect go to standard error as text,
whatever `--json` says, and the run exits `3`.

`index` always builds fresh (`corpus.Build`, not the warm-cache
`corpus.Open`) and writes through `corpus.Write`. It accepts the `--out`
option that `IndexPath` resolves (AC-SPEC-011.1). A member that failed to
parse is recorded in the index like any other, so `index` exits zero
regardless (AC-SPEC-002.5).

`check` opens the corpus warm at the default index path and accepts
`--json`, `--strict`, `--only`, and `--verbose`. Its answer is a
command-level view over the generated `specmodel.FindingReport`, and holds
nothing else: a `summary` counting the findings by severity (`errors`,
`warnings`, and `unverified`, all three always present), then the
`findings`, then, only under `--verbose`, the report's `checks`
(AC-SPEC-006.11, AC-SPEC-006.12). It renders as YAML through `renderYAML`
by default and as JSON under `--json`: one value in two syntaxes, so the
two carry the same values in the same order (AC-SPEC-006.5,
[[cli-output-standards]]).

`--only` narrows the report through `finding.NewReport`'s own accounting
rather than by filtering text. A rule that `--only` names is requested and
then marked ran from its matching findings (if any), and a rule left out is
never requested. So `checks` lists exactly the checks the run was asked
for, each `passed`, `failed`, or `unverified`, and an unrequested check is
absent rather than reported as passed (AC-SPEC-006.8, AC-SPEC-006.9). The
summary counts only the requested rules' findings. The exit status follows
`finding.Report.HasErrors`, whatever view was rendered (AC-SPEC-006.13).
Any error exits 1 unconditionally. Warning or unverified findings alone
exit 0 unless `--strict` is given (AC-SPEC-006.6, AC-SPEC-006.7).

Given a path, `check` reads that document as git HEAD holds it, through
[[module-document-core#Revisions]] (`Repository.FileAt`). It takes the slug
the document declared there with `corpus.DeclaredSlug` and passes it to
`CheckFor` as the former slug. The answer then also carries the references
elsewhere that the document's change broke (AC-SPECREF-002.11). A document
that HEAD doesn't hold has no former slug. Where HEAD can't be read (no
work tree, no runnable `git`, or no commit), `check` writes one `note:`
line to standard error saying those references weren't checked. It answers
with the document's own findings, and the exit status follows them
(AC-SPECREF-002.13).

`checkRules` declares every rule a run can raise, so `--only` can name each
one and a clean run reports each as passed. It holds every `c4/*` rule that
Document Core raises over the C4 diagrams in architecture blueprints
(`c4/parse`, `c4/undeclared-node`, `c4/ambiguous-node`, `c4/bidirectional`,
`c4/duplicate-edge`, `c4/drawing-conflict`, `c4/interaction-conflict`,
`c4/contract-unresolved`, `c4/contract-mismatch`, and `c4/tag-separator`),
so `--only` can select each one. The command surface adds no C4 logic of
its own.

`slug-exists` and `next-id` open the corpus warm and answer
`corpus.Corpus.Slug` and `numbering.Next` directly (AC-SPEC-011.3,
AC-SPEC-011.4). `slug-exists <slug>` answers a mapping that holds `slug`,
`declared`, and `path`, so the command has one key set whatever it finds.
`slug` is spelled as the declaring member writes it, or as given where no
member declares it. `path` is the declaring member's path, or `null` where
no member declares it. The comparison ignores letter case, an undeclared
slug is an answer and exits 0, and `--json` writes the same values in the
same order (AC-SPEC-007.2).

`next-id <type>` answers a mapping that holds `number`, from
`numbering.Next`, and then `slug`, the slug that `numbering.Format` forms
from that number under the type's registered width (for example,
`number: 32` and `slug: adr-032`). It renders as YAML by default and as
JSON under `--json`, with the same values in the same order
(AC-SPEC-007.1). A type that the profile doesn't register as numbered exits
2 with nothing on standard output.

`refs` opens the corpus warm, as `check`, `slug-exists`, and `next-id` do,
and takes one argument, written `<slug>` or `<slug>#<anchor>`. By default
it answers `corpus.Corpus.InboundReferences`: every corpus member whose body
or a registered frontmatter property references the argument
(AC-SPECREF-003.1). Given `--outbound`, it answers
`corpus.Corpus.OutboundReferences` instead: the documents that slug
references (AC-SPECREF-003.2). It converts an argument and a direction into
that one corpus lookup and nothing else, which is the same no-domain-logic
discipline that every subcommand here holds to. It always exits zero: a
slug that the corpus resolves but nothing references, and a slug that no
corpus member declares at all (by the same `corpus.Corpus.Slug` query that
AC-SPEC-007.2 defines for `slug-exists`), are told apart by the `declared`
property, not by the same empty result (AC-SPECREF-003.5).

An anchor qualifies the argument rather than changing what it looks up. The
same inbound set is narrowed to the references whose anchor resolves to
that heading, which is the blast radius of renaming it (AC-SPECANCH-003.1).
Like any anchor, the asked anchor may be a path of headings
(AC-SPECANCH-002.10). Markdown Structure's `SelectPaths` matches it against
the document's own headings. A reference reaches the query when the heading
its `anchor_resolution` records, or one of its ambiguous matches, is one
that the asked anchor selected. Headings are compared by the position of
the heading line, since two headings with one name are two headings. Which
heading a reference reached is read from its own `anchor_resolution`, which
the References component settled during resolution, so this command can't
disagree with the checks.

A reference whose anchor matched more than one heading resolves to none of
them, and is listed for each of them, carrying `ambiguous_anchor`
(AC-SPECANCH-003.4). An anchor that no reference reaches answers with the
empty collection and exits zero (AC-SPECANCH-003.3). An anchored query also
carries `anchor_declared` beside `declared`, in the compact and `--full`
views alike. It's true when the anchor selects some heading of the document
by that same rule, and false when it selects none or the slug is
undeclared. An uncited heading thus reads apart from one that doesn't exist
(AC-SPECANCH-003.6). `--outbound` takes no anchor and refuses one: an
anchor names a heading of the *cited* document, and the outbound direction
asks what the given document cites.

The answer is a mapping of the properties the query established, with the
references beneath it. It renders as YAML, and as JSON under `--json`, per
[[cli-output-standards]]. `--json` changes the syntax of whichever view is
chosen, never the view itself (AC-SPECREF-003.7). By default the references
are the compact view: one entry per referring document, in the order the
corpus returned them, holding a `path` and that document's references. Each
reference is reduced to its `line` and `text`. It carries `origin` only for
a reference in frontmatter, `resolution` (as its bare `state`) only for a
reference that didn't resolve, and `ambiguous_anchor` only where that is
true (AC-SPECREF-003.6).

`--full` answers with every field of every reference instead. Each is the
generated `specmodel.Reference`, carrying its file path and span
(AC-SPECREF-003.3), embedded rather than copied, so that view is over the
generated model ([[adr-016]]) rather than a second declaration of it. The
compact view is derived from the full answer's references after the anchor
has narrowed them, so the two views hold exactly the same references
(AC-SPECREF-003.8). The YAML renderer resolves a value that marshals itself
through its own `MarshalJSON`. That's what makes the two renderings carry
the same keys in the same order for a union such as `resolution` or
`anchor_resolution`.

`links` takes one operand, the path of a Markdown document, and an optional
`--map` that names a link-map file. It doesn't require the operand to be a
corpus member. It extracts the document's references
(`references.Extract`), resolves them against the corpus
(`references.Resolve`), settles each one's Markdown-link target and raises
the two reachability findings (`mdlinks.Resolve`), and rewrites the
document (`mdlinks.Substitute`).

Unlike `refs`, which always exits zero, `links` writes the rewritten
document to standard output only when no finding of severity error was
raised, and exits 1 otherwise (AC-SPECMD-002.2). The exception is an
operand that can't be read (`document/unreadable`) or a `--map` that can't
be read or parsed (`linkmap/invalid`). Those are operands the command can't
act on, and they exit 2 with the finding's message on standard error. A run
whose only findings are warnings still emits the document and exits 0
(AC-SPECMD-002.6, AC-SPECMD-002.8). Findings and the replacement count go
to standard error, never to the stream carrying the document
(AC-SPECMD-003.2). `links` has no `--strict` option. AC-SPECMD-002.8
already fixes the exit status a warning-only run takes, so a flag reversing
it would contradict the criterion rather than narrow it. That reasoning
doesn't apply to `check`'s own `--strict`.

`feature <id> --emit` writes that feature, with its tags, description, and
criteria, as a standalone `.feature` file (AC-GHK-009.1). It looks the
feature up by identifier from the warm corpus index, the same way `refs`
and `slug-exists` look one up. It writes to standard output by default, or
to the path that `--out` names. One feature yields one file, per
[[adr-021]]'s constraint 8. The rendering itself is `gherkin.Render`,
Document Core's inverse of `gherkin.Parse`. This component's own part is
writing the result, and converting an index record's identifier, negation,
manual-verification, and verification-level properties back into the tags
that name them. Those tags are the same vocabulary that `cli/profile`
registers for parsing, read in reverse.

An emitted file always carries the owning-document tag, `@frd-<slug>`, even
where the feature's own source was a fenced block that carried none: the
emitted file is a feature source in its own right, and per [[adr-021]]
constraint 11, a feature source names its owner. An identifier that names
no feature in the index is an operand `--emit` can't act on. The identifier
is named on standard error and the run exits 2, because a producer has no
answer that could carry "not declared".

Querying a feature without `--emit` reads the same warm index by a
different corpus call. `feature <id>` returns that feature's record through
`corpus.Corpus.Feature` (AC-GHK-007.1). `feature --code <CODE>` returns
every feature that `corpus.Corpus.FeaturesByCode` groups under that code.
`feature` with no argument and no `--code` returns every feature that
`corpus.Corpus.Features` holds (AC-GHK-003.3, AC-GHK-007.3). A feature
identifier resolves the same way regardless of which admissible form or
file it sits in, because the lookup reads the index's `Id` field and never
its `Path` (AC-GHK-003.1). An identifier and a code match regardless of
letter case, as a slug does, and the record carries the spelling the corpus
declares rather than the caller's (AC-GHK-007.7).

By default the answer is the compact view: each feature's `id`, `code`,
`name`, `path`, and `criteria_count`. `--full` returns the whole record,
its description and criteria included. Both views derive from the same
selection, so they list the same features in the same order
(AC-GHK-007.10, AC-GHK-007.11). `--full` with `--emit` is refused with exit
`2` (AC-GHK-007.12). The answer has one shape, whatever the arguments
select: a mapping whose one key, `features`, holds a list of the matching
generated `specmodel.Feature` records, a lookup of one identifier included.
It renders as YAML through `renderYAML` by default and as JSON under
`--json`, with the same values in the same order (AC-GHK-007.1,
AC-GHK-007.5, [[cli-output-standards]]).

Given an identifier or `--code`, the mapping also carries `declared` before
`features`. An identifier that names no feature, or a code that no feature
carries, is answered with `declared: false` and `features: []`, and exits
zero. A subject the corpus doesn't hold is an answer, and `declared` is
what tells it apart from an empty selection (AC-GHK-007.4). With neither,
no subject is named and the key is absent. `--code` and `--json` cannot be
combined with `--emit`, and an identifier and `--code` cannot both be
given: the command's `Args` validator and `RunE` branch on `--emit` before
either flag-conversion path runs, so the two never share logic.

`criterion --changed` lists every acceptance criterion whose content hash
differs between the corpus at a git base and the current corpus. It
classifies each as changed, added, or removed by whether its identifier is
present in both, only now, or only at the base (AC-GHK-008.1,
AC-GHK-008.3). The base is the merge-base of HEAD and the repository's
default branch, or of HEAD and the branch, tag, or commit that `--since`
names (AC-GHK-008.4, AC-GHK-008.5, AC-GHK-008.10). The default branch is
the first of these that resolves: `origin/HEAD`'s target, `origin/main`,
`origin/master`, `main`, and `master`. Where the default branch chosen is
remote-tracking and a local branch of the same name is behind it, the
command writes one warning line to standard error and leaves standard
output unchanged (AC-GHK-008.14, AC-GHK-008.15). Each listed criterion
carries its `name` (AC-GHK-008.11). `--full` adds its `record` and, for a
changed or removed criterion, its `base_record`. Without `--changed`,
`--full` is refused with exit 2 (AC-GHK-008.12, AC-GHK-008.16).

`corpus.Build` builds both sides. The current side walks the working tree,
so committed, staged, unstaged, and untracked documents are all compared
(AC-GHK-008.6). The base side walks the spec root as the base commit holds
it, which [[module-document-core#Revisions]] materializes into a temporary
directory. `RepoRoot` and `Root` are rebased there at the same
repository-relative paths, so an unmoved document reports one path on both
sides. Neither side reads, writes, or requires `specs.json`: any query may
rebuild it, so it cannot stand for an earlier state (AC-GHK-008.7). Only
identifier and content hash are compared, never a position or a Markdown
diff, so a criterion that only moved isn't listed (AC-GHK-008.2).

The answer is a mapping that holds `base` and then `criteria`. `base` holds
`revision` (the revision as named, or the default branch found) and
`commit` (the merge-base's full object name). `criteria` holds one
`{id, classification, path}` per listed criterion, sorted by identifier.
`path` is the document's path in the current corpus or, for a removed
criterion, in the base (AC-GHK-008.11). The answer renders as YAML, or as
JSON under `--json`. The listing is a query: it exits 0 whatever it lists,
and 1 for a nonempty list only under `--exit-code` (AC-GHK-008.8). A
comparison that can't be made answers nothing. An unknown `--since`
revision exits 2. These exit 3 (AC-GHK-008.9): no work tree, no runnable
`git`, no default branch, no merge-base (a shallow clone is named as the
cause where it is one), and a spec root outside the work tree, which is
named with the work tree (AC-GHK-008.13). `--since` and `--exit-code`
without `--changed` exit 2 (AC-GHK-008.12).

Querying a criterion without `--changed` reads the same warm index by a
different corpus call. `criterion <id>` returns that criterion's record
through `corpus.Corpus.Criterion` (AC-GHK-007.2). A criterion identifier,
like a feature identifier given to `--of`, matches regardless of letter
case (AC-GHK-007.7). `criterion --of <feature-id>` returns every criterion
that feature holds, through `corpus.Corpus.CriteriaOf`.
`criterion --tag <qualifier>` returns every criterion that
`corpus.Corpus.CriteriaByTag` matches against the active profile's
registered qualifier vocabulary. That vocabulary is the same
`Negative`/`Manual`/`Level` properties and `FeatureTagRules.Qualifiers`
registration that `--emit`'s `criterionTags` reads in the write direction
(AC-GHK-007.3).

The answer has the same one shape: a mapping whose one key, `criteria`,
holds a list of the matching generated `specmodel.Criterion` records, a
lookup of one identifier included. It renders as YAML by default and as
JSON under `--json` (AC-GHK-007.2, AC-GHK-007.5). Given an identifier or
`--of`, the mapping also carries `declared` for that criterion or feature,
as `feature` does: an undeclared one answers `declared: false` and
`criteria: []`, and exits zero (AC-GHK-007.4). `--tag` names a registered
value, not a subject, and its answer carries no `declared`. A qualifier
that no criterion carries is an empty selection. A qualifier written with
Gherkin's leading `@` matches as the bare name does (AC-GHK-007.9). A
qualifier that the profile doesn't register is an invocation error:
`CriteriaByTag` returns `corpus.ErrUnregisteredQualifier`, the value and
every registered qualifier reach standard error with nothing on standard
output, and the run exits 2 (AC-GHK-007.8).

`--changed` cannot be combined with an identifier, `--of`, or `--tag`. An
identifier cannot be combined with `--of` or `--tag`, and `--of` and
`--tag` cannot both be given. `criterion`'s `RunE` validates these
combinations before either flag-conversion path runs, which is the same
separation `feature` keeps between `--emit` and its query surface.

Both `feature` and `criterion` open the corpus warm (`corpus.Open`), as
`check`, `slug-exists`, `next-id`, and `refs` do, so a query answers from
the persisted index without re-walking the corpus when its inputs are
unchanged (AC-GHK-007.6). Every corpus call that the `feature` and
`criterion` queries above name is a linear scan over the warm index's
in-memory `SpecIndex.Features`, not a second on-disk structure. "Warm"
already means that no filesystem walk happens. The two subcommands hold to
the same no-domain-logic discipline as `refs`: each converts its flags into
one corpus lookup, and the result into standard output and an exit status,
with no selection or matching logic of its own.

`headings` answers a heading selection. By default it gives every match's
`anchor`, `ref`, `level`, and `start_line` (AC-SPECSTR-003.16). With
`--full`, it gives each match's file path, heading text, heading
identifier, anchor, ref, level, and the span of the content beneath it
(AC-SPECSTR-003.1). Both views list the same matches in the same order, and
`--full` with `--emit` is refused with exit `2` (AC-SPECSTR-003.17).

The selection is one operand, written `<document>#<heading>`, and either
half can be omitted. `<document>` alone selects every heading in that
document, `#<heading>` selects that heading in every corpus member, and an
absent operand selects every heading in the corpus (AC-SPECSTR-003.3). A
document half that holds a path separator or ends in `.md` is read as a
repository-relative path, and any other as a slug. A heading is matched on
the identifier that `markdown.Identifier` derives, so letter case and
punctuation don't decide a match (AC-SPECSTR-003.2).

The heading half may be a path of headings separated by `#`, such as
`module-document-core#Components#References`. A path selects the heading
named by its last segment beneath a heading named by each earlier segment,
in order and at any depth (AC-SPECSTR-003.10, AC-SPECSTR-003.11). It does
so through Markdown Structure's `SelectPaths`, the same selection that
anchor resolution uses. Each match's `anchor` is its heading path: the
identifier of every enclosing heading from the outermost down, then its
own, joined by `#` (AC-SPECSTR-003.9). Its `ref` is `<slug>#<anchor>`,
using the member's declared slug or, where it declares none, its path
(AC-SPECSTR-003.12). So a `ref` tells two same-named headings apart, and
wherever its anchor has more than one segment, it pastes back as a
selection that names that heading alone: a query spelling some heading's
full path selects the headings with exactly that path before any whose path
merely contains it.

A one-segment path is matched on its identifier, like any one-segment
selection. So some headings have no selection of their own: a top-level
heading of a document with no H1 that shares its identifier with another
heading, or one of two siblings that share text all the way up. Such a
heading's content is read through its span. A matched heading is returned
with every heading nested inside its section, each carrying its own level
(AC-SPECSTR-003.4).

`--emit` instead writes the content that the one matching heading stands
over, read back through the recorded span and written as itself, with no
envelope (AC-SPECSTR-003.5). Its span already covers the headings nested in
it, which are carried but not counted as matches. A raw payload has no
place for a delimiter, so under `--emit`, a selection that matches more
than one heading is refused. The run exits `2` with nothing on standard
output, and names each match on standard error by its `ref`
(AC-SPECSTR-003.7). Under `Select one of:`, it lists those whose selection,
run back through the same selection, matches that heading alone. It lists
apart those that no selection names alone, and for those, the query without
`--emit` returns the span.

A selection that matches nothing renders the empty collection and exits
zero (AC-SPECSTR-003.6). Under `--emit` it's refused instead: the run exits
`2` with nothing on standard output and a message saying whether the
document is undeclared or the heading is absent (AC-SPECSTR-003.13).
Members at a status that the profile registers as no longer in effect are
omitted, from the answer and from the `--emit` refusal's list, unless
`--all` is given (AC-SPECSTR-003.14). A selection that names such a
document by slug or path still answers (AC-SPECSTR-003.15). A selection
with a document half also carries `declared` before `matches`, false when
no corpus member is that document, so a missing heading reads apart from a
missing document (AC-SPECSTR-003.8). A selection with no document half
carries no such key. Like the queries beside it, `headings` opens the
corpus warm and holds no domain logic: the heading tree and the identifier
derivation are Markdown Structure's.

**`headings` is the first command written to [[cli-output-standards]].**
Its answer renders as YAML by default and as JSON under `--json`, both
carrying the same values in the same order. The YAML rendering lives in
`renderYAML`, in `cli/cmd`. It resolves the marshaled struct's fields the
way `encoding/json` resolves them: the same `json` tags, embedded structs
promoted the same way, and the same rule for which of two fields sharing a
name is visible. So the two renderings carry the same keys in the same
order, and neither can quietly gain or lose a value the other has. One
difference between them is required rather than accidental: an empty or
nil sequence renders `[]`, and an empty or nil mapping renders `{}`, per
[[cli-output-standards]], where JSON writes `null` for a nil one.
`cli/cmd/yaml_promotion_test.go` holds the renderer to this by rendering a
table of struct shapes both ways and comparing the key sequences, and
asserts the collection difference as the deliberate one it is.
`renderYAML` isn't a general YAML library. It writes block style with a
two-space indent, `[]` and `{}` for empty collections, block scalars for
multiline values trimmed at the edges, plain scalars where they round-trip
and double quotes where they don't, and no comments, anchors, aliases,
tags, or document markers.

`list` is the second, and is a listing over the frontmatter that every
member already declares. Under the single key `documents`, it holds one
entry per corpus member carrying its `path`, `slug`, `type`, `status`,
`title`, and `description`, in that order (AC-SPECLIST-001.2). The whole
of it is one `corpus.Corpus.DocumentSummaries` call. The summary shape, the
entry order, and both filters are Document Core's, and this component
decides none of them. It opens the corpus warm, as `check` and the queries
beside it do: every value it reports is already in the index, so a listing
needs no walk.

`--type` names the document types to return. It's repeatable and
comma-separated, from one `StringSliceVar`, exactly as `check`'s `--only`
is (AC-SPECLIST-002.2). Per-type boolean flags wouldn't reach a type that a
custom profile registers, which is why the surface takes values rather than
switches. `--all` asks for summaries at every status. Without it, the ones
at a status the active profile registers as no longer in effect are omitted
(AC-SPECLIST-003.2), because a superseded document read as current misleads
an agent rather than merely costing it tokens. `--json` selects the JSON
rendering, carrying the same values in the same order as the YAML one
(AC-SPECLIST-004.1, AC-SPECLIST-004.2).

An empty answer is an answer. A filter that matches nothing and a corpus
that holds no members both render `documents: []` and exit zero
(AC-SPECLIST-001.7). A `--type` value that the active profile doesn't
register is an invocation error, which `DocumentSummaries` reports instead
of answering. The report, holding the value and the registered type names,
reaches standard error, and the run exits 2 (AC-SPECLIST-002.4). This is
the deliberate departure from [[cli-output-standards]]' empty-answer rule,
and it's narrow. The rule is about the world, where a corpus with no
Contracts is a fact. An unregistered type is a malformed question, and an
empty collection would make a typo indistinguishable from an empty
category. Nothing else reaches the output: no count, no heading naming the
command, and no trailing summary (AC-SPECLIST-004.5). The exit status
carries the outcome. Two runs over an unchanged corpus emit the same bytes,
because the order is the index's own (AC-SPECLIST-004.3). A title carrying
a colon and a space comes back double-quoted, because `renderYAML` already
quotes what a plain scalar wouldn't round-trip (AC-SPECLIST-004.4).

`c4` emits the corpus's C4 model and is wired the same way: flags in, one
`corpus.Corpus.C4Model` call, one rendering out. It takes no positional
argument and two flags: `--root <slug>`, which narrows the model to that
element and its descendants, and `--json`. It opens the corpus warm, as
`list` does, and passes `Config.Workspace` as the workspace name. It writes
the model as YAML by default and as JSON under `--json`, both carrying the
same values in the same order (AC-C4-004.1, AC-C4-004.2). The model carries
`schema_version` and the workspace name (AC-C4-004.3), and two runs over an
unchanged corpus emit the same bytes (AC-C4-004.4). It exits zero whatever
the model holds, excluded relationships and `c4/*` findings included
(AC-C4-004.7). A model narrowed by `--root` carries `root_declared`. It's
true when the slug names an element. It's false when the slug names none,
and the model then holds every collection empty: an answer Document Core
gives, exiting zero with nothing on standard error (AC-C4-004.6). A
whole-corpus model carries no `root_declared`.

A bare invocation or `--help` lists the subcommands with their one-line
synopses and, for each subcommand, its own flags. This is cobra's default
rendering of the `Use`, `Short`, and flag declarations above, so the listed
surface and the code that implements it can't drift (AC-SPEC-011.5, and the
Change guidance below). No hand-written help text exists anywhere in this
component.

The snapshot cases under `cli/testdata/cases/`, replayed by
`cli/cmd/snapshot_test.go`, verify observable command-tree behavior, per
[[testing-standards]]. `scripts/regenerate-cli-snapshots.sh` regenerates
their expected artifacts from the current implementation.

### Document Core

See [[module-document-core]] (internal). (Go module) — Corpus discovery,
frontmatter parsing and schema validation, Markdown structure and span
extraction, reference graph construction, numbering, and the index and
finding models. It knows nothing about this framework's document types.

### Specs Profile

See [[module-spec-profile]] (internal). (Go module) — This framework's
document vocabulary: one union schema over every document type, each type's
slug pattern and numbered-ness, and the convention defaults. It registers
itself into Document Core through the extension point that module defines.
Required sections and the validation of element record blocks are designed
in that blueprint and not yet built.

**Command Surface → Specs Profile.** The command
surface activates exactly one profile per invocation and hands it the
configuration it resolved. Which profile is active is a configuration
decision, not a compile-time one.

**Command Surface → Document Core.** The command surface delegates all
indexing, checking, and query work. It converts between the module's
finding and index models and the process's standard output and exit status.

**Specs Profile → Document Core.** The profile consumes Document Core's
extension point: it registers document types, schemas, patterns, and rules
into it, and never the other way round.
Document Core SHALL NOT import the profile.

## Provided Contracts

**Public** (consumable outside the system):

- [[contract-specs-configuration]] — the `specs.toml` format that an
  adopting repository authors, and the profile-registration surface beneath
  it.

## Consumed Contracts

None. The CLI reads files and emits artifacts. It calls no service and
depends on no interface that another team provides.

## Data

The CLI owns no persistent data. It reads the corpus and emits artifacts,
and any cache it keeps is derived and disposable.

- **Index cache** (on-disk, derived) — the built index, kept so that a
  later query answers without rereading the corpus. This is the "warm
  index" that AC-SPEC-009.1 and AC-SPEC-009.2 are measured against.
  - WHEN any input the index was built from has changed, the CLI SHALL
    rebuild rather than serve a stale answer.
  - The CLI SHALL produce identical output whether or not a cache is
    present, so deleting the cache is always safe.

**Built** — there is no separate cache. The emitted index *is* the cache:
`index` writes `specs.json` into the configured spec root, and `check`,
`slug-exists`, `next-id`, and `refs` open that same artifact warm from that
same default path. There is one file, one shape, and one code path, so
"delete the cache" and "delete the artifact" are the same act, and neither
can change an answer. An `index --out` that names another location writes
an artifact for a consumer, not a cache for the reading commands. Those
commands find nothing at the default path and rebuild.

The reading commands **also write it when they had to build it**. That is
what makes AC-SPEC-010.5's own example ("run twice in a row reads
`specs.json` the second time") true of a workflow that never invokes
`index`. Nothing instructs an author to run `index`, so without that write,
every check and query would rebuild from cold and never reach the warm path
that the FEAT-SPEC-009 budgets are stated against. That write is
best-effort and its failure is discarded, because a read-only checkout or a
race with a concurrent writer makes the answer no less correct. A query
that failed because it couldn't refresh a cache would be a worse tool than
a slow one. A failed write changes nothing observable: not the exit status,
not standard output, and not standard error. The index only stays cold
(AC-SPEC-010.7). AC-SPEC-010.4's atomic replace keeps a concurrent reader
from seeing a partial file.

**Staleness is decided by a content hash of the inputs, never by a
modification time.** Every emitted index carries an `input_hash`. It folds
each candidate file's repository-relative path and content digest together
with the inputs that change the index while the files stand still: the
workspace identifier, the active profile's schema and per-type table, and
the schema version. A persisted index is served warm only when it records
the corpus's current hash. Every other condition rebuilds: an absent file,
an unreadable one, one of another `schema_version`, or one carrying no
hash. The index is hashed rather than stat-ed because of the environment
the latency budget is claimed in. A fresh CI checkout gives every file a
new mtime and identical bytes, so an mtime-keyed cache would be cold on
every CI run, and AC-SPEC-009.1 would never hold there. Absolute paths
are deliberately excluded from the hash, so the same content checked out
at another location is still warm.

Invalidation is whole-index, not per-entry. The guarantee above says
"rebuild" rather than "rebuild the affected entries" because a rebuild
costs one corpus walk, and a partial invalidation would buy a fraction of
that while adding a way to be subtly wrong.

The correctness argument rests on one property: the fingerprint and the
build consume the same walk over the same bytes. That property is what
makes "this index was built from exactly these inputs" a checkable claim
rather than a hopeful one.

## Infrastructure and deployment

This container is standalone: there's no System Blueprint, so this section
is authoritative for its deployment.

- **Exclusive infrastructure:** none. The CLI has no database, no queue,
  and no network dependency at runtime.
- **Runtime dependency:** `criterion --changed` alone needs the `git`
  executable on `PATH`, because it reads its base revision through git
  ([[adr-031]]). Its absence is an environment error (exit 3) for that
  command. No other command requires git.
- **Deployment:** the CLI is distributed as a per-platform binary.
  AC-SPEC-009.4 through AC-SPEC-009.6 govern the shape: it runs with
  no language runtime installed beyond itself, it retrieves no
  executable content at install time, and it is installable from at
  least one package registry that supports on-premises mirroring.
  Cross-compilation produces the platform matrix from a single build.
- **Build:** `CGO_ENABLED=0`, so each artifact is one statically linked
  file with no libc to resolve at run time (AC-SPEC-009.4), and
  `-trimpath -ldflags="-s -w"`, so the bytes don't carry the builder's
  absolute paths. Six targets (`linux`, `darwin`, and `windows`, each on
  `amd64` and `arm64`) are cross-compiled on one Linux runner from one
  checkout. `.github/workflows/cli.yml` builds the whole matrix on every
  pull request, so a target that stops cross-compiling fails on the change
  that broke it rather than at release time.
- **Distribution — two channels.** The Go module proxy is the registry
  that satisfies AC-SPEC-009.6: it can be mirrored on premises (GOPROXY,
  Athens, Artifactory), and `go install …/cli/specs@<tag>` needs no
  release artifact at all. For a machine with no Go toolchain, tagged
  releases carry one archive per target plus `SHA256SUMS`. Unpacking one
  runs no install-time code, which is what AC-SPEC-009.5 turns on. Nothing
  here is signed yet. Signing is open, and belongs to the
  release-and-distribution ADR rather than to this blueprint.
- **The entry point is `cli/specs/`, not the module root.** Go names an
  installed binary after the directory holding its main package, so the
  path is what makes `go install` produce `specs` rather than `cli`. For
  the same reason, release tags are `cli/vX.Y.Z`: the module lives in a
  subdirectory, and the proxy resolves a subdirectory module's versions
  only from tags carrying that prefix.
- **Cross-platform reproducibility (AC-SPEC-009.7) is a property of the
  read path, not of the build.** The CLI normalizes CRLF to LF at the
  single point where it reads a candidate file, so a Windows checkout
  under `core.autocrlf` can't move a span's byte offsets or the content
  hash the warm-cache path compares. The consequence is deliberate: where
  the two differ, a reported byte offset indexes the normalized text rather
  than the on-disk file. Line and column, which a human or an editor acts
  on, are unaffected.

## Cross-cutting Standards

This container SHALL comply with:

- [[testing-standards]] — unit tests per declared component,
  harness-independent snapshot cases for observable behavior, and the
  coverage obligations tying both back to acceptance criteria.

[[skill-authoring-standards]] governs skill authoring and doesn't apply to
a compiled binary.

## Change guidance

- WHEN a capability is added to the command surface, that capability
  SHALL be implemented in [[module-document-core]] or
  [[module-spec-profile]] and invoked from the command surface, never
  implemented in the command surface itself.
- [[module-document-core]] SHALL NOT depend on [[module-spec-profile]]
  or on any framework-specific vocabulary. The dependency runs one way,
  and a violation is what would make the core unusable for the Obsidian
  vault case it's designed to serve.
- WHEN the emitted index changes shape, its schema version SHALL be
  updated in the same change, per AC-SPEC-002.2.
- WHEN a new command or flag is added, its `--help` text SHALL be
  derived from the same declaration that defines it, so the two cannot
  drift.
- **Verification gates:** `go test ./cmd/...` for command-surface
  changes, plus the snapshot cases under `cli/testdata/cases/` governed
  by [[testing-standards]]. WHEN the command surface changes in a way
  that alters observable output, the snapshot regeneration script SHALL
  be run and every resulting diff reviewed before commit. Changes
  reaching into a module also run that module's gates. CI runs the full
  suite.

## Key decisions

- [[adr-015]] — Go is the implementation language. The ADR records the
  measured basis and the rejected alternatives.
- [[adr-016]] — Strata is the single source for generated schemas, which
  is why `specs.toml` is validated against a schema generated from
  `schemas/config.strata` rather than one written by hand.
- [[adr-021]] — Gherkin is the notation for features and acceptance
  criteria. Command Surface queries the feature and criterion records
  that Document Core extracts, emits a feature back out as a standalone
  `.feature` file, and reports which criteria changed against a git base.
  The parsing and the rendering are Document Core's, and the tag
  vocabulary is [[module-spec-profile]]'s.
- [[adr-031]] — `criterion --changed` reads its base revision through the
  `git` executable, and builds the spec root as the base commit holds it
  with the same `corpus.Build` that the working tree uses. That's why that
  one command needs `git` on `PATH`.

## Open questions

- [ ] The emitted index is a published artifact with a versioned schema
      (AC-SPEC-002.2, AC-SPEC-002.3) that the Architecture Catalogue
      consumes. That makes it a Contract by this framework's own filing
      rule, but no Contract document exists yet. It's deferred to the
      index-schema ADR, which also has to settle whether references are
      modeled as object properties or as a separate edge collection.
      Decide whether the Contract is authored alongside that ADR or
      before it.

      **Decided, in part.** Promotion to a Contract document, with a
      formal versioning policy, is **deferred** and waits on a consumer or
      catalogue that actually exists. Until one does, the generated JSON
      Schema for `SpecIndex`, emitted from `schemas/index.strata` with
      every other model, is the index's machine-readable specification.
      It's the only consumer-facing statement of the shape.

      What's pinned meanwhile: `schema_version` is `2.2`, and the shape is
      **additive** under it. Features attach headings, extracted record
      blocks, and references to the same entry without changing the
      version. Changing what's already there (a property's meaning, its
      type, or its removal) bumps `schema_version`, the Strata source, and
      the emitted index together, per the Change guidance above. One
      exception: a property that no emitted index has ever carried —
      because the shape holding it has never been emitted with an
      instance — MAY be changed or removed without a bump, since no
      artifact a 1.0 consumer has read contains it. That rule is what
      makes a one-producer artifact safe to ship before its Contract is
      written: a consumer built against 1.0 keeps working until the
      version says otherwise. It's deliberately weaker than a Contract:
      nothing here says how long 1.0 is supported, or what a consumer is
      owed on a bump. Closing that gap is exactly what the deferred
      Contract is for.

- [ ] Whether the CLI ships a `serve` mode that holds the index in memory
      across queries. It isn't required (AC-SPEC-009.1 states a latency
      target, not a gate, and the warm path holds it on a quiet machine),
      but it's the natural home for filesystem watching, which is listed
      as a follow-on capability.

## References

- [[c4-standards]] — C4 mapping rules and diagram conventions
- [[frd-spec-index]] — the requirements this container realizes
- [[module-document-core]] — the general engine it composes
- [[module-spec-profile]] — the framework vocabulary it composes
