# Skills

The Specs plugin includes the following skills. Run each one as a slash
command, such as `/spec`, or let Claude choose it from what you ask for.

| Skill | Description |
|:---|:---|
| **`spec`** | Writes every [document type](document-types.md) from the framework's templates. Writes acceptance criteria in Gherkin, other requirements in [EARS](https://alistairmavin.com/ears/), and architecture in the C4 model. Chooses the right type for what you describe and doesn't create new types. Runs readiness checklists before a document goes to a reviewer. |
| **`groom-feature-blueprint`** | Turns an approved Feature Blueprint into work for agents. Requires an approved blueprint with no open questions, and then breaks its changes into a graph of Work Orders that carry acceptance criteria coverage and a definition of done. If the skill can't plan a blueprint mechanically, it returns the blueprint instead of guessing. |
| **`groom-beads`** | Applies the Definition of Ready to Work Orders stored as beads. Creates the beads, connects their dependencies, assigns each one a capability tier, validates the result, and shows which beads would conflict if they ran in parallel. The goal is that a new subagent with no context can run `bd show <id>` and finish the work without asking questions. |
| **`retro-spec`** | Runs a retrospective on the *documents* that a completed phase used. You can judge a spec's quality through the work that uses it, so this skill runs after a later phase uses the spec. It reports a ranked list of issues against your specs, and you choose which ones to file. |
| **`retro-skill`** | Runs a retrospective on the *skill* that an agent just ran. It cites the instruction that misled the agent, names the script that failed, and records the model that ran. It ranks issues against the skill so that you can file them with the skill's owner as a bead, a GitHub issue, or a Markdown file. |
| **`tech-writer`** | Drafts, revises, and audits technical documentation in your project's style guide, which is the Google developer documentation style guide unless you name another. It adds rules the style guide leaves open, a consistent voice, and a tone set by a reader profile for each document type. The `spec` skill uses it for every document it writes. It also has a reader profile for skills and other docs written only for agents. |

## Related

- [Plan a feature for agents](../how-to/plan-a-feature.md) runs `spec`,
  `groom-feature-blueprint`, and `groom-beads` in order.
- [Run a retrospective](../how-to/run-a-retrospective.md) covers `retro-spec`
  and `retro-skill`.
