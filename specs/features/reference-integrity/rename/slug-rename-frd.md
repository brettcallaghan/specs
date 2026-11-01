---
title: "Spec Corpus Slug Rename"
description: Renames a document's slug and rewrites every reference to it
  across the corpus in one invocation, so a rename is a single command
  rather than a hand sweep that leaves broken links behind
slug: frd-spec-slug-rename
type: frd
status: draft
sources:
  - "frd-spec-reference-integrity"
---

# Spec corpus slug rename

## Overview

Every broken reference this project has observed came from one event.
When [[adr-012]] moved ADR slugs to the number-only form, the declaring
documents were edited and the 26 references to their old names weren't.

[[frd-spec-reference-integrity]] makes that failure visible: the same
rename reports 26 errors instead of silently resolving to nothing. This
feature prevents the failure. A rename is a single command that edits
the declaring document's `slug` and every reference to it together, so
the corpus is never in the intermediate state where the two disagree.

The alternative considered was an alias: a `former_slugs` property
listing a document's previous names, so old references keep resolving
with a warning. It's rejected as the primary mechanism because an alias
is permanent state. It leaves the corpus with two live names for one
document. Every citation of the old name keeps working, so nothing
forces it to be updated, and the alias list grows with every rename and
can never safely be pruned. A rename command leaves no
residue. After it runs, the document has exactly one name and there's
nothing to remember.

This is the first command in the tool that writes to a document. Every
command specified before it reads and reports, and
[[frd-spec-reference-integrity]] states that boundary explicitly. This
feature doesn't weaken the boundary, because the boundary is about the
checker not guessing: a check that repairs what it thinks is wrong acts
on an inference. A rename acts on an instruction. The author says what
the new name is, and the tool doesn't deduce it.

**Out of scope for this feature.** Each of these is deferred on purpose,
not overlooked:

- **Renaming a slug in another workspace.** A rename rewrites the
  corpus it's run against. If another repository cites this one's
  slugs, this command can't reach those references, and they break
  exactly as they do today. This is the one case an alias would solve
  and a rename can't. It's why `former_slugs` is recorded here as
  rejected-for-now rather than rejected: when the merged graph of
  AC-SPEC-002.3 is real, the question should be reopened.
- **Renaming the file on disk.** Under [[adr-012]], a slug is
  location-independent (a reference resolves against frontmatter, never
  a path), so a slug rename implies nothing about the filename. Keeping
  the two independent is what makes this command safe to run without
  touching Git history.
- **Repairing a reference the tool didn't break.**
  [[frd-spec-reference-integrity]] reports a reference that resolves to
  nothing for any reason other than the rename being performed, and this
  feature leaves it alone.
- **Merging two documents, or splitting one.** A rename maps one slug to
  one slug. Anything else is an editorial act.
- **Rewriting anchors.** A heading rename breaks anchors the way a slug
  rename breaks links. The same argument applies and the same solution
  would work, but it depends on the heading data
  [[frd-spec-anchor-references]] consumes, and it's specified there if
  it's specified at all.

## Terminology

- **Declaring document** — the corpus member whose frontmatter `slug`
  holds the slug being renamed.
- **Referring document** — any corpus member holding a reference to the
  slug being renamed, in its body or in a reference-bearing frontmatter
  property.
- **Preview** — the report of every edit a rename would make, produced
  without modifying any file.
- **Rewrite span** — a span the index recorded as a reference, and
  therefore the only text a rename may edit.

## Requirements

A wikilink written anywhere in a Gherkin source is recorded as a
reference ([[adr-021]] constraint 14). So this feature, whose subject is
rewriting wikilinks, writes its notation illustrations in backticked
Markdown beside each block, not in a step, where a literal wikilink would
enter the index as a citation this document never made.

### FEAT-SPECREN-001: Rename a Slug and Its References

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREN-001
Feature: Rename a Slug and Its References

  As a document author
  I want renaming a slug to update every reference to it
  So that the corpus is never left citing a name that no longer exists

  The rename edits only spans the index recorded as references. That is
  what this command has that a shell script cannot: the documents most
  likely to contain the literal text of a slug in a non-reference
  position are the ones that define the conventions, and they are the
  worst documents in the corpus to silently corrupt.

  @AC-SPECREN-001.1
  Scenario: One invocation renames the declaration and every reference
    When the CLI is asked to rename a current slug to a new slug
    Then the CLI rewrites the declaring document's frontmatter slug and every reference to it in the corpus in one invocation

  @AC-SPECREN-001.2 @negative
  Scenario: Only recorded reference spans are rewritten
    When the CLI performs a rename
    Then the CLI rewrites only spans the index recorded as references
    And the CLI rewrites no matching text in a fenced code block or an inline code span

  @AC-SPECREN-001.3
  Scenario: An anchor and display text survive the rewrite
    Given a rewritten reference carrying an anchor or display text
    When the CLI rewrites it
    Then the CLI preserves both unchanged

  @AC-SPECREN-001.4
  Scenario: Reference-bearing frontmatter is rewritten too
    When the CLI performs a rename
    Then the CLI rewrites the value in every frontmatter property the active profile registers as reference-bearing

  @AC-SPECREN-001.5
  Scenario: The new slug is written in the spelling it was given
    When the CLI performs a rename
    Then the CLI writes the new slug in the spelling it was given, regardless of the letter case each reference used

  @AC-SPECREN-001.6 @negative
  Scenario: A slug rename touches no filename
    When the CLI performs a rename
    Then the CLI renames, moves and deletes no file
```

*Examples:* `specs rename adr-011-targeted-verification-gates adr-011`
edits one ADR's frontmatter and the 26 references across 9 other
documents (AC-SPECREN-001.1). [[adr-012]] contains
`` `[[container-<name>#Display Name]]` `` and other backticked
illustrations of the notation. A rename driven by text search corrupts
the document that defines the convention. A rename driven by the
recorded spans of AC-SPECREF-001.1 can't, because AC-SPECREF-001.2 never
recorded those as references (AC-SPECREN-001.2).
`` `[[old-slug#Element Registry|the registry]]` `` becomes
`` `[[new-slug#Element Registry|the registry]]` ``: only the slug half
is the tool's business (AC-SPECREN-001.3). A document's
`sources: ["adr-011-targeted-verification-gates"]` and a retired
document's `superseded_by` are both rewritten, per the properties
AC-SPECREF-004.2 registers. A rename that swept only body links would
leave the relational graph pointing at the old name (AC-SPECREN-001.4).
`` `[[ADR-011]]` `` and `` `[[adr-011]]` `` both become
`` `[[adr-012]]` `` when renamed to `adr-012`. Resolution is
case-insensitive per AC-SPEC-003.4, so both were valid before, and
leaving them divergent preserves a difference that means nothing and
greps badly (AC-SPECREN-001.5). Renaming the slug of
`specs/adrs/adr-011-targeted-verification-gates.md` leaves the path
untouched (AC-SPECREN-001.6).

*Rationale:* AC-SPECREN-001.2 is the criterion this whole feature turns
on. Renaming references is a five-minute job with `sed`, and a permanent
hazard: the documents most likely to contain the literal text of a slug
in a non-reference position are the ones that define the conventions,
and they're the worst documents in the corpus to silently corrupt. This
command belongs in this tool, not in a shell script, because this tool
already knows which occurrences are references (it had to, to check
them), and a script never can.

### FEAT-SPECREN-002: Preflight and Preview

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREN-002
Feature: Preflight and Preview

  As a document author
  I want to see exactly what a rename will change before it changes anything
  So that a corpus-wide edit is a decision I make rather than one I discover

  The destructive step is always the one you opt into. The blast radius
  of this command is the whole corpus and its input is two strings that
  are easy to transpose, so preview is the default.

  @AC-SPECREN-002.1
  Scenario: A rename previews by default and writes only when instructed
    When the CLI is asked to rename a slug
    Then the CLI produces a preview of every edit the rename would make
    And the CLI modifies no file unless explicitly instructed to write

  @AC-SPECREN-002.2
  Scenario: A new slug already in use stops the rename
    Given the new slug is already declared by a corpus member, comparing without regard to letter case
    When the CLI is asked to perform the rename
    Then the CLI reports the conflict
    And the CLI does not perform the rename

  @AC-SPECREN-002.3
  Scenario: A current slug nothing declares stops the rename
    Given the current slug is declared by no corpus member
    When the CLI is asked to perform the rename
    Then the CLI reports that
    And the CLI does not perform the rename

  @AC-SPECREN-002.4
  Scenario: A failed write leaves the corpus untouched
    Given a file that cannot be written
    When the CLI performs the rename
    Then the CLI leaves every file unmodified
    And the CLI reports which write failed
```

*Examples:* `specs rename old new` reports "1 declaration, 26 references
in 9 documents" with each file and span, and changes nothing. The same
command with the write instruction performs the rename
(AC-SPECREN-002.1). Renaming `adr-011` to `adr-012` where `adr-012`
exists reports the collision and names the document holding it. It
doesn't produce the duplicate slug AC-SPEC-003.4 forbids, which would
leave every reference to either one ambiguous (AC-SPECREN-002.2). A
transposed argument order is caught instead of rewriting the corpus
toward a name nothing declares (AC-SPECREN-002.3). A read-only file in
the corpus aborts the rename with nothing changed. Otherwise the
declaring document would be renamed and eleven of twenty-six references
would still cite the old name. That state is strictly worse than not
having run the command, because the corpus would disagree with itself in
a way no single check explains (AC-SPECREN-002.4).

*Rationale:* Every criterion here exists because this command's failure
modes are worse than its success is valuable. A rename that half
completes, or that renames toward a name already taken, produces a
corpus that's harder to repair than the one the author started with.
Preview-by-default (AC-SPECREN-002.1) is the same argument as the
ordering [[frd-spec-index]] sets between indexing and checking:
the destructive step is always the one you opt into.

### FEAT-SPECREN-003: Outcome

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREN-003
Feature: Outcome

  As a reviewing agent
  I want a rename to report what it changed and to leave the corpus resolving
  So that I can verify the result without re-reading every file it touched

  @AC-SPECREN-003.1
  Scenario: A completed rename reports what it modified
    When a rename completes
    Then the CLI reports every document it modified and the number of references rewritten in each

  @AC-SPECREN-003.2
  Scenario: A completed rename leaves the corpus resolving
    When a rename completes
    Then no reference to the renamed slug is left unresolved for the reason the rename introduced

  @AC-SPECREN-003.3
  Scenario: A rename that changes nothing is a successful no-op
    Given a rename that would change nothing
    When it runs
    Then the CLI reports that
    And the CLI exits zero
```

*Examples:* Nine documents listed with their counts, summing to 26, are
the artifact a reviewer reads against the diff (AC-SPECREN-003.1). A
check run immediately after the rename reports no `reference/unresolved`
finding naming either the old or the new slug (AC-SPECREN-003.2).
Renaming a slug to itself, or renaming a document no other document
references, is a successful no-op. That's consistent with
AC-SPECREF-003.5 and AC-SPEC-007.2, where an empty answer is an answer
(AC-SPECREN-003.3).

*Rationale:* AC-SPECREN-003.2 is deliberately written as a property of
the corpus, not as a description of the tool's behavior. A test can
assert it without knowing how the rename was implemented, and it's the
criterion an author actually cares about. The others describe how the
command behaves. This one says that what the command was built to
guarantee is true.

### FEAT-SPECREN-004: Command Surface and Skill Adoption

**Acceptance Criteria:**

```gherkin
@FEAT-SPECREN-004
Feature: Command Surface and Skill Adoption

  As an authoring agent
  I want to know this command exists at the moment I am about to rename a slug
  So that I reach for it instead of editing every reference by hand

  A command that exists and is not named in the surface an agent reads is
  indistinguishable, in practice, from a command that was never built.

  @AC-SPECREN-004.1
  Scenario: A rename subcommand is exposed
    Then the CLI exposes a rename subcommand accepting the current slug and the new slug, previewing by default and writing only when instructed, per FEAT-SPECREN-002

  @AC-SPECREN-004.2
  Scenario: The shipped skill lists the subcommand beside the others
    Then the shipped skill lists the rename subcommand and the question it answers alongside the other subcommands, in the same command surface an agent reads before choosing how to act

  @AC-SPECREN-004.3
  Scenario: The shipped skill says what the command is for
    Then the shipped skill states that renaming a slug is what this command is for, so that an agent prefers it over editing references directly
```

*Examples:* `specs rename adr-011-targeted-verification-gates adr-011`
previews, and the same command with the write flag performs the rename.
The subcommand vocabulary AC-SPEC-011.5 lists gains an entry, and the
existing entries are neither renamed nor repurposed (AC-SPECREN-004.1).
The command table in [[specs-cli-reference]] gains a
`specs rename <current-slug> <new-slug>` row reading "Rename a slug and
every reference to it" (AC-SPECREN-004.2). The guidance sits with the
existing instruction to prefer the CLI for anything it decides ("its
answer is verifiable by whoever reviews the work, where yours is an
assertion"). That instruction applies unchanged to an edit the tool can
make correctly and a hand sweep frequently can't (AC-SPECREN-004.3, and
AC-SPECREN-001.2).

*Rationale:* This requirement advertises the command and gates nothing.
No checkpoint is added to any workflow, and no agent is required to run
the command before a rename, because a rename performed by hand and then
caught by the parent feature's checks is a recoverable outcome: noisy,
but not silent, which was the original defect.

Advertisement is specified here at all, rather than left to
implementation, because the tool has already produced one capability
nobody invoked. A command that exists and isn't named in the surface an
agent reads is indistinguishable, in practice, from a command that was
never built. Naming it in an acceptance criterion is what stops it being
dropped as polish once the interesting part works.

The skill's command surface documents a real binary, so AC-SPECREN-004.2
lands when the subcommand does. Advertising a `rename` that doesn't
exist yet would send an agent to a command that fails, and the skill
describes exactly the subcommands that are built.

## References

- [[frd-spec-reference-integrity]] — the parent feature, whose recorded
  reference spans this command rewrites and whose checks verify the
  result
- [[frd-spec-index]] — the index, corpus discovery, and command
  surface this feature extends
- [[frd-spec-anchor-references]] — the anchor half of a reference, whose
  preservation AC-SPECREN-001.3 requires
- [[contract-specs-configuration]] — declares the reference-bearing
  properties AC-SPECREN-001.4 rewrites
- [[adr-012]] — the `[[slug]]` convention, and the decision that slugs
  are location-independent
- [[adr-021]] — Gherkin, the notation these features and acceptance
  criteria are written in
- [[gherkin-guide]] — how to write them
