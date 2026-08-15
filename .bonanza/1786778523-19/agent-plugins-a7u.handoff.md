# agent-plugins-a7u handoff

All 152 prose acceptance criteria across the six FRDs under `specs/features/`
are now tagged Gherkin — the bead's "73" counted the corpus as it stood on
2026-08-06, before `markdown-links`, `anchors` and `rename` existed. The set of
AC identifiers is byte-identical before and after (verified by diffing the
extracted identifier set against the pre-edit documents: 152 in, 152 out, none
missing, none invented). `spec check` is clean over the whole corpus, every
`gherkin/*` rule passing. `REQ-SPECMD-*` became `FEAT-SPECMD-*`, completing the
FEAT rename that arrived after `markdown-links` merged; `rfcs/` is left alone,
as the earlier rename did. ADR-002 gains a Scope section narrowing it to non-AC
normative statements, and both readiness checklists now name which notation each
surface uses — the document checklist gains a check 9 for Gherkin, the blueprint
checklist states that a delta's behavior lines stay EARS. Three defects the
migration exposed are fixed in the guide itself: any non-empty `Scenario`
description is reported (not only a normative one), a `Rule:` claims every
scenario after it regardless of indentation, and a `[[slug]]` inside a block is a
real reference so wikilink syntax cannot be illustrated there. One emitter defect
is filed rather than fixed: `agent-plugins-h45m`.

Caveat: this closes agent-plugins-oq8 as done-by rather than duplicated — the spec skill's reference table now routes "Specifying a feature" and "Writing a feature or an acceptance criterion" to `references/shared/gherkin-guide.md`, step 3 of the authoring workflow states that an FRD's features and acceptance criteria are tagged Gherkin per adr-021 with EARS keeping everything else, and `scripts/sync-shared-references.sh` vendors the guide into the skill so the copy stays current.
