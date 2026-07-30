## TODO

RFC triage
- What are the themes for this project?
- Always reference the skill(s) the proposal would change.
- Every skill should have a reference to its source repo or a repo to receive retros.

Refine the groom-beads skill, adopting the conventions (e.g. c4, references, etc) of the surrounding skills. Make it belong.

Adopt "work order" or __"dev task"__ or "task" or "bead"?

The one-agent CAB.
- You are a one agent Customer Advisory Board.
- You are the biggest fan of the <product-name> product.
- You are excited about the product and its new releases.
- You use the tool and care about its success.
- You are constantly evaluating the tool against its known use cases, and looking for opportunities to improve.
- Your feedback is constructively critical, honest, and to the point -- that's why the product's creators value you so much.
- You have access only to its documentation, and the tool itself (CLI, web app, etc), not it's source code or implementation details.
- You receive early previews of its upcoming roadmap, and provide feedback on which upcoming features you are most excited about.
- You are constantly evaluating the tool for new use cases, and providing feedback to the Product Team on enhancements that would support the most promising use cases.
- You are a problem solver, and find workarounds or hacks to make the tool work. These hacks turn into how-to guides, and/or enhancement requests when they become cumbersome.

Decision on whether to adopt Gherkin.
- Gherkin files or markdown blocks?
- What parsers are available and can they be worked into scripts (python, node, etc) that we can run?
- Can we tag features and acceptance critera that can be extracted and validated by the parser?
- Does Gherkin have a formal AST in JSON that is easy to extract?

C4 conventions: non-standards to adopt (opinionated).
- Acronyms
    - No internal acronyms.
    - Only common industry-standard acronyms are allowed.
    - Internal acronyms MAY be included in the description, or in parentheses after the title (if there's room).
- Relationships (arrows) point in one direction .
    - Relationships (arrows) tell a story about the most important dependencies or data flow.
    - `<source-title> -> <target-title>` should read like a sentence: “My Web App requests widgets from My Headless API”
    - Relationships should point in the direction of the dependency: client -> server, consumer -> publisher.
- Message buses / async
    - Avoid including message buses in diagrams UNLESS the bus is the context you're trying to diagram.
    - RATIONALE: Diagrams where many containers point at a centralized bus are incomprehensible.
    - Instead, arrows should point from the message consumer to the message producers.
    - When space is available, reference the technology and/or topic name in the relationship.

Bead Bonanza
- If a wave doesn't go well, course correct and adapt.
- Dive into alleged handoff / synchronization issues with the JSONL beads setup (no shared Dolt service).
- Gates for what beads to pull in?
- Have the orchestrator agent (or sub-agent) use `groom-beads` skill on beads created during the bonanza session? Agent that finds the bead should create the bead.
- What will we find in our RFC inbox related to this skill?
- Beads that contain open decisions don't pass the start gate -- run `groom-beads` to complete.
- 

Config
- An enterprise standards repo
- Location for specs (defaults to `@specs/`)
- Location for rfcs (defaults to `@specs/rfcs/`)
- Location for a rfc outbox (defaults to `@specs/rfcs/outbox/`)
- No beads? Location for "tasks" | "work orders" | "dev tasks".
- Checks that should be run after every bead is implemented (e.g. `compile`, `lint`, `check-format`, ...)

`spec-review` skill
- Reviews PRs to ensure specs are compliant with all guidelines and acceptance criteria
- May scripts to perform deterministic checks
