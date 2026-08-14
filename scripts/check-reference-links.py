#!/usr/bin/env python3
"""Check every markdown link in a shipped reference resolves to a co-located file.

A skill's reference directory travels to an adopting repository with no
corpus index behind it. A markdown link — `[text](target)` — that does
not resolve to a file next to (or under) the citing document is dead on
arrival there, and nothing else in this repository catches it: `spec
check` only walks the corpus, and a skill's reference directory is not
corpus.

What this script deliberately does NOT do:

  * Reimplement reference extraction. The Spec CLI's extractor
    (cli/core/references) owns wikilink parsing — including the
    code-span and fenced/indented-block skipping that stops
    illustrative examples from counting as citations. This script reads
    markdown link syntax, a different and simpler grammar, over the raw
    text with none of that exclusion. Do not merge the two parsers.
  * Check wikilinks (`[[slug]]` / `[[slug#anchor]]`). After the
    markdown-link rewrite lands, a shipped copy should contain none; a
    wikilink found in one is the rewrite's defect, not this script's —
    and the `](` this script matches on does not appear in wikilink
    syntax, so none are seen here regardless.
  * Slugify or verify anchors. Only the file half of a link (the part
    before a `#`) is resolved; whatever follows `#` is accepted as
    written.

Scope: every `*.md` under plugin/skills/*/references/**. A link whose
target starts with `http://` or `https://` is external and skipped. A
link with no file half (a bare `#anchor`, same-file) is skipped — there
is nothing to resolve.

Exit status: 1 if any link fails to resolve, or if the glob comes up
under its floor. 0 otherwise.

Usage:
  scripts/check-reference-links.py            # report findings, exit 1 on any
  scripts/check-reference-links.py --quiet    # exit status only
"""
from __future__ import annotations

import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent

SEARCH = "plugin/skills/*/references/**/*.md"

# A floor on how many reference documents must be found. The failure
# this guards against is not "a link is broken" but "this gate is
# pointed at nothing and says so cheerfully" — a smoke alarm for a glob
# that stopped matching, not an inventory to keep in sync. Measured at
# 52 files against this pattern on 2026-08-14, after the reference
# restructure that moved shipped standards and guides out to specs/ and
# left generated `references/shared/` copies in their place. Set well
# below that count so trimming a few files never trips it.
MIN_FILES = 40

LINK = re.compile(r"\[[^\]]*\]\(([^)]+)\)")


def is_external(target: str) -> bool:
    return target.startswith("http://") or target.startswith("https://")


def file_half(target: str) -> str:
    """The part of a link target before its anchor, if any."""
    return target.split("#", 1)[0].strip()


def main() -> int:
    quiet = "--quiet" in sys.argv

    files = sorted(ROOT.glob(SEARCH))
    checked = 0
    links_checked = 0
    errors: list[tuple[str, str]] = []

    for path in files:
        rel = str(path.relative_to(ROOT))
        checked += 1
        text = path.read_text(encoding="utf-8")

        for match in LINK.finditer(text):
            target = match.group(1).strip()
            if is_external(target):
                continue

            target_file = file_half(target)
            if not target_file:
                continue  # bare #anchor: same-file, nothing to resolve

            links_checked += 1
            resolved = (path.parent / target_file).resolve()
            if not resolved.is_file():
                try:
                    shown = str(resolved.relative_to(ROOT))
                except ValueError:
                    shown = str(resolved)
                errors.append((rel, (
                    f"link/unresolved: `{target}` does not resolve to a file "
                    f"co-located with this document (looked for {shown})"
                )))

    if checked < MIN_FILES:
        errors.append(("scripts/check-reference-links.py", (
            f"search/underpopulated: matched {checked} reference documents against "
            f"a floor of {MIN_FILES}. Either the references moved or SEARCH stopped "
            "matching. Fix the pattern — do not lower the floor to make this pass."
        )))

    if not quiet:
        for rel, message in errors:
            print(f"error  {rel}: {message}")
        if errors:
            print()
        print(f"checked {checked} reference documents, {links_checked} links, "
              f"{len(errors)} errors")

    return 1 if errors else 0


if __name__ == "__main__":
    sys.exit(main())
