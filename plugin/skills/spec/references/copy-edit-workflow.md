# Copy-edit workflow

Use this workflow to revise many existing documents at once, such as a
migration to a new template or a copy edit across the corpus. The
orchestrating agent runs it, because a subagent can't start the reviewer that
the `tech-writer` skill requires.

## Group the documents

- Give each editor one host blueprint and the Component Blueprints it
  composes, so one agent reads shared code and makes the terminology choices
  for siblings. Group other document types by directory.
- Give an editor at most about four documents. Split a larger host across
  editors.
- Give every editor the same terms file. An editor records each terminology
  decision there and reads it before choosing a term. A term the project
  glossary defines doesn't go in the file: the glossary wins. When the
  editors finish, the orchestrator moves the file's terms into the glossary.

## Brief each editor

Tell each editor to:

1. Load the `tech-writer` skill and each document's template.
2. Revise each document under the `tech-writer` skill's Revising rules,
   taking the reader profile from the template.
3. For an architecture blueprint, verify every statement against its
   `source_paths` and `test_paths`. Fix a statement the code contradicts and
   cite the file and symbol, or the test. Open each test an "Enforced by:"
   line names and confirm that it checks that invariant.
4. Run `specs check` on each document and fix what it reports.
5. Write a log with these sections: changed sentences, accuracy fixes with
   their evidence, terms recorded, text it couldn't change and the rule that
   stopped it, findings, and judgment calls. The editor doesn't review its
   own draft.

## Review each editor's diff

Start one fresh reviewer per editor. Give it the diff, the editor's log, the
reader profile, and the `tech-writer` skill, not the editor's reasoning. Tell
it to:

- Check each changed requirement statement part by part against the
  original: modal, actor, trigger, condition, scope, and EARS pattern.
- Check each accuracy fix against the evidence the log cites.
- Run the `tech-writer` skill's Patterns to check over the changed text.
- Return each defect with its line and the rule it breaks. Change nothing.

The orchestrator fixes the defects or returns them to the editor, runs
`specs check` on the corpus, and records each agent's token count.
