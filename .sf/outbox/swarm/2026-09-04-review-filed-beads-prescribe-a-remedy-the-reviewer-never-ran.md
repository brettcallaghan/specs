---
title: "Mark a review-filed bead's suggested remedy unverified unless the reviewer ran it"
description: swarm-review files each finding as a bead carrying a prescribed fix the reviewer never measured, and a bead reads as decided work — so the prescription gets implemented rather than re-derived.
slug: issue-review-filed-beads-prescribe-a-remedy-the-reviewer-never-ran
audience: skill
target: swarm
priority: 2
updated_at: 2026-09-04T00:00:00Z

# Evidence — what this issue rests on
findings: ["d7w5.9#2", "d7w5.11#3"]
recurrence: 0
report: docs/retros/2026-09-04-d7w5-spec-corpus-structure-extraction.md

# Provenance — who and what produced this
run: agent-plugins-d7w5
branch: claude/swarm-agent-plugins-5wxrfo
skill_version: "swarm 0.1.0 @ 1a19e8e (brettcallaghan/swarm, 2026-09-02)"
model: "<unrecorded — manifest.models is unset on this run; see d7w5.2.1#3>"
---

## What is wrong, and what it costs

`swarm-review` files each review finding as a bead, and the bead carries the
remedy the reviewer proposed alongside the defect the reviewer measured. The
reviewer measured the defect; it did not run the remedy. The two arrive in one
record with nothing separating them, and a bead is not read the way a review
comment is — it is read as work somebody already decided on. On this run the
reviewer's **evidence was correct every time** and its **prescribed shape was
wrong or unreachable in three of four beads**, so the asymmetry is not
occasional. The cost is paid by the implementing agent, who either implements a
fix that cannot work (`heading.Type()` reads a field the pinned parser does not
have), chases a symptom that is correct behaviour, or ships a partial fix that
leaves the guarantee it was filed to restore still false — and in the one case
where the agent did re-derive it, the cost was an unbudgeted mid-bead scope
expansion of roughly 180 lines. A reviewer that prescribes is doing design
without the measurements design needs; the bead makes that indistinguishable
from design that was done.

## Evidence

- `d7w5.9#2` — "The bead's REJECTED alternative — ask the AST node for the
  heading's written form — was re-measured at the point of use, because the code
  comment I shipped restates it and an inherited claim reads exactly like a
  measured one. evidence: goldmark v1.8.5 in this container,
  /root/go/pkg/mod/github.com/yuin/goldmark@v1.8.5. (1) `sed -n '200,230p'
  ast/block.go` → `type Heading struct { BaseBlock; Level int }` … one kind for
  both forms. (2) `grep -rn \"NewHeading(\" parser/ ast/` →
  parser/setext_headings.go:70, parser/atx_heading.go:95, :103 all call
  `ast.NewHeading(level)`; no other constructor and no form argument. (3) the
  only `SetAttribute` calls are author-written `{#id}` attributes gated on the
  opt-in `WithHeadingAttribute` option; neither parser sets anything that names
  the written form. Conclusion: the public AST genuinely does not carry the
  distinction on the pinned version." The review finding filed as
  `agent-plugins-d7w5.9` had proposed exactly that — its own bead records
  "REJECTED — asking the AST node, which is what the review finding proposed
  (`heading.Type()` / the AST kind)". The prescribed fix had nothing to read.

- `agent-plugins-d7w5.9`, the bead the same review pass filed, records that the
  finding also mis-stated part of the defect it was reporting — corrected at
  grooming, quoted from the bead's Problem section: "NOTE ON THE ORIGINAL
  FINDING, corrected at grooming: the review reported \"A's Section is EMPTY\"
  as part of the defect. With the discriminator fixed, A's Section is still
  empty and that is CORRECT — B's heading line begins immediately after A's, so
  A stands over nothing, and the model documents \"a heading standing over
  nothing carries an empty span rather than an absent one\". The defect to fix
  is the OVERLAPPING Position spans and the list-item Section that excludes
  real content; do not chase the empty section in the blockquote case." The
  measured overlap was real; one of the two symptoms named beside it was not.

- `d7w5.11#3` — "The bead's three named code fixes would have left the
  guarantee still false; two further divergences existed that the bead's
  reproduction did not reach. evidence: the Design named three changes
  (tag-presence in jsonName, the flatten guard, the PkgPath skip). Implementing
  only those leaves structFields as a DEPTH-FIRST walk with no conflict
  handling, while encoding/json resolves fields BREADTH-first and then applies
  dominance rules. … (a) A shallower field does not hide a promoted one of the
  same name. `struct { PromotionShadowed; Name string \`json:\"name\"\` }` —
  JSON: [\"deep\" \"name\"]; YAML before the fix: [\"PromotionShadowed\"
  \"name\"]; **YAML with only the three named fixes: would be [\"name\" \"deep\"
  \"name\"] — the key \"name\" TWICE, carrying two different values.** (b) Two
  embedded structs at one depth offering the same name do not annihilate. …
  DECIDED to port encoding/json's typeFields in full … rather than patch the
  three named spots. … Cost: cli/cmd/yaml.go grew by roughly 180 lines."

## Proposed change

`swarm-review`'s filed bead states the defect and the evidence for it, and
marks any suggested remedy as **unverified** unless the reviewer ran it. In the
bead's own text: the Problem section carries only what was measured — the
reproduction, the observed output, the expected output — and a proposed fix
appears under an explicitly unverified heading that says the reviewer did not
run it and that the implementing agent is to re-derive it. Where the reviewer
did run a remedy, it says so and carries the command and output that show it,
at which point it is a measurement like any other and needs no mark.

Two consequences the wording has to make unavoidable, because both were paid on
this run: a symptom listed beside a measured defect is part of the claim and
carries the same burden (`d7w5.9`'s "A's Section is EMPTY" did not), and an
enumerated remedy ("these three changes") asserts sufficiency, which is a claim
about code the reviewer did not write and must be marked with the rest.

Done is: a bead filed by a review pass contains no imperative fix outside a
section marked unverified, unless the same bead cites the command and output
that establish it — checkable by reading any bead the pass files.

**Overlaps the project-side readiness gate** that this run's retro filed
against `agent-plugins` in the same round: that one requires a work order not
to state a value, signature or fix it derived by inference, and would catch
such a bead at grooming. This issue is the source that produces it; the gate is
the net that catches it. The human accepted both deliberately, so neither is
offered as a substitute for the other.
