---
title: Spec CLI Output Standards
description: What every specs command writes to its output stream — YAML
  by default and JSON under --json, the answer alone, a document payload
  as itself, and the style rules that make both readable
slug: cli-output-standards
type: standards
status: approved
updated_at: 2026-10-08T00:00:00Z
version: "0.2"
---

# Spec CLI Output Standards

These standards govern what every command [[container-spec-cli]] exposes
writes to its output stream. They apply to every subcommand, present and
future, because what a command prints is the command surface's business
rather than any one subcommand's.

**The reader is an agent that knows which command it ran and nothing
else.** Not the command's author, not someone holding this document. The
skills over this CLI are being shrunk toward "run the command and read
the output", and that is only safe where the output needs no key.

**The method is to stop inventing a format.** Structured output is YAML,
so a reader who has met YAML once can read it and a consumer already has
a parser for it. There is no line grammar to reverse-engineer, no column
convention, no punctuation carrying meaning.

**Status: approved, and met by the commands specified since.** The rest
of the surface predates these standards; [Current
conformance](#current-conformance) records what each command prints
today, measured, and what each has to change. A command specified after
this document is written to conform from the start, and `headings` is
the first. `refs` is the first of the commands that predate them to be
migrated, by the feature that next touched it.

---

## Renderings

### YAML by default, JSON under `--json`

Every command SHALL answer with structured output, except where its
answer is a document payload as defined below.

WHERE a command's answer is structured, the CLI SHALL render it as YAML,
and SHALL render it as JSON when `--json` is passed.

**Rationale:** two renderings, one flag, no third form. YAML is the
default because a human and an agent both read it without a schema, and
because this CLI's values routinely carry multi-line text — a criterion's
steps, a finding's message, a document's description — which survives YAML
intact instead of arriving as `\n` inside a quoted string. JSON stays
available so a consumer that already parses it need not acquire a YAML
parser. "Every command" is the point: a caller that has parsed one
command's answer has parsed them all, with no per-command recipe for a
tab-separated line, a bare number or a sentence.

---

The root of every structured answer SHALL be a mapping, in both
renderings.

**Rationale:** a root sequence or scalar has nowhere to put a property
added later, so the first property a command acquires would change its
shape and break every caller. A mapping at the root takes a new key
without disturbing the ones a caller already reads.

---

Both renderings SHALL carry the same values in the same order.

**Rationale:** `--json` selects a syntax, never a depth or a subset. A
caller comparing the two forms is comparing two spellings of one answer.

---

### A document payload is written as itself

WHERE a command's answer is a document, or content taken from a
document, rather than a set of properties, the CLI MAY write that
content raw instead of as a structured answer.

WHERE the CLI writes such content raw, it SHALL write that content
alone, with no envelope, no added heading, and no trailing summary.

**Rationale:** `links` emits a rewritten markdown document, `feature
--emit` a Gherkin source, and a heading query's emit form the text of a
section; any other command that prints content from a document is the
same case. Wrapping any of them in a mapping makes the caller unwrap it,
and quoting a document into a scalar changes the bytes they asked for.
The test is what the caller does next: text they will write to a file or
read as prose is a payload; anything they would index by name is an
answer.

---

### An artifact is not an answer

WHERE a command writes an artifact whose format another document
specifies, that format SHALL govern the artifact, and these standards
SHALL govern only what the command prints about it.

**Rationale:** the index is JSON because [[frd-spec-index]] says so and
consumers parse it as such — including when `--out -` sends it to
standard output, where it is still the artifact rather than a report
about one. What `index` *prints* about a run it wrote to a file is an
answer and follows the rules above.

---

## The Answer

### The answer is the properties the command established

A command's answer SHALL be a YAML mapping whose keys are the properties
the command established and whose values are those properties' values,
and SHALL contain nothing else.

**Rationale:** "nothing else" is doing the work. No header naming the
command — the caller typed it. No count of what follows — the reader can
see the collection. No summary — a summary of a mapping in view is a
second copy of it. No outcome line — the exit status carries the outcome,
and a process printing its own exit status tells the caller what they
already hold.

---

### A collection is a sequence of mappings, never formatted lines

WHERE an answer carries several of a thing, the CLI SHALL render them as
a sequence of mappings with one key per field.

**Rationale:** `path:line:column  slug` is a format a reader must be
told about and a consumer must split on punctuation nothing enforces. A
mapping per reference names every field it carries, and the fields a
later feature adds arrive as new keys rather than as a changed grammar.

---

### One shape per command

A command SHALL answer with the same shape whatever its arguments
select. A lookup by one identifier SHALL return the same sequence of
records as a lookup that matches many, holding one record.

**Rationale:** a caller should not have to branch on what it passed. A
bare record for an identifier and a sequence for a filter are two shapes
for one command, and a caller handed either must first find out which.

---

### An empty answer is an answer

WHERE a query matches nothing, the CLI SHALL render the empty collection
and SHALL exit `0`, as [[cli-exit-status-standards]] requires of every
query that answered.

**Rationale:** nothing found is a fact about the corpus, not a failure of
the invocation. Exiting non-zero would make a caller branch on the shape
of the world rather than on whether the command worked.

### A malformed question is not an empty answer

WHERE an argument names a value from a registered set (a rule
identifier, a document type, a tag qualifier) and the value is not in
that set, the CLI SHALL report the value and every registered name,
SHALL NOT render an answer, and SHALL exit `2`, the invocation-error
status of [[cli-exit-status-standards]].

**Rationale:** an unregistered value is a fact about the invocation, not
about the corpus. Answered as an empty collection or a clean report, a
typo is indistinguishable from a real "nothing found", and a gate that
passes a misspelt rule name to `check --only` passes forever.

---

### Order is the command's, and it is stable

Properties SHALL appear in the order the command established them, and a
collection SHALL appear in the order the command reports it.

**Rationale:** one ordering, chosen by the command, so two runs over
unchanged state produce byte-identical output and a diff of two runs is a
diff of what changed. This extends AC-SPEC-009.7's determinism from the
index to everything the surface prints.

---

## Nothing Is Shortened

A command SHALL render every value it reports in full.

A command SHALL NOT truncate a value, SHALL NOT elide one behind a
description of it, and SHALL NOT answer with an instruction to run the
same command again differently.

**Rationale:** agents read long documents; a line budget protects a
terminal width nobody is looking at. Worse, every form of shortening
turns one invocation into two — and the second costs a round trip, may
observe changed state, and can only be composed by a caller who decoded
the elision notice.

---

### A named view selects fields; it never truncates

WHERE a command offers a named view — a documented subset of the fields
its answer carries, selected by a flag or, where the command's
specification makes it the default, deselected by `--full` — the view
SHALL render every field it selects in full, for every record the full
answer holds, and the command's specification or these standards SHALL
list the fields it keeps.

A named view SHALL NOT be selected by `--json`, which changes only the
syntax of whichever view is rendered.

**Rationale:** the common question an answer serves may need a fraction
of its fields — "which documents cite this?" needs the path, not the byte
offset. A view names that fraction once, so the caller asks for it by
name and knows what it holds; it is not shortening, because every record
remains and every value kept is whole.

#### `refs`: the compact view and `--full`

`refs` answers with the compact view by default and with the full
answer under `--full`. The compact view groups references by citing
document: one entry per document, keyed by `path`, holding that
document's references.

| field of the full answer | in the compact view |
|---|---|
| `path` | kept, as the grouping key — one entry per document |
| `span.start_line` | kept, rendered `line` |
| `text` | kept |
| `origin` | kept only for a reference in frontmatter |
| `resolution.state` | kept only when the reference is not resolved |
| `slug` | dropped — the query names it |
| `anchor` | dropped — `text` carries it |
| `resolution.target_path` | dropped |
| `span.offset`, `span.length`, `span.start_column`, `span.end_line`, `span.end_column` | dropped |
| `anchor_resolution.heading` | dropped |

The properties the query established above the references are rendered
as in the full answer.

---

WHERE a command's design says its output belongs in a file, the command
SHALL write the file and report a property whose name ends `_path` and
whose value is that file's path.

**Rationale:** a per-command design decision, taken when the command is
specified, and never a threshold that engages when a value grows large.
A dynamic rule would make one invocation's shape depend on what it
happened to find, so a caller could not know whether they held a value or
a path without checking — and the check would be the elision notice the
rule above forbids.

---

## YAML Style

Valid YAML is not the bar; readable YAML is. Each rule below removes a
choice, so that two renderings of one value are the same bytes.

The output SHALL use block style throughout, indented two spaces per
level, with a sequence indented one level under the key introducing it.

---

An empty sequence SHALL be rendered `[]` and an empty mapping `{}`.

**Rationale:** the one exception to block style, and it is forced. A
block rendering of an empty collection is a key with nothing after it,
which reads as a key whose value is null — so an empty list, an empty
string and an unset key would be indistinguishable, and a caller could
not tell "nothing is missing" from "nobody looked".

---

A multi-line string SHALL be trimmed of leading and trailing whitespace
and rendered as a block scalar. Whitespace inside the value SHALL NOT be
touched.

**Rationale:** the block scalar is the whole reason for choosing YAML —
the value appears as itself, on its own lines. Trimming is what keeps the
block form always available: a first line indented past the block's own
indentation needs an explicit indicator to survive, and a last line
ending in whitespace is read back differently by different parsers.
Whitespace inside the value is content and is left alone.

---

WHERE a block scalar cannot carry a multi-line string faithfully, the
CLI SHALL double-quote it instead. It cannot when the value carries a
control character other than a line break or a tab.

**Rationale:** readability is why the block scalar is preferred and
correctness outranks it. A block's only structure is the line break, so a
carriage return folds into the break beside it and the value comes back
shorter than it went in.

---

A string SHALL be rendered plain where it round-trips plain, and
double-quoted otherwise. It does not round-trip plain when it is empty,
has leading or trailing whitespace, begins with a YAML indicator
character, contains a control character, contains `: ` or ` #`, ends
with `:`, or reads as a number, a boolean or a null.

**Rationale:** quoting everything is safe and unreadable; quoting nothing
is readable and wrong. This is the exact set of cases where a plain
scalar parses back as something other than what was written, so a reader
can tell a quoted value is quoted because it had to be.

---

Values SHALL be rendered as their own types: a number unquoted, a boolean
as `true` or `false`, an absent value as `null`.

---

The output SHALL NOT contain comments, anchors, aliases, explicit tags, a
document-start marker, or a document-end marker.

**Rationale:** each is a feature a reader must resolve before knowing
what the document says. An alias in particular defines a value somewhere
other than where it is read, which is the opposite of what this output is
for.

---

The output SHALL NOT wrap or fold a long line.

**Rationale:** folding changes the bytes of the value it carries, and any
wrapping decision depends on a terminal width that is not a property of
the data.

---

## What This Surface Does Not Adopt

Stated because the absence is a decision rather than an omission.

- **No result envelope and no `--full-result`.** A specs command reads a
  corpus and answers a question about it; it runs no plan whose steps a
  caller would audit, so there is no provenance layer under the answer to
  reveal. WHEN a command acquires one, it acquires the flag with it.
- **No per-property verdict shape.** This surface already has a finding
  vocabulary — a rule identifier, a severity, a path, a span and a
  message, generated from `schemas/finding.strata` — and a check run's
  answer is the findings it raised. A second verdict shape beside it
  would be two ways to report the same defect.

---

## Current conformance

Measured with a binary built from `main` at `12b2eda`
(`go build -C cli -o ../build/specs ./specs`), run against this
repository's corpus.

| command | prints today | to conform |
|---|---|---|
| `refs <slug>[#<anchor>]` | a YAML mapping of the properties the query established, the references a sequence of mappings beneath it; `--json` the same values in the same order | the compact view by default, today's answer under `--full` |
| `list` | a YAML mapping whose one key is the sequence of documents; `--json` the same values in the same order | conforms |
| `slug-exists <slug>` | a sentence, `adr-001: found at specs/adrs/adr-001-html-comments-for-instructions.md`; no `--json` | a mapping carrying `slug`, `declared` and `path`, with `--json` |
| `next-id <type>` | a YAML mapping carrying `number` and then `slug`, `number: 32` and `slug: adr-032`; `--json` the same values in the same order | conforms |
| `check [path]` | a YAML mapping of `checks` (only the checks the run was asked for) and `findings`; `--json` the same values in the same order | conforms |
| `feature`, `criterion` | a YAML mapping whose one key, `features` or `criteria`, holds a list of records whatever the arguments select, one identifier included; `--json` the same values in the same order | conforms |
| `links <path>` | the rewritten document | conforms — a document payload, written as itself |
| `index` | the artifact, or the bare path it wrote | the artifact is out of scope; the printed path is an answer |
| `headings <selection>` | a YAML mapping whose one key is the sequence of matches; `--json` the same values in the same order; `--emit` the matched sections as themselves, joined when several match | each match carrying its heading path; `--emit` refused for more than one match. Its rendering is the source of the `renderYAML` the other commands adopt |
| `c4 [--root <slug>]` | a YAML mapping of the C4 model; `--json` the same values in the same order | conforms — written against these standards, and renders through `renderYAML` unchanged |

Migration is per command and belongs to whichever feature next touches
it. Every command carries snapshot cases under `cli/testdata/cases/`,
which [[testing-standards]] requires be regenerated and reviewed when
observable output changes.

---

## References

- [[container-spec-cli]] — the container whose command surface these
  standards govern
- [[cli-exit-status-standards]] — the exit status beside every answer;
  this document governs only what is printed
- [[frd-spec-index]] — the command surface, the finding engine and the
  index artifact these standards defer to
- [[frd-spec-structure]] — the heading query, the first command
  specified against these standards
- [[frd-spec-anchor-references]] — the anchor query and checks, likewise
- [[testing-standards]] — the snapshot cases that hold observable output
  still
- [[adr-016]] — Strata as the one source for generated models, which is
  why a rendering is a view over a generated shape rather than a second
  declaration of it
