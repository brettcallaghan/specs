---
title: "FRD: Spec Corpus Slug Rename"
description: Renames a document's slug and rewrites every reference to it
  across the corpus in one invocation, so a rename is a single command
  rather than a hand sweep that leaves broken links behind
slug: frd-spec-slug-rename
type: frd
status: draft
updated_at: 2026-08-05T00:00:00Z
sources:
  - "frd-spec-reference-integrity"
---

# Spec Corpus Slug Rename

## Overview

Every broken reference this project has actually observed came from one
event. When [[adr-012]] moved ADR slugs to the number-only form, the
declaring documents were edited and the 26 references to their old names
were not. Nobody noticed for weeks. The repair was a one-off prototype
and a hand migration.

[[frd-spec-reference-integrity]] makes that failure visible: after it
ships, the same rename reports 26 errors instead of silently resolving
to nothing. This feature makes the failure *not happen*. A rename is a
single command that edits the declaring document's `slug` and every
reference to it together, so the corpus is never in the intermediate
state where the two disagree.

The alternative considered was an alias — a `former_slugs` property
listing a document's previous names, so old references keep resolving
with a warning. It is rejected as the primary mechanism for a reason
worth recording: an alias is permanent state. It leaves the corpus with
two live names for one document, every citation of the old name keeps
working so nothing forces it to be updated, and the alias list grows
with every rename and can never safely be pruned. A rename command has
no residue. After it runs there is exactly one name for the document and
nothing to remember.

This is the first command in the tool that writes to a document.
Everything specified before it reads and reports, and
[[frd-spec-reference-integrity]] states that boundary explicitly. The
boundary is not weakened here, because it was always about the *checker*
not guessing: a check that repairs what it thinks is wrong is a check
that acts on an inference. A rename acts on an instruction. The author
says what the new name is; the tool does not deduce it.

**Out of scope for this feature.** Each of these is deferred
deliberately, not overlooked:

- **Renaming a slug in another workspace.** A rename rewrites the
  corpus it is run against. IF another repository cites this one's
  slugs, THEN this command cannot reach those references, and they break
  exactly as they do today. This is the one case an alias would solve
  and a rename cannot, and it is the reason `former_slugs` is recorded
  here as rejected-for-now rather than rejected: when the merged graph
  of AC-SPEC-002.3 is real, the question should be reopened.
- **Renaming the file on disk.** A slug is location-independent by
  [[adr-012]] — a reference resolves against frontmatter, never a path —
  so a slug rename implies nothing about the filename. Keeping the two
  independent is what makes this command safe to run without touching
  git history.
- **Repairing a reference the tool did not break.** A reference that
  resolves to nothing for any reason other than the rename being
  performed is reported by [[frd-spec-reference-integrity]] and left
  alone here.
- **Merging two documents, or splitting one.** A rename maps one slug to
  one slug. Anything else is an editorial act.
- **Rewriting anchors.** A heading rename breaks anchors the way a slug
  rename breaks links; the same argument applies and the same solution
  would work, but it depends on the heading data
  [[frd-spec-anchor-references]] consumes and is specified there IF it
  is specified at all.

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

### REQ-SPECREN-001: Rename a Slug and Its References

**User Story:** As a document author, I want renaming a slug to update
every reference to it, so that the corpus is never left citing a name
that no longer exists.

**Acceptance Criteria:**

- **AC-SPECREN-001.1:** WHEN the CLI is asked to rename a current slug
  to a new slug, the CLI SHALL rewrite the declaring document's
  frontmatter `slug` and every reference to it in the corpus in one
  invocation.
  - *Example:* `spec rename adr-011-targeted-verification-gates adr-011`
    edits one ADR's frontmatter and the 26 references across 9 other
    documents — the exact migration that was performed by hand and left
    26 broken links behind.
- **AC-SPECREN-001.2:** The CLI SHALL rewrite only spans the index
  recorded as references, and SHALL NOT rewrite matching text in a
  fenced code block or an inline code span.
  - *Example:* [[adr-012]] contains `` `[[container-<name>#Display
    Name]]` `` and other backticked illustrations of the notation. A
    rename driven by text search corrupts the document that defines the
    convention; a rename driven by the recorded spans of
    AC-SPECREF-001.1 cannot, because AC-SPECREF-001.2 never recorded
    those as references.
- **AC-SPECREN-001.3:** WHEN a rewritten reference carries an anchor or
  display text, the CLI SHALL preserve both unchanged.
  - *Example:* `[[old-slug#Element Registry|the registry]]` becomes
    `[[new-slug#Element Registry|the registry]]`. Only the slug half is
    the tool's business; the anchor and the prose belong to the author.
- **AC-SPECREN-001.4:** The CLI SHALL rewrite the value in every
  frontmatter property the active profile registers as
  reference-bearing.
  - *Example:* A document's `related: ["adr-011-targeted-verification-
    gates"]` and a retired document's `superseded_by` are both rewritten,
    per the properties AC-SPECREF-004.2 registers. A rename that swept
    only body links would leave the relational graph pointing at the old
    name — the half of the corpus a reader never sees and a catalogue
    reads first.
- **AC-SPECREN-001.5:** The CLI SHALL write the new slug in the spelling
  it was given, regardless of the letter case each reference used.
  - *Example:* `[[ADR-011]]` and `[[adr-011]]` both become `[[adr-012]]`
    when renamed to `adr-012`. Resolution is case-insensitive per
    AC-SPEC-003.4, so both were valid before; leaving them divergent
    preserves a difference that means nothing and greps badly.
- **AC-SPECREN-001.6:** The CLI SHALL NOT rename, move, or delete any
  file.
  - *Example:* Renaming the slug of
    `specs/adrs/ADR-011-targeted-verification-gates.md` leaves the path
    untouched. The filename is not the identifier — [[adr-012]] made
    slugs location-independent precisely so that these two decisions
    stay separable.

*Rationale:* AC-SPECREN-001.2 is the criterion this whole feature turns
on. Renaming references is a five-minute job with `sed` and a permanent
hazard: the documents most likely to contain the literal text of a slug
in a non-reference position are the ones that define the conventions,
and they are the worst documents in the corpus to silently corrupt. The
reason this command belongs in this tool rather than in a shell script
is that this tool already knows which occurrences are references — it
had to, to check them — and a script never can.

### REQ-SPECREN-002: Preflight and Preview

**User Story:** As a document author, I want to see exactly what a
rename will change before it changes anything, so that a corpus-wide
edit is a decision I make rather than one I discover.

**Acceptance Criteria:**

- **AC-SPECREN-002.1:** The CLI SHALL produce a preview of every edit a
  rename would make, and SHALL NOT modify any file unless explicitly
  instructed to write.
  - *Example:* `spec rename old new` reports "1 declaration, 26
    references in 9 documents" with each file and span, and changes
    nothing. The same command with the write instruction performs it.
    The default is the safe one because the blast radius of this command
    is the whole corpus and its input is two strings that are easy to
    transpose.
- **AC-SPECREN-002.2:** IF the new slug is already declared by a corpus
  member, comparing without regard to letter case, THEN the CLI SHALL
  report the conflict and SHALL NOT perform the rename.
  - *Example:* Renaming `adr-011` to `adr-012` where `adr-012` exists
    reports the collision and names the document holding it, rather than
    producing the duplicate slug AC-SPEC-003.4 forbids and leaving every
    reference to either one ambiguous.
- **AC-SPECREN-002.3:** IF the current slug is declared by no corpus
  member, THEN the CLI SHALL report that and SHALL NOT perform the
  rename.
  - *Example:* A transposed argument order is caught here rather than
    rewriting the corpus toward a name nothing declares.
- **AC-SPECREN-002.4:** IF any file cannot be written, THEN the CLI
  SHALL leave every file unmodified and report which write failed.
  - *Example:* A read-only file in the corpus aborts the rename with
    nothing changed, rather than leaving the declaring document renamed
    and eleven of twenty-six references still citing the old name — a
    state strictly worse than not having run the command, because the
    corpus now disagrees with itself in a way no single check explains.

*Rationale:* Every criterion here exists because this command's failure
modes are worse than its success is valuable. A rename that half
completes, or that renames toward a name already taken, produces a
corpus that is harder to repair than the one the author started with.
Preview-by-default (AC-SPECREN-002.1) is the same argument as the
ordering [[frd-spec-validation-cli]] sets between indexing and checking:
the destructive step is always the one you opt into.

### REQ-SPECREN-003: Outcome

**User Story:** As a reviewing agent, I want a rename to report what it
changed and to leave the corpus resolving, so that I can verify the
result without re-reading every file it touched.

**Acceptance Criteria:**

- **AC-SPECREN-003.1:** WHEN a rename completes, the CLI SHALL report
  every document it modified and the number of references rewritten in
  each.
  - *Example:* Nine documents listed with their counts, summing to 26.
    This is the artifact a reviewer reads against the diff.
- **AC-SPECREN-003.2:** WHEN a rename completes, the CLI SHALL leave no
  reference to the renamed slug unresolved for the reason the rename
  introduced.
  - *Example:* A check run immediately after the rename reports no
    `reference/unresolved` finding naming either the old or the new
    slug. Stated as an outcome rather than as an internal step because
    it is the only property of this command that actually matters, and
    it is verifiable from outside.
- **AC-SPECREN-003.3:** WHEN a rename would change nothing, the CLI
  SHALL report that and exit zero.
  - *Example:* Renaming a slug to itself, or renaming a document no
    other document references, is a successful no-op — consistent with
    AC-SPECREF-003.5 and AC-SPEC-007.2, where an empty answer is an
    answer.

*Rationale:* AC-SPECREN-003.2 is deliberately written as a property of
the corpus rather than as a description of the tool's behavior. It is
the criterion a test can assert without knowing how the rename was
implemented, and it is the one an author actually cares about — the
others describe how the command behaves, this one says the thing it was
built to guarantee is true.

### REQ-SPECREN-004: Command Surface and Skill Adoption

**User Story:** As an authoring agent, I want to know this command
exists at the moment I am about to rename a slug, so that I reach for it
instead of editing every reference by hand.

**Acceptance Criteria:**

- **AC-SPECREN-004.1:** The CLI SHALL expose a `rename` subcommand
  accepting the current slug and the new slug, previewing by default and
  writing only when instructed, per REQ-SPECREN-002.
  - *Example:* `spec rename adr-011-targeted-verification-gates adr-011`
    previews; the same command with the write flag performs it. The
    subcommand vocabulary AC-SPEC-011.5 lists gains a fifth entry, and
    the four established there are neither renamed nor repurposed.
- **AC-SPECREN-004.2:** The shipped skill SHALL list the `rename`
  subcommand and the question it answers alongside the other
  subcommands, in the same command surface an agent reads before
  choosing how to act.
  - *Example:* The command table in
    `plugin/skills/spec/references/spec-cli.md` gains a `spec rename
    <current-slug> <new-slug>` row reading "Rename a slug and every
    reference to it". An agent that does not know the command exists
    writes the hand sweep instead, which is the failure this feature was
    built to remove.
- **AC-SPECREN-004.3:** The shipped skill SHALL state that renaming a
  slug is what this command is for, so that an agent prefers it over
  editing references directly.
  - *Example:* The guidance sits with the existing instruction to prefer
    the CLI for anything it decides — "its answer is verifiable by
    whoever reviews the work, where yours is an assertion" — which
    applies unchanged to an edit the tool can make correctly and a hand
    sweep frequently cannot (AC-SPECREN-001.2).

*Rationale:* This requirement advertises the command; it does not gate
anything. No checkpoint is added to any workflow and no agent is
required to run it before a rename, because a rename performed by hand
and then caught by the parent feature's checks is a recoverable
outcome — noisy, but not silent, which was the original defect.

The reason advertisement is specified here at all, rather than left to
implementation, is that the tool has already produced one capability
nobody invoked. A command that exists and is not named in the surface an
agent reads is indistinguishable, in practice, from a command that was
never built. Naming it in an acceptance criterion is what stops it being
dropped as polish once the interesting part is working.

The skill's command surface is documentation of a real binary, so
AC-SPECREN-004.2 lands when the subcommand does. Advertising a `rename`
that does not yet exist would send an agent to a command that fails, and
the skill currently describes exactly four subcommands because exactly
four are built.

## References

- [[frd-spec-reference-integrity]] — the parent feature, whose recorded
  reference spans this command rewrites and whose checks verify the
  result
- [[frd-spec-validation-cli]] — the index, corpus discovery, and command
  surface this feature extends
- [[frd-spec-anchor-references]] — the anchor half of a reference, whose
  preservation AC-SPECREN-001.3 requires
- [[contract-specs-configuration]] — declares the reference-bearing
  properties AC-SPECREN-001.4 rewrites
- [[adr-012]] — the `[[slug]]` convention, and the decision that slugs
  are location-independent
- [[adr-002]] — EARS, the notation these acceptance criteria are written
  in
