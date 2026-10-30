# Run a retrospective

A retrospective turns the friction from one phase of work into issues you can
file. Specs has two, and they look at different things:

- **`retro-spec`** looks at the *documents* a phase used, such as the FRD and
  architecture blueprints that a Feature Blueprint was written from. Run it
  after a later phase has used those documents, because that's when their
  gaps show.
- **`retro-skill`** looks at the *skill* an agent just ran, such as `spec` or
  `groom-feature-blueprint`, or a skill from another plugin. Run it right
  after the run, especially one where the skill struggled, improvised, or got
  something wrong.

Neither writes a file to your repository. Each one reports in the session,
and you choose what to file.

## Run a retrospective on your specs

1. In the session that finished the phase, run:

   ```
   /retro-spec
   ```

   The `spec` skill also offers this after it writes a document from other
   documents.

2. Read the report. It lists the phase and the documents it examined, then
   candidate issues ranked from priority 0 (critical) to 4 (backlog). Each
   candidate names its target document, the evidence it rests on, and the
   change it proposes.

   Items that didn't clear the bar are listed with the reason. Findings that
   trace back to a skill, template, or script are listed as guidance
   friction.

3. Tell the skill which candidates to file. You can file any of them, change
   their order, or file none.

The skill files each one in your project's tracker: one bead per issue if you
use beads, otherwise one issue in your tracker. If the report listed guidance
friction, run `retro-skill` next.

## Run a retrospective on a skill

1. In the session where the skill ran, run:

   ```
   /retro-skill
   ```

2. Read the report. Each candidate issue cites the instruction that misled
   the agent or names the script that failed, and the report records the
   model that ran the skill.
3. Tell the skill which candidates to file, and change any destination you
   want to.

Where an issue goes depends on who vends the skill:

- **Your project:** a bead in your project's tracker.
- **Another repository that accepts issues the session can file:** an issue
  in that repository, such as a GitHub issue.
- **Anyone else:** a standalone Markdown file outside source control, named
  `<YYYY-MM-DD>-<plugin>-<target>-<slug>.md`, for you to send to the skill's
  owner.

An issue about a skill your project doesn't vend never goes in your own
tracker, because its owners can't act on it there.

## When nothing clears the bar

A retrospective that finds nothing says so, and names what it examined. That
tells you the documents or the skill held up, and there's nothing to file.
