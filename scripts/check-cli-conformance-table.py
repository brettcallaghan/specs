#!/usr/bin/env python3
"""Check that cli-output-standards' conformance table names every command.

`specs/standards/cli-output-standards.md` records, under
"## Current conformance", one row per command saying what it prints and
what it has to change. Whether a command's output conforms is a judgement
the table records and this script does not make. What it can check is the
row set: every command the binary registers has a row, and every row
names a command the binary registers. A command that lands without a row,
or a row left behind by a removed command, fails here.

The command set comes from the built binary, never from a list kept in
this script: it builds `cli/` to a temporary directory and reads the
"Available Commands:" block of `specs --help`, discarding `help` and
`completion`.

A row's first cell names its commands. It is split on ",", backticks are
stripped, and everything from the first space is dropped, so
`refs <slug>[#<anchor>]` names `refs`, `c4 [--root <slug>]` names `c4`,
and `feature`, `criterion` names both.

Exit status: 1 if a registered command has no row, a row names no
registered command, or either set comes up empty. 0 otherwise.

Usage:
  scripts/check-cli-conformance-table.py
  scripts/check-cli-conformance-table.py --doc <path>   # check another copy
"""
from __future__ import annotations

import argparse
import pathlib
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parent.parent

DOC = ROOT / "specs/standards/cli-output-standards.md"
SECTION = "## Current conformance"
HEADER = "| command |"

# Commands cobra adds on its own; they print no answer, so no row.
IGNORED = {"help", "completion"}


def registered_commands() -> set[str]:
    """Build the binary and read the commands its --help lists."""
    with tempfile.TemporaryDirectory() as tmp:
        binary = pathlib.Path(tmp) / "specs"
        build = subprocess.run(
            ["go", "build", "-C", str(ROOT / "cli"), "-o", str(binary), "./specs"],
            capture_output=True,
            text=True,
        )
        if build.returncode != 0:
            sys.exit(f"error: building the specs binary failed:\n{build.stderr}")
        help_run = subprocess.run([str(binary), "--help"], capture_output=True, text=True)
        if help_run.returncode != 0:
            sys.exit(f"error: specs --help exited {help_run.returncode}:\n{help_run.stderr}")

    commands: set[str] = set()
    in_block = False
    for line in help_run.stdout.splitlines():
        if line.strip() == "Available Commands:":
            in_block = True
            continue
        if in_block:
            if not line.strip():
                break
            commands.add(line.split()[0])
    return commands - IGNORED


def table_commands(doc: pathlib.Path) -> dict[str, list[str]]:
    """Map each command a table row names to the first cells naming it."""
    lines = doc.read_text(encoding="utf-8").splitlines()
    try:
        start = lines.index(SECTION)
    except ValueError:
        sys.exit(f"error: {doc} has no {SECTION!r} section")

    named: dict[str, list[str]] = {}
    in_table = False
    for line in lines[start + 1 :]:
        if line.startswith("## "):
            break
        if not in_table:
            in_table = line.startswith(HEADER)
            continue
        if not line.startswith("|"):
            break
        cell = line.strip("|").split("|")[0].strip()
        if set(cell) <= set("-: "):
            continue  # the delimiter row
        for part in cell.split(","):
            name = part.strip().strip("`").split(" ")[0].strip("`")
            if name:
                named.setdefault(name, []).append(cell)
    return named


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--doc", type=pathlib.Path, default=DOC)
    args = parser.parse_args()

    commands = registered_commands()
    rows = table_commands(args.doc)

    failures = []
    if not commands:
        failures.append("specs --help listed no commands under 'Available Commands:'")
    if not rows:
        failures.append(f"{args.doc}: no rows under {SECTION!r}")
    for command in sorted(commands - rows.keys()):
        failures.append(f"{args.doc}: command {command!r} is registered but has no row under {SECTION!r}")
    for name in sorted(rows.keys() - commands):
        for cell in rows[name]:
            failures.append(f"{args.doc}: row {cell!r} names {name!r}, which is not a registered command")

    for failure in failures:
        print(failure)
    if failures:
        return 1
    print(f"RESULT: {len(commands)} registered commands, each with a row; no row names anything else.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
