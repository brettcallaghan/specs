<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: specs/guides/specs-cli-reference.md
     Regenerate with: scripts/sync-shared-references.sh
     Edit the source, then re-run that script. -->

---
title: Specs CLI Reference
description: The specs binary for agents — gotchas, the commands to run at
  each authoring step, each command's answer and exit status, and what a
  clean check doesn't cover
slug: specs-cli-reference
type: guide
status: active
---

# Specs CLI reference

`specs` reads a specification corpus and answers questions about it that would
otherwise mean opening every document. Prefer it to reading the corpus
yourself: its answer is faster, and a reviewer can re-run it.

## Must-know gotchas

1. **Read the answer, not the exit status.** `slug-exists`, `refs`,
   `headings`, `feature`, `criterion`, `list`, and `c4` exit 0 whether or not
   they find the subject. `declared: false` means it doesn't exist.
2. **Never read a status through a pipe.** `specs check | tail` reports
   `tail`'s status, so a failing gate reads as a pass. Redirect to a file,
   then read `$?`.
3. **A clean `specs check` isn't a passing document.** It covers frontmatter,
   slugs, references and anchors, C4 diagrams, and Gherkin. It doesn't check
   required sections, record-block content, line width, or anything that
   needs judgment.
4. **Fix `frontmatter/parse` first.** Nothing else examined that document.
   The usual cause is an unfilled template placeholder, such as
   `visibility: <public | internal>`.
5. **`specs check <path>` reports only that document,** plus references
   elsewhere that a change to it broke: a slug it declared at git HEAD and no
   longer declares, and anchors into it that no longer match a heading. A
   retirement or a changed `type` isn't reported elsewhere. Run
   `specs refs <slug>` before the change and a whole-corpus `specs check`
   after it.
6. **Run `--help` instead of caching flags.** The binary is the source of
   truth.

## The agent loop

- **Need the spec root:** read `[workspace].root` in `specs.toml`.
- **Pick an ADR or other numbered id:** `specs next-id <type>`.
- **Claim a slug:** `specs slug-exists <slug>`.
- **Orient over the corpus:** `specs list [--type <type>]`.
- **Read one section of a long document:**
  `specs headings '<slug>#<Heading>' --emit`.
- **Learn what a document depends on:** `specs refs <slug> --outbound`.
- **Rename, retire, or re-type a document:** `specs refs <slug>` first.
- **Rename a heading:** `specs refs '<slug>#<Heading>'` first.
- **Finish a document:** `specs check <path>`, then fix and re-run.
- **Finish a change that touched several documents:** `specs check`.
- **Read a feature or a criterion:** `specs feature <id>` or
  `specs criterion <id>`.
- **List criteria a branch changed:** `specs criterion --changed`.
- **Render a document outside the corpus:**
  `specs links <path> [--map <file>]`.

To read a `specs check` finding, look its rule id up in the rules
reference generated from `schemas/rules.strata`. It lists every rule, the
severity it reports at, and what it means.

## Prerequisites

- **The binary.** Check with `command -v specs`. The plugin's session hook
  puts the matching release first on `PATH`. If it's missing, see
  [When it isn't installed](#when-it-isnt-installed).
- **A `specs.toml`** declaring `[workspace].root`, `[workspace].name`, and
  `[profile].name`. The CLI uses the nearest one in the working directory or
  above it, up to the repository root. `--config <file>` overrides the
  search. With none found, it names every directory it searched and exits 3.

## Output

- YAML by default, JSON under `--json`, with the same values in the same
  order.
- The root is always a mapping. A collection sits under a named key, such as
  `documents`, `matches`, `features`, `criteria`, or `references`.
- A lookup by one identifier returns a list, the same shape a broader
  selection returns.
- A document payload is written as itself: `headings --emit`,
  `feature --emit`, and `links`.
- `index` writes an artifact, not an answer: JSON, to a file by default.

**Not found is an answer; malformed is a refusal.** A subject the corpus
doesn't declare answers `declared: false` with an empty collection and exits
0. A value outside a set the profile registers, such as a document type, a
rule, or a tag qualifier, writes nothing to standard output, names every
registered value, and exits 2.

```
$ specs slug-exists adr-999-nothing
slug: adr-999-nothing
declared: false
path: null
$ specs criterion --tag nonsense        # exit 2
"nonsense" is not a registered qualifier; registered: negative, manual, level-unit, level-integration, level-e2e
```

## Commands

### `specs check [path]`

Does the corpus, or one document, conform? The answer leads with `summary`
(`errors`, `warnings`, `unverified`), then `findings`, each with its rule,
severity, path, span, and message.

- `--only <rule>` runs only the named rules. Repeatable or comma-separated.
- `--strict` fails on warning and unverified findings too.
- `--verbose` adds `checks`: each rule the run applied and its outcome.

A path that names no corpus member is refused (exit 2). Where the file exists
but was excluded, the message names the reason, such as `no-frontmatter` or
`unknown-type`. Outside a git work tree, a single-document check writes a
`note:` to standard error and skips references elsewhere.

```
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

### `specs index [--out <path>|-]`

Writes every member's frontmatter, headings with section spans, record
blocks, and the reference graph to `specs.json` in the spec root. `--out -`
writes to standard output. Read it when you need many documents' metadata at
once. Other commands refresh the same file on their own, so you never need to
run `index` for speed. The file is disposable and belongs in `.gitignore`.

### `specs next-id <type>`

One more than the highest number in use for a numbered type, and the slug it
forms. This is the only correct way to pick an ADR number: filenames and
slugs can disagree. A type the profile doesn't number exits 2.

```
$ specs next-id adr
number: 32
slug: adr-032
```

### `specs slug-exists <slug>`

Whether any document declares the slug, and where. Case-insensitive. Run it
before you claim a slug. Record-block slugs aren't included.

### `specs refs <slug>[#<Heading>]`

Every member whose body or reference-bearing frontmatter cites the slug, one
entry per document with each reference's line and text. With an anchor, only
references to that heading. `--full` adds each reference's span, origin, and
resolution. `--outbound` lists what the document itself cites, grouped by
cited slug, with the `path` to read each at. `declared: true` with
`references: []` means unreferenced. `declared: false` means no such slug.

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

### `specs list [--type <type>] [--all]`

One entry per member: `path`, `slug`, `type`, `status`, `title`, and
`description`, under `documents`. `--type` is repeatable or comma-separated.
Retired members are omitted unless you pass `--all`. A member whose
frontmatter doesn't parse is listed with its path and the failure only.
Planners triage elements from the `description`, which is why
the frontmatter standards say to write it for triage.

### `specs c4 [--root <slug>]`

The C4 model the corpus declares, from blueprint frontmatter, roster record
blocks, and `mermaid` C4 diagrams. Keys: `schema_version`, `workspace`,
`software_systems`, `orphaned_containers`, `orphaned_components`, `modules`,
`relationships`, and `excluded_relationships` (each with a `reason`).
`--root` narrows to one element's subtree and adds `root_declared`. Exits 0
whatever the model holds. `specs check` judges the diagrams.

### `specs headings [<document>][#<heading>]`

Where headings are, so you can read one section. The document half is a slug,
or a path when it contains `/` or ends in `.md`. The heading half matches on
identifier, so case and punctuation don't matter, and may be a `#`-joined path
of headings, outermost first. Each match carries `anchor`, `ref`, `level`, and
`start_line`. `--full` adds `path`, `text`, `identifier`, and the `section`
span. Retired members are omitted unless you pass `--all`.

`--emit` writes the one matched section as itself. A selection matching
several headings exits 2 and names each by the `ref` that selects it alone:

```
$ specs headings 'module-document-core#References' --emit     # exit 2
"module-document-core#References" matches 2 headings; --emit needs exactly one. Select one of:
    module-document-core#document-core#components#references
    module-document-core#document-core#references
or drop --emit to list the matches
```

### `specs feature [<id>] [--code <CODE>]`

Features from the corpus's Gherkin: `id`, `code`, `name`, `path`, and
`criteria_count`. `--full` returns the whole record. `<id> --emit` writes the
feature as a standalone `.feature` file, to standard output or `--out`.

### `specs criterion [<id>] [--of <feature-id>] [--tag <qualifier>]`

Acceptance criteria, under `criteria`. `--changed` lists each criterion that
differs from a git base as `changed`, `added`, or `removed`. The base is the
merge-base with the default branch, or with `--since <rev>`, and uncommitted
edits count. `--full` with `--changed` adds `record` and `base_record`.
`--exit-code` exits 1 on a non-empty list, for use as a gate. Needs `git`.

### `specs links <path> [--map <file>]`

Prints the document with each `[[slug]]` rewritten as a Markdown link, and
the count of rewritten links on standard error. Without a map, each target is
the cited document's path relative to this one. `--map` takes a YAML or JSON
mapping from slug to a record with `path`, such as
`adr-012: {path: ../adrs/adr-012.md}`. A wikilink inside code is left as
written. `links` writes no file.

### `specs --version`

A YAML mapping of `version`, `revision`, `time`, and `modified`, for bug
reports.

## Exit statuses

- **0:** answered, or the gate passed.
- **1:** a gate failed: findings from `check`, an unrewritable citation in
  `links`, or a list from `criterion --changed --exit-code`. Fix the
  documents.
- **2:** invocation error, such as an unknown flag, a bad argument, or a value
  outside a registered set. Fix the command.
- **3:** environment error, such as no `specs.toml`, an invalid
  configuration, an unknown profile, or an unwritable output path. Fix the
  repository.
- **4:** internal error, including a panic. File a bug.

On 2, 3, and 4, standard output is empty and the reason is on standard error.
On 1, the findings are the answer. Retrying fixes none of them. Findings that
are all `warning` or `unverified` exit 0, and 1 under `--strict`.

## Reference resolution

- A `[[slug]]` resolves against frontmatter `slug`, never a file path, and
  ignores case.
- A wikilink in a code span or a fenced block isn't a reference, except in a
  ```` ```gherkin ```` fence.
- A frontmatter value is a reference only in a property the profile types as
  a slug, other than `slug` itself.
- An anchor matches the cited document's heading identifiers, or a `#`-joined
  heading path. `[[#Heading]]` resolves in the document it's written in.
- A target that exists but isn't a corpus member is reported as excluded,
  with its reason.

## What it doesn't check

- Required sections, record-block content, and line width.
- Backticked relative paths to files outside the spec set.
- References across workspaces: a slug that exists only in another
  repository reports as unresolved.

Findings never suggest or perform a repair.

## When it isn't installed

The plugin pins the release that matches its own version for the session.
When provisioning fails, the session context carries a line starting
`specs plugin:` that names the cause. If that line says not to run another
`specs` on `PATH`, don't. With `SPECS_CLI_PROVISION=off`, the `specs` on
`PATH` runs, whatever its version.

Outside the plugin, install it with:

```
go install github.com/brettcallaghan/specs/cli/specs@latest
```

Without the binary, say so, name the cause, and check by hand. Report the
result as a manual check, because nobody can re-run it.
