# agent-plugins-92u — `.github/workflows/checks.yml` fragment

The container's PAT lacks the `workflow` scope, so this edit is not in the
bead's commit. Apply it through the GitHub App. Everything else the bead
requires (the deletion, `rfcs/specs.toml`, the doc and AGENTS.md updates,
the `.gitignore` entry) is committed on `bb/agent-plugins-92u`.

Three edits, all to `.github/workflows/checks.yml`. Line numbers are from
the file as of `a56b7ca`.

---

## Edit 1 — `checks` job: delete two steps (lines 98–116)

`scripts/validate-corpus.py` is deleted by this bead, and it was the only
consumer of `jsonschema`/`pyyaml` in this workflow (`check-reference-links.py`
and `check-template-drift.py` import stdlib only), so the pip step goes with
it.

**Delete** everything from line 98 through line 116 inclusive — that is, the
`Install schema validation dependencies` step, the comment block that follows
it, and the `Frontmatter conforms outside the corpus root` step:

```yaml
      - name: Install schema validation dependencies
        run: python3 -m pip install --quiet jsonschema pyyaml

      # Validates frontmatter against the generated JSON Schema for each
      # document's type, for the documents that live OUTSIDE the Spec
      # CLI's corpus root — rfcs/ and the skills' bundled references.
      # specs/ is checked by the `corpus` job below, with the real
      # binary. See the script's own docstring for why the split exists.
      #
      # Not gated anywhere yet: that the generated artifacts under cli/
      # and plugin/skills/spec/references/generated/ are current with
      # schemas/*.strata. That check is scripts/generate-schemas.sh
      # --check, and it needs the strata compiler, which has no
      # installable release binary yet — building it from source would
      # add a Rust toolchain and several minutes to every run. Until then
      # regeneration is a local step, and the drift it would catch shows
      # up as an unexplained diff in review.
      - name: Frontmatter conforms outside the corpus root
        run: scripts/validate-corpus.py
```

After the deletion the `checks` job's last step is
`Templates match their type's schema` (currently lines 93–96), and the
`corpus:` job comment block that begins at line 118 follows it.

---

## Edit 2 — `corpus` job: add the replacement step at the end

The replacement needs the `spec` binary, which only the `corpus` job builds,
so it belongs there rather than in `checks`. **Append** after the existing
`Check every document` step (currently the last step in the job, lines
175–180):

```yaml
      # The RFC queue, checked by the same binary through a second
      # specs.toml (rfcs/specs.toml, root "."). RFCs carry the same
      # frontmatter as corpus members and drift the same way, but they
      # are their own slug namespace and cite documents outside that
      # root, so the rule set is narrowed to the document-local rules.
      # The corpus-global rules (reference/*, slug/duplicate) are
      # deliberately not requested — cross-root citations would report as
      # unresolved. The report renders them not-requested rather than
      # passed, so the narrowed scope is visible rather than implied.
      #
      # Replaces scripts/validate-corpus.py, which globbed rfcs/ with a
      # hand-rolled JSON Schema validator and did not check slug patterns
      # at all.
      - name: The RFC queue conforms
        run: |
          ./spec check --config rfcs/specs.toml \
            --only frontmatter/parse,frontmatter/schema,frontmatter/missing-property,slug/pattern-mismatch
```

Verified locally against `bb/agent-plugins-92u` with a binary built from
`cli/`: exit 0, `checked 33 documents`, `no findings`; the four
`rfcs/sprints/**` and `rfcs/retro-log.md` files are reported as
`excluded` (three `no-frontmatter`, one `unknown-type "sprint-report"`),
and the four requested rules all render `passed`.

---

## Edit 3 — header comment (lines 16–24): drop the dead script name

The workflow header still narrates `scripts/validate-corpus.py`.

**Replace** lines 16–24:

```yaml
# The corpus job is the dogfooding one. It exists because the gate that
# preceded it did not do what its name claimed: scripts/validate-corpus.py
# globbed docs/**/*.md, and the specs/ reorganization (4388c06) emptied
# that path. It kept reporting "checked 26 documents" — every one of them
# an RFC or a skill reference — while every document under specs/ went
# unchecked from the reorg until this job landed. A gate that reports a
# healthy count over the wrong files is worse than no gate, so this job
# names its root through specs.toml and the script below now fails rather
# than passing over an empty set.
```

with:

```yaml
# The corpus job is the dogfooding one: this repository's own documents,
# checked by the binary it ships. It names its roots through specs.toml
# — specs/ at the repository root, rfcs/ through rfcs/specs.toml — rather
# than through globs, and the corpus/empty rule fails a root that matched
# nothing rather than reporting a healthy count over an empty set.
```

Also update lines 10–14, which still describe the `checks` job as owning
python3 dependencies and split the two jobs by that boundary. Replace:

```yaml
#   checks — are this repository's skills and vendored copies intact?
#            bash and python3 only.
#   corpus — do this repository's own specification documents conform,
#            according to the Spec CLI it ships? Needs a Go toolchain to
#            build that binary.
```

with:

```yaml
#   checks — are this repository's skills, vendored copies and templates
#            intact? bash and stdlib python3 only, no pip install.
#   corpus — do this repository's own specification documents conform,
#            according to the Spec CLI it ships? Needs a Go toolchain to
#            build that binary. Covers specs/ and rfcs/.
```

---

## Acceptance-criteria mapping

| Criterion | Satisfied by |
|---|---|
| `checks.yml` runs `spec check` over `rfcs/` with the four `--only` rules in place of the "Frontmatter conforms outside the corpus root" step | Edits 1 + 2 |
| That step passes green | verified locally, exit 0 (`agent-plugins-8ko` already closed) |
| No second root over `plugin/` | Edit 2 adds one root, `rfcs/`, only |
| No remaining reference to `validate-corpus.py` in CI | Edits 1 + 3 |
