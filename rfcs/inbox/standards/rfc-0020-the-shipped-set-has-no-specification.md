---
title: "RFC: The shipped set has no specification — write down the vendoring
  model, the invariant that governs it, and the citation rule its depths break"
description: The move that made specs/ the single source established an
  invariant and a two-depth vendoring model that appear in no document, and it
  falsified the stated rationale of the citation standard's relative-path rule.
slug: rfc-0020
type: rfc
status: ready-for-review
updated_at: 2026-08-14T09:55:00Z
origin: self
skill: bead-bonanza
skill_version: "1.5"
model: claude-opus-5
effort: unknown
runs_observed: 1
severity: blocker
plugin_version: "0.17.0"
---

# RFC: The shipped set has no specification

## Summary

Epic `agent-plugins-eaa` moved 14 documents into `specs/`, inverted
`scripts/sync-shared-references.sh` so `specs/` is the single source, and made
every copy inside a skill generated. That is a new architectural model for this
repository: it decides what ships to adopters, it now governs 26 map entries,
and it is the reason the plugin tree no longer needs a corpus walk.

It is written down nowhere. The invariant that governs it — *a shipped document
may cite only other shipped documents* — exists only in the description of a
bead that is now closed. No document under `specs/` mentions the vendoring
model, the `GENERATED COPY` marker, or the map's role as the statement of the
shipped set.

Three consequences already showed up during the run that created it: an
implementer resolved a transitive cascade by making a citation inert and had to
record the reasoning in a comment; two implementers independently hit a
citation that cannot be spelled correctly under the new model, and both invented
the same workaround; and the citation standard's stated rationale is now false.

This RFC does not restate `rfc-0019` P4, which covers the separate case of a
document crossing into the corpus.

## Provenance

| Field | Value |
|---|---|
| Phase executed | Implementation — 16 beads across two epics, run `1786686364-15`, branch `claude/validate-corpus-spec-cli-0t0cxo`, draft PR #22 |
| Documents consumed | `[[cross-document-references]]` v0.3; `[[template-authoring-standards]]` v0.13; `[[findings-capture-protocol]]` v0.4; the `agent-plugins-eaa` epic and its five children as work orders |
| Documents produced | 14 documents relocated under `specs/standards/` and `specs/guides/`; an inverted `scripts/sync-shared-references.sh` MAP (26 entries) |
| Skills/templates that guided it | `bead-bonanza` v1.5; `plan-sprint` v0.1; `groom-beads` (work-order format) |
| Plugin version | 0.17.0 |
| Evidence base | 64 `retro:` comments across all 15 dispatched beads. `rfc-0019` rests on `ia5.1`–`ia5.5` plus DoD findings; this RFC rests on the nine beads that RFC does not cover — `wu1`, `8ko`, `cb9`, `eaa.1`–`eaa.4`, `47b`, `92u` |
| Run dates | 2026-08-14 |

## Evidence

1. **The invariant that decides what ships is not in any document.** The
   `agent-plugins-eaa` epic states it: "A shipped document may cite only other
   shipped documents. After the inversion the sync map IS the shipped set, so
   it doubles as the marker — no new mechanism needed." Searching `specs/` and
   `scripts/sync-shared-references.sh` for that rule returns nothing, and no
   document under `specs/` mentions `sync-shared-references` or
   `GENERATED COPY` at all. The bead is closed, so the rule now governs the
   repository from a tracker record rather than from a specification.

2. **The invariant is not closed under transitivity, and the implementer
   discovered that at consume time.** `agent-plugins-eaa.2`:

   > `retro: Silent decision — frontmatter-standards.md's two
   > [[document-relationship-standards]] wikilinks were backticked (inert
   > prose) rather than shipping document-relationship-standards.md via the
   > map, because that document itself cites adr-009/adr-012/two FRDs that
   > aren't shipped — adding it would cascade the ships-only-what-it-cites
   > invariant outward instead of resolving it.`

   Nothing in any document says which resolution is correct. The one chosen
   satisfies the invariant's letter while defeating its purpose: the adopter
   still reads a named document they cannot reach, and now cannot even tell it
   was meant to be a citation. The epic's own motivation was to fix exactly
   that reader-facing defect.

3. **A citation can be unspellable under the new model, and the standard's
   rationale says otherwise.** `[[cross-document-references]]` v0.3,
   "Paths are relative to the file that contains them" (lines 108–124):

   > "A backticked path SHALL be written relative to the **file it appears
   > in**, not to the skill root."
   >
   > "**Rationale:** ... Writing the path from the reader's position is the
   > only spelling that works from every location the file is read."

   Vendoring falsifies the rationale. A file vendored into another skill lives
   at `references/shared/`, one level below its authored location, so one
   relative string cannot be right in both. `agent-plugins-eaa.3` hit it:

   > `retro: work-order-template.md (a template, not moved) bare-cited
   > work-order-field-guide.md by filename. After the move that citation is
   > only correct from a directory where both files are flat siblings — true
   > in every OTHER skill's vendored references/shared/, but not from the
   > authored spec/references/ location, where the guide's copy is now one
   > level down in shared/. A single relative-path string can't be correct in
   > both places`

   `agent-plugins-eaa.4` hit the same thing on the same two files and reached
   the same workaround independently, naming it as a general shape:

   > `Shape 2 (a relative path that can't work from both the authored spot and
   > a shared/ copy one level deeper): feature-blueprint-template.md and
   > work-order-template.md are each vendored into a second skill ... so their
   > c4-guide.md citations became [[c4-standards]] wikilinks, mirroring
   > eaa.3's fix to the same two files`

   Two independent implementers converging on the same unwritten rule is the
   definition of a rule that should be written. Note the escape hatch works
   only because the target happened to be a corpus member; a template citing
   another template has no correct spelling at all.

4. **A measurement in a spec went stale between authoring and execution, by
   17x.** The epic and `agent-plugins-eaa.1` both state "~50" illustrative
   wikilinks needing backticks. `eaa.1` measured the truth:

   > `The bead's ~50-dangling estimate was measured before agent-plugins-wu1
   > landed (which trimmed c4-guide.md's wikilink count sharply); the actual
   > pre-fix dangling count in the 14-doc corpus was 3, not ~50.`

   The number was correct when written and wrong when consumed, because a
   sibling bead in the same queue changed the thing being counted. Nothing in
   the document records when it was measured, so a consumer cannot tell a live
   figure from a stale one.

5. **A work order's acceptance criterion excluded the only correct
   implementation.** `agent-plugins-92u`'s criterion said the new CI step goes
   "in place of" the deleted one, inside the `checks` job.
   `agent-plugins-92u`:

   > `The bead should have pinned WHICH JOB the new step lands in. The
   > replacement needs the spec binary and the checks job is bash+python3 only
   > with no Go toolchain, so the step has to go in the corpus job that already
   > builds it — a job move, not an in-place swap, which the acceptance
   > criterion's "in place of" wording does not admit.`

   The implementer did the right thing and the criterion had to be read
   against its own wording.

## Proposals

### P1 — Specify the shipped set as a standards document

**Problem.** Finding 1: the model that decides what this plugin ships to
adopters has no document. There is no text to quote here, because none exists —
that is the finding. The rule currently lives in the description of
`agent-plugins-eaa`, a closed bead, and in the shape of a shell script's MAP.

**Change.** Add `specs/standards/shipped-reference-set.md` (type `standards`),
stating at minimum:

> **The map is the shipped set.** `scripts/sync-shared-references.sh`'s MAP is
> the authoritative statement of which documents ship inside a skill. A
> document under `specs/` that is absent from the map is internal to this
> repository by construction; no other marker is kept, and none is needed.
>
> **`specs/` is the single source.** Every copy under
> `plugin/skills/*/references/shared/` is generated and carries the
> `GENERATED COPY` header naming its source. Editing a copy is always wrong.
> This includes the copies inside the skill that authors the document.
>
> **The header is load-bearing twice.** It tells a human which file to edit,
> and it is what makes a copy `no-frontmatter`-excluded from a corpus walk —
> which is why the plugin tree needs no corpus checking at all.
>
> **A shipped document may cite only other shipped documents.** A citation to
> a document that does not ship is dead in the adopter's repository, which is
> the defect this model exists to remove.

Backward compatible: it documents what the tree already does.

### P2 — Give the invariant a rule for the transitive case

**Problem.** Finding 2. The invariant in P1's last clause has two resolutions
when a shipped document cites a non-shipped one — ship the target, or stop
citing it — and they have very different costs. Shipping can cascade
arbitrarily far; backticking silently degrades the reader's experience while
technically satisfying the rule.

**Change.** Add to the document P1 creates:

> WHEN a shipped document cites a document that does not ship, resolve it in
> this order and record which was chosen:
>
> 1. **Ship the target**, IF its own citations are already satisfied or are
>    satisfiable without further cascade.
> 2. **Rewrite the sentence** so it no longer needs the citation — state the
>    fact rather than pointing at where the fact lives.
> 3. **Demote the citation to a backticked name**, ONLY WHEN 1 and 2 both
>    fail, and say in the sentence that the document is not shipped.
>
> Resolution 3 is a last resort, not a default. It leaves the adopter reading
> a pointer they cannot follow — the exact defect this invariant exists to
> prevent — so a bare demotion with no acknowledgement in the prose is a
> regression dressed as compliance.
>
> **Rationale:** the cascade is real. `frontmatter-standards.md` cites
> `document-relationship-standards`, which cites two ADRs and two FRDs that do
> not ship; resolution 1 would have pulled all five in.

Backward compatible: the one existing instance was resolved as 3 and would need
its prose annotated, nothing more.

### P3 — Fix the relative-path rationale, and rule for a file that lives at two depths

**Problem.** Finding 3. `[[cross-document-references]]` v0.3 asserts that a
path relative to the citing file "is the only spelling that works from every
location the file is read." Under the vendoring model that is false for any
file vendored to a different depth than its authored location, and this run
produced two such files.

**Change.** In `[[cross-document-references]]`, "Paths are relative to the file
that contains them", replace the Rationale's final sentence and add a rule:

> **Rationale:** an agent resolves a path from the file it is currently
> reading, so a `references/`-prefixed path written *inside* `references/`
> resolves to nothing. Writing the path from the reader's position is correct
> for a file that is read from exactly one location.
>
> **A file read from two depths has no correct relative path.** A document
> vendored into another skill is read both from its authored location and from
> that skill's `references/shared/`, one level deeper. A single relative string
> cannot be right in both. WHEN a citation appears in a file that is itself
> vendored, it SHALL NOT be a relative path:
>
> - IF the target is a corpus member, cite it as `[[slug]]` — position-
>   independent by construction.
> - IF the target is not a corpus member, the citation SHALL be made
>   position-independent some other way (name it without a path, or make the
>   target a corpus member). Do not write a relative path and hope one of the
>   two readers is the one that matters.

Backward compatible: existing citations in non-vendored files are unaffected.
The two known instances were already converted to wikilinks during this run.

### P4 — A measurement in a document carries its as-of point

**Problem.** Finding 4: a 17x-wrong figure, correct when written, consumed as
if live. `CLAUDE.md`'s "Historical prose" section already directs measurements
into the bead rather than the skill file, so the *location* is settled; what is
missing is that a measurement records when it was taken.

**Change.** Add to `[[template-authoring-standards]]`:

> WHEN a document states a measured figure — a count, a ratio, a timing — it
> SHALL name what the measurement was taken against: a commit sha, a date, or
> the command that produces it. A bare figure cannot be distinguished from a
> stale one by the agent consuming it, and a queue in which sibling items
> change the measured population makes staleness the normal case rather than
> the exception.
>
> Prefer the command over the number where one exists: a consumer can re-run
> it, and a number nobody can reproduce is an assertion.

Backward compatible: applies to figures added or edited from adoption onward;
no sweep of existing documents is implied.

### P5 — An acceptance criterion names the outcome, not the edit site

**Problem.** Finding 5: "in place of" pinned a location that could not work,
and the implementer had to read past the criterion's wording to satisfy its
intent. `[[work-order-definition-of-ready]]` check 4 already says criteria are
"Outcome-focused, checkable, implementation-agnostic ... Describe the
observable end state, not the steps" — the failure is that "in place of the X
step" reads as an outcome while actually pinning an edit site.

**Change.** Add to check 4's guidance:

> A criterion naming WHERE an edit lands — a file, a job, a function — is a
> step wearing an outcome's clothes. Name what must be observably true
> instead, and leave the site to the implementer, who can see constraints
> grooming cannot: "CI runs `spec check` over `rfcs/` on every run" is
> checkable and admits the only correct implementation; "the `checks` job runs
> it in place of the old step" forbids it, because that job has no Go
> toolchain to build the binary with.

Backward compatible: guidance on authoring new criteria.

## Considered and Not Proposed

- **Converting every stale backticked path in `specs/` prose.** Already owned
  by `agent-plugins-jn6` (38 citations, open) and by `rfc-0019` P4's rule. Not
  re-proposed here.
- **Requiring `document-relationship-standards` to ship.** Rejected on the
  evidence in finding 2: its own citations cascade to five further documents.
  P2 gives the decision procedure instead of pre-deciding one instance.
- **A gate that resolves backticked paths inside `specs/`.** Attractive — it is
  the check whose absence lets these rot — but it is a CLI capability
  proposal, not a document change, and it needs its own design (which paths are
  citations vs. illustrations). Worth a bead; out of scope for a retro of the
  documents.
- **`eaa.1`'s stale `.pyc` shadowing a corrected script.** A container
  artifact, not a document defect. One-off.
- **`wu1` leaving `guide-c4` strings in retired RFCs and sprint reports.**
  Correct judgement, explicitly reasoned ("rewriting them would edit history
  rather than fix a live reference"), and `92u` reached the same conclusion
  independently about `validate-corpus.py` mentions in archived records. No
  defect — if anything this is evidence the historical-record boundary is
  already well understood.
- **`92u`'s discovery that a second spec root produces a second index artifact
  needing a `.gitignore` entry.** Real, and it would have staged a 99KB build
  artifact, but it is a consequence of a capability added in this same run
  rather than a defect in a document the phase consumed.

## Non-Goals

- Any change to the markdown-links feature's documents. `rfc-0019` covers that
  phase and this RFC deliberately rests on the nine beads it does not.
- Changing `scripts/sync-shared-references.sh` itself. P1 documents the model
  the script implements; the script behaved correctly throughout, and
  `--check` exited 0 at every gate.
- The skills and templates that guided the run. Three `retro-skill` RFCs were
  filed separately against `bead-bonanza`, `plan-sprint` and `recover-sprint`.

## Adoption

P1 is the foundation: P2 extends the document it creates, so P1 and P2 land
together or P2 waits. P3, P4 and P5 are independent of P1 and of each other,
each amending a different existing standard.

Order by impact: P1+P2 first — an undocumented invariant governing what ships
to adopters is the largest exposure, and the one instance already resolved the
weakest way. P3 next, since two implementers hit it in one run. P4 and P5 are
authoring-quality improvements with no urgency.

Every proposal leaves existing documents valid; none requires a sweep. The only
migration implied is annotating one demoted citation under P2.

## Resolution
