---
title: Swarm Command Output
description: What every swarm command prints — YAML by default and JSON
  under --json, the answer alone unless the full result is asked for, and
  the style rules that make both readable without a grammar to decode
slug: swarm-output-standards
type: standards
status: active
updated_at: 2026-08-30T12:00:00Z
version: "1.0"
---

# Swarm Command Output

These standards govern what every command [[container-swarm-cli]]
provides writes to its output stream. They apply to every subcommand
without exception, present and future, because what a command prints is
the Swarm Command Surface's business and not any subcommand's.

**The reader is an agent that knows which command it ran and nothing
else.** Not the command's author, not someone holding this document.
That reader is the point: the family of skills over this CLI is being
shrunk toward "run the command and read the output", and that is only
safe if the output needs no key.

**The whole method is to stop inventing a format.** The output is YAML,
so a reader who has met YAML once already knows how to read it and a
consumer already has a parser for it. There is no line grammar, no
reserved keyword set, no punctuation to reverse-engineer. Every
requirement below either says which YAML to emit or removes something a
reader would otherwise have to be told.

**Status.** Active, and met. [Conformance](#conformance) lists every
requirement with the invocation that measures it, and the
[worked examples](#worked-examples) below are what a built binary prints,
copied out of its output rather than composed.

---

## Renderings

### YAML by default, JSON under `--json`

Every command SHALL render its output as YAML, and SHALL render it as
JSON when `--json` is passed.

**Rationale:** two renderings, one flag, no third form. YAML is the
default because a human and an agent both read it without a schema, and
because multi-line text — which this CLI's values routinely carry —
survives it intact instead of arriving as `\n` inside a quoted string.
JSON stays available because a consumer that already parses JSON should
not have to acquire a YAML parser to keep working.

---

This SHALL hold for the answer and for the full result alike.

**Rationale:** `--json` selects a syntax and `--full-result` selects a
depth. Two independent choices, four combinations, all four defined. A
flag that also changed the depth would make the caller's two questions
one.

---

### The answer is what a command prints; the full result is asked for

A command SHALL print its answer, and SHALL print the full result only
when `--full-result` is passed.

**Rationale:** this is the requirement the rest of the document serves.
The rejected default was the whole report — a header, the plan
transcript, one line per fact with the invocation that established it
appended, one line per executed step with its digests and duration, the
summary, the resolution, and the outcome. Measured on `swarm status
--run fixture-run`, that form was 24 lines and 3,914 bytes to report a
run that has two beads in it; the same invocation's answer is 22 lines
and 765 bytes — an 80% reduction, and every byte removed is a byte about
how the command worked rather than about what it found.

The provenance is not lost. It is one flag away, and the flag is named
for what it returns.

---

## The Answer

### A clear invocation renders the properties and their values

WHEN a command's outcome is clear, its answer SHALL be a YAML mapping
whose keys are the properties the command established and whose values
are those properties' values, and SHALL contain nothing else.

**Rationale:** "nothing else" is doing the work. No header naming the
command — the caller typed it. No transcript — the caller did not ask
what was run. No summary — a summary of a mapping the reader can see is
a second copy of it. No outcome line — the exit status carries the
outcome, and a process that prints its own exit status is telling the
caller something the caller already has.

```yaml
branch: claude/bead-bonanza-fixture
trailers: |-
  Do re me
  fa so la ti
  do
```

---

### A property is the thing itself, never a verdict about it

A property's value SHALL be what the property holds. A command SHALL NOT
report a verdict about a property in the position where its value
belongs.

**Rationale:** the rejected form was `swarm manifest require` printing
`trailers = false`, which did not mean the manifest key `trailers` holds
`false` — it meant the key is not written. The name is the subject and
the value was a different quantity about it, with nothing on the line to
say so. `require` reports the same values `get` reports; whether they
were all present is the exit status, which is where a verdict belongs. It
costs `require` nothing, because it already reads the whole document in
one invocation and the values are in hand.

---

### Anything but a clear invocation renders a verdict per property

WHEN a command's outcome is not clear, its answer SHALL be a YAML
mapping whose keys are every property the command was asked about, each
mapping to `ok`, `value` and `issues`.

- `ok` is a boolean: whether this property was established.
- `value` is what it holds, and is `null` when `ok` is false.
- `issues` is a sequence of one-line strings, empty when `ok` is true.

**Rationale:** a caller whose invocation failed has exactly two
questions — which one failed, and why — and this shape answers both by
being read. The properties that succeeded stay in the document with
`ok: true`, so the reader sees the whole set rather than a filtered one
and does not have to diff against what they asked for to find the gap.

`ok`, `value` and `issues` are keys of **one** mapping. Written instead
as a sequence — `trailers:` then `- ok: false`, `- value: null` — they
parse as three separate single-key mappings in a list, and a reader has
to index by position to reach a field that has a name.

```yaml
findings_path:
  ok: false
  value: null
  issues:
    - the key is not written
    - "write it with: swarm manifest set --run fixture-run --key findings_path --value <value>"
trailers:
  ok: true
  value: |-
    Do re me
    fa so la ti
    do
  issues: []
```

---

The exit status SHALL be what selects between the two shapes: clear
(exit 0) renders the flat mapping, and every other status renders the
per-property mapping.

**Rationale:** a reader needs to know which shape they are holding
before they parse it, and the status is a value they already have and
already route on. The rejected alternative was one shape for both — the
per-property form always — which would put three lines around every
value in the common case and is exactly the verbosity these standards
exist to remove.

---

A failing property's `issues` SHALL name what would resolve it, when
anything would.

**Rationale:** the resolution is the actionable half and it has to reach
the caller. It goes inside the property rather than beside the mapping
because the answer is a mapping of properties and nothing else — a
sibling `resolution` key would be a tenth thing to know about, and would
collide the first time a command reports a property by that name.

---

### A refusal that names no property

WHEN a command refuses before establishing or naming any property, its
answer SHALL be a mapping carrying `issues` alone.

**Rationale:** a malformed invocation — a missing `--run`, two mode
flags at once — is about the invocation and not about any property, so
there is no property to key on. The shape stays decodable because its
`issues` maps to a sequence where a property would map to a mapping.

```yaml
issues:
  - swarm manifest get is scoped to one run; pass --run <run-epic-id>
```

---

### Order

Properties SHALL appear in the order the command established them.

**Rationale:** one ordering, chosen by the command, so two runs over
unchanged state produce byte-identical output and a diff of two runs is
a diff of what changed. Alphabetical was rejected: it would reorder a
command whose report has a deliberate sequence, and the mapping is not
read by scanning for a name.

---

## The Full Result

### `--full-result` renders `CommandResult`

WHEN `--full-result` is passed, the command SHALL render the
`CommandResult` [[contract-swarm-json]] declares, field for field.

**Rationale:** the full result already exists and already carries every
fact anyone has asked for — the mode, the outcome, the summary, the
resolution, each fact with the invocations that established it, each
plan step with its status, exit status, duration and stream digests, and
the transcript. Defining a second structure over the same facts would
create two things to keep in agreement and give a reader two documents
to learn. There is one structure, it is generated from
`schemas/swarm/output.strata`, and `--full-result` prints it.

---

The full result SHALL be rendered as YAML by default and as JSON under
`--json`, like every other output.

---

### Every subcommand carries `--full-result`

Every subcommand SHALL carry `--full-result`.

**Rationale:** every subcommand has an answer smaller than its report, so
every one of them has a use for the flag. A flag declared only by the
subcommands that happened to want it would leave a caller of any other
one with no way back to the provenance. It is declared by the surface
rather than by each subcommand, which is what makes "carried by every
subcommand" a property of the code; the surface's own test walks the
command tree and fails on a leaf that does not carry it.

---

### `--explain` and `--dry-run` imply it

WHEN `--explain` or `--dry-run` is passed, `--full-result` SHALL be in
effect whether or not it was passed.

**Rationale:** both flags ask what the command would do, and what it
would do is the plan — which lives in the full result's `invocations`
and `transcript`. In explain mode nothing is executed, so there are no
established properties and the answer alone would be an empty mapping.
Requiring a second flag to see the only thing these modes produce would
be a trap, and the implication costs a caller nothing: they asked to see
the working.

---

## Nothing Is Shortened

### No budget, no elision, no companion invocation

A command SHALL render every value it reports in full.

A command SHALL NOT truncate a value, SHALL NOT elide a value behind a
description of it, and SHALL NOT answer with an instruction to re-run
the same command differently.

**Rationale:** agents read long lines and long documents; a line budget
protects a terminal width nobody is looking at. Worse, every form of
shortening converts one invocation into two — and the second invocation
costs a round trip, may observe changed state, and can only be
constructed by a caller who understood the elision notice. A command
asked for a value returns the value.

---

### Bulk is surfaced as a `*_path` property

WHERE a command produces a body of output large enough that inlining it
would be wrong, the command SHALL write it to a file and report a
property whose name ends `_path` and whose value is that file's path.

**Rationale:** the two rules are about different actors and do not
conflict. This one binds the command *producing* bulk: it decides once,
at the point of production, that the artifact is a file. The rule above
binds the *renderer*: whatever a command does report, it reports whole.
A path is a value like any other — short, complete, and re-readable —
so surfacing one is not an elision.

---

## YAML Style

Valid YAML is not the bar; readable YAML is. Every rule here removes a
choice, so that two renderings of the same value are the same bytes.

---

The output SHALL use block style throughout, indented two spaces per
level, with a sequence indented one level under the key that introduces
it.

**Rationale:** one style, so a reader learns one set of quoting rules
rather than the different ones block and flow contexts impose. The cost
is accepted where it shows most — a list of argv vectors in the full
result renders as nested block sequences rather than as one line — and
it is the right trade for a form the caller opted into by asking for the
full result.

---

An empty sequence SHALL be rendered `[]` and an empty mapping `{}`.

**Rationale:** the single exception to block style, and it is forced.
There is no block rendering of an empty collection: it would be a key
with nothing after it, which is a key whose value is null. The rejected
text form had exactly this defect — `scope_missing = ` was
indistinguishable from an empty string and from an unset key.

---

A multi-line string SHALL be rendered as a block scalar, with the
chomping indicator that reproduces the value exactly: `|-` when it ends
with no newline, `|` when it ends with exactly one, `|+` when it ends
with more.

**Rationale:** a block scalar is the whole reason for choosing YAML —
the value appears as itself, on its own lines, instead of as `\n` inside
a quoted string. The indicator is not decoration: plain `|` clips to
exactly one trailing newline, so using it on a value that has none adds
a byte and the rendering no longer round-trips.

---

A multi-line string that a block scalar cannot carry exactly SHALL be
double-quoted instead. It cannot when the value carries a control
character other than a line break, or when its last content line ends in
white space.

**Rationale:** readability is the reason for the block scalar and
correctness outranks it. A block's only structure is the line break, so a
carriage return is folded into the break beside it and the value comes
back shorter. Trailing white space on the last content line is the
sharper case, because the specification permits it and readers disagree:
measured, `v: |-` over `a` then `b ` reads back `a\nb ` through PyYAML and
`a\nb` through goccy/go-yaml, which is the parser this repository's own
tests use. A renderer that cannot know which reader it is writing for does
not write the document they disagree about.

---

A string SHALL be rendered plain (unquoted) where it round-trips plain,
and double-quoted otherwise. It does not round-trip plain when it is
empty, has leading or trailing whitespace, begins with a YAML indicator
character, contains a control character, contains `: ` or ` #`, ends
with `:`, or reads as a number, a boolean or a null.

**Rationale:** quoting everything is safe and unreadable; quoting
nothing is readable and wrong. The list is the exact set of cases where
a plain scalar parses back as something other than what was written, so
a reader can tell that a quoted value is quoted *because it had to be*.

---

Values SHALL be rendered as their own types: a number unquoted, a
boolean as `true` or `false`, an absent value as `null`.

**Rationale:** a consumer parsing the YAML gets the type the fact
carries, and a human reading it sees the same thing. `null` is written
out rather than left as an empty value or `~` because the reader of a
failing answer is looking for it.

---

The output SHALL NOT contain comments, anchors, aliases, explicit tags,
a document-start marker, or a document-end marker.

**Rationale:** each is a feature a reader would have to resolve before
they know what the document says. An alias in particular makes a value
defined somewhere other than where it is read, which is the opposite of
what this output is for.

---

The output SHALL NOT wrap or fold a long line.

**Rationale:** folding (`>`) changes the bytes of the value it carries,
and any wrapping decision depends on a terminal width that is not a
property of the data. This is the line-budget rule, stated where the
renderer would otherwise be tempted.

---

## `require` and `validate`

The two commands answer different questions and both are needed.

**`validate` is stage-relative.** It checks the record against the rules
that apply at the stage the record has reached, or at the stage
`--stage` names. A key that belongs to no stage is never required, so
`validate` neither requires nor mentions it.

**`require` overrides optionality.** It fails on any named key that is
not written, whatever the stage and whatever the key's optionality. It
is how a caller says "I need this one, now" about a key the record does
not owe it.

Measured against the restored `core` fixture, whose `findings_path` key
is not written and belongs to no stage the run has reached:

```console
$ swarm manifest validate --run fixture-run
stage: dispatch
stage_source: derived from the keys the record carries
# exit 0 — findings_path is never mentioned

$ swarm manifest require --run fixture-run --key branch --key findings_path
branch:
  ok: true
  value: claude/bead-bonanza-fixture
  issues: []
findings_path:
  ok: false
  value: null
  issues:
    - the key is not written
    - "write each with: swarm manifest set --run fixture-run --key <key> --value <value>"
# exit 1
```

---

`require` checks presence, not validity. This is recorded here as an
[open question](#open-questions) rather than settled, because the two
readings imply different commands.

**Evidence:** `Require` in `cli/swarm/manifest/validate.go` performs two
tests per key — that the key is declared, and that `doc.Get(name)`
reports it present — and nothing else. `grep -c checkType
cli/swarm/manifest/validate.go` returns 3, and none of the three is
inside `Require`.

A property whose value is present but ill-typed therefore reports
`ok: true` today. Were the second reading adopted, that property would
report `ok: false` with the type error among its `issues`, and nothing
about the shape would change.

---

## Shape

**The answer needs no new type in `schemas/swarm/output.strata`.** Its
keys are property names, which no schema fixes — a command chooses them
— so the answer is a projection of `CommandResult.facts` rather than a
structure of its own. `--full-result` renders `CommandResult` itself.
Nothing here asks for a second declaration.

**Where the answer's failing form would be declared**, it is a map from
property name to a small closed struct, and it compiles today:

```strata
"""One property's verdict in a failing answer."""
struct PropertyVerdict {
  """Whether this property was established."""
  ok: Bool

  """What it holds. Absent when ok is false."""
  value: FactValue?

  """Why it is not established, and what would establish it. Empty when
  ok is true."""
  issues: [String]
}

"""The answer document a failing command renders."""
alias FailingAnswer: [String: PropertyVerdict]
```

Measured: that block, placed in a package alongside the real
`schemas/swarm/output.strata` and this repository's own
`strata.config.json` (`allowAdditionalProperties: false`), compiles for
both targets — `strata build . --target json-schema` and `--target go`
each exit 0. The emitted `FailingAnswer` is `{"additionalProperties":
{"$ref": "#/definitions/PropertyVerdict"}, "propertyNames": {"type":
"string"}, "type": "object"}` beside its description, and
`PropertyVerdict`'s `required` is `["ok", "issues"]` — `value` is
optional, which is what lets it be `null` on a failing property.

The `alias` matters: it emits a **bare root mapping**, which is the wire
form these standards require. A `struct` wrapping the same map nests it
one level deeper under a key nothing in the output would carry.

**What does not compile is a union, and it is the reason `FactValue`
looks the way it does.** A `@discriminator` union under
`allowAdditionalProperties: false` emits each variant *without* the
discriminator key, so the emitted JSON Schema rejects the documents its
own Go target produces. Measured on the `strata` build this repository
provisions, with a two-variant probe:

- the Go target's `MarshalJSON` writes `{"kind":"text", …}` — the
  discriminator is prepended to the variant's own object;
- the JSON Schema's `TextAnswer` is
  `{"additionalProperties": false, "properties": {"text": …},
  "required": ["text"]}` — no `kind`;
- the union's `allOf` branch is `then: {"$ref": "…/TextAnswer"}`, so a
  document carrying `kind` fails `additionalProperties: false`. Both
  variants, and by construction every variant.

This is upstream and unfixed. `FactValue` already carries the
workaround: a `kind` enum beside one optional field per kind, with a
coherence rule ("exactly one field, and it is the one the kind names")
enforced in code because the shape cannot state it.

**This is the constraint that blocks the one shape gap these standards
expose.** A manifest value that is a JSON object — `models`,
`standing_auth` — has no kind among the four, so it is reported as text
carrying its JSON and renders as an escaped one-line string in the
middle of otherwise readable YAML. Measured, in the answer to `swarm
manifest get --run fixture-run`:

```yaml
models: "{\"advanced\":\"opus\",\"frontier\":\"fable\",\"light\":\"haiku\",\"standard\":\"sonnet\"}"
```

The clean modelling is a fifth kind whose value is one of the concrete
structs `schemas/swarm/manifest.strata` already declares — `TierModels`,
`StandingAuth` — which is a discriminated union and does not compile
correctly. Strata offers no `Any` type to fall back on: a probe
declaring `value: Any` fails compilation with `Unknown type 'Any'`. So
an object-valued property waits on upstream, or on the same hand-rolled
workaround extended by one field per concrete type. **These standards do
not require it, and no proposal here should be read as compiling
cleanly.**

---

## Worked Examples

Every example below is what a binary built from this repository's source
(`cd cli && go build ./swarm`) printed, run against the restored `core`
end-to-end fixture, whose run epic is `fixture-run`. Each is copied out of
that output rather than composed, and the byte counts are `len(stdout)`
for the same invocation.

### 1. A value with newlines in it

The value was written through the real tracker:

```console
$ swarm manifest set --run fixture-run --key trailers --value '"Do re me\nfa so la ti\ndo"'
```

`swarm manifest get --run fixture-run --key branch --key trailers` — 79
bytes, exit 0:

```yaml
branch: claude/bead-bonanza-fixture
trailers: |-
  Do re me
  fa so la ti
  do
```

The block scalar is the whole reason for choosing YAML: the value appears
as itself, on its own lines, rather than as `\n` inside a quoted string.
`|-` and not `|` because the value ends with no newline, and `|` would
clip to exactly one — adding a byte the value does not carry.

---

### 2. The same command, one key missing

`swarm manifest get --run fixture-run --key trailers --key findings_path`
— 258 bytes, exit 1:

```yaml
findings_path:
  ok: false
  value: null
  issues:
    - the key is not written
    - "write it with: swarm manifest set --run fixture-run --key findings_path --value <value>"
trailers:
  ok: true
  value: |-
    Do re me
    fa so la ti
    do
  issues: []
```

The property that succeeded stays in the document with `ok: true`, so the
reader sees the whole set rather than a filtered one. The resolution is
inside the property it would resolve.

---

### 3. The whole manifest

`swarm manifest get --run fixture-run` — 32 lines and 1,218 bytes, exit 0:

```yaml
admitted:
  - fixture-run.1
  - fixture-run.2
base_sha: c004e53c76b2459e9c8f89c17605adc5c2b27b2e
branch: claude/bead-bonanza-fixture
checks_baseline: |
  # Checks baseline

  Everything passed on the untouched tree.
checks_baseline_status: 0
checks_ids:
  - format
  - vet
  - build
  - test-go
collision_edges: .beads/plans/fixture.edges.jsonl
concurrency_cap: 20
effort: unrecorded
excluded: []
ledger:
  - the fixture's tracker restores from a committed export — proved by `bd import`, observed the same ids and timestamps as the export carries
models: "{\"advanced\":\"opus\",\"frontier\":\"fable\",\"light\":\"haiku\",\"standard\":\"sonnet\"}"
moved: []
partition_preview: accepted — no collision pairs co-ready
review_budget: 2
review_passes: 0
standing_auth: "{\"admit_review_findings\":{\"source\":\"default\",\"value\":false},\"merge_to_default\":{\"source\":\"default\",\"value\":false},\"move_foreign_beads\":{\"source\":\"default\",\"value\":false},\"pull_request\":{\"source\":\"default\",\"value\":\"draft\"},\"review_budget\":{\"source\":\"default\",\"value\":2},\"tier_escalation\":{\"source\":\"default\",\"value\":\"once\"}}"
target: fixture-run.target
trailers: |-
  Do re me
  fa so la ti
  do
```

Three things are visible here: `checks_baseline` is a multi-line document
and reads as one, with `|` because it ends in exactly one newline;
`excluded` and `moved` are empty *lists* rather than possibly-empty
strings; and `models` and `standing_auth` are objects the four fact kinds
cannot carry, so they arrive as escaped JSON. The third is the shape gap
the [Shape](#shape) section measures, and the only requirement here that
is met by a workaround rather than outright.

---

### 4. A command whose report is its answer

`swarm status --run fixture-run` — 22 lines and 765 bytes, exit 0. The
rejected text form was 24 lines and 3,914 bytes for the same run:

```yaml
stall_after: 30m0s
observability: whether a subagent is still working cannot be observed from outside the harness; this reports the standing worktree count, the last commit's age and the ready front's depth in its place
worktrees_standing: 1
dispatch_worktrees_standing: 0
last_commit_sha: ac4a37c7aadfbc9a0f7af911b4b95558e3989eda
last_commit_subject: fixture-run.1 P1 the first bead of the run
last_commit_at: 2026-01-01T01:00:00Z
last_commit_age_seconds: <age-seconds>
ready_front_depth: 0
completion_total: 2
completion_done: 1
completion_blocked: 0
completion_claimed: 1
completion_open: 0
completion_deferred: 0
completion_unknown: 0
scope_additions: []
scope_missing: []
scope_excluded: []
dispatches:
  - "dev-fixture-1: in progress: fixture-run.2"
conflicts: []
```

The saving is larger than the line count suggests: in the text form each
of those lines carried the argv of every invocation behind it, appended
after `    <- `, and the same commands repeated across lines. On the real
run epic `agent-plugins-bd97` that form measured 35,856 bytes for 25
lines. Not one of those bytes was part of the answer.

`last_commit_age_seconds` is measured against the real clock, so it is
the one value here that moves between runs and the one shown as a
placeholder.

---

### 5. The full result

`swarm manifest get --run fixture-run --key branch --key trailers
--full-result` — 1,018 bytes, exit 0:

```yaml
contract_version: "2.0"
command: swarm manifest get
run: fixture-run
mode: execute
outcome: clear
status: 0
summary: 2 key(s) written
resolution: ""
facts:
  - name: branch
    established: true
    reason: ""
    value:
      kind: text
      text: claude/bead-bonanza-fixture
    invocations:
      - - bd
        - show
        - fixture-run
        - "--json"
  - name: trailers
    established: true
    reason: ""
    value:
      kind: text
      text: |-
        Do re me
        fa so la ti
        do
    invocations:
      - - bd
        - show
        - fixture-run
        - "--json"
invocations:
  - step: read
    argv:
      - bd
      - show
      - fixture-run
      - "--json"
    dir: ""
    status: ran
    reason: ""
    exit_status: 0
    duration_ms: <duration>
    stdout_digest: sha256:<digest>
    stderr_digest: sha256:<digest>
    error: ""
transcript:
  - bd show fixture-run --json
```

Three values move between runs and are shown as placeholders rather than
as the figures one run happened to produce: the elapsed time, and the two
stream digests, which cover the tracker's own output. The committed
end-to-end artifacts normalize the same three and assert every other
byte.

That is the `CommandResult` [[contract-swarm-json]] declares, field for
field, with nothing added. It is verbose, and that is correct: the caller
asked for it. `"--json"` is quoted because a plain scalar may not begin
with an indicator character, and `-` is one.

---

### 6. An explained plan

`swarm manifest get --run fixture-run --key branch --explain` — 361
bytes, exit 0. `--explain` implies `--full-result`, which is what makes
the plan visible: in explain mode nothing runs, so `facts` is empty and
the answer alone would be `{}`.

```yaml
contract_version: "2.0"
command: swarm manifest get
run: fixture-run
mode: explain
outcome: clear
status: 0
summary: explained; nothing was executed
resolution: ""
facts: []
invocations: []
transcript:
  - "# read fixture-run's manifest"
  - "# 1. read — read fixture-run's run manifest [read; stop on failure; safe to re-run]"
  - bd show fixture-run --json
```

---

### 7. A refusal about the invocation

`swarm manifest get --key branch`, with no `--run` — 78 bytes, exit 3,
having established nothing:

```yaml
issues:
  - swarm manifest get is scoped to one run; pass --run <run-epic-id>
```

---

## Conformance

Each row was measured against a binary built from this repository's
source (`cd cli && go build ./swarm`), run against the restored `core`
end-to-end fixture. The "enforced by" column names the check that fails
when the row stops holding, because a table nobody runs is a list of
properties enforced by nobody.

| requirement | measurement | enforced by |
|---|---|---|
| YAML by default | every invocation in [Worked Examples](#worked-examples) | `TestSwarmCases`, over the committed artifacts |
| JSON under `--json` | `manifest get --run fixture-run --key branch --json` → `{\n  "branch": "claude/bead-bonanza-fixture"\n}` | `TestTheAnswerAsJSONCarriesTheSameValue` |
| the answer is the default | example 1: the answer, not the report | `TestAClearAnswerIsThePropertiesAndTheirValues` |
| a clear answer is properties and values | example 1 | `TestAClearAnswerCarriesTheEstablishedPropertiesInReportOrder` |
| a property is never a verdict | `manifest require … --key branch --key findings_path` reports `claude/bead-bonanza-fixture`, not `true` | `TestRequireNamesEveryAbsentKey` |
| a failing answer is per-property `ok`/`value`/`issues` | example 2 | `TestAFailingAnswerCarriesAVerdictPerProperty` |
| a refusal that names no property carries `issues` alone | example 7 | `TestARefusalThatNamesNoPropertyCarriesIssuesAlone` |
| properties appear in report order | example 4 — `stall_after` first, `conflicts` last | `TestAClearAnswerCarriesTheEstablishedPropertiesInReportOrder` |
| `--full-result` renders `CommandResult` | example 5 | `TestFullResultRestoresTheReport`, `TestWhatTheSurfaceEmitsConformsToTheEnvelopeSchema` |
| every subcommand carries `--full-result` | `manifest require --run fixture-run --key branch --full-result` → exit 1 with the answer, no flag error | `TestEverySubcommandInTheTreeCarriesTheContractFlags` |
| `--explain` and `--dry-run` imply `--full-result` | example 6; `manifest set --dry-run` in `manifest-set-dry-run` | `TestExplainAndDryRunImplyTheFullResult` |
| nothing is truncated, elided or wrapped | example 4's `observability`, one line of 185 characters | `TestALongValueIsRenderedWhole` |
| bulk surfaces as a `*_path` property | `findings_path` follows it | not enforced — see [Open Questions](#open-questions) |
| empty collections are distinguishable | example 3's `excluded: []` and `moved: []` | `TestAnEmptyCollectionIsDistinguishableFromNothing` |
| multi-line values are block scalars that round-trip | examples 1 and 3 | `TestAMultiLineValueIsABlockScalarThatRoundTrips` |
| a string is quoted exactly where it has to be | example 5's `"--json"`, example 3's `models` | `TestAStringIsQuotedExactlyWhereItHasToBe` |
| block style throughout, two spaces per level | example 5's nested argv sequences | `TestTheOutputIsBlockStyleThroughout` |
| no comments, anchors, aliases, tags or document markers | every example | `TestTheDocumentCarriesNothingToResolve` |
| values render as their own types | example 3: `review_passes: 2`, `excluded: []` | `TestAValueIsRenderedAsItsOwnType` |
| both syntaxes carry the same values in the same order | — | `TestTheJSONRenderingKeepsTheOrderTheCommandReported`, `TestTheJSONFormCarriesTheSameFactsAsTheYAMLForm` |

Every named test lives in `cli/swarm/cli` except `TestSwarmCases`, which
is in `cli/swarm/e2e` and replays the committed artifacts under
`cli/testdata/swarm-cases` against a real tracker and real git.

---

## Open Questions

- [ ] **Does `require` mean "is set" or "is set to a valid value"?**
      The command was described as requiring "a valid value is set",
      which is not what it does. It means the first: `Require` tests
      declaredness and presence and calls no type check. If the second is
      wanted, the answer's `issues` gains a second population — a key
      written with a value of the wrong type — and `require` starts
      overlapping `validate` in a way that needs saying. Not decided
      here; the shape would not change either way.

- [ ] **Should an object-valued property render as nested YAML?**
      It cannot today: the four fact kinds carry no object, so
      `models` and `standing_auth` arrive as escaped JSON strings. The
      fix is a fifth kind, which is a discriminated union, which does
      not compile correctly under this repository's schema
      configuration. See [Shape](#shape).

- [ ] **What enforces "bulk surfaces as a `*_path` property"?**
      Nothing does. It binds a command at the point where it decides to
      produce bulk, and no check can tell a value that should have been a
      file from one that should not. `collision_edges` is a path not
      named for one, and `checks_baseline` is inlined bulk — 18,715 bytes
      over 339 lines on the run epic `agent-plugins-bd97`. Both predate
      this standard and neither is renamed here, because a manifest key
      rename is a breaking change to [[contract-swarm-json]].

**Settled here, and recorded because the answer is not obvious:**
**which stream carries a failing answer.** Both shapes go to standard
output. A caller redirecting stdout captures the answer whether the
command succeeded or failed, and routes on the exit status to know which
shape they are holding — which is exactly what a failing answer put on
the error stream would take away from them. Standard error carries what
the process itself could not do: an unparseable invocation cobra refuses
before any command runs.

---

## References

- [[container-swarm-cli]] — the container whose commands these
  standards govern
- [[contract-swarm-json]] — the `CommandResult` structure
  `--full-result` renders, generated from
  `schemas/swarm/output.strata`
- [[frd-swarm-core]] — the acceptance criteria the emitted properties
  come from
- [[adr-016]] — Strata as the one source for schemas and generated
  models, which is why the shape discussion is about what compiles
- [[module-bead-bonanza]] — the consumer whose skills read this output
- [YAML 1.2.2](https://yaml.org/spec/1.2.2/) — the specification the
  style rules above are a subset of
