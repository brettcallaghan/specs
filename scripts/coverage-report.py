#!/usr/bin/env python3
"""Merge Go coverage profiles and render the result into the README.

Reads one text-format coverage profile per test layer, merges them into a
single profile, and rewrites the generated regions of a Markdown file
with the resulting statistics. Standard library only.

The layers are merged rather than reported side by side alone: a
statement executed by any layer is covered, and the union is the only
honest total. Reporting a per-layer column as well is what makes the
table diagnostic — a package covered only by the end-to-end layer is
reached through the process boundary and nowhere else.

Usage:
  coverage-report.py --layer Unit=unit.cov --layer E2E=e2e.cov \\
      --module github.com/example/repo/cli \\
      --merged-out build/coverage/coverage.out \\
      --readme README.md [--check] [--min 70]

Exit status:
  0  the README was written, or (with --check) is already current
  1  --check found the README stale, or --min was not met
  2  bad invocation, or a profile could not be parsed
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

# The markers delimiting each generated region. Everything between a
# start and its end marker is replaced; the markers themselves and all
# surrounding prose are left alone.
REGIONS = {
    "badge": ("<!-- coverage:badge:start -->", "<!-- coverage:badge:end -->"),
    "table": ("<!-- coverage:table:start -->", "<!-- coverage:table:end -->"),
}

# A profile line: "<file>:<startLine>.<col>,<endLine>.<col> <numStmt> <count>".
PROFILE_LINE = re.compile(r"^(.+:\d+\.\d+,\d+\.\d+) (\d+) (\d+)$")


class ProfileError(Exception):
    """A coverage profile could not be parsed."""


def parse_profile(path: Path) -> dict[str, tuple[int, bool]]:
    """Parse a text coverage profile into {block: (numStmt, covered)}.

    A profile written by `go test -coverprofile` over several packages
    is a concatenation, so the same block can appear more than once with
    different counts. Keying by block and OR-ing the counts is the merge
    the cover tool itself performs, and doing it here is what keeps the
    statement totals below from counting a block once per test binary.
    """
    blocks: dict[str, tuple[int, bool]] = {}
    with path.open(encoding="utf-8") as handle:
        for lineno, line in enumerate(handle, start=1):
            line = line.strip()
            if not line or line.startswith("mode:"):
                continue
            match = PROFILE_LINE.match(line)
            if not match:
                raise ProfileError(f"{path}:{lineno}: unparsable profile line: {line}")
            block, num_stmt, count = match.group(1), int(match.group(2)), int(match.group(3))
            covered = count > 0
            if block in blocks:
                previous_stmt, previous_covered = blocks[block]
                if previous_stmt != num_stmt:
                    raise ProfileError(
                        f"{path}:{lineno}: block {block} declares {num_stmt} statements "
                        f"but was already seen with {previous_stmt}"
                    )
                covered = covered or previous_covered
            blocks[block] = (num_stmt, covered)
    if not blocks:
        raise ProfileError(f"{path}: no coverage blocks (was the test run cached?)")
    return blocks


def import_path_of(block: str) -> str:
    """Return the import path of the file a block belongs to.

    A block key is "<import path>/<file>.go:<startLine>.<col>,...".
    """
    return block.rsplit(":", 1)[0]


def package_of(block: str, module: str) -> str:
    """Return the package a block belongs to, named as a repository path."""
    directory = import_path_of(block).rsplit("/", 1)[0]
    prefix = module.rsplit("/", 1)[0] + "/"
    if directory.startswith(prefix):
        return directory[len(prefix) :]
    return directory


# Go's convention for machine-written source, as recognized by the
# toolchain and by every linter that skips it: the first line of the
# file, verbatim apart from the tool's name.
GENERATED_MARKER = re.compile(r"^// Code generated .* DO NOT EDIT\.$")


def source_file_of(import_path: str, module: str, source_root: Path) -> Path:
    """Map a file's import path to the file it was compiled from."""
    if not import_path.startswith(module + "/"):
        raise ProfileError(
            f"{import_path} is not inside the module {module}; --source-root cannot locate it"
        )
    return source_root / import_path[len(module) + 1 :]


def generated_files(blocks: dict[str, tuple[int, bool]], module: str, source_root: Path) -> set[str]:
    """Return the import paths whose source carries the generated marker.

    A missing file is an error rather than a default, because the
    fallback — treating what cannot be read as hand-written — would
    silently move statements back into the denominator and change the
    published number without saying so.
    """
    generated: set[str] = set()
    for path in {import_path_of(block) for block in blocks}:
        source = source_file_of(path, module, source_root)
        try:
            with source.open(encoding="utf-8") as handle:
                first = handle.readline().rstrip("\n")
        except OSError as err:
            raise ProfileError(f"reading {source} to classify {path}: {err}") from err
        if GENERATED_MARKER.match(first):
            generated.add(path)
    return generated


def percent(covered: int, total: int) -> float:
    return 100.0 * covered / total if total else 0.0


def format_percent(covered: int, total: int) -> str:
    return f"{percent(covered, total):.1f}%" if total else "—"


def badge_color(pct: float) -> str:
    for threshold, color in ((90, "brightgreen"), (80, "green"), (70, "yellowgreen"), (60, "yellow"), (50, "orange")):
        if pct >= threshold:
            return color
    return "red"


def render_badge(covered: int, total: int) -> str:
    pct = percent(covered, total)
    # A static shields.io badge rather than a coverage service: the value
    # is committed alongside the numbers it summarizes, so the badge and
    # the table can never disagree, and no third party has to be trusted
    # or paid to keep serving it.
    return (
        f"[![coverage](https://img.shields.io/badge/"
        f"cli%20coverage-{pct:.1f}%25-{badge_color(pct)})](#test-coverage)"
    )


def union(layers: dict[str, dict[str, tuple[int, bool]]]) -> dict[str, tuple[int, bool]]:
    """Combine every layer: a statement is covered if any layer reached it."""
    combined: dict[str, tuple[int, bool]] = {}
    for blocks in layers.values():
        for block, (num_stmt, covered) in blocks.items():
            previous = combined.get(block)
            combined[block] = (num_stmt, covered or (previous[1] if previous else False))
    return combined


def totals(blocks: dict[str, tuple[int, bool]]) -> tuple[int, int]:
    """Return (statements, covered statements)."""
    return (
        sum(num_stmt for num_stmt, _ in blocks.values()),
        sum(num_stmt for num_stmt, covered in blocks.values() if covered),
    )


def render_table(
    layers: dict[str, dict[str, tuple[int, bool]]],
    module: str,
    footnote: str = "",
) -> str:
    """Render the per-layer and per-package tables."""
    combined = union(layers)
    total_stmt, total_covered = totals(combined)

    lines: list[str] = []
    lines.append("| Test layer | Statements covered | Coverage |")
    lines.append("|:---|---:|---:|")
    for name, blocks in layers.items():
        layer_covered = sum(num_stmt for num_stmt, covered in blocks.values() if covered)
        lines.append(
            f"| {name} | {layer_covered:,} / {total_stmt:,} "
            f"| {format_percent(layer_covered, total_stmt)} |"
        )
    lines.append(
        f"| **All layers combined** | **{total_covered:,} / {total_stmt:,}** "
        f"| **{format_percent(total_covered, total_stmt)}** |"
    )
    lines.append("")

    packages: dict[str, list[int]] = {}
    for block, (num_stmt, covered) in combined.items():
        entry = packages.setdefault(package_of(block, module), [0, 0])
        entry[0] += num_stmt
        if covered:
            entry[1] += num_stmt

    lines.append("| Package | Statements | Covered | Coverage |")
    lines.append("|:---|---:|---:|---:|")
    for name in sorted(packages):
        stmt, covered = packages[name]
        lines.append(f"| `{name}` | {stmt:,} | {covered:,} | {format_percent(covered, stmt)} |")
    lines.append(
        f"| **Total** | **{total_stmt:,}** | **{total_covered:,}** "
        f"| **{format_percent(total_covered, total_stmt)}** |"
    )
    if footnote:
        lines.append("")
        lines.append(footnote)
    return "\n".join(lines)


def replace_region(text: str, region: str, body: str) -> str:
    start, end = REGIONS[region]
    start_at = text.find(start)
    end_at = text.find(end)
    if start_at < 0 or end_at < 0:
        raise SystemExit(f"the README has no {start} … {end} region to write into")
    if end_at < start_at:
        raise SystemExit(f"the README's {end} appears before {start}")
    return text[: start_at + len(start)] + "\n" + body + "\n" + text[end_at:]


def write_merged(path: Path, blocks: dict[str, tuple[int, bool]]) -> None:
    """Write a merged profile that `go tool cover` can read.

    Counts are clamped to 0 or 1 because the profile is written in `set`
    mode, where any non-zero count means the same thing.
    """
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as handle:
        handle.write("mode: set\n")
        for block in sorted(blocks):
            num_stmt, covered = blocks[block]
            handle.write(f"{block} {num_stmt} {1 if covered else 0}\n")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument(
        "--layer",
        action="append",
        required=True,
        metavar="NAME=PROFILE",
        help="a test layer's display name and its coverage profile; repeatable, order preserved",
    )
    parser.add_argument("--module", required=True, help="the Go module path the profiles were produced for")
    parser.add_argument(
        "--source-root",
        type=Path,
        help="the module's directory on disk; required by --exclude-generated",
    )
    parser.add_argument(
        "--exclude-generated",
        action="store_true",
        help="leave machine-written files out of the statistics, and say so in the output",
    )
    parser.add_argument("--readme", type=Path, help="the Markdown file whose generated regions are rewritten")
    parser.add_argument("--merged-out", type=Path, help="where to write the merged profile")
    parser.add_argument(
        "--markdown-out",
        type=Path,
        help="where to write the rendered tables on their own, for a CI job summary",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="do not write the README; exit non-zero if it is not already current",
    )
    parser.add_argument("--min", type=float, default=None, metavar="PCT", help="fail if total coverage is below PCT")
    args = parser.parse_args()

    layers: dict[str, dict[str, tuple[int, bool]]] = {}
    for spec in args.layer:
        if "=" not in spec:
            parser.error(f"--layer expects NAME=PROFILE, got {spec!r}")
        name, _, profile = spec.partition("=")
        try:
            layers[name] = parse_profile(Path(profile))
        except (OSError, ProfileError) as err:
            print(f"error: {err}", file=sys.stderr)
            return 2

    # The merged profile is written before any exclusion, so that
    # `go tool cover -html` annotates every file the tests compiled —
    # excluded or not, the source is still worth reading.
    if args.merged_out:
        write_merged(args.merged_out, union(layers))

    footnote = ""
    if args.exclude_generated:
        if not args.source_root:
            parser.error("--exclude-generated requires --source-root")
        try:
            generated = generated_files(union(layers), args.module, args.source_root)
        except ProfileError as err:
            print(f"error: {err}", file=sys.stderr)
            return 2
        if generated:
            excluded_stmt = sum(
                num_stmt
                for block, (num_stmt, _) in union(layers).items()
                if import_path_of(block) in generated
            )
            layers = {
                name: {
                    block: value
                    for block, value in blocks.items()
                    if import_path_of(block) not in generated
                }
                for name, blocks in layers.items()
            }
            files = ", ".join(f"`{path.split('/')[-1]}`" for path in sorted(generated))
            footnote = (
                f"Generated code is excluded from these figures: {excluded_stmt:,} "
                f"statements in {len(generated)} "
                f"{'file' if len(generated) == 1 else 'files'} ({files}) carrying the "
                "`// Code generated … DO NOT EDIT.` marker. It is still instrumented, "
                "and appears in the HTML report `scripts/coverage.sh` writes."
            )

    combined = union(layers)
    total_stmt, total_covered = totals(combined)

    table = render_table(layers, args.module, footnote)
    badge = render_badge(total_covered, total_stmt)
    print(table)
    print()

    if args.markdown_out:
        args.markdown_out.parent.mkdir(parents=True, exist_ok=True)
        args.markdown_out.write_text(f"{badge}\n\n{table}\n", encoding="utf-8")

    status = 0
    if args.readme:
        original = args.readme.read_text(encoding="utf-8")
        updated = replace_region(original, "badge", badge)
        updated = replace_region(updated, "table", table)
        if args.check:
            if updated != original:
                print(
                    f"error: {args.readme} does not match the measured coverage.\n"
                    "       regenerate it with scripts/coverage.sh and commit the result.",
                    file=sys.stderr,
                )
                status = 1
        elif updated != original:
            args.readme.write_text(updated, encoding="utf-8")
            print(f"updated {args.readme}")
        else:
            print(f"{args.readme} is already current")

    if args.min is not None and percent(total_covered, total_stmt) < args.min:
        print(
            f"error: total coverage {format_percent(total_covered, total_stmt)} "
            f"is below the required {args.min:.1f}%",
            file=sys.stderr,
        )
        status = 1
    return status


if __name__ == "__main__":
    sys.exit(main())
