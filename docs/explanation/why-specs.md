# Why specs

Agents can plan, write, and test code faster than a team can by hand. An AI
software factory builds what its inputs describe. Feature and architecture
specs are those inputs. They help agents build the software that you
intended.

## Agency, not supervision

Vibe coding works well for exploring an idea or building a demo. To build
systems, teams need to know how to adopt agentic development at scale and how
to trust what the agents produce.

More supervision doesn't answer that question. Agency does. If an agent must
stop and ask a person every time the plan and reality diverge, that person is
on the critical path. If an agent has no written intent, it doesn't stop. It
guesses.

An agent can act safely on its own only when it can make good decisions, and
specs guide those decisions. The following table shows what an agent does
when the plan meets reality during a run:

| Situation during a run | Without specs, the agent | With specs, the agent |
|:---|:---|:---|
| An acceptance criterion is missing or contradicts another feature. | Picks one interpretation and continues without saying so. | Detects the conflict against the FRD and returns it before it writes code. |
| The architecture must change in a way that planning didn't anticipate. | Invents a structure that fits the file that it's editing. | Reads the change guidance in the Container or Module Blueprint and stays inside the boundary, or reports that the boundary must move. |
| An unexpected blocker appears. | Works around it and invents new behavior. | Checks the blocker against the Standards and ADRs that already settled the question. |
| The scope grows. | Absorbs the extra work, and the diff grows with it. | Recognizes that the work is outside the blueprint's deltas and files it as new work. |

Clear intent leads to good decisions, and good decisions lead to a better
implementation.

## Why a CLI

Specs take effort for agents to write and maintain. When an agent runs a
conformance check by reading documents, the check is expensive, and nobody
can verify it: an agent that claims a checklist passed produces the same
output as an agent that ran it.

The `specs` CLI moves that work into deterministic commands. A reviewer can
rerun any check the CLI made and get the same answer. For the commands, see
[The `specs` CLI](../reference/cli.md).

## Related

- [Workflow](workflow.md): how work moves from a product definition to a
  reviewed pull request.
- [How the document model works](document-model.md): the document types and
  how they fit together.
