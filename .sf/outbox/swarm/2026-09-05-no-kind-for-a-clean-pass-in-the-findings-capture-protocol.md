---
title: "Say how a clean pass is recorded, and that a kindless retro: comment is not a finding"
description: The findings capture protocol says a clean run records nothing and gives no way to record the affirmation an agent wants to leave, so a kindless retro: comment routes nowhere and shows up as an unrouted finding.
slug: issue-no-kind-for-a-clean-pass-in-the-findings-capture-protocol
audience: skill
target: findings-capture-protocol
priority: 4
updated_at: 2026-09-05T00:00:00Z

# Evidence — what this issue rests on
findings: ["rc14.3#3"]
recurrence: 0
report: docs/retros/2026-09-05-59ic-spec-corpus-anchor-references.md

# Provenance — who and what produced this
run: agent-plugins-59ic
session: https://claude.ai/code/session_0185zfzFm9maHZjPX4FQqhbZ
branch: sprint/agent-plugins-59ic
skill_version: "findings-capture-protocol 0.6, vendored into swarm-retro from brettcallaghan/agent-plugins at commit 81ae08ece301 (2026-09-03)"
model: "<derived, not observed — agent-plugins-rc14.3 carries tier `standard`, which this project's [models] maps to sonnet; nothing records what actually served the dispatch>"
---

**Routing note for whoever receives this.** The `swarm-retro` skill vendors
this standard and harvests against it, which is why it arrives through the
outbox — but the fix does not land in `brettcallaghan/swarm`. The vendored copy
at `references/shared/findings-capture-protocol.md` opens `GENERATED COPY — DO
NOT EDIT. Source of truth: brettcallaghan/agent-plugins
specs/standards/findings-capture-protocol.md (version 0.6)`, so the edit is
made there and re-vendored with `scripts/sync-shared-references.sh`. This issue
was filed to the outbox on an explicit human ruling, over the retro's own merge
test, which put it in the originating project's tracker. Route it accordingly
rather than acting on it in place.

## What is wrong, and what it costs

The protocol says "**A clean run is fine — record nothing.** Do not manufacture
findings." It offers no kind for an agent that nevertheless wants to record
that it looked and found nothing — which is a reasonable thing to want, because
"no finding was recorded" and "the agent checked and there was nothing" are
different states and the record cannot tell them apart. An agent that writes
the affirmation anyway has nothing to write it as, so it writes a `retro:`
comment with no kind.

The cost is small and exactly one thing: **that comment routes nowhere.** The
retro's routing is a lookup on the kind, and a kind that is missing or unknown
lands in the UNROUTED bucket, which the skill treats as the taxonomy's health
metric — "a rising one means the kinds are drifting from what agents actually
hit". One kindless affirmation per run is not drift, but it is
indistinguishable from drift at the point where the count is read, and it
inflates the one number that is supposed to be a signal.

## Evidence

- `rc14.3#3`, the only unrouted finding in 42 on this run, in full:
  "`retro: clean — no scope gaps found beyond the acceptance criteria.
  AC-SPECANCH-001.1/.2/.5 each got a citation comment on a dedicated (or
  extended) unit test in cli/core/references/references_test.go; the
  AC-SPECANCH-001.6 (frontmatter anchor) and AC-SPECANCH-002.7 (embed)
  participation tests were added without an AC: citation, per this bead's own
  note that those citations belong to SREF and ARES respectively.`" It carries
  no kind, and the agent was doing nothing wrong.

- The protocol's own text, under "The kinds": "**A clean run is fine — record
  nothing. Do not manufacture findings.**" Nine kinds are defined and none of
  them is an affirmation.

- The routing rule this collides with: "kind missing or unknown → UNROUTED — a
  count and a list, never dropped", and "Unrouted findings are the taxonomy's
  health metric. Report the count every run."

## Proposed change

Add one sentence to the protocol, in the "A clean run is fine" paragraph:

> A clean pass is recorded by recording nothing; there is no kind for it. A
> `retro:` comment carrying no kind is therefore not a finding, and a harvest
> counts it as neither routed nor unrouted.

Done is: the standard says it, `scripts/sync-shared-references.sh` has
re-vendored the copy, and a run whose only kindless `retro:` comment is a
clean-pass affirmation reports an unrouted count of **0** rather than 1.
