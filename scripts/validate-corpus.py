#!/usr/bin/env python3
"""Validate this repository's documents against the generated JSON Schemas.

This is an interim gate. The Spec CLI described by [[frd-spec-validation-cli]]
will subsume it and do far more — spans, reference integrity, required
sections, line width. Until that binary exists, this script proves the
generated schemas are correct and catches frontmatter defects in CI.

What it checks, and only this:

  * every markdown file whose frontmatter parses and declares a known
    `type` is validated against that type's definition in
    cli/core/schema/spec-model.schema.json
  * specs.toml is validated against cli/core/schema/specs-config.schema.json

Files with no frontmatter, and files declaring a type outside the
profile's vocabulary, are reported as excluded rather than dropped
silently — the same treatment AC-SPEC-001.3 and AC-SPEC-001.4 require of
the CLI, so that a document which should be a specification but is
missing frontmatter is visible rather than absent.

Requires: pip install jsonschema pyyaml

Usage:
  scripts/validate-corpus.py            # report findings, exit 1 on any
  scripts/validate-corpus.py --quiet    # exit status only
"""
from __future__ import annotations

import json
import pathlib
import sys
import tomllib

try:
    import yaml
    from jsonschema import Draft7Validator
except ImportError:  # pragma: no cover
    sys.exit("error: pip install jsonschema pyyaml")

ROOT = pathlib.Path(__file__).resolve().parent.parent

# Where documents are looked for. The corpus proper is docs/; the RFC
# queue and the spec skill's own reference documents are checked too,
# because they carry the same frontmatter and drift the same way.
SEARCH = ["docs/**/*.md", "rfcs/**/*.md", "plugin/skills/*/references/*.md"]

# Templates carry deliberate placeholder slugs and type values, and are
# not corpus members. AC-SPEC-004.4 names this explicitly.
EXCLUDE_NAME_CONTAINS = ["template"]

# type -> definition name in the bundled schema. This mapping is what a
# profile registers; it lives here until the Go profile module exists.
TYPE_TO_DEF = {
    "adr": "Adr",
    "frd": "Frd",
    "contract": "Contract",
    "standards": "Standards",
    "system-blueprint": "SystemBlueprint",
    "container-blueprint": "ContainerBlueprint",
    "module-blueprint": "ModuleBlueprint",
    "component-blueprint": "ComponentBlueprint",
    "feature-blueprint": "FeatureBlueprint",
    "rfc": "Rfc",
    "guide": "Guide",
}


def read_frontmatter(path: pathlib.Path) -> str | None:
    """Return the raw frontmatter text, or None if the file has none."""
    text = path.read_text(encoding="utf-8")
    if not text.startswith("---\n"):
        return None
    end = text.find("\n---", 4)
    return text[4:end] if end != -1 else None


def main() -> int:
    quiet = "--quiet" in sys.argv

    bundle_path = ROOT / "cli/core/schema/spec-model.schema.json"
    config_path = ROOT / "cli/core/schema/specs-config.schema.json"
    if not bundle_path.exists():
        return err("generated schemas are missing — run scripts/generate-schemas.sh")

    bundle = json.loads(bundle_path.read_text())

    findings: list[tuple[str, str]] = []
    excluded: list[tuple[str, str]] = []
    checked = 0

    seen: set[pathlib.Path] = set()
    for pattern in SEARCH:
        for path in sorted(ROOT.glob(pattern)):
            if path in seen:
                continue
            seen.add(path)
            if any(s in path.name for s in EXCLUDE_NAME_CONTAINS):
                continue
            rel = str(path.relative_to(ROOT))

            raw = read_frontmatter(path)
            if raw is None:
                excluded.append((rel, "no-frontmatter"))
                continue
            try:
                data = yaml.safe_load(raw)
            except yaml.YAMLError as exc:
                findings.append((rel, f"frontmatter/parse: {str(exc).splitlines()[0]}"))
                continue
            if not isinstance(data, dict) or "type" not in data:
                excluded.append((rel, "no-type"))
                continue

            declared = str(data["type"])
            defname = TYPE_TO_DEF.get(declared)
            if defname is None:
                excluded.append((rel, f"unknown-type: {declared}"))
                continue

            checked += 1
            schema = dict(bundle)
            schema["$ref"] = f"#/definitions/{defname}"
            # YAML coerces timestamps and versions before a validator can
            # see the source text; default=str renders them back to the
            # string form the schema expects. It does NOT recover the
            # author's original spelling — validating raw property text
            # is AC-SPEC-003.2, and belongs to the CLI, not here.
            payload = json.loads(json.dumps(data, default=str))
            for e in sorted(Draft7Validator(schema).iter_errors(payload),
                            key=lambda e: list(e.absolute_path)):
                loc = "/".join(map(str, e.absolute_path)) or "<root>"
                findings.append((rel, f"frontmatter/schema: {loc}: {e.message}"))

    # The configuration file itself.
    specs_toml = ROOT / "specs.toml"
    if not specs_toml.exists():
        findings.append(("specs.toml", "config/missing: no specs.toml at the repository root"))
    elif config_path.exists():
        with specs_toml.open("rb") as fh:
            cfg = tomllib.load(fh)
        cfg_schema = json.loads(config_path.read_text())
        for e in Draft7Validator(cfg_schema).iter_errors(cfg):
            loc = "/".join(map(str, e.absolute_path)) or "<root>"
            findings.append(("specs.toml", f"config/schema: {loc}: {e.message}"))

    if not quiet:
        for rel, msg in findings:
            print(f"error  {rel}: {msg}")
        if excluded:
            print()
            for rel, reason in excluded:
                print(f"excluded  {rel}: {reason}")
        print()
        print(f"checked {checked} documents, {len(findings)} findings, {len(excluded)} excluded")

    return 1 if findings else 0


def err(msg: str) -> int:
    print(f"error: {msg}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
