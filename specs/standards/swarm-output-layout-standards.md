---
title: Swarm Command Output Layout
description: The layout every swarm command's output keeps in both its
  text and JSON forms, so an agent that knows only which command it ran
  can read the result without investigating the format
slug: swarm-output-layout-standards
type: standards
status: draft
updated_at: 2026-08-29T00:00:00Z
version: "0.1"
---

# Swarm Command Output Layout

These standards govern the output of every command
[[container-swarm-cli]] provides, in both forms it renders — the text
form and the `--json` form — and the two streams it renders them on.
They apply to every subcommand without exception, present and future,
because the layout is the Swarm Command Surface's and not any
subcommand's.

**The reader they are written for is an agent that knows which command
it ran and nothing else.** Not the command's author, not someone holding
this document, not someone who has read the source. That reader is the
whole point: the family of skills over this CLI is being shrunk toward
"run the command and read the output", and that is only safe if the
output explains itself. Where this document has to choose between a
shorter rendering and one a cold reader can decode, it chooses the
second.

**Why a standards document.** The shape of the JSON is already fixed by
[[contract-swarm-json]], generated from Strata, and a second document
restating shape would be a second source. What is not fixed anywhere is
the *reading contract* — which kinds of line exist, how a reader tells
them apart, and what may be assumed from each. That is a cross-cutting
constraint over every command rather than a requirement of one feature,
which is what this type is for. An ADR was considered and rejected: this
is not one decision with alternatives but a set of obligations each
command must satisfy, and the decisions inside it are recorded as
rationales against the requirement they justify.

**Status.** Draft. The requirements below are the target layout; the
renderer at the time of writing satisfies some and not others. The
[Conformance](#conformance) section states which, with the measurement
behind each, so the gap is a checklist rather than a surprise.

---

## The Defect This Corrects

The text form today has at least seven kinds of line and distinguishes
them only by punctuation a reader has to reverse-engineer. This is the
whole of one real report, measured:

```console
$ swarm manifest require --run agent-plugins-bd97 --key trailers --key branch
swarm manifest require — run agent-plugins-bd97 (execute)
bd show agent-plugins-bd97 --json
branch = true    <- bd show agent-plugins-bd97 --json
trailers = false    <- bd show agent-plugins-bd97 --json
read: ran exit 0 472ms sha256:72f3b0a38d202ddef4dfff826b0b2b371fb7e93ccab52ef3c4a9a2f89f732cbc sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855    <- bd show agent-plugins-bd97 --json
summary: trailers: not written
what would resolve this: write each with: swarm manifest set --run agent-plugins-bd97 --key <key> --value <value>
outcome: attention (exit 1)
```

Eight lines, seven different grammars, and nothing on any line says
which grammar it is in. A reader has to induce all of the following:

- Line 1 is a header because it is first, and for no other reason.
- Line 2 is a tool invocation that was run, because it looks like one.
- Lines 3–4 are facts because they contain ` = `.
- Line 5 is a step record because it contains `: ` and its first word
  happens not to be one of the three reserved phrases below.
- Lines 6 and 8 are reserved single words. Line 7 is a reserved phrase
  of five words, so a reader recognising the reserved set has to
  recognise it as a set of phrases rather than of tokens.
- The `    <- …` suffix means two different things. On lines 3–4 it is
  **provenance** — the invocation that established the value. On line 5
  it is **identity** — the argv the step itself ran. One glyph, two
  relations.
- `trailers = false` does not mean the manifest key `trailers` holds
  `false`. It means the key is not written. The name is the subject and
  the value is a different quantity about it.
- `sha256:e3b0c442…` is the digest of the empty string, so that step
  wrote nothing to standard error. Nothing in the line says so, and no
  reader recovers it without knowing the constant.

Every requirement below removes one of those inductions.

---

## Reading Contract

### The closed set of line kinds

The text form SHALL consist only of lines drawn from the kinds in this
table, and each line SHALL begin at column 0 with the keyword naming its
kind, followed by whitespace.

**Rationale:** the kind of a line is the first thing a reader needs and
the last thing the current form supplies. A keyword is decidable by a
human at a glance and by a parser with `split(maxsplit=1)`, needs no
lookahead, and cannot be confused with content because content never
appears at column 0.

| keyword | kind | what it asserts |
|---|---|---|
| `command` | header | which invocation produced this report, and in which mode |
| `source` | source declaration | an invocation, named, so facts can cite it once |
| `step` | step record | one planned invocation and what became of it |
| `argv` | step argv | the exact argv the step above ran or would run |
| `fact` | fact | one thing the command established, and its value |
| `from` | provenance | which sources established the fact above |
| `summary` | summary | what the command concluded, in one line |
| `resolve` | resolution | what would resolve an attention outcome |
| `outcome` | outcome | the verdict and the exit status it implies |

There SHALL be no tenth kind, and no line outside this set.

**Rationale:** a closed set is what lets a reader treat an unrecognised
first word as a defect rather than as something they have not learned
yet. It is also what makes the correspondence with the JSON form total
rather than approximate.

---

### What a reader may assume

For a reader who knows only which command they ran:

| element | may assume | may NOT assume |
|---|---|---|
| `command` | this is the invocation whose result follows, and the mode named is the mode that actually ran | that re-running it now gives the same result |
| `source` | the argv named is exactly what was run, quoted for POSIX `sh`, and can be pasted | that it was run once; a source may be cited by many facts |
| `step` | the command intended to run this, in this position; the status word says what became of it | that a step which ran changed anything, or that its digests cover anything the reader holds |
| `argv` | this is the resolved argv of the `step` immediately above | that an argument shown was a literal in the plan; it may have come from an earlier step |
| `fact` | the command established this, and the value is in the kind the line names | that an absent fact means a false or empty value — absence is reported, never implied |
| `from` | every name here is declared elsewhere in this same report, as a `step` or a `source` | that the invocation was run *for* this fact; one invocation commonly establishes many |
| `summary` | this is the command's own reading of what it found | that it is exhaustive; it is a reading, and the facts are the evidence |
| `resolve` | doing this is expected to clear the condition | that it is the only way, or that it is safe to do unattended |
| `outcome` | this is the verdict, and the exit status is derived from it and cannot disagree | that `clear` means nothing needs attention anywhere — it means *this command's* verdict is clear |

**Rationale:** the second column is the half that stops a reader
inventing a meaning. Every question that prompted this document was of
the form "may I assume X", and the answers differed: the `<- …` suffix
*is* provenance on a fact line and *is not* on a step record; the
`set-guard: ran exit 0 …` lines *are* process steps; and a summary
*may not* span more than one line. A table that only listed what may be
assumed would have answered one of the three.

---

### The three depths

A reader SHALL be able to stop at any of three depths and be correct
about what they read:

1. **`outcome` alone** — route on it. This is what a coordinator does,
   and it is why the exit status is a contract.
2. **`summary` and `outcome`** — know what happened. Enough for a log
   line or a hand-off.
3. **`fact`, `from`, `step`, `argv`** — check the verdict against its
   evidence, and re-run any single invocation.

A reader at depth 1 or 2 SHALL reach what they need by reading the
report's last three lines, whatever the report contains.

**Rationale:** this is the requirement the current form breaks worst. A
single fact carrying a multi-line value pushed the `outcome` line to
line 344 of one measured report (see [Conformance](#conformance)), and
its `summary` to line 343 — reachable by a reader who reads backwards,
and by nobody who reads forwards. One line per element is what makes
"the last three lines" a place rather than a hope.

---

## Line Requirements

### Order

A report SHALL emit its lines in this order and no other: one `command`;
then every `source`; then every `step`, each followed by its `argv`;
then every `fact`, each followed by its `from`; then one `summary`; then
one `resolve` where the outcome is `attention`; then one `outcome`.

**Rationale:** a fixed order makes position redundant with the keyword
rather than load-bearing, which is what lets a reader who has lost their
place recover it from the line in front of them. It also puts the
verdict where "the last three lines" can find it, and the evidence
before the conclusion it supports.

---

### Every line is one line

No element SHALL span more than one line, and no line SHALL carry more
than one element.

**Rationale:** this is the direct answer to "may a summary contain
multiple lines" — **no**, and not only a summary. A continuation line
has no keyword, so admitting continuations means either a tenth marker
to learn, or a rule that a line without a keyword belongs to the line
above — which fails the moment a value's own text begins with something
that looks like a keyword, which the manifest's `checks_baseline` value
already does. Forbidding continuations also bounds the report's length,
which is what keeps `outcome` reachable.

---

WHEN a value does not fit within the line budget, the command SHALL
render its kind, its size, and an invocation that yields it whole,
instead of the value.

**Rationale:** eliding is the command's decision to make, not the
terminal's. [[contract-swarm-json]] forbids output whose meaning depends
on the width of the producing process's terminal, so the budget is a
constant in this document rather than `$COLUMNS`.

---

A line SHALL NOT exceed 200 characters.

**Rationale:** a number, because "reasonable" is not testable. 200 is
wide enough for a manifest value, a sha-prefixed digest pair and a
typical argv, and narrow enough that a reader scanning a column of
keywords is not scrolling. A command that cannot fit an element in 200
characters elides it under the requirement above.

---

### `command`

The first line of every report SHALL be a `command` line, and it SHALL
name the invocation path, the run scope where the command is
run-scoped, and the mode that ran.

---

The mode named SHALL be the mode that ran.

**Rationale:** not a restatement. A refusal taken before the mode is
resolved currently reports the zero-valued mode, so a command with no
mode flag at all reports `(explain)` — see [Conformance](#conformance).
A header that misreports the mode is worse than one that omits it,
because a reader takes it as evidence that nothing was executed.

---

### `source` and `step`

A `step` line SHALL carry, in fixed order: the step's position in the
plan, its name, its status, and the fields its status defines.

WHEN a step's status is `ran`, the line SHALL carry its exit status, its
duration, and a digest of each of its two streams, separately.

WHEN a step's status is `refused` or `skipped`, the line SHALL carry the
reason in one clause, and SHALL NOT carry an exit status or a duration.

**Rationale:** a refused step has no exit status, and rendering a zero
for it invites a reader to conclude it succeeded.

---

Every `step` line SHALL be followed immediately by exactly one `argv`
line carrying the resolved argv of that step.

**Rationale:** a step and its argv are one record split over two lines
rather than one line of two grammars. Splitting them is what frees the
`from` keyword to mean provenance and nothing else — which is the
ambiguity that prompted this document.

---

WHEN a stream is empty, its digest field SHALL read `empty`.

**Rationale:** every step record measured while writing this document
carried the same `stderr` digest, beginning `sha256:e3b0c44298fc`. That
is the digest of the empty string — `printf '' | sha256sum` prints
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` — so
every one of them meant "the tool wrote nothing to standard error". No
reader recovers that without knowing the constant, and the fact is
worth reporting.

---

A digest SHALL render as `sha256:` followed by the first 12 hexadecimal
characters of the digest, and the JSON form SHALL carry it in full.

**Rationale:** the digest exists to detect a step whose output changed
between two runs; 48 bits does that at a glance. Two full digests are
142 characters, which is most of a line's budget spent on something
nobody reads character by character. The full value stays one `--json`
away for anything that compares them mechanically.

---

An invocation cited by a `from` line SHALL be declared exactly once in
the report, as either a `step` (by name) or a `source` (by name).

**Rationale:** repeating an argv per fact is what makes the largest
report unreadable: one measured `swarm status` report was 24 lines and
33,329 bytes, of which nine fact lines carried the *same* 53 `bd list
--parent … --json` invocations — 477 copies of 53 distinct commands.
Declaring each invocation once and citing it by name is the same
information in an order of magnitude less text.

---

WHEN a command's invocations are plan steps, it SHALL cite them by step
name and SHALL NOT emit `source` lines for them.

**Rationale:** the step block already carries the argv. A `source`
declaration beside it would be a second copy to keep in agreement, which
is the defect this whole surface is built to avoid.

---

### `fact` and `from`

A `fact` line SHALL name the fact, then ` = `, then the value's kind,
then its rendering.

The kinds are exactly `text`, `count`, `flag` and `list` — the four
[[contract-swarm-json]] declares — plus `absent`, which is not a fifth
value kind but the rendering of a fact that was not established.

**Rationale:** the four value kinds are already fixed by
[[contract-swarm-json]] and enforced where a value is built. The text
form currently discards the kind, which is why `checks_ids = format,
vet, build` is indistinguishable from a single string containing commas,
and why `scope_missing = ` (an empty list, measured) is
indistinguishable from an empty string and from a missing value. Naming
the kind costs one word and removes three ambiguities.

---

A `list` value SHALL render as its length followed by its compact JSON.

**Rationale:** joining a list with `, ` is unrecoverable the moment an
item contains a comma, and the manifest's `ledger` key — measured — is
seven prose entries every one of which does. JSON is the only rendering
every reader already parses, and the length in front means a reader who
stops at the first line still knows how many there are.

---

An unestablished fact SHALL render with the kind `absent` and a reason,
and SHALL NOT render as an empty value of its type.

---

A fact's name SHALL name what was measured, not what it was measured
about, wherever those differ.

**Rationale:** two measured cases. `swarm manifest require` reports
`trailers = false`, where the name is a manifest key and the value is a
different quantity about it — whether the key is written; a reader who
takes the line at face value concludes the key holds `false`. `swarm
manifest set --dry-run --key retro_filed --value 0` reports
`retro_filed = 8`, which is the value the key holds *now*, under a
summary that says what *would* be written; the reader has to know which
of the two it is. Naming the quantity — `trailers written`,
`retro_filed before` — removes both without changing what is reported.

---

Every `fact` line SHALL be followed immediately by exactly one `from`
line naming the sources that established it.

**Rationale:** [[frd-swarm-core]] already requires that every reported
fact names the command a reader can re-run to reproduce it, and the
surface refuses to render a fact with no invocation behind it. This
document fixes only where that goes: on its own line, under the fact,
introduced by a keyword that means provenance and never means anything
else.

---

### `summary`, `resolve`, `outcome`

Every report SHALL carry exactly one `summary` line.

**Rationale:** the summary is optional today, and a report with no
summary makes a reader who wanted depth 2 fall through to the facts.
There is always something to say; where a command found nothing, saying
so is the summary.

---

WHEN the outcome is `attention`, the report SHALL carry exactly one
`resolve` line, and it SHALL name an action, not restate the condition.

**Rationale:** the surface already refuses to render an attention result
with an empty resolution, which is necessary and not sufficient.
Measured, exit 1:

```text
$ swarm manifest set --run agent-plugins-bd97 --dry-run --key review_passes --value 3
swarm manifest set — run agent-plugins-bd97 (dry-run)
summary: plan: step "set-guard": manifest/beads: refusing the write — "review_passes" would be 3 against a budget of 2
what would resolve this: plan: step "set-guard": manifest/beads: refusing the write — "review_passes" would be 3 against a budget of 2
outcome: attention (exit 1)
```

The two strings are identical, so the check passed and the caller was
told nothing they did not already have. A resolution is an imperative —
a command to run, or an action to take — and here it would be *raise
the review budget, or accept the run's current one*.

---

The last line of every report SHALL be an `outcome` line naming the
outcome and the exit status derived from it.

The four outcomes and their statuses are fixed by [[frd-swarm-core]],
which is normative; they are restated here because a reader who reaches
this document by reading output has the `outcome` line in front of them
and needs to know what it means:

| outcome | exit | what the caller should do |
|---|---|---|
| `clear` | 0 | the command answered; its verdict is clear |
| `attention` | 1 | the command answered; what it found needs attention, and the `resolve` line says what would fix it |
| `stop` | 2 | the command found a condition the caller cannot repair in place |
| `malformed` | 3 | the invocation itself was wrong — a missing scope, an unknown flag, a bad argument |

---

## Streams and the Bare Form

### Which stream carries what

**Standard output carries what the caller asked for. Standard error
carries why they cannot have it.** Every rule below follows from that
one sentence, and there is no case it does not decide.

---

A command with no bare form SHALL render its report on standard output
at every exit status, including a non-zero one.

**Rationale:** for such a command the report *is* what the caller asked
for, including when the report is of a refusal.
[[contract-swarm-json]] guarantees that a consumer parsing `--json`
never has to fall back to reading text to learn why; moving refusals to
standard error would force that consumer to merge two streams to find
the envelope, and merging streams is the hazard [[frd-swarm-core]]
forbids elsewhere. The alternative — errors on standard error, as a
shell tool conventionally does — was rejected on that ground; the cost
accepted is that a caller who wants only failures must route on the exit
status rather than on a stream, which they must do anyway.

---

### When a subcommand may declare a bare form

A subcommand MAY declare a bare form when, and only when, all three
hold:

1. **Every clear invocation has exactly one answer** — one value, not
   one per line. Several keys answer with one object; they do not answer
   with several lines a reader would have to split back apart.
2. **The answer is smaller than the report**, and a caller could
   otherwise obtain it only by parsing the report.
3. **The answer is retrieved rather than derived.** A lookup returns
   something the command read. A command whose output is its own
   conclusion — a verdict, a count it computed, a status — has no answer
   smaller than its report and SHALL NOT declare a bare form.

**Rationale:** test 3 is the one that decides the cases the other two
leave open. `swarm manifest get` is a lookup and is bare. `swarm status`
computes a verdict from twenty measurements; its report is its answer,
so a bare form would have to invent one. `swarm manifest require` and
`swarm manifest validate` answer with a verdict already carried by the
exit status, so a bare form would be a second copy of it.

---

WHERE a subcommand declares a bare form, it SHALL carry `--full-result`,
and under `--full-result` it SHALL behave exactly as a subcommand with
no bare form.

WHILE a bare-form subcommand's outcome is clear, standard output SHALL
carry the answer alone, terminated by exactly one newline, and standard
error SHALL be empty.

IF a bare-form subcommand's outcome is not clear, THEN standard output
SHALL be empty and standard error SHALL carry the full report.

The exit status SHALL be the same whether or not the bare form is in
effect.

**Rationale:** the split is what lets `$(swarm manifest get …)` capture
a value with no parser, while a caller who gets nothing still reads why
without having polluted the value they were capturing. The status
carries the distinction either way, so nothing is lost by the report
moving off standard output. Measured, both halves:

```console
$ swarm manifest get --run agent-plugins-bd97 --key branch 2>/dev/null
claude/bead-bonanza-agent-plugins-ih18-2yjc8j
$ swarm manifest get --run agent-plugins-bd97 --key trailers 2>/dev/null; echo "exit=$?"
exit=1
```

---

A bare answer SHALL carry no provenance, no framing and no trailing
report.

**Rationale:** provenance is one flag away under `--full-result`, and a
caller substituting the command into a shell wants the value and nothing
else. A stored string is printed without its JSON quoting, so a value
carrying whitespace, a newline or a leading dash survives command
substitution; every other shape is printed as its compact JSON. Under
`--json` the answer is JSON, so a string arrives quoted:

```console
$ swarm manifest get --run agent-plugins-bd97 --key branch --json
"claude/bead-bonanza-agent-plugins-ih18-2yjc8j"
```

---

### Refusals the surface did not build

Every exit SHALL be rendered by the surface, including a refusal raised
by argument parsing before any subcommand runs.

**Rationale:** [[contract-swarm-json]] guarantees a conforming result on
every exit status. An unknown flag currently escapes that guarantee
entirely — see [Conformance](#conformance) — which means a consumer
routing on `--json` gets unparseable text at exit 3 and has no way to
learn why except by reading it.

---

## The JSON Form

Shape is [[contract-swarm-json]]'s and is generated from
`schemas/swarm/output.strata`; this document adds only what layout
requires.

Every text line kind SHALL correspond to exactly one place in the JSON
envelope:

| text keyword | JSON |
|---|---|
| `command` | `command`, `run`, `mode` |
| `source`, `step`, `argv` | one element of `invocations` |
| `fact`, `from` | one element of `facts`, with its `invocations` |
| `summary` | `summary` |
| `resolve` | `resolution` |
| `outcome` | `outcome`, `status` |

The text form SHALL be derivable from the JSON form of the same
invocation, with no fact, kind or relation present in one and absent
from the other.

**Rationale:** the container already renders both from one value, so the
facts cannot disagree. What this adds is that the *kinds* cannot either:
a reader who has learned one form has learned both, and a test can
assert the correspondence rather than a reviewer eyeballing it.

---

WHERE a value is elided in the text form under the line budget, the JSON
form SHALL carry it whole.

**Rationale:** the elision exists so a human-and-agent-readable report
stays readable. Nothing about a machine-readable form wants it.

---

### What this asks of the envelope

Two of the requirements above cannot be met by the renderer alone, and
both are breaking changes to [[contract-swarm-json]] under its own
policy. They are stated here so the cost is visible before the layout is
agreed, not discovered during implementation:

- **`transcript` is removed.** The correspondence table has no row for
  it. In execute and dry-run modes the transcript is the same argv
  vectors the `invocations` array already carries; in explain mode it is
  the only body there is, and under this layout explain emits `step` and
  `argv` records like every other mode. It is also not the live stream
  it reads as: the surface captures the plan runner's output into a
  buffer and re-emits it after the command finishes, so nothing about it
  survives an interruption.
- **`facts[].invocations` becomes a list of source names** rather than
  a list of argv vectors, and the argv lives once in `invocations`. That
  is what the `from` requirement means on the wire, and it is what
  removes the 477 duplicated commands measured in Example 5.

Both raise the contract's major version. Neither is proposed as
compatible.

---

## Worked Examples

Every "measured" block below is the literal output of a freshly built
binary — `cd cli && go build -o <path> ./swarm` — run against
`agent-plugins-bd97`, a populated run. Every "required" block is
constructed from the measured block above it by applying this document's
requirements and changing nothing else; none has been produced by a
binary, because no binary implements this document yet.

### 1. A verdict that needs attention

**Measured**, exit 1:

```console
$ swarm manifest require --run agent-plugins-bd97 --key trailers --key branch
swarm manifest require — run agent-plugins-bd97 (execute)
bd show agent-plugins-bd97 --json
branch = true    <- bd show agent-plugins-bd97 --json
trailers = false    <- bd show agent-plugins-bd97 --json
read: ran exit 0 472ms sha256:72f3b0a38d202ddef4dfff826b0b2b371fb7e93ccab52ef3c4a9a2f89f732cbc sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855    <- bd show agent-plugins-bd97 --json
summary: trailers: not written
what would resolve this: write each with: swarm manifest set --run agent-plugins-bd97 --key <key> --value <value>
outcome: attention (exit 1)
```

**Required**, exit 1:

```text
command  swarm manifest require --run agent-plugins-bd97 (execute)
step     1  read  ran  exit 0  472ms  stdout sha256:72f3b0a38d20  stderr empty
argv     bd show agent-plugins-bd97 --json
fact     branch written = flag true
from     read
fact     trailers written = flag false
from     read
summary  1 of 2 required keys is not written: trailers
resolve  swarm manifest set --run agent-plugins-bd97 --key trailers --value <value>
outcome  attention (exit 1)
```

Ten lines against eight, and nothing left to induce: every line names
its own kind, `from read` resolves to the `step` two lines up, `stderr
empty` says what the empty digest meant, `written` says what the flag is
about, and the resolution names one action instead of a template for
both keys.

---

### 2. An explained plan

**Measured**, exit 0:

```console
$ swarm manifest set --run agent-plugins-bd97 --explain --key review_passes --value 3
swarm manifest set — run agent-plugins-bd97 (explain)
# write review_passes into agent-plugins-bd97's manifest
# 1. set-guard — check the record still permits this write [read; stop on failure; safe to re-run]
bd show agent-plugins-bd97 --json
# 2. set-write — merge review_passes into agent-plugins-bd97's run manifest [write; stop on failure; safe to re-run]
bd update agent-plugins-bd97 --metadata '{"review_passes":3}' --json
summary: explained; nothing was executed
outcome: clear (exit 0)
```

**Required**, exit 0:

```text
command  swarm manifest set --run agent-plugins-bd97 --key review_passes --value 3 (explain)
step     1  set-guard  would-run  read; stop on failure; safe to re-run; check the record still permits this write
argv     bd show agent-plugins-bd97 --json
step     2  set-write  would-run  write; stop on failure; safe to re-run; merge review_passes into the run manifest
argv     bd update agent-plugins-bd97 --metadata '{"review_passes":3}' --json
summary  2 steps planned; nothing was executed
outcome  clear (exit 0)
```

The `#` comment convention disappears, and with it the question of
whether a `#` line is a comment the tool wrote or output a tool
produced. The plan's name is dropped: it restated the invocation the
`command` line already carries.

---

### 3. A dry run

**Measured**, exit 0:

```console
$ swarm manifest set --run agent-plugins-bd97 --dry-run --key retro_filed --value 0
swarm manifest set — run agent-plugins-bd97 (dry-run)
bd show agent-plugins-bd97 --json
# refused, dry run writes nothing: set-write
retro_filed = 8    <- bd show agent-plugins-bd97 --json
set-guard: ran exit 0 533ms sha256:72f3b0a38d202ddef4dfff826b0b2b371fb7e93ccab52ef3c4a9a2f89f732cbc sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855    <- bd show agent-plugins-bd97 --json
set-write: refused (dry run: the step writes)
summary: would write retro_filed as one merge; nothing was written
outcome: clear (exit 0)
```

**Required**, exit 0:

```text
command  swarm manifest set --run agent-plugins-bd97 --key retro_filed --value 0 (dry-run)
step     1  set-guard  ran  exit 0  533ms  stdout sha256:72f3b0a38d20  stderr empty
argv     bd show agent-plugins-bd97 --json
step     2  set-write  refused  a dry run does not execute a step that writes
argv     bd update agent-plugins-bd97 --metadata '{"retro_filed":0}' --json
fact     retro_filed before = count 8
from     set-guard
summary  would write retro_filed as one merge; nothing was written
outcome  clear (exit 0)
```

This is the reading the original prompt asked about directly: `set-guard:
ran exit 0 …` is a **step record**, and under this layout the `step`
keyword says so rather than leaving it to be inferred from a colon. The
fact is renamed `retro_filed before`, because `retro_filed = 8` under a
summary saying "would write" is the one line in this report a reader
most reliably misreads. The refused step gains its argv, which the
measured form omits entirely — a reader could not see what would have
been written.

---

### 4. A value too large for a line

**Measured**: `swarm manifest get --run agent-plugins-bd97
--full-result --key checks_baseline` exits 0 and prints 344 lines. Its
first four and last four, with the middle 336 elided:

```text
swarm manifest get — run agent-plugins-bd97 (execute)
bd show agent-plugins-bd97 --json
checks_baseline = # Checks baseline
<blank line — line 4 of the output, and line 2 of the value>
    … 336 further lines of the value, including its own fenced code …
    <- bd show agent-plugins-bd97 --json
read: ran exit 0 442ms sha256:72f3…f732cbc sha256:e3b0…852b855    <- bd show agent-plugins-bd97 --json
summary: 1 key(s) written
outcome: clear (exit 0)
```

One `fact` element occupies lines 3 to 341 — 339 lines, 18,715 bytes of
value. Its provenance annotation arrives alone on line 341, indented
four spaces, 338 lines below the `checks_baseline =` that owns it, and
directly after a line consisting of three backticks that came out of
the value itself. The `outcome` line is line 344. This is the case that
settles the multi-line question.

**Required**, exit 0:

```text
command  swarm manifest get --run agent-plugins-bd97 --full-result --key checks_baseline (execute)
step     1  read  ran  exit 0  442ms  stdout sha256:72f3b0a38d20  stderr empty
argv     bd show agent-plugins-bd97 --json
fact     checks_baseline = text (339 lines, 18715 bytes) — read it with: swarm manifest get --run agent-plugins-bd97 --key checks_baseline
from     read
summary  1 key read
outcome  clear (exit 0)
```

Seven lines against 344. The value is not lost: the bare form yields it
whole with no framing at all, which is what the elision names, and
`--json` carries it whole as well.

---

### 5. Provenance repeated per fact

**Measured**: `swarm status --run agent-plugins-bd97` is 24 lines and
33,329 bytes. Nine of those lines are facts whose provenance is the
*same* 53 `bd list --parent … --json` invocations — 477 copies of 53
distinct commands — and each of those lines is between 3,425 and 3,614
characters. Abbreviated:

```text
completion_total = 45    <- bd list --parent agent-plugins-bd97 --all --limit 0 --json ; bd list --parent agent-plugins-bd97.6 --all --limit 0 --json ; … (51 more)
completion_done = 27    <- bd list --parent agent-plugins-bd97 --all --limit 0 --json ; bd list --parent agent-plugins-bd97.6 --all --limit 0 --json ; … (51 more)
```

**Required**:

```text
source   walk  bd list --parent agent-plugins-bd97 --all --limit 0 --json
source   walk  bd list --parent agent-plugins-bd97.6 --all --limit 0 --json
source   walk  … 51 further invocations, one per level of the run's tree
fact     completion_total = count 45
from     walk
fact     completion_done = count 27
from     walk
```

A `source` name may be declared over several invocations where they are
one logical read, which is what a tree walk of unknown depth is. Cited
by name, the 53 argv vectors appear once in the report instead of nine
times.

---

### 6. An empty list

**Measured**, from the same report:

```text
scope_missing =     <- bd show agent-plugins-bd97 --json
```

**Required**:

```text
source   manifest-read  bd show agent-plugins-bd97 --json
fact     scope_missing = list (0) []
from     manifest-read
```

Three readings collapse into one. The measured line cannot distinguish
an empty list from an empty string from a key that is present and unset.
`swarm status` supplies no plan, so its invocations are `source`
declarations rather than `step` records — which is the case the
`source` keyword exists for.

---

### 7. A list whose items contain the separator

**Measured**, abbreviated — the manifest's `ledger` is seven prose
entries joined with `, `, every one of which contains commas of its own.
The rendered fact line is 2,856 characters:

```text
ledger = An issue's parent is exposed in `bd ... --json` under the key `parent`; `parent_id` is present but ALWAYS null. Proved by …, A `bd gate create -t human --blocks <task>` on the scope epic's task child holds the ENTIRE phase chain. Proved by …, …    <- bd show agent-plugins-bd97 --json
```

**Required**:

```text
fact     ledger = list (7) — read it with: swarm manifest get --run agent-plugins-bd97 --key ledger
from     read
```

The length is the part a reader of the report actually needs; the items
are 2.8 KB and belong in the bare form or in `--json`. Where a list does
fit the budget it renders as its length and its compact JSON, which
keeps an item containing a comma recoverable.

---

## Conformance

Every row below was measured against a binary built from `cli/` at
`80c9113` — `cd cli && go build -o <path> ./swarm` — invoked against
run epic `agent-plugins-bd97`. Two rows are established by reading the
renderer rather than by running it, and say so. This is the checklist
the renderer's conformance work takes; it is not a claim that the
renderer was wrong to ship first.

### Already satisfied

| requirement | evidence |
|---|---|
| Every fact names an invocation | The surface refuses to render a fact with none, and every measured fact line carries one |
| The attention outcome names a resolution | The surface refuses to render an attention result with an empty resolution |
| Both forms carry the same facts | Read from the renderer: both forms are rendered from one `Result` value, not by two renderers kept in agreement |
| The bare form splits the streams by verdict | `get --key branch 2>/dev/null` prints the value; `get --key trailers 2>/dev/null` prints nothing and exits 1 |
| A bare `--json` answer is JSON | `get --key branch --json` prints `"claude/bead-bonanza-agent-plugins-ih18-2yjc8j"` |
| No terminal-width-dependent rendering | Read from the renderer: it pads and aligns nothing, and reports a duration as an integer count of milliseconds |

### Not yet satisfied

| requirement | measured today |
|---|---|
| Every line begins with a keyword naming its kind | No line does. Seven grammars, none self-identifying |
| `from` means provenance and nothing else | `    <- …` is provenance on a fact line and identity on a step record |
| No element spans more than one line | One fact spanned 339 lines and pushed `outcome` to line 344 |
| A value states its kind | The kind is discarded; `scope_missing = ` is an empty list rendering as nothing |
| A list is recoverable | Joined with `, `; the `ledger` value's own items contain commas |
| An invocation is declared once | `swarm status` repeats the same 53-invocation list on nine fact lines — 477 copies, 33,329 bytes total |
| An empty stream reads `empty` | Rendered as the digest of the empty string on every measured step |
| A digest is 12 hex characters | Rendered in full; two digests are 142 characters |
| A fact names what was measured | `trailers = false` means the key is not written; `retro_filed = 8` is the value before the write |
| Every report carries a summary | Optional today |
| A resolution names an action | `manifest set --dry-run --key review_passes --value 3` emitted an identical `summary` and `what would resolve this`, both restating the condition |
| The header names the mode that ran | `swarm manifest get --key branch` — no mode flag, no `--run` — reports `(explain)`, because the scope refusal is taken before the mode is resolved and the zero-valued mode is explain |
| Every exit is rendered by the surface | `swarm manifest get --run agent-plugins-bd97 --bogus` exits 3 having printed `Error: unknown flag: --bogus` on standard error and nothing on standard output — no envelope, and `--json` does not change it |

---

## Open Questions

- [ ] Should `command` restate the full invocation, given the reader is
      assumed to know what they ran? It costs one line and makes a
      report pasted into a transcript self-contained; it is redundant
      for a reader watching their own terminal.
- [ ] Does the `step` line's fixed-column grammar hold for a step with a
      long name and a long reason, or should the reason move to its own
      keyword?
- [ ] Is 200 characters the right budget? It is chosen to fit a digest
      pair and a typical argv, not measured against a corpus of real
      values.

---

## References

- [[container-swarm-cli]] — the provider whose Swarm Command Surface
  owns this layout
- [[contract-swarm-json]] — the machine-readable shape, which this
  document defers to and does not restate
- [[frd-swarm-core]] — the requirements the reported facts and the exit
  vocabulary come from
- [[adr-024]] — plans are values, which is why a step record exists as a
  kind of line at all
- [[adr-025]] — argv rather than a shell, which is why an argv can be
  rendered as a pasteable line
- [[testing-standards]] — the snapshot-case tier the worked examples
  above become fixtures in
