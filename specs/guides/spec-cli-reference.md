---
title: Spec CLI Reference
description: The specs binary's command surface — what each command
  decides, the rules check applies, what it does not yet check, exit
  statuses, and how to install it — for skills that delegate corpus
  queries and conformance checks
slug: spec-cli-reference
type: guide
status: active
version: "0.3"
---

# Spec CLI Reference

`specs` reads a specification corpus and answers questions about it that
would otherwise mean opening every document. Prefer it over reading the
corpus yourself for anything it decides: it is faster, and its answer is
verifiable by whoever reviews the work, where yours is an assertion.

## Contents

- Prerequisites — what must exist before any command works
- Output — the one shape every answer takes
- Commands — the eleven commands and the question each answers
- Reference checks — the six rules `check` applies to `[[slug]]` links
  and reference-bearing frontmatter, and the severity each carries
- C4 diagram checks — the ten `c4/*` rules over architecture blueprint
  diagrams
- Gherkin checks — the fourteen `gherkin/*` rules over acceptance
  criteria
- Metadata checks — the frontmatter, slug and empty-corpus rules
- Exit statuses — how to read a run's result
- What it does not check — the gap you still cover by hand
- When it is not installed

## Prerequisites

Two things, both cheap to confirm:

- **The binary.** `command -v specs`. If it is missing, see the last
  section — do not silently fall back to reading documents.
- **A `specs.toml`**, declaring `[workspace].root` (the corpus
  directory), `[workspace].name`, and `[profile].name`. Every command
  resolves the corpus through it. The CLI reads the nearest `specs.toml`
  in the working directory or a directory above it, searching up to the
  repository root — or up to `/` when the working directory is not in a
  repository — so a command run from a subdirectory finds the same file.
  `--config <file>` names one instead and overrides the search. Its
  absence is not a default to infer: the CLI names every directory it
  searched and exits 3, the message beginning `no specs.toml found in`
  and ending `pass --config to name one`.

`[workspace].root` is also the answer to "where do documents go" — read
it rather than assuming `docs/`.

## Output

Every answer has the same shape, so the parsing you write for one
command carries to the next:

- **YAML by default, JSON under `--json`**, holding the same values in
  the same order. Every command that answers a question takes `--json`.
- **The root is always a mapping.** A collection sits under a named key
  (`documents`, `matches`, `features`, `criteria`, `references`) as a
  list of mappings, never as formatted lines.
- **A lookup by one identifier returns a list**, the same list a
  broader selection would — `specs feature FEAT-SPEC-010` answers
  `features:` holding one record.
- **A document payload is written as itself**, not wrapped: the section
  `headings --emit` reads, the Gherkin `feature --emit` writes, and the
  rewritten document `links` prints.
- **`index` writes an artifact, not an answer**: always JSON, to a file
  by default.

**Not found is an answer; malformed is a refusal.** A subject the corpus
does not declare — an identifier, a feature code, a slug, a document
named in a `headings` selection — is answered with `declared: false` and
an empty collection, and exits 0. A value outside a set the profile
registers — a document type, a rule, a tag qualifier — is a malformed
question: the command names the value and every registered one, writes
nothing to standard output, and exits 2.

```
$ specs slug-exists adr-999-nothing
slug: adr-999-nothing
declared: false
path: null
$ specs criterion --tag nonsense        # exit 2
"nonsense" is not a registered qualifier; registered: negative, manual, level-unit, level-integration, level-e2e
```

## Commands

| Command | Answers |
|---------|---------|
| `specs check [path]` | Does the corpus — or one document — conform? |
| `specs index [--out <path>\|-]` | The whole corpus's frontmatter, structure and reference graph, as JSON |
| `specs next-id <type>` | What number does the next document of this type take? |
| `specs slug-exists <slug>` | Is this slug already taken, and by which file? |
| `specs refs <slug>[#<anchor>]` | Who references this document or heading — or what does it reference? |
| `specs list [--type <type>] [--all]` | What does this corpus hold, and what is each document about? |
| `specs c4 [--root <slug>]` | What systems, containers, components and modules does the corpus document, and what depends on what? |
| `specs headings [<document>#<heading>]` | Where are this document's headings, and what does one section say? |
| `specs feature [<id>]` | What does this feature require, with its criteria? |
| `specs criterion [<id>]` | What does this acceptance criterion say — and which changed? |
| `specs links <path>` | This document with its wikilinks rewritten as markdown links |

**`specs --version`** answers which build is running — what a bug
report needs. It writes a YAML mapping of `version`, `revision`, `time`
and `modified` and exits 0, with no subcommand and nothing on standard
error; it has no `--json` form. `version` is the version stamped into a
release build, else the module version (`(devel)` for a local build);
`revision`, `time` and `modified` are the binary's version-control build
settings, each `null` where the build recorded none (a `go install` of a
tagged module records none).

```
$ specs --version
version: (devel)
revision: 3f1c0a9e2b7d4c5a8e6f1b0d9c2a7e4f5b6a8c1d
time: 2026-10-08T20:51:03Z
modified: false
```

**`specs check`** validates every document's frontmatter against the
active profile's schema and slug patterns, reports duplicate slugs, and
applies the reference, C4, Gherkin and metadata rules in the sections
below. Its answer leads with a `summary` — the counts of `errors`,
`warnings` and `unverified` findings, all three always present — and then
the `findings` (each with its rule, severity, path, span and message), so
the head of a failing run's output is its findings. `--json` writes the
same values in the same order. `--strict` makes warning and unverified
findings fail too. `--verbose` adds `checks` after the findings: every
rule the run was asked for, each `passed`, `failed` or `unverified`. The
answer's shape never changes the exit status. `--only <rule>` narrows to
named rules — repeatable or comma-separated — and the summary then counts
only their findings, and `checks` lists only those:

```
$ specs check
summary:
  errors: 0
  warnings: 0
  unverified: 0
findings: []

$ specs check --verbose --only reference/unresolved
summary:
  errors: 0
  warnings: 0
  unverified: 0
findings: []
checks:
  - rule: reference/unresolved
    outcome: passed
```

**`specs check <path>` reports the findings located in that file, and
the references elsewhere that your change to it broke.** It resolves
the document's references against the whole corpus and reports every
finding that sits in it — an unresolved reference or anchor, a retired
target, a type mismatch, a slug or identifier it shares with another
document — at its own line. It also reports, each at its line in the
citing document, every reference elsewhere that cites a slug the
document declared at git HEAD and no longer declares — an uncommitted
rename — and every anchor elsewhere into the document that no longer
names a heading of it. Nothing else located in another document is
reported: an unrelated broken link elsewhere stays out. Outside a git
work tree (or without git) there is no HEAD to compare with, so the
check writes one `note:` line to standard error saying the references
elsewhere were not checked and answers with the document's own findings.
A changed `type` or a retirement is not reported elsewhere, because the
citing reference still resolves. Before you rename a slug, retire a
document or change its `type`, run `specs refs <slug>`: it names every
document citing it, which is the blast radius of the change. After it, a
whole-corpus `specs check` reports what is broken anywhere.

**A path that names no corpus member is a refusal, not a pass.** Where
the file exists but was excluded, the message names the reason
(`no-frontmatter`, `unknown-type`):

```
$ specs check docs/typo.md              # exit 2, nothing on stdout
docs/typo.md is not a corpus member: no document at that path was discovered under the spec root
```

**`specs index`** writes the corpus index — every member's frontmatter,
headings with their section spans, record blocks, and the reference
graph — to `specs.json` inside the spec root. `--out <path>` writes it
elsewhere; `--out -` sends it to standard output. Read it when you need
many documents' metadata at once instead of opening files. Indexing
never fails on a broken reference: an unresolved reference is recorded
in the index and reported by `check`, so a repository whose links are
broken still contributes its graph.

Every other command except `links` refreshes that same file when it had
to rebuild, so the second of two runs is served from it. You never need
to run `specs index` to make the others fast — run it when you want the
artifact itself. It is derived and disposable: deleting it changes no
answer, and it belongs in `.gitignore`. Writing it is best-effort: in a
read-only checkout every command answers and exits exactly as it would
in a writable one, and only stays cold.

**`specs next-id <type>`** answers one greater than the highest number in
use for a numbered type, across the whole corpus, and the slug that
number forms. This is the only correct way to pick an ADR number;
scanning filenames misses a document whose filename and slug disagree. A
type the profile does not number is refused (exit 2).

```
$ specs next-id adr
number: 32
slug: adr-032
```

**`specs slug-exists <slug>`** answers whether any document declares that
slug, and where. Case-insensitive, because a case-insensitive filesystem
cannot hold two files that differ only in case. The slug is written as the
document declares it where one does, so `ADR-012` answers `adr-012`; where no
document declares it, the slug is written as given. It exits 0 either way.
Run it before claiming a slug.

```
$ specs slug-exists adr-012
slug: adr-012
declared: true
path: specs/adrs/adr-012-reference-conventions.md
```

**`specs refs <slug>`** answers the blast-radius question. By default it
lists every corpus member whose body or reference-bearing frontmatter
names that slug — run it *before* retiring or renaming a document, not
after. The answer is compact: one entry per referring document, each
reference reduced to its line and text. `--full` returns every field of
every reference instead — its span, its origin (body or which
frontmatter property) and what it resolved to — so a rewrite opens only
the lines that carry the reference. `--outbound` reverses the direction
and lists what that document itself references, which is what to read
before editing one. Its compact answer groups by cited slug instead of by
document: one entry per slug, carrying the `path` to read it at (`null`
when it resolves to nothing, with `resolution.state`) and the `lines`
that cite it.

```
$ specs refs adr-016
slug: adr-016
direction: inbound
declared: true
references:
  - path: specs/adrs/adr-018-references-as-index-edges.md
    references:
      - line: 31
        text: adr-016
```

**`specs refs '<slug>#<Heading>'`** narrows to the references whose
anchor resolves to that heading of the document — the blast radius of
renaming the heading. A heading is matched on its identifier, so letter
case and punctuation do not decide a match; a reference whose anchor
matches more than one heading is reported for each, marked as such. The
answer then carries `anchor` and `anchor_declared`. `--outbound` takes
no anchor.

```
$ specs refs 'container-spec-cli#Command Surface'
slug: container-spec-cli
anchor: Command Surface
direction: inbound
declared: true
anchor_declared: true
references:
  - path: specs/features/gherkin-adoption/gherkin-adoption-blueprint.md
    references:
      - line: 537
        text: container-spec-cli#Command Surface
```

Every form exits zero when the answer is empty. "No corpus member
references this slug" (`declared: true`, `references: []`) is reported
distinctly from "no corpus member declares this slug" (`declared:
false`) — the first means the document is unreferenced, the second means
you asked about a slug that does not exist. Read the output, not the
status.

**`specs list`** answers "what does this corpus hold" without opening a
document or the index. It returns one entry per corpus member, each
carrying — in this order — the member's `path`, `slug`, `type`, `status`,
`title` and `description`. The path is repository-relative with forward
slashes, so the next step is a path you were handed rather than a slug
you have to resolve. The answer is a mapping under a single `documents`
key, with no count, no heading naming the command and no trailing
summary around it.

Two filters narrow it, and they compose:

- **`--type`** names the document types to return. It is repeatable and
  accepts a comma-separated value, so `--type adr --type frd` and
  `--type adr,frd` ask the same question. A value the active profile does
  not register is a malformed question rather than an empty category: the
  command names the unregistered value and every registered type name,
  and exits 2.
- **`--all`** asks for every status. Without it, members at a retired
  status are omitted — a retired document is not in effect, and an agent
  orienting over the corpus that reads one has been actively misled
  rather than merely charged for tokens. Every other status is reported
  by default, and each entry carries its own `status`.

A filter matching no member is an empty collection and exit 0, not a
failure.

Two kinds of file are treated differently, and the difference is what the
index recorded. A member the index recorded with a **parse error** is
reported as an entry naming its path and the failure, carrying no slug,
type, status, title or description — it is a member, so omitting it would
make the listing disagree with the corpus, and a null value would assert
the document declares none, which is not known, so the fields are absent
rather than null. Under a `--type` filter it is not reported at all,
because it declares no type for a filter to match. A file the index
recorded as an **exclusion** is never reported: it is not a member, and
whether it should be one is `specs check`'s judgement rather than this
command's.

Entries come back in the index's own order, fixed by repository-relative
path, so two runs against an unchanged corpus produce byte-identical
output.

`specs list` is what makes `description` a consumed interface — see
`shared/frontmatter-standards.md`, which tells an author to write it to
be triaged from for exactly this reason.

**`specs c4`** emits the corpus's C4 model: the architecture the corpus
declares, read from blueprint frontmatter, roster record blocks and the
`mermaid` C4 diagrams in blueprints. It takes no positional argument and
two flags:

- **`--root <slug>`** narrows the model to the element with that slug and
  its descendants. The root's ancestors appear only as the path down to
  it, and only relationships with both endpoints inside the subtree are
  kept. The model then carries `root_declared`; a slug naming no element
  is answered with `root_declared: false` and every collection empty, and
  exits 0.
- **`--json`** writes JSON instead, holding the same values in the same
  order.

```
$ specs c4 --root nothing-here
schema_version: "0.1"
workspace: agent-plugins
root_declared: false
software_systems: []
orphaned_containers: []
orphaned_components: []
modules: []
relationships: []
excluded_relationships: []
```

The model is a mapping whose top-level keys are `schema_version`,
`workspace` (the name in `specs.toml`), and these collections, each
ordered by slug so that two runs over an unchanged corpus produce
byte-identical output:

- `software_systems` — systems, each holding its containers, and each
  container its components and the slugs of its modules.
- `orphaned_containers` — containers whose blueprint declares no `system`.
- `orphaned_components` — components no container or module declares.
- `modules` — every module once, whatever the number of its hosts.
- `relationships` — the dependency edges drawn in blueprint diagrams
  between two elements.
- `excluded_relationships` — every `Rel` or `BiRel` statement the model
  does not admit, with its `path`, `span`, `text` and a `reason`:
  `external_endpoint` (a `Person` or `_Ext` node), `passive_store` (a
  `Db` or `Queue` naming no element), `undeclared_node` (a label
  matching no element in scope), `ambiguous_node` (a label matching more
  than one) or `bidirectional` (a `BiRel`, which states no single
  dependency).

The command exits 0 whatever the model holds: excluded relationships are
facts recorded in the artifact, and judging the diagrams is `specs check`'s
job (`c4/*` findings). The JSON output validates against the `C4Model`
definition in the generated model JSON Schema.

**`specs headings`** answers where a document's headings are, so you can
read one section without opening the whole document. The selection is
one operand, `<document>#<heading>`, either half optional: `<document>`
alone selects every heading in it, `#<heading>` alone selects that
heading in every corpus member, and no operand selects every heading in
the corpus. A document half holding a path separator or ending in `.md`
is a repository-relative path; otherwise it is a slug. A heading is
matched on its identifier, so letter case and punctuation do not decide
a match. The heading half may be a path of headings joined by `#`,
outermost first, each nested anywhere inside the one before it. Each
match is compact by default: its `anchor`, its `ref`, its `level` and the
`start_line` its heading stands on. `--full` adds the file `path`, `text`,
`identifier` and the `section` span of the content beneath it. A match is
followed by every heading nested inside its section.

```
$ specs headings 'module-document-core#Components#References' --full
declared: true
matches:
  - path: specs/modules/document-core/document-core-module.md
    text: References
    identifier: references
    anchor: document-core#components#references
    ref: module-document-core#document-core#components#references
    level: 3
    section:
      offset: 55158
      length: 13357
      start_line: 943
```

Without `--full` the same match is `anchor`, `ref`, `level: 3` and the
`start_line` of the heading line itself.

**`--emit`** writes the section's content as itself instead — the
cheapest way to read one part of a long document. It requires the
selection to match exactly one heading. A selection matching several is
refused (exit 2), naming each match by its `ref` that selects it
alone, which you paste back (`ref` is the declared slug, `#`, and the
`anchor`; a member declaring no slug is named by its path in place of
the slug):

```
$ specs headings 'module-document-core#References' --emit     # exit 2
"module-document-core#References" matches 2 headings; --emit needs exactly one. Select one of:
    module-document-core#document-core#components#references
    module-document-core#document-core#references
or drop --emit to list the matches
```

A document the corpus does not declare answers `declared: false`; a
heading matching nothing in a declared document answers `declared: true`
with no matches. Both exit 0. Under `--emit`, where raw text cannot say
either, both are refused (exit 2, nothing on standard output) with a
message naming the selection and whether the document is undeclared or
the heading absent.

Like `specs list`, `headings` omits members at a status no longer in
effect unless `--all` is given, from the answer and from the `--emit`
refusal's list alike. A selection that names such a document by slug or
path still answers with its headings.

**`specs feature`** returns features from the Gherkin in the corpus. The
answer is compact by default — each feature's `id`, `code`, `name`, `path`
and `criteria_count`; `--full` returns the whole record, description and
criteria included. `feature <id>` returns that one, `--code <CODE>`
every feature grouped under that code, and no argument every feature.
Identifiers and codes match regardless of letter case. The answer is a
`features` list in every case, preceded by `declared` where the
arguments name a subject. `feature <id> --emit` writes the feature as a
standalone Gherkin `.feature` file — tags, description and criteria — to
standard output, or to the path `--out` names; an unknown id there is
refused (exit 2), since there is nothing to write, and `--full` there is
refused too.

```
$ specs feature FEAT-SPEC-010 --emit
@frd-spec-index @FEAT-SPEC-010
Feature: Index Output and Persistence
```

**`specs criterion`** returns acceptance criteria. `criterion <id>`
returns that one, `--of <feature-id>` every criterion beneath a feature,
and `--tag <qualifier>` every criterion carrying a qualifier, written
with or without the leading `@`. The answer is a `criteria` list in
every case, preceded by `declared` given an id or `--of`.

```
$ specs criterion ac-spec-010.7
declared: true
criteria:
  - id: AC-SPEC-010.7
    path: specs/features/spec-index/spec-index-frd.md
    feature_id: FEAT-SPEC-010
```

**`criterion --changed`** lists every acceptance criterion whose content
differs from the corpus at a git base, classifying each as `changed`,
`added` or `removed`. The base is the merge-base of `HEAD` and the
repository's default branch, or of `HEAD` and the branch, tag or commit
`--since` names, and the current side includes uncommitted edits. The
default branch is the first that exists of `origin/HEAD`'s target,
`origin/main`, `origin/master`, `main` and `master`, so a stale local
`main` never wins over `origin/main`; when a local branch of the same
name is behind the one chosen, standard error carries `warning: local
main (<short>) is behind origin/main (<short>); comparing against
origin/main` and standard output is unchanged. The
answer holds `base` (the revision and the merge-base commit) and then
`criteria`, each carrying `id`, `name` (omitted where the criterion has
none; the base's name for a removed one), `classification` and `path`.
**`--full`** adds `record` (the criterion exactly as `specs criterion
<id>` lists it) for an added or changed criterion and `base_record` (the
same taken from the base) for a changed or removed one; `--full` without
`--changed` exits 2. It exits 0 whatever it lists; `--exit-code` makes a
non-empty list exit 1, for a gate. It needs `git` on `PATH`.

```
$ specs criterion --changed
base:
  revision: main
  commit: 88b49ac97d2d2374d9bea22ca0fc14fadadf32b9
criteria:
  - id: AC-C4-004.6
    name: Every node is declared before an edge cites it
    classification: changed
    path: specs/features/spec-c4/spec-c4-frd.md
```

**`specs links <path>`** prints the document with each `[[slug]]`
rewritten as an ordinary markdown link, for rendering outside the
corpus, and the count of rewritten links on standard error. With no
map, each link targets the cited document's path relative to the one
being rewritten. **`--map <file>`** names the targets instead: a YAML
mapping from each corpus slug to a record declaring `path`, such as
`adr-012: {path: ../adrs/adr-012.md}`. A record's other properties are
ignored, and a map written in JSON is accepted. A map of any other shape
is refused (exit 2) with the expected shape in the message. A cited slug
the map omits raises `links/unmapped` (error) and no document is
printed; a slug matching no corpus member raises `links/unresolved`
(warning) and is left as written. A wikilink inside code — an inline
span, an indented block, or any fenced block, a ```` ```gherkin ````
fence included — is left as written and not counted. `links` creates,
modifies and deletes no file.

```
$ specs links specs/adrs/adr-018-references-as-index-edges.md >out.md   # stderr:
no findings
13 link(s) rewritten
```

## Reference checks

`specs check` applies six rules to references. All six resolve a
reference against the target's frontmatter `slug`, never a file path,
and without regard to letter case.

| Rule | Severity | Reports |
|------|----------|---------|
| `reference/unresolved` | error | A `[[slug]]` or reference-bearing frontmatter value matching no corpus member |
| `reference/unresolved-anchor` | error; unverified | An anchor matching no heading of the cited document (error). An anchor into a document whose headings are not recorded — it could not look (unverified) |
| `reference/ambiguous-anchor` | error | An anchor whose identifier is the identifier of more than one heading of the cited document |
| `reference/retired-target` | warning | A live document referencing one at `status: retired`, quoting its `superseded_by` when it has one |
| `reference/pairing-incomplete` | error | One half of a paired relationship whose target does not declare the counterpart |
| `reference/type-mismatch` | error | A typed reference whose target declares a different `type` — a feature blueprint's `frd` naming something that is not an FRD |

Four things about reference resolution decide whether you trust it:

- **A backticked `[[slug]]` is not a reference.** Text inside a fenced
  code block or an inline code span is an illustration of the notation,
  and contributes nothing. Neither does a `[[slug]]` written in a
  frontmatter *value* — frontmatter contributes references only through
  the properties the profile registers as reference-bearing. A
  ```` ```gherkin ```` fence is the exception: its wikilinks are
  references, checked and indexed like prose.
- **An anchor is resolved against the cited document's headings.**
  `[[container-x#Some Heading]]` resolves on `container-x`, then its
  anchor is matched on heading identifier — so case and punctuation do
  not matter — or on a `#`-joined heading path. A miss raises
  `reference/unresolved-anchor`; more than one match raises
  `reference/ambiguous-anchor`. A bare `[[#Heading]]` resolves against
  the document it is written in.
- **A target that exists but is not a corpus member reports as such** —
  "target exists but is not a corpus member: excluded: no-frontmatter"
  rather than as a slug matching nothing, because the reader's next step
  differs.
- **A target whose frontmatter does not parse raises nothing against the
  citing document.** The parse failure is reported once, against the
  document that has it. A slug reaches such a document by its file name,
  or by a prefix of it followed by `-`: `[[adr-012]]` reaches
  `adr-012-reference-conventions.md`.

`reference/retired-target` is the only reference rule that emits
`warning`. A reference to a retired document still resolves — the reader
reaches a real document that says it has been replaced — so it does not
fail a build unless you pass `--strict`. It is also the one finding that
tells an author what to cite instead.

Which frontmatter properties are reference-bearing is **derived from the
profile's schema**, not a hand-kept list: every property typed as a
document slug is one, except `slug` itself, which names its own document.
`superseded_by`, `sources`, `frd`, `system`, `modules`, `provider` and
`contracts_consumed` are all reference-bearing
for that reason, and a slug-typed property added to the vocabulary
becomes one the moment it is declared. This framework's own profile
registers **no** pairing, so `reference/pairing-incomplete` is inert
against its own corpus by design — it exists for adopting profiles whose
relationships are declared from both ends.

## C4 diagram checks

`specs check` reads the C4 diagrams in architecture blueprints (system,
container, module and component) and applies ten `c4/*` rules. Every
finding carries the span of the statement it concerns. Diagrams in other
document types are not checked. The statement subset the parser reads
and the reserved tags are defined in [[c4-standards]].

| Rule | Severity | Reports |
|------|----------|---------|
| `c4/parse` | error; unverified | A statement outside the parsed subset (error). A `mermaid` block whose diagram type is not read, including `C4Dynamic` and `C4Deployment` — it was not looked inside (unverified) |
| `c4/undeclared-node` | error | A non-external node that resolves to no element in the drawing document's scope and is not a passive store |
| `c4/ambiguous-node` | error | A node whose label matches the entity names of more than one element in scope, naming each |
| `c4/bidirectional` | error | A `BiRel` statement |
| `c4/duplicate-edge` | error | A second relationship statement in one diagram with the same source and target, after `Rel_Back` is normalised |
| `c4/drawing-conflict` | warning | One relationship drawn in two places with differing technology, description or tags, naming the other drawing |
| `c4/interaction-conflict` | error | A relationship tagged both `sync` and `async` |
| `c4/contract-unresolved` | error | A `contract:` tag whose slug no contract document declares |
| `c4/contract-mismatch` | warning | A `contract:` tag naming a contract the source element does not consume or the target element does not provide |
| `c4/tag-separator` | warning | A tags argument that joins tags with a comma instead of `+` |

Like every rule, each `c4/*` rule is narrowed to the file you name.
`c4/drawing-conflict` still compares that file's drawings
against every architecture blueprint's, but reports only the findings that
sit in the named file.

## Gherkin checks

`specs check` reads every acceptance criterion written in Gherkin — in a
```` ```gherkin ```` fence in a document, or in a standalone `.feature`
file under the spec root — and applies fourteen `gherkin/*` rules.

| Rule | Severity | Reports |
|------|----------|---------|
| `gherkin/parse` | error | A Gherkin source that could not be parsed |
| `gherkin/feature-untagged` | error | A Feature carrying no identifier tag |
| `gherkin/criterion-untagged` | error | A Scenario carrying no identifier tag |
| `gherkin/identifier-malformed` | error | An identifier tag that does not match the pattern for its kind |
| `gherkin/identifier-multiple` | error | A Feature or Scenario carrying more than one identifier tag |
| `gherkin/identifier-duplicate` | error | A feature or criterion identifier declared twice in the corpus |
| `gherkin/criterion-code-mismatch` | error | A criterion whose identifier's code and number are not its feature's |
| `gherkin/owner-missing` | error | A feature in a standalone `.feature` file carrying no owning-document tag |
| `gherkin/owner-mismatch` | error | A feature in a fenced block whose owning-document tag names a different document |
| `gherkin/owner-unresolved` | error | An owning-document tag naming no corpus member |
| `gherkin/criterion-empty` | error | A criterion with no steps |
| `gherkin/examples-partially-tagged` | error | A Scenario Outline whose Examples blocks carry an identifier on some and not others |
| `gherkin/scenario-prose` | warning | A Scenario description that reads as a normative statement — a probable untagged criterion |
| `gherkin/tag-unregistered` | warning | A tag the profile's vocabulary assigns no meaning, such as a misspelt `@negatve`, naming the registered qualifiers |

`gherkin/tag-unregistered` is a warning because a project's own tags
coexist with the vocabulary; only a misspelt one loses a meaning. Pass
`--strict` where an unknown tag should block.

## Metadata checks

| Rule | Severity | Reports |
|------|----------|---------|
| `frontmatter/parse` | error | A frontmatter block that could not be read as a document envelope — most often a `status` or other enumerated value outside its vocabulary, or a `created_at` that is not an RFC 3339 date-time — so nothing else examined the document |
| `frontmatter/schema` | error | A frontmatter value that violates its type's schema, naming the property, the value and the constraint — `/reviewer: "someone" does not satisfy additionalProperties` |
| `frontmatter/missing-property` | error | A required frontmatter property the document does not declare |
| `frontmatter/inline-comment` | warning | A YAML comment trailing a frontmatter value on the same line — `description: a # b` — located at the `#`: unquoted, the text after it is not part of the value. Quote the value to keep it |
| `slug/pattern-mismatch` | error | A slug that does not match its type's pattern |
| `slug/duplicate` | error | A slug declared by more than one document, naming the other |
| `corpus/empty` | error | A run whose spec root holds no corpus member at all, so a clean report would mean nothing was examined |

Fix a `frontmatter/parse` first: it means nothing else could examine the
document, so the other checks have not cleared it. The commonest cause
by far is a template placeholder (`visibility: <public | internal>`)
left unfilled.

## Exit statuses

Every command exits with one of five statuses, each naming who acts:

- **0 — success.** The command did what was asked: a query answered, a
  gate passed. For `check`, no error-severity findings (and, under
  `--strict`, none at all).
- **1 — gate failed.** A gate ran to completion and the corpus failed
  it: `check` found something, `links` met a citation it cannot rewrite,
  or `criterion --changed --exit-code` listed something. Fix the
  documents.
- **2 — invocation error.** The command could not run as invoked: an
  unknown command or flag, a missing or surplus argument, a value outside
  a registered set, an operand it cannot act on. Fix the command.
- **3 — environment error.** The invocation was correct but its
  surroundings prevent it: no `specs.toml`, an invalid configuration, an
  unknown profile, an unreadable spec root, an output path it cannot
  write. Fix the repository.
- **4 — internal error.** The CLI reached a state it does not handle,
  a panic included. File a bug.

On 2, 3 or 4 nothing is written to standard output and the reason goes
to standard error, so a caller parsing the output never parses half an
answer. On 1 the findings are the answer, on standard output. Retrying
is correct for none of them.

A run whose findings are all `warning` or `unverified` — such as
`reference/retired-target`, `gherkin/tag-unregistered`,
`c4/drawing-conflict`, or a diagram `c4/parse` did not look inside —
exits **0** without `--strict` and **1** with it. That is where the two
differ, so `--strict` is worth passing in CI where any of them should
block.

**Read a query's answer, not its status.** `slug-exists`, `refs`,
`headings`, `feature`, `criterion` and `c4` exit 0 whether or not they
find what you asked about; `declared` is what says. `list`, whose empty
answer is a fact about the corpus, exits 0 too. Read the status
directly, never through a pipe: `specs check | tail` reports `tail`'s
status, so a failing gate reads as a pass.

## What it does not check

`specs check` covers frontmatter properties, slug patterns, slug
uniqueness, reference and anchor integrity, C4 diagrams and Gherkin
criteria. It does **not** yet check required sections, the content of
element record blocks, or line width. Those are specified and not built.

Within references, two things are deliberately out of scope and a clean
run says nothing about them: backticked **relative paths** to files
outside the spec set, and resolution **across workspaces** — a slug that
exists only in another repository's index reports as unresolved.
Reference findings also never suggest or perform a repair; they state
the fact and leave the fix to you.

This matters because a clean `check` looks like a clean document. It is
not: it means the mechanized rules passed. The rules that are still
yours are named in `shared/document-readiness-checklist.md`, which marks
each check with who decides it.

## When it is not installed

Say so, and name the install, rather than reading the corpus by hand and
reporting a result you cannot show your work for:

```
go install github.com/brettcallaghan/agent-plugins/cli/specs@latest
```

or unpack a release archive from the repository's releases page onto
`PATH` (a single static binary; nothing runs at install time).

Reading the corpus by hand is the correct fallback for a project that
has chosen not to install it — but it is a fallback, and the human
should know that is what happened. The checks in
`shared/document-readiness-checklist.md` are defined independently of
the tool, so all of them can be run by hand; the cost is that the result
is an assertion rather than a verifiable one.
