---
title: "Container Blueprint: Spec CLI"
entity_name: Spec CLI
description: The command-line binary that indexes a specification corpus
  and reports conformance findings, wrapping the Document Core and
  profile modules in a command surface
slug: container-spec-cli
type: container-blueprint
status: approved
updated_at: 2026-10-08T18:00:00Z
contracts_consumed: []
modules:
  - "module-document-core"
  - "module-spec-profile"
---

# Spec CLI

## Overview

The Spec CLI is the single deployable artifact realizing this
framework's specification tooling. [[frd-spec-index]] is the
first of the feature requirements documents it serves, and this
blueprint describes the container as a whole rather than that one
feature. It walks a specification corpus, builds a
machine-readable index of every document's metadata, structure and
relationships, and reports deterministic conformance findings against
the rules the active profile registers.

The container itself is deliberately thin. Every capability lives in one
of two composed modules — [[module-document-core]], which knows about
markdown documents with frontmatter that reference each other, and
[[module-spec-profile]], which supplies this framework's document
vocabulary. The container contributes only the command surface: argument
parsing, configuration loading, output rendering, and exit codes. That
split exists so the same capabilities can later be wrapped in an HTTP
API or an MCP server without extracting them from a CLI first.

- **Responsibility:** present the indexing and conformance capabilities
  of its composed modules as a command-line interface, and own the
  process-level concerns those modules deliberately do not — argument
  parsing, configuration discovery, output format, and exit status.
- **Technology:** Go, `cobra` for the command tree. See [[adr-015]] for
  the language decision and its measured basis.
- **Runtime:** a single statically linked binary, invoked per command.
  No daemon, no persistent process, no network listener.

## Container Context

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
document before review, or to ask what identifier is next. The surface
is the command tree and its `--help` output.

**Coding Agent → Spec CLI.** An agent invokes the CLI as a subprocess
and parses its JSON output, rather than reading documents itself. This
is the interaction the latency requirements in FEAT-SPEC-009 exist for:
the agent asks one question and acts on the answer.

**CI Pipeline → Spec CLI.** An adopting repository runs the CLI in its
build and fails on a non-zero exit status. The CLI is invoked as a
subprocess with no runtime installed beyond the binary itself.

**Architecture Catalogue → Spec CLI.** A central service consumes the
emitted index to compose an organization-wide architecture graph. The
CLI does not push; it emits an artifact that a pipeline uploads, so the
CLI has no network dependency and no knowledge of the catalogue. The
index schema that governs this exchange is not yet a contract document —
see Open Questions.

**Spec CLI → Spec Corpus.** The CLI reads markdown documents and one
configuration file, and writes nothing. Read-only operation is a
property of this feature, not an implementation detail:
[[frd-spec-index]] places document modification out of scope.
`links` re-affirms rather than narrows this: its rewritten document goes
only to standard output, and it creates, modifies and deletes no file
(AC-SPECMD-005.3) — including the derived index cache the "Data" section
below describes `check`, `refs`, `slug-exists` and `next-id` writing on a
cold cache, which `links` opens warm without persisting for exactly this
reason.

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

(Go, `cobra`) — Command tree, flag parsing, configuration discovery,
output rendering and exit codes. This is the only component the
container owns outright. It resolves `specs.toml`, activates the
profile, dispatches to Document Core, and turns findings into either
human-readable text or the JSON that AC-SPEC-006.5 requires. It holds no
domain logic: a check that lives here rather than in a module is a
layering defect, because it would be unavailable to a future API or MCP
wrapper.

**Built (configuration resolution)** — `cli/cmd`. `Load` reads and
checks `specs.toml` against the generated
`cli/core/schema/specs-config.schema.json`, reporting every defect in one
run — an offending key with its span (AC-SPEC-008.4), never defaulted
(AC-SPEC-008.5) — through the same `finding.Report` a corpus check
produces its findings through. It resolves `[workspace].root` relative to
the configuration file's own directory (AC-SPEC-008.1) and, on a clean
run, returns the spec root, workspace identifier, activated profile and
resolved conventions the rest of the container needs (AC-SPEC-008.2). An
absent `specs.toml` is reported as `*cmd.ErrConfigNotFound` rather than a
finding, per [[contract-specs-configuration]]'s "no file to carry a span
into" (AC-SPEC-008.3), and exits 3.

Every command reaches `Load` through `loadConfig`, which reads the file
`--config` names and searches for no other (AC-SPEC-008.8), or, with no
`--config`, runs `discoverConfig`: the working directory and then each
directory above it, the nearest `specs.toml` winning (AC-SPEC-008.7). The
search stops after the repository root, the nearest directory holding a
`.git` entry — a directory in a clone, a file in a worktree or submodule
— and, outside any repository, runs to the filesystem root
(AC-SPEC-008.9). The root is found by looking for `.git` rather than by
running `git`, so discovery costs no subprocess and works where `git` is
not installed; [[adr-031]]'s `git` dependency is change detection's
alone. A failed search names every directory it checked, nearest first,
and `--config` as the remedy (AC-SPEC-008.3). The discovered path is
handed on to `corpusOptions`, so the paths the file declares resolve
beside it and every reported path is relative to its directory
(AC-SPEC-008.1). Discovery runs inside a command, never before one, so
help and invocation errors need no `specs.toml`.

`Registry` is the profile registry [[contract-specs-configuration]]
names as the customization seam: it maps a `[profile].name` value to a
`corpus.Profile`, and `BuiltinRegistry` is the one place in this
container that imports `cli/profile` to register `specs` as
the built-in. `[profile].name` naming no entry in the registry is
reported as a finding listing the registered names, not a bare lookup
failure.

Once `[profile].name` resolves to a registered profile, `Load` also runs
`corpus.ValidateRegistration` against it. A registration is compiled in
and has no position of its own, so every defect it reports is turned
into a finding positioned at `[profile].name` — the one place in
`specs.toml` that selected it — naming the offending key and, where the
defect belongs to one document type rather than the profile's lifecycle
registration, that type (AC-SPECREF-004.3). As with every other
`specs.toml` defect, a registration defect is never defaulted around: no
`Config` is resolved while one is reported.

**Built (link-map reading)** — also `cli/cmd`, beside `Load` and reading
the same way: `LoadLinkMap` takes a caller-named YAML file — a mapping
whose keys are corpus document slugs and whose value per slug is a
record carrying a `path` string — and resolves it to a `LinkMap`
(AC-SPECMD-005.4). A property other than `path` is ignored, so the file
can carry a caller's own per-document data the CLI never interprets
(AC-SPECMD-005.6); decoding with `goccy/go-yaml` ([[adr-015]]) means a
map written in JSON is accepted unchanged, JSON being valid YAML
(AC-SPECMD-005.7). The link map is caller-supplied file input, not
repository configuration, so unlike `specs.toml` there is no
not-found contract of its own: a file that cannot be read or parsed, or
a record declaring no `path`, is reported as a finding naming the file
and the fault, with no map resolved at all — never partially, the same
discipline `Load` applies to `specs.toml` (AC-SPECMD-005.8). A file that
parses but has the wrong shape — a top-level list or scalar, an entry
that is not a record, a `path` that is not a string — is reported per
offending entry, naming what was found and stating the shape a link map
holds, so the message itself says how to fix it (AC-SPECMD-005.13). A slug
repeated as a top-level key is a parse fault `goccy/go-yaml` itself
rejects as a duplicate mapping key, rather than one record silently
winning. `LoadLinkMap` performs no interpretation, resolution or
normalization of a `path` value; it is read and returned exactly as
written.

**Built (command tree)** — also `cli/cmd`, via `cobra`
([[adr-015]]): `Execute` builds the command tree — `index`, `check`,
`slug-exists`, `next-id`, `refs`, `links`, `feature`, `criterion`,
`headings`, `list`, `c4` — and converts each one's flags into calls against Document Core
(`corpus.Build`, `corpus.Open`, `corpus.Load`, `corpus.Corpus.Check`/`CheckFor`,
`corpus.Corpus.Slug`, `corpus.Corpus.InboundReferences`/`OutboundReferences`,
`corpus.Corpus.Feature`/`Features`/`FeaturesByCode`,
`corpus.Corpus.Criterion`/`CriteriaOf`/`CriteriaByTag`,
`corpus.Corpus.DocumentSummaries`, `corpus.Corpus.C4Model`,
`numbering.Next`, `references.Extract`/`Resolve`,
`mdlinks.Resolve`/`Substitute`, `gherkin.Render`) and its result back into
stdout and an exit status. No file here decides what makes a document
valid or scores a check; every one of those calls is the whole of what a
subcommand does with what it receives.

**Built (exit statuses)** — `cli/cmd/exit.go`, per
[[cli-exit-status-standards]]. Every run exits with one of five statuses
and no other: `0` success; `1` a gate failed (the corpus author acts);
`2` an invocation error — an unknown command or flag, a missing or
surplus argument, a value outside a registered set, an operand the
command cannot act on (the caller acts); `3` an environment error — no
`specs.toml`, a `specs.toml` defect, an output path that cannot be
written (the repository's maintainer acts); `4` an internal error (this
tool's maintainer acts). A gate failing is not an error: it sets the
status and returns its findings as the answer. A run ending `2`, `3` or
`4` returns a classified error instead (`invocationError`,
`environmentError`, `internalError`); an error cobra raises before any
command runs is `2`, an unclassified `*fs.PathError` is `3`, and any
other error nobody classified is `4`, so a missing classification
surfaces as the bug it is rather than as a plausible status. Document
Core keeps returning its own sentinel errors; classifying them is this
component's job, because a future API or MCP wrapper would map the same
sentinels to its own codes. Standard output is buffered and reaches the
process only when the run ends `0` or `1`, so a run ending `2`, `3` or
`4` writes nothing there — not even half an answer written before it
failed — and writes its reason to standard error, beginning with the
offending value or condition and carrying no `Error: ` prefix. A panic
in any command is recovered and exits `4`, its value and stack on
standard error; an unrecovered Go panic would exit `2` and read as "you
called me wrong". Where a run meets more than one class, the highest
wins (`4`, `3`, `2`, `1`). A `specs.toml` defect's findings go to
standard error as text whatever `--json` says, and the run exits `3`.

`index` always builds fresh (`corpus.Build`, not the warm-cache
`corpus.Open`) and writes through `corpus.Write`, accepting the
`--out` option `IndexPath` resolves (AC-SPEC-011.1); a member that
failed to parse is recorded in the index like any other, so `index`
exits zero regardless (AC-SPEC-002.5). `check` opens the corpus warm at
the default index path and accepts `--json`, `--strict`, `--only` and
`--verbose`. Its answer is a command-level view over the generated
`specmodel.FindingReport`: a `summary` counting the findings by severity
(`errors`, `warnings`, `unverified`, all three always present), then the
`findings`, then — only under `--verbose` — the report's `checks`, and
nothing else (AC-SPEC-006.11, AC-SPEC-006.12). It is rendered as YAML by
`renderYAML` by default and as JSON under `--json`, one value in two
syntaxes so the two carry the same values in the same order
(AC-SPEC-006.5, [[cli-output-standards]]). `--only` narrows a report built from
`finding.NewReport`'s own accounting rather than by filtering text: a
rule named by `--only` is requested and then marked ran from the
matching findings (if any), and a rule left out is never requested, so
`checks` lists exactly the checks the run was asked for, each `passed`,
`failed` or `unverified`, and an unrequested check is absent rather than
reported passed (AC-SPEC-006.8, AC-SPEC-006.9); the summary counts only
the requested rules' findings. Exit status
follows `finding.Report.HasErrors`: any error exits 1 unconditionally;
warning or unverified findings alone exit 0 unless `--strict` is given
(AC-SPEC-006.6, AC-SPEC-006.7), whatever view was rendered
(AC-SPEC-006.13). Given a path, `check` reads that document as git HEAD
holds it through [[module-document-core#Revisions]]
(`Repository.FileAt`), takes the slug it declared there with
`corpus.DeclaredSlug`, and passes it to `CheckFor` as the former slug, so
the answer also carries the references elsewhere the document's change
broke (AC-SPECREF-002.11). A document HEAD does not hold has no former
slug. Where no work tree, no runnable `git` or no commit lets HEAD be
read, it writes one `note:` line to standard error saying those
references were not checked and answers with the document's own
findings, the exit status following them (AC-SPECREF-002.13). `checkRules` declares every rule a run
can raise, so `--only` can name each and a clean run reports each as
passed. It holds all ten `c4/*` rules Document Core raises over the C4
diagrams in architecture blueprints — `c4/parse`, `c4/undeclared-node`,
`c4/ambiguous-node`, `c4/bidirectional`, `c4/duplicate-edge`,
`c4/drawing-conflict`, `c4/interaction-conflict`,
`c4/contract-unresolved`, `c4/contract-mismatch` and `c4/tag-separator`
— so each is selectable with `--only`; the command surface adds no c4
logic of its own. `slug-exists` and `next-id` open the
corpus warm and answer `corpus.Corpus.Slug` and `numbering.Next`
directly (AC-SPEC-011.3, AC-SPEC-011.4). `slug-exists <slug>` answers a
mapping holding `slug`, as asked, `declared`, and `path`, the declaring
member's path or `null` where no member declares it, so the command has
one key set whatever it finds; the comparison ignores letter case, an
undeclared slug is an answer and exits 0, and `--json` writes the same
values in the same order (AC-SPEC-007.2). `next-id <type>` answers a
mapping holding `number`, from `numbering.Next`, then `slug`, the slug
`numbering.Format` forms from it under the type's registered width —
`number: 32`, `slug: adr-032` — as YAML by default and as JSON under
`--json`, the same values in the same order (AC-SPEC-007.1). A type the
profile does not register as numbered exits 2 with nothing on standard
output.

`refs` opens the corpus warm, the same as `check`, `slug-exists` and
`next-id`, and takes one argument written `<slug>` or `<slug>#<anchor>`:
by default it answers `corpus.Corpus.InboundReferences` — every corpus
member whose body or a registered frontmatter property references it
(AC-SPECREF-003.1) — and, given `--outbound`,
`corpus.Corpus.OutboundReferences` instead, the documents that slug
references (AC-SPECREF-003.2). It converts an argument and a direction
into that one corpus lookup and nothing else — the same no-domain-logic
discipline every subcommand here holds to — and always exits zero: a
slug the corpus resolves but nothing references, and a slug no corpus
member declares at all — the same `corpus.Corpus.Slug` query
AC-SPEC-007.2 defines for `slug-exists` — are told apart by the
`declared` property rather than by the same empty result
(AC-SPECREF-003.5).

An anchor qualifies the argument rather than changing what it looks up:
the same inbound set is narrowed to the references whose anchor
resolves to that heading, the blast radius of renaming it
(AC-SPECANCH-003.1). The asked anchor may be a path of headings, like any
anchor (AC-SPECANCH-002.10): it is matched against the document's own
headings by Markdown Structure's `SelectPaths`, and a reference reaches
the query when the heading its `anchor_resolution` records — or one of
its ambiguous matches — is one the asked anchor selected, compared by
the heading line's position, since two headings of one name are two
headings. Which heading a reference reached is read off its own
`anchor_resolution` — the References component settled it during
resolution — so this command cannot disagree with the checks. A reference whose anchor matched more
than one heading resolves to none of them and is listed for each of
them, carrying `ambiguous_anchor` (AC-SPECANCH-003.4); an anchor no
reference reaches answers with the empty collection and exits zero
(AC-SPECANCH-003.3). An anchored query also carries `anchor_declared`
beside `declared`, in the compact and `--full` views alike: true when the
anchor selects some heading of the document by that same rule, false when
it selects none or the slug is undeclared, so an uncited heading reads
apart from one that does not exist (AC-SPECANCH-003.6). `--outbound` takes no anchor and refuses one: an
anchor names a heading of the *cited* document, and the outbound
direction asks what the given document cites.

The answer is a mapping of the properties the query established, with
the references beneath it, rendered as YAML and as JSON under `--json`
per [[cli-output-standards]]; `--json` changes the syntax of whichever
view is chosen and never the view (AC-SPECREF-003.7). By default the
references are the compact view: one entry per referring document, in
the order the corpus returned them, holding a `path` and that document's
references, each reduced to its `line` and `text`, with `origin` only for
a reference in frontmatter, `resolution` only — as its bare `state` —
for a reference that did not resolve, and `ambiguous_anchor` only where
it is true (AC-SPECREF-003.6). `--full` answers with every field of every
reference instead: each is the generated `specmodel.Reference` — carrying
its file path and span (AC-SPECREF-003.3) — embedded rather than copied,
so that view is over the generated model ([[adr-016]]) rather than a
second declaration of it. The compact view is derived from the full
answer's references after the anchor has narrowed them, so the two
views hold exactly the same references (AC-SPECREF-003.8). The YAML renderer resolves a value that
marshals itself through its own `MarshalJSON`, which is what makes the
two renderings carry the same keys in the same order for a union such
as `resolution` or `anchor_resolution`.

`links` takes one operand, the path of a markdown document, and an
optional `--map` naming a link-map file; it does not require the operand
to be a corpus member. It extracts the document's references
(`references.Extract`), resolves them against the corpus
(`references.Resolve`), settles each one's markdown-link target and
raises the two reachability findings (`mdlinks.Resolve`), and rewrites
the document (`mdlinks.Substitute`). Unlike `refs`, which always exits
zero, `links` writes the rewritten document to standard output only when
no finding of severity error was raised, and exits 1 otherwise
(AC-SPECMD-002.2) — except that an operand that cannot be read
(`document/unreadable`) and a `--map` that cannot be read or parsed
(`linkmap/invalid`) are operands the command cannot act on, which exit 2
with the finding's message on standard error; a run whose only findings are warnings still emits
the document and exits 0 (AC-SPECMD-002.6, AC-SPECMD-002.8). Findings
and the replacement count go to standard error, never to the stream
carrying the document (AC-SPECMD-003.2). `links` carries no `--strict`
option: AC-SPECMD-002.8 already fixes the exit status a warning-only run
takes, so a flag reversing it would contradict the criterion rather than
narrow it — the same reasoning `check`'s own `--strict` does not apply
to.

`feature <id> --emit` writes that feature — its tags, description and
criteria, looked up by identifier from the warm corpus index the same
way `refs` and `slug-exists` look one up — as a standalone `.feature`
file: to standard output by default, to the path `--out` names
otherwise (AC-GHK-009.1). One feature yields one file, per
[[adr-021]]'s constraint 8. The rendering itself is `gherkin.Render`,
Document Core's inverse of `gherkin.Parse`; this component's own part is
converting an index record's identifier, negation, manual-verification
and verification-level properties back into the tags that name them —
the same vocabulary `cli/profile` registers for parsing, read in
reverse — and writing the result. An emitted file always carries the
owning-document tag, `@frd-<slug>`, even where the feature's own source
was a fenced block that carried none: the emitted file is a feature
source in its own right, and per [[adr-021]] constraint 11 a feature
source names its owner. An identifier naming no feature in the index is
an operand `--emit` cannot act on: it is named on standard error and the
run exits 2, since a producer has no answer that could carry "not
declared".

Querying a feature without `--emit` reads the same warm index by a
different corpus call. `feature <id>` returns that feature's record,
via `corpus.Corpus.Feature` (AC-GHK-007.1);
`feature --code <CODE>` returns every feature `corpus.Corpus.FeaturesByCode`
groups under that code, and `feature` with no argument and no `--code`
every feature `corpus.Corpus.Features` holds (AC-GHK-003.3, AC-GHK-007.3).
A feature identifier resolves the same way regardless of which admissible
form or file it sits in, because the lookup reads the index's `Id` field
and never its `Path` (AC-GHK-003.1). An identifier and a code match
regardless of letter case, as a slug does, and the record carries the
spelling the corpus declares rather than the caller's (AC-GHK-007.7).
The answer is the compact view by default — each feature's `id`, `code`,
`name`, `path` and `criteria_count` — and `--full` returns the whole
record, its description and criteria included, derived from the same
selection so both list the same features in the same order
(AC-GHK-007.10, AC-GHK-007.11); `--full` with `--emit` is refused with
exit `2` (AC-GHK-007.12). The answer has one shape whatever
the arguments select: a mapping whose one key, `features`, holds a list
of the matching generated `specmodel.Feature` records, one identifier's
lookup included, rendered as YAML by `renderYAML` by default and as JSON
under `--json` with the same values in the same order (AC-GHK-007.1,
AC-GHK-007.5, [[cli-output-standards]]). Given an identifier or
`--code`, the mapping also carries `declared` before `features`: an
identifier naming no feature, or a code no feature carries, is answered
with `declared: false` and `features: []`, and exits zero — a subject the
corpus does not hold is an answer, and `declared` is what tells it apart
from an empty selection (AC-GHK-007.4). With neither, no subject is named
and the key is absent. `--code` and `--json` cannot be
combined with `--emit`, and an identifier and `--code` cannot both be
given: the command's `Args` validator and `RunE` branch on `--emit` before
either flag-conversion path runs, so the two never share logic.

`criterion --changed` lists every acceptance criterion whose content
hash differs between the corpus at a git base and the current corpus,
classifying each as changed, added or removed by whether its identifier
is present in both, only now, or only at the base (AC-GHK-008.1,
AC-GHK-008.3). The base is the merge-base of HEAD and the repository's
default branch — the first that resolves of `origin/HEAD`'s target,
`origin/main`, `origin/master`, `main` and `master` — or of HEAD and the
branch, tag or commit `--since` names (AC-GHK-008.4, AC-GHK-008.5,
AC-GHK-008.10). Where the default branch chosen is remote-tracking and a
local branch of the same name is behind it, the command writes one
warning line to standard error and leaves standard output unchanged
(AC-GHK-008.14, AC-GHK-008.15). Each listed criterion carries its `name`
(AC-GHK-008.11); `--full` adds its `record` and, for a changed or removed
criterion, its `base_record`, and is refused with exit 2 without
`--changed` (AC-GHK-008.12, AC-GHK-008.16). Both sides are built by `corpus.Build`. The current side
walks the working tree, so committed, staged, unstaged and untracked
documents are all compared (AC-GHK-008.6); the base side walks the spec
root as the base commit holds it, which
[[module-document-core#Revisions]] materializes into a temporary
directory, with `RepoRoot` and `Root` rebased there at the same
repository-relative paths so an unmoved document reports one path on
both sides. Neither side reads, writes or requires `specs.json`: any
query may rebuild it, so it cannot stand for an earlier state
(AC-GHK-008.7). Only identifier and content hash are compared, never a
position or a markdown diff, so a criterion that only moved is not
listed (AC-GHK-008.2). The answer is a mapping holding `base` —
`revision`, the revision as named or the default branch found, and
`commit`, the merge-base's full object name — then `criteria`, one
`{id, classification, path}` per listed criterion sorted by identifier,
`path` being the document's in the current corpus or, for a removed
criterion, in the base (AC-GHK-008.11); it renders as YAML, or as JSON
under `--json`. The listing is a query: it exits 0 whatever it lists,
and 1 for a non-empty list only under `--exit-code` (AC-GHK-008.8). A
comparison that cannot be made answers nothing: an unknown `--since`
revision exits 2, while no work tree, no runnable `git`, no default
branch, no merge-base (a shallow clone is named as the cause where it
is one) and a spec root outside the work tree, named with the work tree
(AC-GHK-008.13), exit 3 (AC-GHK-008.9). `--since` and `--exit-code`
without `--changed` exit 2 (AC-GHK-008.12).

Querying a criterion without `--changed` reads the same warm index by a
different corpus call. `criterion <id>` returns that criterion's record
via `corpus.Corpus.Criterion` (AC-GHK-007.2), and a criterion identifier,
like a feature identifier given to `--of`, matches regardless of letter
case (AC-GHK-007.7). `criterion --of <feature-id>`
returns every criterion `corpus.Corpus.CriteriaOf` that feature holds, and
`criterion --tag <qualifier>` every criterion `corpus.Corpus.CriteriaByTag`
matches against the active profile's registered qualifier vocabulary —
the same `Negative`/`Manual`/`Level` properties and `FeatureTagRules.Qualifiers`
registration `--emit`'s `criterionTags` reads in the write direction
(AC-GHK-007.3). The answer has the same one shape: a mapping whose one
key, `criteria`, holds a list of the matching generated
`specmodel.Criterion` records, one identifier's lookup included, as YAML
by default and as JSON under `--json` (AC-GHK-007.2, AC-GHK-007.5).
Given an identifier or `--of`, the mapping also carries `declared` for
that criterion or feature, as `feature` does: an undeclared one answers
`declared: false` and `criteria: []`, and exits zero (AC-GHK-007.4).
`--tag` names a registered value, not a subject, and carries no
`declared`: a qualifier no criterion carries is an empty selection, a
qualifier written with Gherkin's leading `@` matches as the bare name
does (AC-GHK-007.9), and a qualifier the profile does not register is an
invocation error — `CriteriaByTag` returns
`corpus.ErrUnregisteredQualifier`, and the value and every registered
qualifier reach standard error with nothing on standard output and exit
2 (AC-GHK-007.8). `--changed` cannot be combined
with an identifier, `--of` or `--tag`; an identifier cannot be
combined with `--of` or `--tag`; and `--of` and `--tag` cannot both be
given — `criterion`'s `RunE` validates these combinations before either
flag-conversion path runs, the same separation `feature` keeps between
`--emit` and its query surface.

Both `feature` and `criterion` open the corpus warm (`corpus.Open`), the
same as `check`, `slug-exists`, `next-id` and `refs`, so a query answers
from the persisted index without re-walking the corpus when its inputs
are unchanged (AC-GHK-007.6). Every corpus call this paragraph and the one
above it name is a linear scan over the warm index's in-memory
`SpecIndex.Features`, not a second on-disk structure — "warm" already
means no filesystem walk happens, and the two subcommands hold to the
same no-domain-logic discipline `refs` does: each converts its flags into
one corpus lookup and the result into stdout and an exit status, with no
selection or matching logic of its own.

`headings` answers a heading selection with every match's `anchor`,
`ref`, `level` and `start_line` by default (AC-SPECSTR-003.16), and with
`--full` its file path, heading text, heading identifier, anchor, ref,
level and the span of the content beneath it (AC-SPECSTR-003.1); both
views list the same matches in the same order, and `--full` with `--emit`
is refused with exit `2` (AC-SPECSTR-003.17). The selection is one operand written
`<document>#<heading>`, either half omittable: `<document>` alone selects
every heading in that document, `#<heading>` that heading in every corpus
member, and an absent operand every heading in the corpus
(AC-SPECSTR-003.3). A document half holding a path separator or ending in
`.md` is read as a repository-relative path and otherwise as a slug. A
heading is matched on the identifier `markdown.Identifier` derives, so
letter case and punctuation do not decide a match (AC-SPECSTR-003.2). The
heading half may be a path of headings separated by `#` —
`module-document-core#Components#References` — selecting the heading
named by the last segment beneath a heading named by each earlier one, in
order and at any depth (AC-SPECSTR-003.10, AC-SPECSTR-003.11), through
Markdown Structure's `SelectPaths`, the selection anchor resolution uses.
Each match's `anchor` is its heading path — the identifier of every
enclosing heading from the outermost down, then its own, joined by `#`
(AC-SPECSTR-003.9) — and its `ref` is `<slug>#<anchor>`, the member's
declared slug or, where it declares none, its path (AC-SPECSTR-003.12),
so it tells two same-named headings apart and pastes back as a selection
naming that heading alone wherever its anchor has more than one segment: a query
spelling some heading's full path selects the headings with exactly that
path before any whose path merely contains it. A one-segment path is
matched on its identifier like any one-segment selection, so a top-level
heading of a document with no H1 that shares its identifier with another
heading — or one of two siblings sharing text all the way up — has no
selection of its own, and its content is read through its span.
A matched heading is returned with every heading nested inside its
section, each carrying its own level (AC-SPECSTR-003.4). `--emit` writes
the content the one matching heading stands over instead, read back
through the recorded span and written as itself with no envelope
(AC-SPECSTR-003.5); its span already covers the headings nested in it,
which are carried but not counted as matches. A raw payload has no place
for a delimiter, so a selection matching more than one heading is refused
under `--emit` with exit `2`, nothing on standard output, and each match
named on standard error by its `ref` (AC-SPECSTR-003.7):
under `Select one of:` those whose selection, run back through the same
selection, matches that heading alone, and listed apart those no
selection names alone, for which the query without `--emit` returns the
span. A selection matching nothing renders the empty
collection and exits zero (AC-SPECSTR-003.6); under `--emit` it is
refused instead, exit `2` with nothing on standard output and a message
saying whether the document is undeclared or the heading absent
(AC-SPECSTR-003.13). Members at a status the profile registers as no longer in
effect are omitted, from the answer and from the `--emit` refusal's list, unless
`--all` is given (AC-SPECSTR-003.14); a selection naming such a document by slug
or path still answers (AC-SPECSTR-003.15). A selection with a document
half also carries `declared` before `matches`, false when no corpus
member is that document, so a missing heading reads apart from a missing
document (AC-SPECSTR-003.8); a selection with none carries no such key. Like the queries beside it,
it opens the corpus warm and holds no domain logic: the heading tree and
the identifier derivation are Markdown Structure's.

**`headings` is the first command written to [[cli-output-standards]].**
Its answer renders as YAML by default and as JSON under `--json`, both
carrying the same values in the same order. The YAML rendering lives in
`cli/cmd`, in `renderYAML`, and resolves the marshalled struct's fields
the way `encoding/json` resolves them — the same `json` tags, embedded
structs promoted the same way, and the same rule deciding which of two
fields sharing a name is visible — so the two renderings carry the same
keys in the same order and neither can quietly gain or lose a value the
other has. One difference between them is required rather than
accidental: an empty or nil sequence renders `[]` and an empty or nil
mapping `{}` per [[cli-output-standards]], where JSON writes `null` for a
nil one. `cli/cmd/yaml_promotion_test.go` holds the renderer to this by
rendering a table of struct shapes both ways and comparing the key
sequences, and asserts the collection difference as the deliberate one it
is. It is not a general YAML library: block style, two-space indent, `[]`
and `{}` for empty collections, block scalars for multi-line values
trimmed at the edges, plain scalars where they round-trip and double
quotes where they do not, and no comments, anchors, aliases, tags or
document markers.
Every other command on this surface still prints what it printed before;
migrating them is per command, and belongs to whichever feature next
touches each one.

`list` is the second, and is a listing over the frontmatter every member
already declares: one entry per corpus member carrying its `path`,
`slug`, `type`, `status`, `title` and `description`, in that order
(AC-SPECLIST-001.2), under the single key `documents`. The whole of it is
one `corpus.Corpus.DocumentSummaries` call — the summary shape, the entry
order and both filters are Document Core's, and this component decides
none of them. It opens the corpus warm, as `check` and the queries beside
it do: every value it reports is already in the index, so a listing needs
no walk.

`--type` names the document types to return, repeatable and
comma-separated from one `StringSliceVar` exactly as `check`'s `--only`
is (AC-SPECLIST-002.2). Per-type boolean flags would not reach a type a
custom profile registers, which is why the surface takes values rather
than switches. `--all` asks for summaries at every status; without it the
ones the active profile registers as no longer in effect are omitted
(AC-SPECLIST-003.2), because a superseded document read as current
misleads an agent rather than merely costing it tokens. `--json` selects
the JSON rendering, carrying the same values in the same order as the
YAML one (AC-SPECLIST-004.1, AC-SPECLIST-004.2).

An empty answer is an answer: a filter matching nothing, and a corpus
holding no members, both render `documents: []` and exit zero
(AC-SPECLIST-001.7). A `--type` value the active profile does not
register is an invocation error, which `DocumentSummaries` reports rather
than answering; the report — the value and the registered type names —
reaches standard error and the run exits 2 (AC-SPECLIST-002.4).
That is the deliberate departure from [[cli-output-standards]]' empty-answer
rule, and it is narrow: the rule is about the world, where a corpus with
no contracts is a fact, whereas an unregistered type is a malformed
question and an empty collection would make a typo indistinguishable from
an empty category. Nothing else reaches the output — no count, no heading
naming the command, no trailing summary (AC-SPECLIST-004.5); the exit
status carries the outcome. Two runs over an unchanged corpus emit the
same bytes, because the order is the index's own (AC-SPECLIST-004.3), and
a title carrying a colon and a space comes back double-quoted because
`renderYAML` already quotes what a plain scalar would not round-trip
(AC-SPECLIST-004.4).

`c4` emits the corpus's C4 model and is wired the same way: flags in, one
`corpus.Corpus.C4Model` call, one rendering out. It takes no positional
argument and two flags, `--root <slug>`, which narrows the model to that
element and its descendants, and `--json`. It opens the corpus warm, as
`list` does, passes `Config.Workspace` as the workspace name, and writes
the model as YAML by default and as JSON under `--json`, both carrying the
same values in the same order (AC-C4-004.1, AC-C4-004.2). The model
carries `schema_version` and the workspace name (AC-C4-004.3), and two
runs over an unchanged corpus emit the same bytes (AC-C4-004.4). It exits
zero whatever the model holds, excluded relationships and `c4/*` findings
included (AC-C4-004.7). A model narrowed by `--root` carries
`root_declared`: true when the slug names an element, and false when it
names none, the model then holding every collection empty — an answer
Document Core gives, exiting zero with nothing on standard error
(AC-C4-004.6). A whole-corpus model carries no `root_declared`.

A bare invocation or `--help` lists the subcommands with their one-line
synopses and, per subcommand, its own flags — cobra's default rendering
of the `Use`, `Short` and flag declarations above, so the listed surface
and the code that implements it cannot drift (AC-SPEC-011.5, and the
Change Guidance below). No hand-written help text exists anywhere in
this component.

Observable command-tree behavior is verified by the snapshot cases
under `cli/testdata/cases/`, replayed by `cli/cmd/snapshot_test.go`,
per [[testing-standards]]; `scripts/regenerate-cli-snapshots.sh`
regenerates their expected artifacts from the current implementation.

### Document Core

See [[module-document-core]] (internal). (Go module) — Corpus
discovery, frontmatter parsing and schema validation, markdown structure
and span extraction, reference graph construction, numbering, and the
index and finding models. Knows nothing about this framework's document
types.

### Specs Profile

See [[module-spec-profile]] (internal). (Go module) — This framework's
document vocabulary: one union schema over every document type, each
type's slug pattern and numbered-ness, and the convention defaults.
Registers itself into Document Core through the extension point that
module defines. Required sections, element record blocks and lifecycle
rules are designed in that blueprint and not yet registered.

**Command Surface → Specs Profile.** The command
surface activates exactly one profile per invocation and hands it the
configuration it resolved. Which profile is active is a configuration
decision, not a compile-time one.

**Command Surface → Document Core.** All indexing, checking and query
work is delegated. The command surface converts between the module's
finding and index models and the process's stdout and exit status.

**Specs Profile → Document Core.** The profile is a
consumer of Document Core's extension point: it registers document
types, schemas, patterns and rules, and never the other way round.
Document Core SHALL NOT import the profile.

## Provided Contracts

**Public** (consumable outside the system):

- [[contract-specs-configuration]] — the `specs.toml` format an adopting
  repository authors, and the profile-registration surface beneath it.

## Consumed Contracts

None. The CLI reads files and emits artifacts; it calls no service and
depends on no interface another team provides.

## Data

The CLI owns no persistent data. It reads the corpus and emits
artifacts, and any cache it keeps is derived and disposable.

- **Index cache** (on-disk, derived) — the built index, retained so a
  subsequent query answers without re-reading the corpus. This is the
  "warm index" FEAT-SPEC-009.1 and FEAT-SPEC-009.2 are measured against.
  - WHEN any input the index was built from has changed, the CLI SHALL
    rebuild rather than serve a stale answer.
  - The CLI SHALL produce identical output whether or not a cache is
    present, so deleting the cache is always safe.

**Built** — there is no separate cache. The emitted index *is* the cache:
`index` writes `specs.json` into the configured spec root, and `check`,
`slug-exists`, `next-id` and `refs` open that same artifact warm from
that same default path. One file, one shape, one code path, so "delete
the cache" and "delete the artifact" are the same act and neither can
change an answer. An `index --out` naming somewhere else writes an
artifact for a consumer rather than a cache for the reading commands,
which simply find nothing at the default path and rebuild.

The reading commands **also write it when they had to build it**, which
is what makes AC-SPEC-010.5's own example — "run twice in a row reads
`specs.json` the second time" — true of a workflow that never invokes
`index`. Until they did, only `index` ever created the artifact, and
since nothing instructs an author to run it, every check and query in
practice rebuilt from cold and the warm path the FEAT-SPEC-009 budgets
are stated against went unreached (`agent-plugins-9ur`). That write is
best-effort and its failure is discarded: a read-only checkout or a race
with a concurrent writer makes the answer no less correct, and a query
that failed because it could not refresh a cache would be a worse tool
than a slow one. A failed write changes nothing observable — not the
exit status, not standard output, not standard error; the index only
stays cold (AC-SPEC-010.7). AC-SPEC-010.4's atomic replace keeps a concurrent
reader from seeing a partial file.

**Staleness is decided by a content hash of the inputs, never by a
modification time.** Every emitted index carries an `input_hash` folding
each candidate file's repository-relative path and content digest,
together with the inputs that change the index while the files stand
still: the workspace identifier, the active profile's schema and per-type
table, and the schema version. A persisted index is served warm only when
it records the corpus's current hash; every other condition — an absent
file, an unreadable one, one of another `schema_version`, one carrying no
hash — rebuilds. The reason for hashing rather than stat-ing is the
environment the latency budget is claimed in: a fresh CI checkout gives
every file a new mtime and identical bytes, so an mtime-keyed cache would
be cold on every CI run and FEAT-SPEC-009.1 would never hold there.
Absolute paths are deliberately excluded from the hash, so the same
content checked out at another location is still warm.

Invalidation is whole-index, not per-entry: the guarantee above says
"rebuild" rather than "rebuild the affected entries" because a rebuild
costs one corpus walk — 70ms cold over this repository's own 26
documents, against 26ms warm — and a partial invalidation would buy a
fraction of that while adding a way to be subtly wrong.

The correctness argument rests on one property — the fingerprint and the
build consume the same walk over the same bytes — which is what makes
"this index was built from exactly these inputs" a checkable claim rather
than a hopeful one.

## Infrastructure and Deployment

This container is standalone: there is no system blueprint, so this
section is authoritative for its deployment.

- **Exclusive infrastructure:** none. The CLI has no database, no queue,
  and no network dependency at runtime.
- **Runtime dependency:** `criterion --changed` alone needs the `git`
  executable on `PATH`, because it reads its base revision through git
  ([[adr-031]]). Its absence is an environment error (exit 3) for that
  command; no other command runs git.
- **Deployment:** the CLI is distributed as a per-platform binary.
  FEAT-SPEC-009.4 through FEAT-SPEC-009.6 govern the shape: it runs with
  no language runtime installed beyond itself, it retrieves no
  executable content at install time, and it is installable from at
  least one package registry that supports on-premises mirroring.
  Cross-compilation produces the platform matrix from a single build.
- **Build:** `CGO_ENABLED=0`, so each artifact is one statically linked
  file with no libc to resolve at run time (AC-SPEC-009.4), and
  `-trimpath -ldflags="-s -w"`, so the bytes do not carry the builder's
  absolute paths. Six targets — `linux`, `darwin` and `windows`, each on
  `amd64` and `arm64` — are cross-compiled on one Linux runner from one
  checkout. `.github/workflows/cli.yml` builds the whole matrix on every
  pull request, so a target that stops cross-compiling fails on the
  change that broke it rather than at release time.
- **Distribution — two channels.** The Go module proxy is the registry
  AC-SPEC-009.6 is satisfied by: it is mirrorable on premises (GOPROXY,
  Athens, Artifactory), and `go install …/cli/specs@<tag>` needs no
  release artifact at all. For a machine with no Go toolchain, tagged
  releases carry one archive per target plus `SHA256SUMS`; unpacking one
  runs no install-time code, which is what AC-SPEC-009.5 turns on.
  Nothing here is signed yet, and the binaries carry no version string —
  both are open, and both belong to the release-and-distribution ADR
  rather than to this blueprint.
- **The entry point is `cli/specs/`, not the module root.** Go names an
  installed binary after the directory holding its main package, so the
  path is what makes `go install` produce `specs` rather than `cli`. For
  the same reason release tags are `cli/vX.Y.Z`: the module lives in a
  subdirectory, and the proxy resolves a subdirectory module's versions
  only from tags carrying that prefix.
- **Cross-platform reproducibility (AC-SPEC-009.7) is a property of the
  read path, not of the build.** The CLI normalizes CRLF to LF at the
  single point where it reads a candidate file, so a Windows checkout
  under `core.autocrlf` cannot move a span's byte offsets or the
  content hash the warm-cache path compares. The consequence is
  deliberate: a reported byte offset indexes the normalized text rather
  than the on-disk file where the two differ, while line and column —
  what a human or an editor acts on — are unaffected.

## Cross-Cutting Standards

This container SHALL comply with:

- [[testing-standards]] — unit tests per declared component,
  harness-independent snapshot cases for observable behavior, and the
  coverage obligations tying both back to acceptance criteria.

The repository's other standards document,
`specs/standards/skill-authoring-standards.md`, governs skill authoring and
does not apply to a compiled binary.

## Change Guidance

- WHEN a capability is added to the command surface, that capability
  SHALL be implemented in [[module-document-core]] or
  [[module-spec-profile]] and invoked from the command surface, never
  implemented in the command surface itself.
- [[module-document-core]] SHALL NOT depend on [[module-spec-profile]]
  or on any framework-specific vocabulary. The dependency runs one way,
  and a violation is what would make the core unusable for the Obsidian
  vault case it is designed to serve.
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
  reaching into a module additionally run that module's gates. CI runs
  the full suite.

## Key Decisions

- [[adr-015]] — Go is the implementation language; the measured basis
  and the rejected alternatives are recorded there.
- [[adr-021]] — Gherkin is the notation for features and acceptance
  criteria. Command Surface queries the feature and criterion records
  Document Core extracts, emits a feature back out as a standalone
  `.feature` file, and reports which criteria changed against a git
  base; the parsing and the rendering are Document Core's, and the tag
  vocabulary is [[module-spec-profile]]'s.
- [[adr-031]] — `criterion --changed` reads its base revision through
  the `git` executable, building the spec root as the base commit holds
  it with the same `corpus.Build` the working tree uses, which is why
  that one command needs `git` on `PATH`.

## Open Questions

- [ ] The emitted index is a published artifact with a versioned schema
      (AC-SPEC-002.2, AC-SPEC-002.3) consumed by the Architecture
      Catalogue. That makes it a contract by this framework's own filing
      rule, but no contract document exists yet. It is deferred to the
      index-schema ADR, which also has to settle whether references are
      modelled as object properties or as a separate edge collection
      (`agent-plugins-dgn`). Decide whether the contract is authored
      alongside that ADR or before it.

      **Decided, in part.** Promotion to a contract document, with a
      formal versioning policy, is **deferred to `agent-plugins-h6c`**
      and waits on a consumer or catalogue that actually exists. Until
      one does, the generated JSON Schema for `SpecIndex` — emitted from
      `schemas/index.strata` with every other model — is the index's
      machine-readable specification, and it is the only consumer-facing
      statement of the shape.

      What is pinned meanwhile: `schema_version` is `2.1`, and the shape
      is **additive** under it. Later features attach headings, extracted
      record blocks and references to the same entry without changing the
      version; changing what is already there — a property's meaning, its
      type, its removal — bumps `schema_version`, the Strata source and
      the emitted index together, per the Change Guidance above. One
      exception: a property that no emitted index has ever carried —
      because the shape holding it has never been emitted with an
      instance — MAY be changed or removed without a bump, since no
      artifact a 1.0 consumer has read contains it. `RecordBlock.content`
      was removed under this exception: `record_blocks` had been emitted
      empty for every entry. That
      rule is what makes a one-producer artifact safe to ship before its
      contract is written: a consumer built against 1.0 keeps working
      until the version says otherwise. It is deliberately weaker than a
      contract — nothing here says how long 1.0 is supported, or what a
      consumer is owed on a bump — and closing that gap is exactly what
      `agent-plugins-h6c` is for.
- [ ] Whether the CLI ships a `serve` mode holding the index in memory
      across queries. Not required — FEAT-SPEC-009.1 states a latency
      target, not a gate, and the warm path holds it on a quiet machine —
      but it is the natural home for filesystem watching, which is listed
      as a follow-on capability.

## References

- `c4-guide.md` — C4 mapping rules and diagram conventions
- [[frd-spec-index]] — the requirements this container realizes
- [[module-document-core]] — the general engine it composes
- [[module-spec-profile]] — the framework vocabulary it composes
