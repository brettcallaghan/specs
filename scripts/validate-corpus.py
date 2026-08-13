#!/usr/bin/env python3
"""Validate documents OUTSIDE the Spec CLI's corpus root against the schemas.

The Spec CLI now exists, and `spec check` runs over `specs/` — the root
this repository's specs.toml declares — in the `corpus` job of
.github/workflows/checks.yml. That is the real gate, with spans, exit
codes, and every rule the active profile registers.

This script covers what that root deliberately excludes: the RFC queue
and the reference files bundled inside each skill. Those carry the same
frontmatter and drift the same way, but they are not corpus members —
the `spec` skill's own rule is that everything under a skill's
`references/` is cited by backticked path, never by `[[slug]]` — so
widening the CLI's root to reach them would make them corpus members in
every consumer's repository too, which is not what they are.

The two gates do not overlap. If this script's SEARCH ever grows a
pattern under the CLI's root, delete it here instead: a document checked
twice is a document whose two gates can disagree.

What it checks, and only this:

  * every markdown file whose frontmatter parses and declares a known
    `type` is validated against that type's definition in
    cli/core/schema/spec-model.schema.json
  * specs.toml is validated against cli/core/schema/specs-config.schema.json

It does NOT check reference integrity, required sections, or line width;
those are the CLI's, and are specified but not yet built.

Files with no frontmatter, and files declaring a type outside the
profile's vocabulary, are reported as excluded rather than dropped
silently — the same treatment AC-SPEC-001.3 and AC-SPEC-001.4 require of
the CLI, so that a document which should be a specification but is
missing frontmatter is visible rather than absent.

Files opening with a GENERATED COPY header are not validated here. They
are DELEGATED, not skipped: this script runs ORPHAN_AUTHORITY below and
fails on its exit status, so the header cannot confer "somebody else
checks this" without somebody else actually having checked it.

Requires: python3 -m pip install jsonschema pyyaml

Usage:
  scripts/validate-corpus.py            # report findings, exit 1 on any
  scripts/validate-corpus.py --quiet    # exit status only
"""
from __future__ import annotations

import json
import pathlib
import subprocess
import sys
import tomllib

try:
    import yaml
    from jsonschema import Draft7Validator
except ImportError:  # pragma: no cover
    sys.exit("error: python3 -m pip install jsonschema pyyaml")

ROOT = pathlib.Path(__file__).resolve().parent.parent

# Where documents are looked for: everything outside the CLI's corpus
# root that carries the same frontmatter. See the module docstring.
#
# This list used to lead with "docs/**/*.md". The specs/ reorganization
# moved every specification document out of docs/ and nobody updated the
# glob, so from that commit until the CLI gate landed this script matched
# exactly one file under docs/ (session-bootstrap.md, no frontmatter) and
# reported a healthy "checked 26 documents" drawn entirely from the other
# two patterns. MIN_CHECKED below is the guard against a repeat.
# Not listed: plugin/skills/*/references/shared/*.md, the vendored copies
# sync-shared-references.sh writes. Their authored sources are checked by
# the second pattern, and that script's --check compares the copies byte
# for byte against those sources, which is stronger than re-validating a
# copy. A vendored copy that lands inside SEARCH anyway is handled by the
# delegation below, not dropped.
SEARCH = [
    "rfcs/**/*.md",
    "plugin/skills/*/references/*.md",
]

# A file opening with this comment declares itself a generated copy of a
# source elsewhere in the repo. Its frontmatter sits below the header, so
# read_frontmatter would call it "no-frontmatter" — a wrong reason for a
# file that has frontmatter three lines down.
GENERATED_COPY_HEADER = "<!-- GENERATED COPY"

# The single authority on whether a GENERATED COPY header is honest: it
# diffs every managed copy against its source, and fails on any tracked
# *.md carrying the header that names no generator this repo can run.
#
# This script does not restate that logic — it runs it. Without that, the
# header would confer "someone else checks this" in two places while only
# one place checked, and a file could claim to be generated, be passed
# over here, and have nothing regenerate it.
#
# Consequence to expect: the authority answers one question for the whole
# repository, so when it fails this script reports its entire output,
# including staleness that is not about any file matched by SEARCH. That
# is the price of delegating instead of duplicating; filtering its output
# would mean reimplementing its judgement here.
ORPHAN_AUTHORITY = "scripts/sync-shared-references.sh"
ORPHAN_AUTHORITY_ARGS = ["--check"]

# The corpus root the CLI owns, read from specs.toml rather than hardcoded.
# A SEARCH pattern reaching inside it is a configuration error, not a
# broader check: see the module docstring.
CLI_ROOT_KEY = ("workspace", "root")

# A floor on how many documents this script must find. The failure it
# guards against is not "a document is wrong" but "this script is
# pointed at nothing and says so cheerfully" — the exact way the stale
# docs/ glob survived review. The value is deliberately well below the
# current count (26 at the time of writing): it is a smoke alarm for a
# glob that stopped matching, not an inventory to keep in sync.
MIN_CHECKED = 20

# Templates carry deliberate placeholder slugs and type values, and are
# not corpus members. AC-SPEC-004.4 names this explicitly. Templates are
# identified by the -template.md suffix, not by substring "template": see
# check-template-drift.py, which uses the same rule.
EXCLUDE_TEMPLATE_SUFFIX = "-template.md"

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


def opens_with_generated_copy_header(path: pathlib.Path) -> bool:
    """True if the file's first line declares it a generated copy."""
    with path.open(encoding="utf-8") as fh:
        return fh.readline().startswith(GENERATED_COPY_HEADER)


def delegated_orphan_findings(delegated: list[str]) -> list[tuple[str, str]]:
    """Run ORPHAN_AUTHORITY over the files this script hands to it.

    Returns findings, so a header that names no runnable generator fails
    this gate too rather than only the sibling one. A missing or
    non-executable authority is itself a finding: the delegation is the
    only thing standing behind the files skipped above, so it may not
    fail open.
    """
    authority = ROOT / ORPHAN_AUTHORITY
    quoted = ", ".join(delegated)
    if not authority.exists():
        return [(ORPHAN_AUTHORITY, (
            f"generated-copy/no-authority: orphan detection for {len(delegated)} "
            f"generated-copy file(s) ({quoted}) is delegated to "
            f"`{ORPHAN_AUTHORITY}`, which is not present. Restore it, or check "
            "those files here."
        ))]

    proc = subprocess.run(
        ["bash", str(authority), *ORPHAN_AUTHORITY_ARGS],
        cwd=ROOT, capture_output=True, text=True,
    )
    if proc.returncode == 0:
        return []

    output = (proc.stdout + proc.stderr).strip() or f"exit status {proc.returncode}"
    detail = "\n".join(f"    | {line}" for line in output.splitlines())
    return [(ORPHAN_AUTHORITY, (
        f"generated-copy/delegated: {len(delegated)} file(s) in this script's search "
        f"scope ({quoted}) open with a GENERATED COPY header, so they are checked by "
        f"`{ORPHAN_AUTHORITY} {' '.join(ORPHAN_AUTHORITY_ARGS)}` instead of here. "
        f"That check failed (exit {proc.returncode}):\n{detail}"
    ))]


def cli_corpus_root() -> str | None:
    """Return [workspace].root from specs.toml, or None if unreadable.

    None rather than an exception: a repository with no specs.toml has no
    CLI gate to overlap with, so there is nothing for the caller to
    enforce and no reason to fail here.
    """
    specs_toml = ROOT / "specs.toml"
    if not specs_toml.exists():
        return None
    with specs_toml.open("rb") as fh:
        cfg = tomllib.load(fh)
    for key in CLI_ROOT_KEY:
        cfg = cfg.get(key) if isinstance(cfg, dict) else None
        if cfg is None:
            return None
    return str(cfg).strip("/")


def main() -> int:
    quiet = "--quiet" in sys.argv

    bundle_path = ROOT / "cli/core/schema/spec-model.schema.json"
    config_path = ROOT / "cli/core/schema/specs-config.schema.json"
    if not bundle_path.exists():
        return err("generated schemas are missing — run scripts/generate-schemas.sh")

    # No SEARCH pattern may reach into the root `spec check` owns.
    cli_root = cli_corpus_root()
    if cli_root:
        overlapping = [p for p in SEARCH if p.split("/", 1)[0] == cli_root]
        if overlapping:
            return err(
                f"SEARCH patterns {overlapping} reach inside the Spec CLI's corpus "
                f"root ({cli_root}/). Those documents are checked by 'spec check' "
                "in the checks.yml corpus job; remove them here rather than "
                "gating them twice."
            )

    bundle = json.loads(bundle_path.read_text())

    findings: list[tuple[str, str]] = []
    excluded: list[tuple[str, str]] = []
    delegated: list[str] = []
    checked = 0

    seen: set[pathlib.Path] = set()
    for pattern in SEARCH:
        for path in sorted(ROOT.glob(pattern)):
            if path in seen:
                continue
            seen.add(path)
            if path.name.endswith(EXCLUDE_TEMPLATE_SUFFIX):
                continue
            rel = str(path.relative_to(ROOT))

            # Delegated, not skipped. Validating a copy's frontmatter here
            # would gate the same bytes twice; whether the copy should
            # exist at all is ORPHAN_AUTHORITY's question, and it is asked
            # below on this list rather than assumed.
            if opens_with_generated_copy_header(path):
                delegated.append(rel)
                continue

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

    if delegated:
        findings.extend(delegated_orphan_findings(delegated))

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

    # A glob that stopped matching reports zero findings, which reads as
    # success. Refuse to be that gate.
    if checked < MIN_CHECKED:
        findings.append((
            "scripts/validate-corpus.py",
            f"search/underpopulated: matched {checked} documents against a floor of "
            f"{MIN_CHECKED}. Either SEARCH has a pattern that no longer matches "
            "anything, or documents moved. Fix the patterns — do not lower the floor "
            "to make this pass.",
        ))

    if not quiet:
        for rel, msg in findings:
            print(f"error  {rel}: {msg}")
        if excluded:
            print()
            for rel, reason in excluded:
                print(f"excluded  {rel}: {reason}")
        # Printed on success too. A reader of this script's output alone
        # must be able to see that these files were handed to a gate that
        # ran, not passed over.
        if delegated:
            print()
            for rel in delegated:
                print(f"delegated  {rel}: generated copy — checked by "
                      f"{ORPHAN_AUTHORITY} {' '.join(ORPHAN_AUTHORITY_ARGS)}")
        print()
        print(f"checked {checked} documents, {len(findings)} findings, "
              f"{len(excluded)} excluded, {len(delegated)} delegated")

    return 1 if findings else 0


def err(msg: str) -> int:
    print(f"error: {msg}", file=sys.stderr)
    return 1


if __name__ == "__main__":
    sys.exit(main())
