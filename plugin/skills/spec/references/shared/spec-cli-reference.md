<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/spec-cli-reference.md (version unversioned)
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Spec CLI Reference
description: The specs binary's command surface — what each command
  decides, the rules check applies, what it does not yet check, exit
  statuses, and how to install it — for skills that delegate corpus
  queries and conformance checks
slug: spec-cli-reference
type: guide
status: active
---

# Spec CLI reference

`specs` reads a specification corpus and answers questions about it that
would otherwise mean opening every document. This reference is for skills
that delegate corpus queries and conformance checks to it. Prefer it over
reading the corpus yourself for anything it decides. It's faster, and
whoever reviews the work can verify its answer, where yours is an
assertion.

## Contents

- Prerequisites: what must exist before any command works
- Output: the one shape every answer takes
- Commands: the 11 commands and the question each answers
- Reference checks: the six rules `check` applies to `[[slug]]` links and
  reference-bearing frontmatter, and the severity each carries
- C4 diagram checks: the 10 `c4/*` rules over architecture blueprint
  diagrams
- Gherkin checks: the 14 `gherkin/*` rules over acceptance criteria
- Metadata checks: the frontmatter, slug, and empty-corpus rules
- Exit statuses: how to read a run's result
- What it does not check: the gap you still cover by hand
- When it is not installed

## Prerequisites

Every command needs two things, and both are cheap to confirm:

- The `specs` binary. Check for it with `command -v specs`.
- A `specs.toml` that declares `[workspace].root` (the corpus directory),
  `[workspace].name`, and `[profile].name`. Every command resolves the
  corpus through it.

The specs plugin's session hook puts the release matching the plugin's
version first on the PATH, so in a session with the plugin the binary is
normally there. If it's missing, see the last section. Don't silently fall
back to reading documents.

The CLI reads the nearest `specs.toml` in the working directory or a
directory above it. It searches up to the repository root, or up to `/`
when the working directory isn't in a repository, so a command run from a
subdirectory finds the same file. `--config <file>` names one instead and
overrides the search. Don't infer a default when the file is absent. The
CLI names every directory it searched and exits 3, with a message that
begins `no specs.toml found in` and ends `pass --config to name one`.

`[workspace].root` also answers "where do documents go?" Read it rather
than assuming `docs/`.

## Output

Every answer has the same shape, so the parsing you write for one
command carries to the next:

- Answers are YAML by default and JSON under `--json`, with the same
  values in the same order. Every command that answers a question takes
  `--json`.
- The root is always a mapping. A collection sits under a named key
  (`documents`, `matches`, `features`, `criteria`, `references`) as a
  list of mappings, never as formatted lines.
- A lookup by one identifier returns a list, the same list a broader
  selection would. For example, `specs feature FEAT-SPEC-010` answers
  `features:` holding one record.
- A document payload is written as itself, not wrapped: the section
  `headings --emit` reads, the Gherkin `feature --emit` writes, and the
  rewritten document `links` prints.
- `index` writes an artifact, not an answer. Its output is always JSON,
  written to a file by default.

Not found is an answer, and malformed is a refusal. A command answers a
subject the corpus doesn't declare with `declared: false` and an empty
collection, and exits 0. Such a subject is an identifier, a feature code,
a slug, or a document named in a `headings` selection. A value outside a
set the profile registers, such as a document type, a rule, or a tag
qualifier, makes the question malformed. The command names the value and
every registered one, writes nothing to standard output, and exits 2.

```
$ specs slug-exists adr-999-nothing
slug: adr-999-nothing
declared: false
path: null
$ specs criterion --tag nonsense        # exit 2
"nonsense" is not a registered qualifier; registered: negative, manual, level-unit, level-integration, level-e2e
```

## Commands

The table lists each command and what it answers:

| Command | Answers |
|---------|---------|
| `specs check [path]` | Does the corpus, or one document, conform? |
| `specs index [--out <path>\|-]` | The whole corpus's frontmatter, structure, and reference graph, as JSON |
| `specs next-id <type>` | What number does the next document of this type take? |
| `specs slug-exists <slug>` | Is this slug already taken, and by which file? |
| `specs refs <slug>[#<anchor>]` | Who references this document or heading, or what does it reference? |
| `specs list [--type <type>] [--all]` | What does this corpus hold, and what is each document about? |
| `specs c4 [--root <slug>]` | What systems, containers, components, and modules does the corpus document, and what depends on what? |
| `specs headings [<document>#<heading>]` | Where are this document's headings, and what does one section say? |
| `specs feature [<id>]` | What does this feature require, with its criteria? |
| `specs criterion [<id>]` | What does this acceptance criterion say, and which criteria changed? |
| `specs links <path>` | This document with its wikilinks rewritten as Markdown links |

**`specs --version`** answers which build is running, which is what a bug
report needs. It writes a YAML mapping of `version`, `revision`, `time`,
and `modified` and exits 0, with no subcommand and nothing on standard
error. It has no `--json` form. `version` is the version stamped into a
release build, or otherwise the module version (`(devel)` for a local
build). `revision`, `time`, and `modified` are the binary's
version-control build settings. Each is `null` where the build recorded
none, and a `go install` of a tagged module records none.

```
$ specs --version
version: (devel)
revision: 3f1c0a9e2b7d4c5a8e6f1b0d9c2a7e4f5b6a8c1d
time: 2026-10-08T20:51:03Z
modified: false
```

**`specs check`** validates every document's frontmatter against the
active profile's schema and slug patterns, reports duplicate slugs, and
applies the reference, C4, Gherkin, and metadata rules in the sections
below. Its answer leads with a `summary`: the counts of `errors`,
`warnings`, and `unverified` findings, all three always present. The
`findings` follow, each with its rule, severity, path, span, and message,
so the head of a failing run's output is its findings. `--json` writes the
same values in the same order.

`--strict` makes warning and unverified findings fail too. `--verbose`
adds `checks` after the findings: every rule the run was asked for, each
`passed`, `failed`, or `unverified`. The answer's shape never changes the
exit status. `--only <rule>` narrows the run to named rules, given
repeatedly or comma-separated. The summary then counts only their
findings, and `checks` lists only those rules:

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

`specs check <path>` reports the findings located in that file, and the
references elsewhere that your change to it broke. It resolves the
document's references against the whole corpus. It reports every finding
that sits in the document at its own line: an unresolved reference or
anchor, a retired target, a type mismatch, or a slug or identifier it
shares with another document.

It also reports two kinds of reference elsewhere, each at its line in the
citing document:

- Every reference that cites a slug the document declared at git HEAD and
  no longer declares, which is an uncommitted rename.
- Every anchor into the document that no longer names one of its
  headings.

Nothing else located in another document is reported, so an unrelated
broken link elsewhere stays out. Outside a git work tree, or without git,
there's no HEAD to compare with. The check then writes one `note:` line to
standard error saying that it didn't check references elsewhere, and
answers with the document's own findings. A changed `type` or a
retirement isn't reported elsewhere, because the citing reference still
resolves.

Before you rename a slug, retire a document, or change its `type`, run
`specs refs <slug>`. It names every document citing the slug, which is the
blast radius of the change. After the change, a whole-corpus `specs check`
reports what's broken anywhere.

A path that names no corpus member is refused, not passed. Where the file
exists but was excluded, the message names the reason (`no-frontmatter`,
`unknown-type`):

```
$ specs check docs/typo.md              # exit 2, nothing on stdout
docs/typo.md is not a corpus member: no document at that path was discovered under the spec root
```

**`specs index`** writes the corpus index to `specs.json` inside the spec
root. The index holds every member's frontmatter, headings with their
section spans, record blocks, and the reference graph. `--out <path>`
writes it elsewhere, and `--out -` sends it to standard output. Read it
when you need many documents' metadata at once instead of opening files.
Indexing never fails on a broken reference. The index records an
unresolved reference and `check` reports it, so a repository whose links
are broken still contributes its graph.

Every other command except `links` refreshes that same file when it had
to rebuild, so the second of two runs is served from it. You never need
to run `specs index` to make the others fast. Run it when you want the
artifact itself.

The file is derived and disposable: deleting it changes no answer, and it
belongs in `.gitignore`. Writing it is best effort. In a read-only
checkout every command answers and exits exactly as it would in a
writable one, and the file only stays cold.

**`specs next-id <type>`** answers one greater than the highest number in
use for a numbered type across the whole corpus, and the slug that number
forms. It's the only correct way to pick an ADR number, because scanning
filenames misses a document whose filename and slug disagree. A type the
profile doesn't number is refused (exit 2).

```
$ specs next-id adr
number: 32
slug: adr-032
```

**`specs slug-exists <slug>`** answers whether any document declares that
slug, and where. The match is case-insensitive, because a
case-insensitive file system can't hold two files that differ only in
case. Where a document declares the slug, the answer writes it as that
document declares it, so `ADR-012` answers `adr-012`. Where no document
declares it, the answer writes the slug as given. It exits 0 either way.
Run it before you claim a slug.

```
$ specs slug-exists adr-012
slug: adr-012
declared: true
path: specs/adrs/adr-012-reference-conventions.md
```

**`specs refs <slug>`** answers the blast-radius question. By default it
lists every corpus member whose body or reference-bearing frontmatter
names that slug. Run it before you retire or rename a document, not
after. The answer is compact: one entry per referring document, with each
reference reduced to its line and text. `--full` returns every field of
every reference instead: its span, its origin (body, or which frontmatter
property), and what it resolved to. A rewrite then opens only the lines
that carry the reference.

`--outbound` reverses the direction and lists what that document itself
references, which is what to read before you edit one. Its compact answer
groups by cited slug instead of by document. It has one entry per slug,
carrying the `path` to read it at (`null` when it resolves to nothing,
with `resolution.state`) and the `lines` that cite it.

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
anchor resolves to that heading of the document, which is the blast
radius of renaming the heading. A heading is matched on its identifier,
so letter case and punctuation don't decide a match. A reference whose
anchor matches more than one heading is reported for each, marked as
such. The answer then carries `anchor` and `anchor_declared`.
`--outbound` takes no anchor.

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

Every form exits 0 when the answer is empty. The answer reports "no
corpus member references this slug" (`declared: true`, `references: []`)
distinctly from "no corpus member declares this slug" (`declared: false`).
The first means the document is unreferenced. The second means you asked
about a slug that doesn't exist. Read the output, not the status.

**`specs list`** answers "what does this corpus hold?" without opening a
document or the index. It returns one entry per corpus member. Each entry
carries, in this order, the member's `path`, `slug`, `type`, `status`,
`title`, and `description`. The path is repository-relative with forward
slashes, so the next step is a path you were handed rather than a slug
you have to resolve. The answer is a mapping under a single `documents`
key. Nothing surrounds it: no count, no heading naming the command, and no
trailing summary.

Two filters narrow it, and they compose:

- **`--type`** names the document types to return. It's repeatable and
  accepts a comma-separated value, so `--type adr --type frd` and
  `--type adr,frd` ask the same question. A value the active profile
  doesn't register is a malformed question, not an empty category. The
  command names the unregistered value and every registered type name,
  and exits 2.
- **`--all`** asks for every status. Without it, the answer omits members
  at a retired status. A retired document isn't in effect, and an agent
  orienting over the corpus that reads one has been actively misled, not
  merely charged for tokens. The answer reports every other status by
  default, and each entry carries its own `status`.

A filter that matches no member returns an empty collection and exits 0.
That isn't a failure.

The command treats two kinds of file differently, according to what the
index recorded.

A member the index recorded with a parse error is reported as an entry
naming its path and the failure. The entry carries no slug, type, status,
title, or description. It's a member, so omitting it would make the
listing disagree with the corpus. A null value would assert that the
document declares none, which isn't known, so the fields are absent
rather than null. Under a `--type` filter the entry isn't reported at
all, because it declares no type for a filter to match.

A file the index recorded as an exclusion is never reported. It isn't a
member, and whether it should be one is for `specs check` to judge, not
this command.

Entries come back in the index's own order, fixed by repository-relative
path, so two runs against an unchanged corpus produce byte-identical
output.

`specs list` is what makes `description` a consumed interface. For this
reason, `shared/frontmatter-standards.md` tells an author to write the
description so that a reader can triage from it.

**`specs c4`** emits the corpus's C4 model: the architecture the corpus
declares, read from blueprint frontmatter, roster record blocks, and the
`mermaid` C4 diagrams in blueprints. It takes no positional argument and
two flags:

- **`--root <slug>`** narrows the model to the element with that slug and
  its descendants. The root's ancestors appear only as the path down to
  it, and the model keeps only relationships with both endpoints inside
  the subtree. The model then carries `root_declared`. A slug naming no
  element is answered with `root_declared: false` and every collection
  empty, and exits 0.
- **`--json`** writes JSON instead, with the same values in the same
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
`workspace` (the name in `specs.toml`), and the following collections.
Each collection is ordered by slug, so two runs over an unchanged corpus
produce byte-identical output:

- `software_systems`: systems, each holding its containers, and each
  container its components and the slugs of its modules.
- `orphaned_containers`: containers whose blueprint declares no `system`.
- `orphaned_components`: components no container or module declares.
- `modules`: every module once, whatever the number of its hosts.
- `relationships`: the dependency edges drawn in blueprint diagrams
  between two elements.
- `excluded_relationships`: every `Rel` or `BiRel` statement the model
  doesn't admit, with its `path`, `span`, `text`, and a `reason`. The
  reason is one of these:
  - `external_endpoint`: a `Person` or `_Ext` node.
  - `passive_store`: a `Db` or `Queue` naming no element.
  - `undeclared_node`: a label matching no element in scope.
  - `ambiguous_node`: a label matching more than one element.
  - `bidirectional`: a `BiRel`, which states no single dependency.

The command exits 0 whatever the model holds. Excluded relationships are
facts recorded in the artifact, and judging the diagrams is the job of
`specs check` (`c4/*` findings). The JSON output validates against the
`C4Model` definition in the generated model JSON Schema.

**`specs headings`** answers where a document's headings are, so you can
read one section without opening the whole document. The selection is
one operand, `<document>#<heading>`, and either half is optional:

- `<document>` alone selects every heading in it.
- `#<heading>` alone selects that heading in every corpus member.
- No operand selects every heading in the corpus.

A document half that holds a path separator or ends in `.md` is a
repository-relative path. Otherwise it's a slug. A heading is matched on
its identifier, so letter case and punctuation don't decide a match. The
heading half may be a path of headings joined by `#`, outermost first,
each nested anywhere inside the one before it.

Each match is compact by default: its `anchor`, its `ref`, its `level`,
and the `start_line` its heading stands on. `--full` adds the file
`path`, `text`, `identifier`, and the `section` span of the content
beneath it. Every heading nested inside a match's section follows the
match.

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

Without `--full` the same match is `anchor`, `ref`, `level: 3`, and the
`start_line` of the heading line itself.

**`--emit`** writes the section's content as itself instead, which is the
cheapest way to read one part of a long document. It requires the
selection to match exactly one heading. A selection matching several is
refused (exit 2). The refusal names each match by the `ref` that selects
it alone, which you paste back. A `ref` is the declared slug, `#`, and
the `anchor`. A member that declares no slug is named by its path in
place of the slug:

```
$ specs headings 'module-document-core#References' --emit     # exit 2
"module-document-core#References" matches 2 headings; --emit needs exactly one. Select one of:
    module-document-core#document-core#components#references
    module-document-core#document-core#references
or drop --emit to list the matches
```

A document the corpus doesn't declare answers `declared: false`. A
heading matching nothing in a declared document answers `declared: true`
with no matches. Both exit 0. Under `--emit`, raw text can't say either,
so both are refused (exit 2, nothing on standard output). The message
names the selection and whether the document is undeclared or the
heading absent.

Like `specs list`, `headings` omits members at a status no longer in
effect unless you give `--all`. That applies to the answer and to the
`--emit` refusal's list alike. A selection that names such a document by
slug or path still answers with its headings.

**`specs feature`** returns features from the Gherkin in the corpus. The
answer is compact by default: each feature's `id`, `code`, `name`,
`path`, and `criteria_count`. `--full` returns the whole record,
description and criteria included. `feature <id>` returns that one
feature, `--code <CODE>` returns every feature grouped under that code,
and no argument returns every feature. Identifiers and codes match
regardless of letter case. The answer is a `features` list in every case,
preceded by `declared` where the arguments name a subject.

`feature <id> --emit` writes the feature as a standalone Gherkin
`.feature` file (tags, description, and criteria) to standard output, or
to the path `--out` names. An unknown id is refused there (exit 2),
since there's nothing to write, and `--full` is refused there too.

```
$ specs feature FEAT-SPEC-010 --emit
@frd-spec-index @FEAT-SPEC-010
Feature: Index Output and Persistence
```

**`specs criterion`** returns acceptance criteria. `criterion <id>`
returns that one criterion, `--of <feature-id>` returns every criterion
beneath a feature, and `--tag <qualifier>` returns every criterion
carrying a qualifier, written with or without the leading `@`. The
answer is a `criteria` list in every case, preceded by `declared` when
you give an id or `--of`.

```
$ specs criterion ac-spec-010.7
declared: true
criteria:
  - id: AC-SPEC-010.7
    path: specs/features/spec-index/spec-index-frd.md
    feature_id: FEAT-SPEC-010
```

**`criterion --changed`** lists every acceptance criterion whose content
differs from the corpus at a git base, and classifies each as `changed`,
`added`, or `removed`. The base is the merge-base of `HEAD` and the
repository's default branch, or of `HEAD` and the branch, tag, or commit
that `--since` names. The current side includes uncommitted edits.

The default branch is the first that exists of `origin/HEAD`'s target,
`origin/main`, `origin/master`, `main`, and `master`, so a stale local
`main` never wins over `origin/main`. When a local branch of the same
name is behind the one chosen, standard output is unchanged and standard
error carries `warning: local
main (<short>) is behind origin/main (<short>); comparing against
origin/main`.

The answer holds `base` (the revision and the merge-base commit) and then
`criteria`. Each criterion carries `id`, `name`, `classification`, and
`path`. `name` is omitted where the criterion has none, and a removed
criterion carries the base's name. `--full` adds `record` (the criterion
exactly as `specs criterion <id>` lists it) for an added or changed
criterion. It adds `base_record` (the same, taken from the base) for a
changed or removed one. `--full` without `--changed` exits 2. The command
exits 0 whatever it lists, and `--exit-code` makes a non-empty list exit
1, for a gate. It needs `git` on `PATH`.

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
rewritten as an ordinary Markdown link, for rendering outside the corpus.
It prints the count of rewritten links on standard error. With no map,
each link targets the cited document's path relative to the one being
rewritten.

`--map <file>` names the targets instead. The map is a YAML mapping from
each corpus slug to a record declaring `path`, such as
`adr-012: {path: ../adrs/adr-012.md}`. The command ignores a record's
other properties and accepts a map written in JSON. A map of any other
shape is refused (exit 2) with the expected shape in the message. A cited
slug the map omits raises `links/unmapped` (error), and no document is
printed. A slug matching no corpus member raises `links/unresolved`
(warning) and is left as written.

A wikilink inside code is left as written and not counted. Code here
means an inline span, an indented block, or any fenced block, including
a ```` ```gherkin ```` fence. `links` creates, modifies, and deletes no file.

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
| `reference/unresolved-anchor` | error; unverified | An anchor matching no heading of the cited document (error). An anchor into a document whose headings aren't recorded, so the check couldn't look (unverified) |
| `reference/ambiguous-anchor` | error | An anchor whose identifier is the identifier of more than one heading of the cited document |
| `reference/retired-target` | warning | A live document referencing one at `status: retired`, quoting its `superseded_by` when it has one |
| `reference/pairing-incomplete` | error | One half of a paired relationship whose target doesn't declare the counterpart |
| `reference/type-mismatch` | error | A typed reference whose target declares a different `type`, such as a feature blueprint's `frd` naming something that isn't an FRD |

Four things about reference resolution decide whether you trust it:

- A backticked `[[slug]]` isn't a reference. Text inside a fenced code
  block or an inline code span illustrates the notation and contributes
  nothing. Neither does a `[[slug]]` written in a frontmatter value.
  Frontmatter contributes references only through the properties the
  profile registers as reference-bearing. A ```` ```gherkin ```` fence is
  the exception: its wikilinks are references, checked and indexed like
  prose.
- An anchor is resolved against the cited document's headings.
  `[[container-x#Some Heading]]` resolves on `container-x`. Its anchor is
  then matched on heading identifier, so case and punctuation don't
  matter, or on a `#`-joined heading path. A miss raises
  `reference/unresolved-anchor`, and more than one match raises
  `reference/ambiguous-anchor`. A bare `[[#Heading]]` resolves against
  the document it's written in.
- A target that exists but isn't a corpus member reports as such: "target
  exists but is not a corpus member: excluded: no-frontmatter". It doesn't
  report as a slug matching nothing, because the reader's next step
  differs.
- A target whose frontmatter doesn't parse raises nothing against the
  citing document. The parse failure is reported once, against the
  document that has it. A slug reaches such a document by its file name,
  or by a prefix of the file name followed by `-`: `[[adr-012]]` reaches
  `adr-012-reference-conventions.md`.

`reference/retired-target` is the only reference rule that emits
`warning`. A reference to a retired document still resolves, because the
reader reaches a real document that says it has been replaced. So it
doesn't fail a build unless you pass `--strict`. It's also the one
finding that tells an author what to cite instead.

The profile's schema determines which frontmatter properties are
reference-bearing, and no list is kept by hand. Every property typed as a
document slug is one, except `slug` itself, which names its own document.
For that reason, `superseded_by`, `sources`, `frd`, `system`, `modules`,
`provider`, and `contracts_consumed` are all reference-bearing. A
slug-typed property added to the vocabulary becomes one the moment it's
declared.

This framework's own profile registers no pairing, so
`reference/pairing-incomplete` is inert against its own corpus by design.
The rule exists for adopting profiles whose relationships are declared
from both ends.

## C4 diagram checks

`specs check` reads the C4 diagrams in architecture blueprints (system,
container, module, and component) and applies 10 `c4/*` rules. Every
finding carries the span of the statement it concerns. Diagrams in other
document types aren't checked. The statement subset the parser reads and
the reserved tags are defined in [[c4-standards]].

| Rule | Severity | Reports |
|------|----------|---------|
| `c4/parse` | error; unverified | A statement outside the parsed subset (error). A `mermaid` block whose diagram type isn't read, including `C4Dynamic` and `C4Deployment`, so the check didn't look inside (unverified) |
| `c4/undeclared-node` | error | A non-external node that resolves to no element in the drawing document's scope and isn't a passive store |
| `c4/ambiguous-node` | error | A node whose label matches the entity names of more than one element in scope, naming each |
| `c4/bidirectional` | error | A `BiRel` statement |
| `c4/duplicate-edge` | error | A second relationship statement in one diagram with the same source and target, after `Rel_Back` is normalized |
| `c4/drawing-conflict` | warning | One relationship drawn in two places with differing technology, description, or tags, naming the other drawing |
| `c4/interaction-conflict` | error | A relationship tagged both `sync` and `async` |
| `c4/contract-unresolved` | error | A `contract:` tag whose slug no contract document declares |
| `c4/contract-mismatch` | warning | A `contract:` tag naming a contract the source element doesn't consume or the target element doesn't provide |
| `c4/tag-separator` | warning | A tags argument that joins tags with a comma instead of `+` |

Like every rule, each `c4/*` rule is narrowed to the file you name.
`c4/drawing-conflict` still compares that file's drawings against every
architecture blueprint's, but reports only the findings that sit in the
named file.

## Gherkin checks

`specs check` reads every acceptance criterion written in Gherkin and
applies 14 `gherkin/*` rules. It finds them in two places: in a
```` ```gherkin ```` fence in a document, or in a standalone `.feature`
file under the spec root.

| Rule | Severity | Reports |
|------|----------|---------|
| `gherkin/parse` | error | A Gherkin source that couldn't be parsed |
| `gherkin/feature-untagged` | error | A Feature carrying no identifier tag |
| `gherkin/criterion-untagged` | error | A Scenario carrying no identifier tag |
| `gherkin/identifier-malformed` | error | An identifier tag that doesn't match the pattern for its kind |
| `gherkin/identifier-multiple` | error | A Feature or Scenario carrying more than one identifier tag |
| `gherkin/identifier-duplicate` | error | A feature or criterion identifier declared twice in the corpus |
| `gherkin/criterion-code-mismatch` | error | A criterion whose identifier's code and number aren't its feature's |
| `gherkin/owner-missing` | error | A feature in a standalone `.feature` file carrying no owning-document tag |
| `gherkin/owner-mismatch` | error | A feature in a fenced block whose owning-document tag names a different document |
| `gherkin/owner-unresolved` | error | An owning-document tag naming no corpus member |
| `gherkin/criterion-empty` | error | A criterion with no steps |
| `gherkin/examples-partially-tagged` | error | A Scenario Outline whose Examples blocks carry an identifier on some and not others |
| `gherkin/scenario-prose` | warning | A Scenario description that reads as a normative statement, which is probably an untagged criterion |
| `gherkin/tag-unregistered` | warning | A tag the profile's vocabulary assigns no meaning, such as a misspelled `@negatve`, naming the registered qualifiers |

`gherkin/tag-unregistered` is a warning because a project's own tags
coexist with the vocabulary, and only a misspelled one loses a meaning.
Pass `--strict` where an unknown tag should block.

## Metadata checks

`specs check` also applies these rules to frontmatter, slugs, and the
corpus as a whole:

| Rule | Severity | Reports |
|------|----------|---------|
| `frontmatter/parse` | error | A frontmatter block that couldn't be read as a document envelope (most often a `status` or other enumerated value outside its vocabulary), so nothing else examined the document |
| `frontmatter/schema` | error | A frontmatter value that violates its type's schema, naming the property, the value, and the constraint: `/reviewer: "someone" does not satisfy additionalProperties` |
| `frontmatter/missing-property` | error | A required frontmatter property the document doesn't declare |
| `frontmatter/inline-comment` | warning | A YAML comment trailing a frontmatter value on the same line, such as `description: a # b`, located at the `#`. Unquoted, the text after it isn't part of the value. Quote the value to keep it |
| `slug/pattern-mismatch` | error | A slug that doesn't match its type's pattern |
| `slug/duplicate` | error | A slug declared by more than one document, naming the other |
| `corpus/empty` | error | A run whose spec root holds no corpus member at all, so a clean report would mean nothing was examined |

Fix a `frontmatter/parse` finding first. It means nothing else could
examine the document, so the other checks haven't cleared it. By far the
most common cause is a template placeholder
(`visibility: <public | internal>`) left unfilled.

## Exit statuses

Every command exits with one of five statuses, and each status names who
acts:

- **0: success.** The command did what was asked, such as answering a
  query or passing a gate. For `check`, it means no error-severity
  findings, and under `--strict`, no findings at all.
- **1: gate failed.** A gate ran to completion and the corpus failed it:
  `check` found something, `links` met a citation it can't rewrite, or
  `criterion --changed --exit-code` listed something. Fix the documents.
- **2: invocation error.** The command couldn't run as invoked. The cause
  is an unknown command or flag, a missing or surplus argument, a value
  outside a registered set, or an operand it can't act on. Fix the
  command.
- **3: environment error.** The invocation was correct, but its
  surroundings prevent it from running. The cause is a missing
  `specs.toml`, an invalid configuration, an unknown profile, an
  unreadable spec root, or an output path it can't write. Fix the
  repository.
- **4: internal error.** The CLI reached a state it doesn't handle, a
  panic included. File a bug.

On 2, 3, or 4, nothing is written to standard output and the reason goes
to standard error, so a caller parsing the output never parses half an
answer. On 1, the findings are the answer, on standard output. Don't
retry on any of these statuses.

A run whose findings are all `warning` or `unverified` exits 0 without
`--strict` and 1 with it. Such findings include `reference/retired-target`,
`gherkin/tag-unregistered`, `c4/drawing-conflict`, and a diagram that
`c4/parse` didn't look inside. That's where the two runs differ, so pass
`--strict` in CI wherever any of them should block.

Read a query's answer, not its status. `slug-exists`, `refs`,
`headings`, `feature`, `criterion`, and `c4` exit 0 whether or not they
find what you asked about, and `declared` tells you which. `list` also
exits 0, because its empty answer is a fact about the corpus. Read the
status directly, never through a pipe: `specs check | tail` reports
`tail`'s status, so a failing gate reads as a pass.

## What it does not check

`specs check` covers frontmatter properties, slug patterns, slug
uniqueness, reference and anchor integrity, C4 diagrams, and Gherkin
criteria. It doesn't yet check required sections, the content of element
record blocks, or line width. Those checks are specified but not built.

Within references, two things are deliberately out of scope, and a clean
run says nothing about them:

- Backticked relative paths to files outside the spec set.
- Resolution across workspaces. A slug that exists only in another
  repository's index reports as unresolved.

Reference findings also never suggest or perform a repair. They state
the fact and leave the fix to you.

A clean `check` looks like a clean document, but it only means the
mechanized rules passed. `shared/document-readiness-checklist.md` names
the rules that are still yours and marks each check with who decides it.

## When it is not installed

The specs plugin installs the CLI when a session starts. It pins the
release whose version equals the plugin's own and puts it first on the
session's PATH. On a user's machine, a `specs` of another version already
on the PATH stays installed and is shadowed for the session only. In a
cloud session, the pinned binary also replaces `/usr/local/bin/specs`.

When provisioning fails, the session's context carries one line beginning
`specs plugin:` that names the cause. If that line also says not to run
another `specs` left on the PATH, don't run it. It isn't the version these
skills are written against, so check by hand as described later in this
section. When a user has set `SPECS_CLI_PROVISION=off`, the `specs` on
their PATH is the one that runs, whatever its version.

Without the binary, say so, and name the cause and how to install it.
Don't read the corpus by hand and report a result you can't show
your work for. Outside the plugin, install it with:

```
go install github.com/brettcallaghan/specs/cli/specs@latest
```

You can also unpack a release archive from the repository's releases page
onto `PATH`. It's a single static binary, and nothing runs at install
time.

Reading the corpus by hand is the correct fallback when the binary is
unavailable. It's still a fallback, and the human should know that's what
happened. The checks in `shared/document-readiness-checklist.md` are
defined independently of the tool, so you can run all of them by hand.
The cost is that the result is an assertion rather than a verifiable one.
