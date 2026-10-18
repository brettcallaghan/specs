---
title: Spec CLI Exit Status Standards
description: What every specs command's exit status means — success and
  four failure classes, which one a run reports, and why a query never
  answers through its status while a gate never passes on a question it
  could not ask
slug: cli-exit-status-standards
type: standards
status: draft
---

# Spec CLI Exit Status Standards

These standards govern the exit status of every command
[[container-spec-cli]] exposes, present and future. They are the
companion to [[cli-output-standards]]: that document governs what a
command prints, this one governs the one value a caller reads before
reading anything printed.

**The reader is a caller deciding what to do next** — an agent, a hook, a
CI step — that has the status and may not have looked at the output yet.
The status exists to answer one question: *who has to act?* The corpus
author, the caller, the repository's maintainer, or this tool's
maintainer. Each answer is a different status, because each sends the
caller somewhere different. Today every failure exits 1, so a hook cannot
tell "the document you wrote is broken" from "you invoked me wrong".

---

## Outcome Classes

### The five statuses

The CLI SHALL exit with one of the following statuses, and SHALL NOT
exit with any other status of its own choosing.

| Status | Class | Who acts | Meaning |
|---|---|---|---|
| `0` | success | nobody | The command did what was asked. A query answered; a gate passed |
| `1` | gate failed | the corpus author | A gate ran to completion and the corpus failed it |
| `2` | invocation error | the caller | The command could not run as invoked: an unknown command or flag, a missing or surplus argument, a value outside a registered set, an operand the command cannot act on |
| `3` | environment error | the repository's maintainer | The command was invoked correctly but its surroundings prevent it: no `specs.toml`, an invalid configuration, an unknown profile, an unreadable spec root, an output path it cannot write |
| `4` | internal error | this tool's maintainer | The CLI reached a state it does not handle: a panic, an encoder refusing a value the model holds, an invariant broken |

**Rationale:** four failure classes, because there are four parties a
caller can hand a failure to, and anything finer is a distinction no
caller acts on differently. `1` keeps the meaning it has in `grep`,
`diff` and every linter — "ran fine, the answer is no" — so a CI step
written for any of them reads this one correctly. Statuses above `125`
are left alone because the shell owns them for signals.

---

### A panic is an internal error, never a usage error

The CLI SHALL recover a panic in any command and exit `4`, writing the
panic's message to standard error.

**Rationale:** an unrecovered Go panic exits `2`, which this standard
assigns to invocation errors. Without recovery a crash reads as "you
called me wrong", and the caller retries a correct invocation instead of
reporting a bug. This is the trap the obvious numbering walks into.

---

### One status per run, by precedence

IF a run meets the conditions of more than one non-zero class, the CLI
SHALL exit with the class highest in the order `4`, `3`, `2`, `1`.

**Rationale:** the higher class makes the lower one unreliable. Findings
from a run that also hit an environment error were produced against a
corpus the CLI could not fully read, so reporting `1` would send the
author to fix documents that may not be broken.

---

### The same input gives the same status

The CLI SHALL exit with the same status for the same corpus content,
configuration and invocation, on every platform.

**Rationale:** this extends AC-SPEC-009.7's determinism from findings to
the status, so a gate that passes locally passes in CI. A status that
depends on whether the cache was warm is a defect.

---

## Queries and Gates

Every command is a **query** (it answers a question about the corpus) or
a **gate** (it decides whether the corpus meets a bar). The two read the
same facts and must report them through the status differently.

| Kind | Commands today |
|---|---|
| gate | `check` |
| query | `list`, `refs`, `headings`, `feature`, `criterion`, `next-id`, `slug-exists`, `c4` |
| producer | `index`, `links`, `feature --emit`, `headings --emit` — a query whose answer is an artifact |

### A query never answers through its status

A query SHALL exit `0` whenever it produced an answer, whatever the
answer is.

**Rationale:** "no document declares this slug", "nothing references
this heading" and "no criterion changed" are facts about the corpus,
and [[cli-output-standards]] already requires them rendered as answers.
Encoding them in the status as well would make a caller branch on what
the corpus holds rather than on whether the command worked, and would
make every query a gate nobody asked for. A caller that wants a
yes-or-no reads the answer's property — `declared: false` — not the
status.

---

### A subject the corpus does not hold is an answer

WHEN a query names a subject that the corpus could hold but does not — a
slug, a feature or criterion identifier, a feature code, a heading — the
CLI SHALL render the empty answer with a property stating that the
subject is not declared, and SHALL exit `0`.

**Rationale:** whether `AC-SPEC-099.1` exists is a fact about this
corpus today; it will exist once someone writes it. That is the same
kind of fact as an empty search, and it gets the same status. This is
what moves `criterion <id>` and `c4 --root <slug>` from `1` to `0`.

---

### A value outside a registered set is an invocation error

WHEN an argument names a value from a set the active profile or the CLI
registers — a document type, a rule identifier, a tag qualifier — and
the value is not in that set, the CLI SHALL report the value and every
registered name on standard error and SHALL exit `2`.

WHERE the same value is named by the configuration file rather than by
an argument, the CLI SHALL exit `3`.

**Rationale:** the line between this rule and the previous one is
whether the corpus could ever make the value valid. No document an
author writes turns `adrx` into a document type; a configuration change
does. So a misspelt type is a mistake in the question, and answering it
with an empty list would make a typo indistinguishable from "none
exist", and a gate running `check --only reference/unresolvd` would
pass forever. The caller cannot fix a value written in `specs.toml`, so
the same unknown name there — `[profile] name`, for one — is the
repository maintainer's, which is why an unknown profile is listed under
`3`.

---

### A gate never passes on a question it could not ask

WHEN a gate is given an operand it cannot act on — a path that is not a
corpus member, a rule it cannot run — the CLI SHALL exit `2`, and SHALL
NOT exit `0`.

**Rationale:** the asymmetry with queries is deliberate. A query that
cannot find its subject has still answered. A gate that cannot inspect
its subject has decided nothing, and a `0` from it would be read as a
pass: a write-time hook handed a mistyped path would approve every
edit.

---

### A gate fails only on what reaches its bar

A gate SHALL exit `1` WHEN it raises at least one finding at a severity
its bar fails, and SHALL exit `0` otherwise.

Without `--strict` the bar is `error`. Under `--strict` the bar is every
severity, `warning` and `unverified` included.

**Rationale:** an `unverified` finding records a check that could not
look, and is the reason `--strict` exists: in CI, "could not tell" must
block; in an editing loop it must not. Folding it into the default bar
would make every unverifiable diagram fail every hook.

---

### A producer succeeds when its artifact is written

A producer SHALL exit `0` WHEN it wrote its artifact, including an
artifact that records defects in the corpus it was built from.

A producer SHALL exit `3` WHEN it cannot write its artifact to the
destination named, and SHALL exit `4` WHEN it cannot encode the artifact
from the corpus it read.

**Rationale:** an index of a corpus with broken links is a correct
index, and AC-SPEC-002.5 already requires it to exit zero; judging the
corpus is the gate's job. Failing to encode, by contrast, is never the
corpus's fault: a corpus missing a required property is a finding for
`check`, and an encoder that refuses it is a bug in this tool.

---

### A gate flag on a query is opt-in and named

WHERE a query offers a caller the option to fail on a non-empty answer,
the CLI SHALL expose it as a flag named `--exit-code`, under which a
non-empty answer exits `1`, and SHALL NOT change the query's default
status.

**Rationale:** "fail the build if any criterion changed" is a reasonable
CI step, and `git diff --exit-code` is the convention callers already
know. Naming one flag for it everywhere keeps the default rule above
intact instead of growing a per-command exception.

---

## Streams

### A failed invocation answers nothing

IF a run exits `2`, `3` or `4`, the CLI SHALL write nothing to standard
output, and SHALL write the reason to standard error.

**Rationale:** a caller that captures standard output and parses it must
never parse half an answer. The exception is status `1`, whose findings
*are* the answer and are written to standard output as
[[cli-output-standards]] requires.

---

### The reason names the class

WHEN the CLI exits non-zero, the message on standard error SHALL begin
with the offending value or condition, and SHOULD say what the caller
can change.

**Rationale:** the status says who acts; the message says what to do.
"no specs.toml found at specs.toml" tells the maintainer the file is
missing, but not that the CLI searched only the working directory.

---

## Guidance for Callers

Stated here because the status is only as good as the way it is read.

- **Read the status directly, never through a pipe.** `specs check | tail`
  reports `tail`'s status, so a failing gate reads as a pass. Redirect to a
  file and read `$?`, or capture the output in a variable.
- **Branch on the class, not on non-zero.** `1` means fix the documents;
  `2` means fix the command; `3` means fix the repository; `4` means file
  a bug. A retry is correct for none of them.
- **Read a query's answer, not its status**, for any yes-or-no question
  about the corpus.

---

## Current conformance

Measured with a binary built from `main` at `31b3079`, run against this
repository's corpus; the `headings` row at `12b2eda`. The
`headings … --emit` row was re-measured on the change that refuses a
selection matching more than one heading: exit `2`, nothing on standard
output, both matches named on standard error. The four not-declared and
qualifier rows were re-measured on the change that answers an undeclared
subject: each as its "today" cell reads, nothing on standard error for
the `0` rows. The last row was
re-measured on the change that moved `criterion --changed` onto a git
base, in a directory outside any repository: exit `3`, nothing on
standard output. The missing-enum row was re-measured on the change that
omits an undeclared enum from the index, over `index-default`'s corpus
with an ADR's `status` deleted: `index --out -` and `list --json` each
exit `0` with the answer written.

| invocation | today | under these standards |
|---|---|---|
| `check` with error findings | `1` | `1` |
| `specs chek`, `specs refs` (no argument) | `1` | `2` |
| `list --type nope`, `check --only nope`, `next-id frd` | `1` | `2` |
| `check specs/nope.md` | `1` | `2` |
| `--config nope.toml list` | `1` | `3` |
| `criterion AC-X-1` | `0`, `declared: false` | `0`, with the subject marked not declared |
| `c4 --root zzz` | `0`, `root_declared: false` | `0`, with the subject marked not declared |
| `criterion --tag zzz` | `2`, the value and every registered qualifier on standard error | `2` — qualifiers are a registered set |
| `slug-exists zzz`, `feature --code ZZ`, `headings 'zzz#q'` | `0`, `declared: false` | `0`, with the subject marked not declared |
| `list --json` or `index` on a document missing a required enum | `0`, the artifact written | `0` and the artifact written |
| a panic | `2` (Go's default) | `4` |
| `headings 'module-document-core#references' --emit`, two matches | `2`, each match named on standard error | `2`, each match named on standard error |
| `criterion --changed` outside a git repository | `3` | `3` |

---

## References

- [[cli-output-standards]] — what a command prints; this document governs
  only the status beside it
- [[container-spec-cli]] — the container whose command surface these
  standards govern
- [[frd-spec-index]] — AC-SPEC-002.5, AC-SPEC-006.7 and AC-SPEC-009.7,
  which these standards generalize
