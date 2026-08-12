#!/usr/bin/env python3
"""Check every skill template's property set against its type's schema.

A template's values are placeholders, so it cannot be validated as an
instance — `updated_at: <YYYY-MM-DDTHH:MM:SSZ>` is deliberately not a
timestamp. Its structure can be: every document authored from a template
inherits the template's property set, so that set must match the
definition the template's declared `type` selects.

What it reports, and only this:

  * a property the type's definition requires that the template omits
  * a property the template declares that the definition does not define
  * a declared `type` outside the DocumentType vocabulary

Templates ship inside the plugin, so a drifted one is a defect
distributed to every adopting repository, authored into every document
copied from it.

Reads the frontmatter keys and the generated schema bundle, nothing else:
no corpus, no third-party dependency, and deliberately no YAML parse.
Template frontmatter is not required to be valid YAML and several blocks
are not — a placeholder such as `<one line: what drove the choice>` puts
a bare `:` in a value — so keys are read from the source text.

Usage:
  scripts/check-template-drift.py            # report findings, exit 1 on any
  scripts/check-template-drift.py --quiet    # exit status only
"""
from __future__ import annotations

import json
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent

# Templates are identified by suffix, never by the substring "template":
# template-authoring-standards.md is a standards document, and a
# substring match silently excludes it from whatever gate applies the
# rule.
SEARCH = "plugin/skills/*/references/*-template.md"

# Not matched by SEARCH, one directory deeper: the vendored copies under
# references/shared/. Their authored sources are checked here, and
# sync-shared-references.sh --check compares the copies against those
# sources byte for byte, which is stronger than re-deriving this
# judgement about a copy.

# The generated bundle, and the enum naming the document types a profile
# may register.
BUNDLE = "cli/core/schema/spec-model.schema.json"
VOCABULARY = "DocumentType"

# A floor on how many templates must be found. The failure this guards
# against is not "a template is wrong" but "this gate is pointed at
# nothing and says so cheerfully". Deliberately well below the current
# count: a smoke alarm for a glob that stopped matching, not an inventory
# to keep in sync.
MIN_TEMPLATES = 8

# A frontmatter key: a property name at column zero, before its colon.
# Anything indented belongs to a value, and a commented-out key is not
# declared.
KEY = re.compile(r'^["\']?([A-Za-z_][A-Za-z0-9_]*)["\']?[ \t]*:', re.MULTILINE)

TYPE = re.compile(r"^type:[ \t]*(\S+)", re.MULTILINE)


def frontmatter(path: pathlib.Path) -> str | None:
    """Return the raw frontmatter text, or None if the file has none."""
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        return None
    end = text.find("\n---", 4)
    return text[4:end] if end != -1 else None


def definition_name(declared: str) -> str:
    """Return the schema definition a document type selects.

    Derived from the type's own spelling rather than transcribed into a
    table, so a type added to the vocabulary is covered here with no
    second list to maintain.
    """
    return "".join(part.capitalize() for part in declared.split("-"))


def main() -> int:
    quiet = "--quiet" in sys.argv

    bundle_path = ROOT / BUNDLE
    if not bundle_path.exists():
        print(f"error: {BUNDLE} is missing — run scripts/generate-schemas.sh",
              file=sys.stderr)
        return 1
    bundle = json.loads(bundle_path.read_text(encoding="utf-8"))
    definitions = bundle.get("definitions", {})
    vocabulary = set(definitions.get(VOCABULARY, {}).get("enum", []))

    findings: list[tuple[str, str]] = []
    checked = 0

    for path in sorted(ROOT.glob(SEARCH)):
        rel = str(path.relative_to(ROOT))
        checked += 1

        raw = frontmatter(path)
        if raw is None:
            findings.append((rel, "template/no-frontmatter: a template declares the "
                                  "frontmatter every document authored from it carries"))
            continue

        match = TYPE.search(raw)
        if match is None:
            findings.append((rel, "template/no-type: no `type` property, so there is no "
                                  "definition to check the property set against"))
            continue

        declared = match.group(1)
        if declared not in vocabulary:
            findings.append((rel, (
                f"template/unknown-type: `{declared}` is not in the {VOCABULARY} "
                f"vocabulary ({', '.join(sorted(vocabulary))}), so no schema governs "
                "documents authored from this template"
            )))
            continue

        defname = definition_name(declared)
        definition = definitions.get(defname)
        if definition is None:
            findings.append((rel, (
                f"template/no-definition: `{declared}` is in the {VOCABULARY} "
                f"vocabulary but the bundle defines no {defname}"
            )))
            continue

        keys = set(KEY.findall(raw))
        required = set(definition.get("required", []))
        known = set(definition.get("properties", {}))

        for name in sorted(required - keys):
            findings.append((rel, (
                f"template/missing-required: `{name}` is required of every {declared}, "
                "so a document authored from this template is born non-conforming"
            )))
        for name in sorted(keys - known):
            findings.append((rel, (
                f"template/unknown-property: `{name}` is not defined for {declared}, "
                "so a document authored from this template carries a property no "
                "schema governs"
            )))

    if checked < MIN_TEMPLATES:
        findings.append(("scripts/check-template-drift.py", (
            f"search/underpopulated: matched {checked} templates against a floor of "
            f"{MIN_TEMPLATES}. Either the templates moved or SEARCH stopped matching. "
            "Fix the pattern — do not lower the floor to make this pass."
        )))

    if not quiet:
        for rel, message in findings:
            print(f"error  {rel}: {message}")
        if findings:
            print()
        print(f"checked {checked} templates, {len(findings)} findings")

    return 1 if findings else 0


if __name__ == "__main__":
    sys.exit(main())
