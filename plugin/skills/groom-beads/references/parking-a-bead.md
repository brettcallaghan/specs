# Parking a bead that cannot be brought to Ready

Read this at step 2, when a bead will not reach the Definition of Ready.

**Parking takes two writes and only the second one works.** `bd ready` reads
`status`, **not** `grooming_status`. A bead flagged only with the metadata
stays dispatchable and will be handed to a fresh subagent exactly as if it were
ready — which is the failure the flag exists to prevent, made worse by looking
handled.

```
bd update <id> --set-metadata grooming_status=<value>
bd defer <id>          # status -> deferred; drops out of `bd ready`
```

Verify the park took effect: the bead SHALL NOT appear in `bd ready`
afterwards.

## Which value

| value | means | what the next pass owes |
|---|---|---|
| `blocked-needs-info` | a specific fact is missing, or the Design leaves a choice open | get the fact, or put the choice to a human. **The choice is never yours to guess** — a groomer that picks one has silently made a decision the bead exists to surface |
| `blocked-ungroomable` | the bead needs redesign, not information | send it back to whoever scoped it. No amount of grooming fixes a bead that describes the wrong work |

The distinction is what tells the next pass whether to go looking or to go
back. Recording neither leaves it to re-derive both.

## Then say so

Flag every parked bead in the grooming report, with its value and its reason.
A queue that silently narrowed is the failure; one that narrowed and said so is
the feature.
