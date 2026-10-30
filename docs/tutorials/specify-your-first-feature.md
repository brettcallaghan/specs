# Specify your first feature

In this tutorial, you install the Specs plugin, set up a spec corpus in a
repository, and write your first feature requirements document (FRD) with the
`spec` skill. Then you check it with the `specs` CLI.

You need the following:

- Claude Code.
- A repository to work in. A new, empty one is fine.
- A feature you want to build, small enough to describe in a sentence or two.

## Install the plugin

1. In Claude Code, add the `brettcallaghan` marketplace and install the
   plugin:

   ```
   /plugin marketplace add brettcallaghan/agent-plugins
   /plugin install specs@brettcallaghan
   ```

2. Start a new session in your repository. When the session starts, the
   plugin puts the `specs` CLI release that matches its own version on the
   session's `PATH`. It downloads that release the first time and reuses it
   after that.

3. Ask Claude to run `specs --version`.

   The CLI prints its version, revision, and build time as YAML. If Claude
   can't find `specs`, ask it for the line that begins `specs plugin:` in the
   session's start-up context. That line names the cause and the fix. For more
   help, see [Install the plugin and the CLI](../how-to/install.md).

## Create the spec corpus

The skills and the CLI find your specs through a `specs.toml` file.

1. At the root of your repository, create `specs.toml` with the following
   content. Replace `my-project` with your project's name.

   ```toml
   [workspace]
   name = "my-project"
   root = "specs"

   [profile]
   name = "specs"
   ```

   `root` is the spec root: the directory that holds every spec. The `specs`
   profile supplies the document types that the `spec` skill writes.

2. Create the `specs` directory.

3. Ask Claude to run `specs check`.

   The check fails with one error, because the corpus is still empty:

   ```
   message: "spec root \"specs\" holds no corpus members: the run checked zero documents"
   ```

   That error means the CLI found your `specs.toml` and read the spec root
   from it. The next step gives it something to check.

## Write the FRD

An FRD describes one feature: why users need it, as a user story, and what
it must do, as acceptance criteria. It never describes how to build the
feature.

Describe your feature to the `spec` skill. For example:

```
/spec Users can export a report as a CSV file.
```

The skill confirms that the document is an FRD, writes it from the FRD
template, and saves it as
`specs/features/<feature-name>/<feature-name>-frd.md`. Before it sets the FRD
to `ready-for-review`, it runs `specs check` on the file and then the
document readiness checklist, and fixes what both report.

## Read what the skill wrote

Open the new FRD and look for the following parts:

- **Front matter:** the `title`, `description`, `slug`, `type: frd`, and
  `status`. Every spec carries these five properties, which is what lets the
  CLI index the corpus.
- **The user story:** who wants the feature, and why.
- **Acceptance criteria:** tagged Gherkin scenarios. Each feature has a
  `FEAT-<CODE>-NNN` tag, and each scenario has an `AC-<CODE>-NNN.N` tag. Tests
  cite these ids to show which criterion they verify.

If a criterion doesn't say what you meant, tell the skill what to change.
This is the cheapest moment to fix it: before any design or code depends on
it.

## Check the FRD

1. Ask Claude to run `specs check` again.

   The `summary` block at the top of the output counts `errors`, `warnings`,
   and `unverified` findings. Each finding names its `rule`, the `path` and
   line of the problem, and a `message` that says what is wrong.

2. If the check reports findings, ask Claude to fix them, and then run the
   check again. Repeat until it reports zero errors.

3. Ask Claude to run `specs list`.

   The list shows your FRD's path, slug, type, status, title, and
   description, all read from its front matter.

A reviewer can run the same two commands and get the same answer. That's the
difference between a check the CLI ran and a check an agent says it ran.

A clean `specs check` doesn't mean the FRD is finished, though. The document
readiness checklist names the checks the CLI doesn't make, and the skill
makes those by reading the document.

## Approve the FRD

When the acceptance criteria say what you want built, change the FRD's
`status` to `approved`. An approved FRD is ready for the next
step: an architecture design and a Feature Blueprint that maps the feature
onto it.

## Next steps

You've installed the plugin, set up a spec corpus, and written and checked an
FRD. From here:

- To take the feature on to a design and a plan that agents can run, see
  [Plan a feature for agents](../how-to/plan-a-feature.md).
- To see the other document types, see
  [Document types](../reference/document-types.md).
- To understand why an FRD comes first, see
  [Why specs](../explanation/why-specs.md).
