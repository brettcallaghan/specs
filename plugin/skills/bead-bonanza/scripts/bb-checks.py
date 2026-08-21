#!/usr/bin/env python3
"""bb-checks.py — run the checks a project declares in sf.toml.

USAGE
  bb-checks.py                    run every declared check
  bb-checks.py <id> [<id>...]     run only these, in DECLARATION order
  bb-checks.py --ids              print declared ids, one per line
  bb-checks.py --report <path>    also write `<id>\t<exit>` per check there
  bb-checks.py --self-test

THE DECLARATION lives in `sf.toml` at the repository root — an array of
tables, two keys each, nothing else accepted:

    [[checks]]
    id = "test-go"
    command = 'cd cli && go test ./...'

`id` is `[a-z0-9][a-z0-9-]*` and unique. `command` is any non-empty string,
run through `sh -c` with the repository root as its working directory and the
environment inherited. Prefer TOML *literal* strings (single quotes): they do
no escape processing, so a shell one-liner survives verbatim.

WHAT BELONGS IN IT is the quality bar for a single commit, not the release
bar. Every bead runs these before its own close, so a check that takes minutes
— or that only means anything across the whole repository at release time —
belongs in CI instead.

TWO RULES THE RUNNER ENFORCES so no project has to remember them:
  * Serial, in declaration order, ALWAYS — including when ids are given as
    arguments. Declaration order is the only place a project can say "this
    must run before that", so argument order never overrides it.
  * Every requested check runs, even after one fails. Chaining with `&&` stops
    at the first failure and hides the rest, which is how one known-red check
    masks everything behind it.

EXIT CODES — the caller acts on these:
  0  every requested check passed
  1  at least one failed
  3  invocation or configuration error. Nothing ran: a bad declaration must
     not half-execute.

OUTPUT is the tools' own bytes, untouched. The runner adds `=== <id>` before a
check and `=== <id> FAILED (exit N)` after a failing one, both on stderr — the
verdict is the one thing the tools do not say themselves.

**Those delimiters are for a human reading the log; they are not a machine
interface.** They share a stream with output the runner does not control, so a
tool is free to print a line identical to one — measured: a check echoing
`=== other-check FAILED (exit 9)` makes an in-band reader report a passing
check as failed. A caller that needs verdicts passes `--report <path>` and gets
them out of band, one `<id>\t<exit-status>` line per check as it completes.
"""

import os
import re
import subprocess
import sys
import tempfile

try:
    import tomllib
except ModuleNotFoundError:  # pragma: no cover - depends on interpreter
    sys.stderr.write(
        "bb-checks: this interpreter has no `tomllib` (stdlib since Python "
        "3.11); run bb-checks.py under python3.11 or newer\n"
    )
    sys.exit(3)

RC_OK, RC_FAILED, RC_USAGE = 0, 1, 3

ID_RE = re.compile(r"^[a-z0-9][a-z0-9-]*$")
ALLOWED_KEYS = {"id", "command"}


def die(msg):
    sys.stderr.write(f"bb-checks: {msg}\n")
    sys.exit(RC_USAGE)


def repo_root():
    try:
        out = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            capture_output=True, text=True, check=True,
        )
        return out.stdout.strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return os.getcwd()


def load(root):
    """Return [(id, command)] in declaration order, or exit 3 saying why.

    The whole array is validated before anything runs. An unrecognised key is
    a hard error rather than an ignored one: a typo like `comand = ...` would
    silently drop a check, and a dropped check reads downstream exactly like a
    check that passed.
    """
    path = os.environ.get("BB_SF_TOML") or os.path.join(root, "sf.toml")
    if not os.path.isfile(path):
        die(f"no {path} — declare the project's checks as [[checks]] entries there")
    try:
        with open(path, "rb") as fh:
            doc = tomllib.load(fh)
    except tomllib.TOMLDecodeError as exc:
        die(f"{path} is not valid TOML: {exc}")

    if "checks" not in doc:
        die(f"{path} declares no [[checks]] — add one entry per check, "
            "each with an `id` and a `command`")
    raw = doc["checks"]
    if not isinstance(raw, list) or not all(isinstance(e, dict) for e in raw):
        die(f"{path}: `checks` must be an array of tables ([[checks]]), "
            f"got {type(raw).__name__}")
    if not raw:
        die(f"{path} declares an EMPTY [[checks]] array — a project with "
            "nothing to run cannot be gated; declare at least one check")

    seen, out = {}, []
    for i, entry in enumerate(raw):
        where = f"{path}: [[checks]] #{i + 1}"
        extra = set(entry) - ALLOWED_KEYS
        if extra:
            die(f"{where} has unrecognised key(s) {sorted(extra)} — only "
                "`id` and `command` are accepted, and an ignored typo would "
                "silently drop this check")
        for key in ("id", "command"):
            if key not in entry:
                die(f"{where} has no `{key}`")
            if not isinstance(entry[key], str):
                die(f"{where}: `{key}` must be a string")
        cid, cmd = entry["id"], entry["command"]
        if not ID_RE.match(cid):
            die(f"{where}: id {cid!r} must match [a-z0-9][a-z0-9-]*")
        if not cmd.strip():
            die(f"{where}: `command` is empty")
        if cid in seen:
            die(f"{where}: duplicate id {cid!r}, already declared as "
                f"#{seen[cid] + 1}")
        seen[cid] = i
        out.append((cid, cmd))
    return out


def select(declared, wanted):
    """Filter to `wanted`, keeping DECLARATION order rather than argument order."""
    known = {cid for cid, _ in declared}
    unknown = [w for w in wanted if w not in known]
    if unknown:
        die(f"unknown check id(s) {unknown} — declared: "
            f"{' '.join(cid for cid, _ in declared)}")
    return [(cid, cmd) for cid, cmd in declared if cid in set(wanted)]


def run(root, selected, report_path=None):
    rc = RC_OK
    report = None
    if report_path:
        try:
            report = open(report_path, "w")
        except OSError as exc:
            die(f"cannot write the report at {report_path}: {exc}")
    try:
        for cid, cmd in selected:
            sys.stderr.write(f"=== {cid}\n")
            sys.stderr.flush()
            sys.stdout.flush()
            # The child inherits our streams, so its output is passed through
            # byte for byte rather than captured and re-emitted.
            status = subprocess.run(cmd, shell=True, cwd=root).returncode
            if status != 0:
                sys.stderr.write(f"=== {cid} FAILED (exit {status})\n")
                sys.stderr.flush()
                rc = RC_FAILED
            if report:
                # Written as each check finishes, so a run that is killed
                # partway still says how far it got.
                report.write(f"{cid}\t{status}\n")
                report.flush()
    finally:
        if report:
            report.close()
    return rc


# --------------------------------------------------------------------- tests

FIXTURES = {
    "two-pass": """
[[checks]]
id = "alpha"
command = 'echo alpha-ran'
[[checks]]
id = "beta"
command = 'echo beta-ran'
""",
    "first-fails": """
[[checks]]
id = "alpha"
command = 'echo alpha-ran; exit 7'
[[checks]]
id = "beta"
command = 'echo beta-ran'
""",
    "duplicate-id": """
[[checks]]
id = "alpha"
command = 'echo one'
[[checks]]
id = "alpha"
command = 'echo two'
""",
    "unknown-key": """
[[checks]]
id = "alpha"
command = 'echo one'
comand = 'echo typo'
""",
    "missing-command": """
[[checks]]
id = "alpha"
""",
    "bad-id": """
[[checks]]
id = "Alpha Check"
command = 'echo one'
""",
    "malformed": """
[[checks]
id = "alpha"
""",
    "no-array": """
[models]
light = "haiku"
""",
    "empty-array": """
checks = []
""",
    "not-tables": """
checks = ["alpha", "beta"]
""",
    "spoofer": """
[[checks]]
id = "green"
command = 'echo green-output'
[[checks]]
id = "liar"
command = 'echo "=== green FAILED (exit 9)"'
""",
    "cd-then-next": """
[[checks]]
id = "wanders"
command = 'cd / && echo wandered'
[[checks]]
id = "checks-cwd"
command = 'test -f cwd-marker && echo started-at-root'
""",
}


def self_test():
    failures = []

    def check(name, cond, detail=""):
        if cond:
            print(f"  ok    {name}")
        else:
            print(f"  FAIL  {name}  {detail}")
            failures.append(name)

    with tempfile.TemporaryDirectory() as tmp:
        # A real git repo, so repo_root() resolves here and the fixture
        # sf.toml is what gets read.
        subprocess.run(["git", "init", "-q", tmp], check=True)
        for name, body in FIXTURES.items():
            with open(os.path.join(tmp, f"{name}.toml"), "w") as fh:
                fh.write(body.lstrip())
        # Only reachable from the repo root, so a check that finds it proves
        # the previous check's `cd` did not follow it.
        open(os.path.join(tmp, "cwd-marker"), "w").close()

        def invoke(fixture, *args):
            env = dict(os.environ)
            env["BB_SF_TOML"] = os.path.join(tmp, f"{fixture}.toml")
            return subprocess.run(
                [sys.executable, os.path.abspath(__file__), *args],
                cwd=tmp, env=env, capture_output=True, text=True,
            )

        r = invoke("two-pass")
        check("two passing checks exit 0", r.returncode == RC_OK, r.returncode)
        check("both ran", "alpha-ran" in r.stdout and "beta-ran" in r.stdout)
        check("delimiters on stderr", "=== alpha" in r.stderr and "=== beta" in r.stderr)

        r = invoke("first-fails")
        check("a failure exits 1", r.returncode == RC_FAILED, r.returncode)
        check("the check AFTER a failure still ran", "beta-ran" in r.stdout,
              "this is the rule the runner exists to enforce")
        check("failure is named with its status", "=== alpha FAILED (exit 7)" in r.stderr)

        for fixture, label in [
            ("duplicate-id", "duplicate id"),
            ("unknown-key", "unrecognised key"),
            ("missing-command", "missing command"),
            ("bad-id", "malformed id"),
            ("malformed", "malformed TOML"),
            ("no-array", "no [[checks]] array"),
            ("empty-array", "empty [[checks]] array"),
            ("not-tables", "checks is not an array of tables"),
        ]:
            r = invoke(fixture)
            check(f"{label} exits 3", r.returncode == RC_USAGE, r.returncode)
            check(f"{label} ran nothing", r.stdout.strip() == "",
                  f"stdout was {r.stdout!r}")

        r = invoke("no-array")
        r2 = invoke("empty-array")
        check("absent and empty arrays are worded differently",
              r.stderr.strip() != r2.stderr.strip())

        r = invoke("two-pass", "beta", "alpha")
        check("arguments run in DECLARATION order, not argument order",
              r.stdout.index("alpha-ran") < r.stdout.index("beta-ran"),
              r.stdout.replace("\n", " | "))

        r = invoke("two-pass", "alpha")
        check("a subset runs only what was asked",
              "alpha-ran" in r.stdout and "beta-ran" not in r.stdout)

        r = invoke("two-pass", "gamma")
        check("unknown id exits 3", r.returncode == RC_USAGE, r.returncode)
        check("unknown id ran nothing", r.stdout.strip() == "")

        r = invoke("two-pass", "--ids")
        check("--ids prints declared ids in order",
              r.stdout.split() == ["alpha", "beta"], r.stdout)
        r = invoke("duplicate-id", "--ids")
        check("--ids on an invalid declaration exits 3", r.returncode == RC_USAGE)

        r = invoke("cd-then-next")
        check("a check that cd's does not move the next one",
              "started-at-root" in r.stdout, r.stdout + r.stderr)

        rp = os.path.join(tmp, "report.tsv")
        r = invoke("first-fails", "--report", rp)
        with open(rp) as fh:
            rows = [ln.split("\t") for ln in fh.read().splitlines()]
        check("--report writes one row per check, in order",
              [row[0] for row in rows] == ["alpha", "beta"], rows)
        check("--report carries each check's real exit status",
              [row[1] for row in rows] == ["7", "0"], rows)

        rp2 = os.path.join(tmp, "spoof.tsv")
        r = invoke("spoofer", "--report", rp2)
        with open(rp2) as fh:
            spoofed = dict(ln.split("\t") for ln in fh.read().splitlines())
        check("a tool CAN forge the in-band delimiter",
              "=== green FAILED (exit 9)" in r.stdout,
              "if this stops being true the next assertion proves nothing")
        check("--report is not fooled by it", spoofed["green"] == "0", spoofed)

    print()
    if failures:
        print(f"RESULT: FAIL — {len(failures)} of the assertions above")
        return 1
    print("RESULT: PASS")
    return 0


def main(argv):
    if argv and argv[0] in ("-h", "--help"):
        print(__doc__)
        return RC_OK
    if argv and argv[0] == "--self-test":
        if len(argv) > 1:
            die("--self-test takes no other arguments")
        return self_test()

    report_path = None
    if "--report" in argv:
        i = argv.index("--report")
        if i + 1 >= len(argv):
            die("--report needs a path")
        report_path = argv[i + 1]
        argv = argv[:i] + argv[i + 2:]

    root = repo_root()
    declared = load(root)

    if argv and argv[0] == "--ids":
        if len(argv) > 1:
            die("--ids takes no other arguments")
        for cid, _ in declared:
            print(cid)
        return RC_OK

    for a in argv:
        if a.startswith("-"):
            die(f"unknown option {a!r} — see --help")

    selected = select(declared, argv) if argv else declared
    return run(root, selected, report_path)


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
