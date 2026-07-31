# Retro note — FRD authoring phase, 2026-07-31

**Outcome: no project-doc finding cleared the bar.** All friction
observed in this phase root-caused to the guidance, not to the documents
the phase consumed.

## Phase

The `spec` skill (v1.4, plugin 0.9.0) was run twice in one session:
first to author `docs/adrs/ADR-014-gherkin-for-acceptance-criteria.md`,
then to author `docs/frds/FRD-spec-validation-cli.md` (9 requirements,
58 acceptance criteria). Model `claude-opus-5`; harness Claude Code on
the web. Effort not observable from within the session.

## Documents examined

| Consumed | Held up? |
|----------|----------|
| `docs/plans/spec-validation-cli-brief.md` | Yes — every scope decision in the FRD traces to a question answered in it |
| `docs/adrs/ADR-014-gherkin-for-acceptance-criteria.md` | Yes — its deferral and its reversibility constraint were both usable as written |
| `docs/adrs/ADR-008-element-record-blocks.md` | Yes — verified during the retro. Its Consequences → Negative already carry "Tooling must extract and validate blocks in addition to frontmatter" and "The block-XOR-link invariant must be enforced or a promoted unit could leave a stale block behind" — the first is what REQ-SPEC-002.6 specifies, the second what REQ-SPEC-005.7 partially covers |
| `docs/adrs/ADR-012-reference-conventions.md` | Yes — `:57` covers backticked paths for framework guides |
| `docs/adrs/ADR-002-ears-for-requirements.md` | Yes — cited for the notation the acceptance criteria are written in |

## Where the friction went

Five findings were harvested. All five root-caused to the `spec` skill,
its templates, or their contradiction with a standard — none to a
project document's content. They are filed as
`rfcs/inbox/rfc-self-spec-reference-bundles-and-adr-slug.md`
(2 blockers, 2 snags, 1 polish), per this skill's step 3 routing rule.

## Considered and not proposed

- **The 11 non-conforming ADR slugs.** A real defect in consumed
  project documents: `[[adr-002]]` and `[[adr-008]]` both dangle from
  the FRD authored in this phase, so the defect is now propagating into
  new documents rather than sitting still. Not filed here because it is
  already tracked as bead `agent-plugins-5nz` with the fix specified,
  and its root cause — the ADR template's slug instruction — is P1 of
  the skill RFC. A project RFC would add a third record of one defect.
- **`docs/adrs/README.md` is missing an entry for ADR-013.** Observed
  while adding ADR-014's entry. Not filed as a project finding because
  it caused no friction in this phase — it was noticed incidentally,
  not encountered as an obstacle — and its likely cause is the
  template's broken path (`docs/adr/README.md`), which is P3 of the
  skill RFC. The missing line is a one-line content fix for whoever
  next touches that index.
- **The FRD ships with `sources: []` despite being derived from a
  600-line research brief.** The document is not at fault: the brief is
  correctly outside the specification set and correctly has no slug, so
  there is nothing for `sources` to hold. The gap is in what the
  frontmatter property accepts, which is P4 of the skill RFC.

## Note

A clean project-doc retro this early is weak evidence, not strong: the
corpus is small, this phase consumed five documents, and four of the
five ADRs were read narrowly for a single fact each. The result says
those documents held up for what this phase asked of them — not that
they would hold up for a phase that leaned on them harder.
