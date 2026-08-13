#!/usr/bin/env bash
#
# Regenerate every artifact derived from schemas/*.strata.
#
# The .strata files are the single source of truth. Everything this
# script writes is generated and must not be hand-edited — each output
# carries a generated-by banner saying so.
#
# Usage:
#   scripts/generate-schemas.sh          # regenerate in place
#   scripts/generate-schemas.sh --check  # fail if any output is stale
#
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

SRC="schemas"
SCHEMA_OUT="cli/core/schema"
GO_OUT="cli/core/specmodel"
MD_OUT="plugin/skills/spec/references/generated"

if ! command -v strata >/dev/null 2>&1; then
  echo "error: strata not found on PATH." >&2
  echo "  The session-start hook installs it; see .claude/skills/strata/SKILL.md" >&2
  echo "  to build it by hand from a strata-lang/strata checkout." >&2
  exit 1
fi

CHECK=0
[[ "${1:-}" == "--check" ]] && CHECK=1

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

# 1. Validate. --target ast runs the parser and every semantic check but
#    emits only JSON, so it is the cheapest correctness gate. Nothing is
#    written to the tree until this passes.
strata build "$SRC" --target ast --output "$WORK/ast" --error-format json

# 2. Bundled JSON Schema covering every type. The profile registers each
#    document type against a JSON Pointer into this bundle.
strata build "$SRC" --target json-schema --output "$WORK/schema"
mv "$WORK/schema/SpecModel.schema.json" "$WORK/schema/spec-model.schema.json"

# 2.5. Sanity assertion: every document-type struct must retain a non-empty
#      `required` list. strata commit 44c1955 (predating strata-lang/strata's
#      fix in 8997153, "Fix: allowAdditionalProperties: false silently
#      dropped struct inheritance (#289)") emitted structs like
#      `Adr: DocumentFrontmatter {}` as
#      {"additionalProperties": false, "properties": {}} — every inherited
#      envelope property AND the required list silently gone — while strata
#      itself still exited 0. `strata --version` cannot tell a good build
#      from a bad one (both report strata-cli 0.7.0), so this checks the
#      emitted SHAPE instead of the version string, which catches any future
#      regression of this kind rather than only this one commit
#      (agent-plugins-j7f).
#
#      The DocumentType enum (kebab-case values, e.g. "system-blueprint") is
#      used to derive the document-type struct names (PascalCase, e.g.
#      "SystemBlueprint") rather than hardcoding them, so the check stays
#      correct as document types are added or renamed in schemas/*.strata.
#      The enum itself is a plain string enum with no struct inheritance, so
#      it is not vulnerable to the bug this check guards against and remains
#      a reliable anchor even when the structs it names have been corrupted.
python3 - "$WORK/schema/spec-model.schema.json" <<'PYEOF'
import json
import sys

path = sys.argv[1]
with open(path) as f:
    schema = json.load(f)

defs = schema.get("definitions", {})

doc_type_enum = defs.get("DocumentType", {}).get("enum")
if not doc_type_enum:
    print(
        "error: generated schema has no definitions.DocumentType enum to "
        "derive the document-type list from",
        file=sys.stderr,
    )
    sys.exit(1)

missing = []
broken = []
for value in doc_type_enum:
    name = "".join(word.capitalize() for word in value.split("-"))
    definition = defs.get(name)
    if definition is None:
        missing.append(name)
        continue
    if not definition.get("required"):
        broken.append(name)

if missing:
    print(
        "error: generated schema is missing definitions for document "
        f"type(s): {', '.join(missing)}",
        file=sys.stderr,
    )
    sys.exit(1)

if broken:
    print(
        "error: generated schema has document type(s) with an empty or "
        f"missing 'required' list: {', '.join(broken)}\n"
        "  This is the shape of the regression fixed upstream by "
        "strata-lang/strata commit 8997153 ('Fix: allowAdditionalProperties: "
        "false silently dropped struct inheritance (#289)') — struct "
        "inheritance (e.g. 'struct Adr: DocumentFrontmatter {}') silently "
        "drops every inherited property and the required list, while strata "
        "still exits 0 and 'strata --version' still reports strata-cli "
        "0.7.0 either way.\n"
        "  Check the strata build on PATH: cat "
        "~/.cache/strata-cli/built-commit names the commit it was built "
        "from. See .claude/skills/strata/SKILL.md to rebuild if needed.",
        file=sys.stderr,
    )
    sys.exit(1)
PYEOF

# 3. The standalone specs.toml schema, filtered to SpecsConfig and its
#    transitive dependencies. This is the artifact contract-specs-
#    configuration.md names as normative for the shape of specs.toml.
strata build "$SRC" --target json-schema \
  --config "$SRC/overlays/specs-config.json" \
  --output "$WORK/cfg"
mv "$WORK/cfg/SpecModel.schema.json" "$WORK/schema/specs-config.schema.json"

# 3.5. A small Go wrapper embedding both bundles as a package, so a
#      consumer that needs a schema at runtime (the spec-driven-dev
#      profile) reads no file. cli/core/schema has no .strata source of
#      its own — the file below is generated by this script rather than
#      by strata — but it is still generated content and is written here
#      so it stays in cli/core/schema's synced output rather than a
#      hand-maintained file a regeneration could silently leave stale.
#      go:embed cannot reach outside its own package directory, which is
#      why the embedding lives beside the JSON rather than in the
#      package that consumes it.
cat >"$WORK/schema/embed.go" <<'GOFILE'
// Code generated by scripts/generate-schemas.sh. DO NOT EDIT.

// Package schema embeds the JSON Schema bundles this package holds, so a
// consumer needs no file access at runtime to read them.
package schema

import _ "embed"

// SpecModel is the union schema covering every document type this
// framework defines: spec-model.schema.json.
//
//go:embed spec-model.schema.json
var SpecModel []byte

// SpecsConfig is the standalone schema for specs.toml, filtered to
// SpecsConfig and its transitive dependencies: specs-config.schema.json.
//
//go:embed specs-config.schema.json
var SpecsConfig []byte
GOFILE

# 4. Go structs for the CLI and its modules.
strata build "$SRC" --target go --output "$WORK/go"

# 5. Markdown reference shipped with the plugin, describing the shape of
#    frontmatter to anyone authoring a document by hand.
strata build "$SRC" --target markdown --output "$WORK/md"

cat >"$WORK/md/README.md" <<'BANNER'
# Generated schema reference

**Generated — do not edit.** Every file in this directory is produced by
`scripts/generate-schemas.sh` from the `.strata` sources in `schemas/`.
Edit those and re-run the script; a change made here is lost on the next
regeneration.

These pages describe the shape of the frontmatter and configuration a
document author writes. The normative artifacts derived from the same
sources are the JSON Schemas under `cli/core/schema/`.
BANNER

sync_dir() {
  local src="$1" dest="$2"
  if [[ $CHECK -eq 1 ]]; then
    if ! diff -r -q "$src" "$dest" >/dev/null 2>&1; then
      echo "error: $dest is stale — run scripts/generate-schemas.sh" >&2
      diff -r "$src" "$dest" >&2 || true
      return 1
    fi
  else
    rm -rf "$dest"
    mkdir -p "$(dirname "$dest")"
    cp -r "$src" "$dest"
  fi
}

rc=0
sync_dir "$WORK/schema" "$SCHEMA_OUT" || rc=1
sync_dir "$WORK/go" "$GO_OUT" || rc=1
sync_dir "$WORK/md" "$MD_OUT" || rc=1

if [[ $rc -ne 0 ]]; then
  exit 1
fi

if [[ $CHECK -eq 1 ]]; then
  echo "generated artifacts are up to date"
else
  echo "regenerated:"
  echo "  $SCHEMA_OUT  (JSON Schema)"
  echo "  $GO_OUT  (Go)"
  echo "  $MD_OUT  (Markdown)"
fi
