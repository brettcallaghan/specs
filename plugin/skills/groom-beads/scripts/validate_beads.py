#!/usr/bin/env python3
"""
validate_beads.py — validate a groomed-bead JSONL file (the file-emit fallback
used when the `bd` CLI isn't available).

Two layers of checking:
  1. STRUCTURE — each record against references/bead.schema.json. This is what
     stops field-name improvisation (`deps` vs `dependencies`, `slug` vs `id`,
     a top-level `parent`): `additionalProperties: false` + required canonical
     names + edge-type enums. Uses the `jsonschema` package if installed; falls
     back to a built-in structural check that mirrors the same rules so the
     script still works offline.
  2. GRAPH & SEMANTICS — things a per-record schema can't express: dependency /
     parent IDs resolve, no cycles in the blocks graph, C4 component names are in
     the controlled vocabulary from your C4 docs, model is non-null on ready
     beads, ready features/bugs carry acceptance criteria, plus a session-phrase
     lint and a wave-collision advisory.

Exit status: 0 = clean (warnings allowed), 1 = hard errors, 2 = bad invocation.
With --strict, warnings are promoted to errors.

Usage:
  validate_beads.py groomed-backlog.jsonl \
      --c4-docs docs/c4/components.md [--schema references/bead.schema.json] [--strict]

The script never mutates beads; it only reports. Grooming should fix what it
surfaces and re-run — the same emit -> validate -> fix loop the real `bd` gives.
"""

import argparse
import json
import os
import re
import sys

CANON_TOP_KEYS = {"id", "title", "type", "priority", "status", "grooming_status",
                  "description", "design", "acceptance", "dependencies", "metadata"}
TYPES = {"epic", "feature", "task", "bug", "chore"}
STATUSES = {"open", "ready", "in_progress", "blocked", "closed"}
GROOMING = {"ready", "blocked-needs-info", "blocked-ungroomable"}
EDGE_TYPES = {"blocks", "blocked-by", "parent", "related", "discovered-from"}
MODELS = {"opus", "sonnet", "fable"}
SIZES = {"XS", "S", "M", "L", "XL", None}
BLOCKED = {"blocked-needs-info", "blocked-ungroomable"}

# Common alternative names agents drift to — call them out by name so the fix is obvious.
KNOWN_ALIASES = {"deps": "dependencies", "slug": "id", "parent": "a dependency of type 'parent'",
                 "accept": "acceptance", "acceptance_criteria": "acceptance",
                 "blocked_by": "a dependency {type:'blocked-by'}", "blockedBy": "dependencies"}

SESSION_PHRASES = ["as we discussed", "as discussed", "like we talked about", "like we said",
                   "the thing above", "the code above", "you know the one", "you know the",
                   "from earlier", "the above", "as above", "per our chat", "per our conversation"]
# lines that legitimately quote the original vague wording are exempt from the lint
QUOTE_MARKERS = ["original wording", "original text", "needs from requester", "ungroomable",
                 "flagged for the requester", "original note", "resolved from"]


class Report:
    def __init__(self):
        self.errors = []
        self.warnings = []
        self.infos = []
    def err(self, bead, msg): self.errors.append((bead, msg))
    def warn(self, bead, msg): self.warnings.append((bead, msg))
    def info(self, bead, msg): self.infos.append((bead, msg))


def load_beads(path):
    beads, bad = [], []
    with open(path, encoding="utf-8") as f:
        for i, line in enumerate(f, 1):
            line = line.strip()
            if not line:
                continue
            try:
                beads.append(json.loads(line))
            except json.JSONDecodeError as e:
                bad.append((i, str(e)))
    return beads, bad


def parse_c4_vocab(paths):
    """Extract component names from C4 docs: inline-code tokens shaped like a
    component id (…-service, …-gateway, …-spa, …-db, …-worker)."""
    vocab = set()
    pat = re.compile(r'`([a-z0-9]+(?:-[a-z0-9]+)*-(?:service|gateway|spa|db|worker|queue|api))`')
    for p in paths or []:
        if not os.path.exists(p):
            continue
        text = open(p, encoding="utf-8", errors="ignore").read()
        vocab.update(m.group(1) for m in pat.finditer(text))
    return vocab


def structural_check(bead, rep, schema_path):
    """Try jsonschema; fall back to a built-in check mirroring the schema."""
    bid = bead.get("id", "<no-id>")
    used_jsonschema = False
    if schema_path and os.path.exists(schema_path):
        try:
            import jsonschema  # type: ignore
            schema = json.load(open(schema_path, encoding="utf-8"))
            validator = jsonschema.Draft202012Validator(schema)
            for e in sorted(validator.iter_errors(bead), key=lambda e: e.path):
                loc = "/".join(str(x) for x in e.path) or "(root)"
                rep.err(bid, f"schema: {loc}: {e.message}")
            used_jsonschema = True
        except ImportError:
            pass
        except Exception as e:  # malformed schema shouldn't abort the run
            rep.warn(bid, f"schema load failed ({e}); using built-in checks")
    if used_jsonschema:
        # jsonschema covered shape; still name-check aliases for a friendly message
        _alias_hints(bead, rep, bid)
        return
    _builtin_structural(bead, rep, bid)


def _alias_hints(bead, rep, bid):
    for k in bead:
        if k in KNOWN_ALIASES:
            rep.err(bid, f"non-canonical field '{k}' — use {KNOWN_ALIASES[k]}")


def _builtin_structural(bead, rep, bid):
    # unknown / aliased top-level keys (mirrors additionalProperties:false)
    for k in bead:
        if k in KNOWN_ALIASES:
            rep.err(bid, f"non-canonical field '{k}' — use {KNOWN_ALIASES[k]}")
        elif k not in CANON_TOP_KEYS:
            rep.err(bid, f"unknown field '{k}' (schema forbids additional properties)")
    for req in ("id", "title", "type", "priority", "status", "description"):
        if not bead.get(req) and bead.get(req) != 0:
            rep.err(bid, f"missing required field '{req}'")
    if bead.get("type") not in TYPES:
        rep.err(bid, f"type '{bead.get('type')}' not in {sorted(TYPES)}")
    if bead.get("status") not in STATUSES:
        rep.err(bid, f"status '{bead.get('status')}' not in {sorted(STATUSES)}")
    p = bead.get("priority")
    if not (isinstance(p, int) and 0 <= p <= 4):
        rep.err(bid, f"priority must be integer 0..4, got {p!r}")
    gs = bead.get("grooming_status")
    if gs is not None and gs not in GROOMING:
        rep.err(bid, f"grooming_status '{gs}' not in {sorted(GROOMING)}")
    for dep in bead.get("dependencies", []) or []:
        if not isinstance(dep, dict) or set(dep) - {"type", "id"} or "type" not in dep or "id" not in dep:
            rep.err(bid, f"dependency must be {{type,id}}, got {dep!r}")
        elif dep["type"] not in EDGE_TYPES:
            rep.err(bid, f"edge type '{dep['type']}' not in {sorted(EDGE_TYPES)}")
    md = bead.get("metadata")
    if md is not None:
        if not isinstance(md, dict):
            rep.err(bid, "metadata must be an object")
        else:
            if "model" in md and md["model"] not in MODELS:
                rep.err(bid, f"metadata.model '{md['model']}' not in {sorted(MODELS)}")
            if md.get("est_size") not in SIZES:
                rep.err(bid, f"metadata.est_size '{md.get('est_size')}' not in {sorted(x for x in SIZES if x)}")


def semantic_checks(beads, rep, vocab, have_c4_docs):
    ids = {b.get("id") for b in beads if b.get("id")}
    by_id = {b.get("id"): b for b in beads if b.get("id")}

    for b in beads:
        bid = b.get("id", "<no-id>")
        typ = b.get("type")
        gs = b.get("grooming_status")
        blocked = gs in BLOCKED
        md = b.get("metadata") or {}

        # conditional completeness (schema can't express the grooming_status exemption cleanly)
        if not blocked and typ != "epic":
            if not b.get("design"):
                rep.err(bid, "ready non-epic bead is missing 'design'")
            if not md.get("model"):
                rep.err(bid, "ready bead has no metadata.model (must be non-null; opus|sonnet|fable)")
            if not md.get("model_reason"):
                rep.err(bid, "metadata.model set without a 'model_reason' (unauditable)")
        if not blocked and typ in ("feature", "bug") and not (b.get("acceptance") or []):
            rep.err(bid, f"ready {typ} has no acceptance criteria")
        if typ == "bug" and not blocked:
            blob = (b.get("description", "") + " " + " ".join(b.get("acceptance", []) or [])).lower()
            if not any(w in blob for w in ("reproduce", "repro", "steps", "expected", "actual")):
                rep.warn(bid, "bug has no visible repro / expected-vs-actual — a fresh agent can't verify the fix")
        if blocked and md.get("model") and md.get("model") in MODELS:
            rep.info(bid, f"{gs}: carries a placeholder model — fine, but must not be dispatched until specified")

        # dependency / parent referential integrity + self-reference
        parents = []
        for dep in b.get("dependencies", []) or []:
            if not isinstance(dep, dict):
                continue
            tid = dep.get("id")
            if tid == bid:
                rep.err(bid, f"self-referential dependency ({dep.get('type')})")
            elif tid not in ids:
                rep.err(bid, f"dependency {dep.get('type')} -> '{tid}' does not resolve to any bead")
            if dep.get("type") == "parent":
                parents.append(tid)
        for ptid in parents:
            par = by_id.get(ptid)
            if par is not None and par.get("type") != "epic":
                rep.warn(bid, f"parent '{ptid}' is a {par.get('type')}, not an epic")

        # C4 vocabulary
        if have_c4_docs:
            for field in ("c4_changes", "c4_impacts"):
                for comp in md.get(field, []) or []:
                    if comp not in vocab:
                        sugg = _closest(comp, vocab)
                        hint = f" (did you mean '{sugg}'?)" if sugg else " (greenfield, or a typo)"
                        rep.warn(bid, f"{field} '{comp}' not in C4 docs{hint}")

        # session-phrase lint (skip lines that explicitly quote/flag the original)
        for field in ("description", "design"):
            text = b.get(field) or ""
            for ln in text.splitlines():
                low = ln.lower()
                if any(q in low for q in QUOTE_MARKERS):
                    continue
                for ph in SESSION_PHRASES:
                    if ph in low:
                        rep.warn(bid, f"{field} may contain live session-dependent text: '{ph}' — inline the referent or mark it as a quote")
                        break

    # cycle detection over blocks graph:  X blocks Y  and  Y blocked-by X  both mean edge X->Y
    adj = {i: set() for i in ids}
    for b in beads:
        bid = b.get("id")
        for dep in b.get("dependencies", []) or []:
            if not isinstance(dep, dict):
                continue
            t, tid = dep.get("type"), dep.get("id")
            if tid not in ids:
                continue
            if t == "blocks":
                adj[bid].add(tid)
            elif t == "blocked-by":
                adj[tid].add(bid)
    cyc = _find_cycle(adj)
    if cyc:
        rep.err(cyc[0], "cycle in blocks graph: " + " -> ".join(cyc))

    # wave-collision advisory: ready beads sharing a c4_changes component with no
    # blocks edge between them cannot safely share a wave.
    ready = [b for b in beads if (b.get("grooming_status") or "ready") == "ready" and b.get("id")]
    reach = _transitive(adj)
    for i in range(len(ready)):
        for j in range(i + 1, len(ready)):
            a, c = ready[i], ready[j]
            ca = set((a.get("metadata") or {}).get("c4_changes", []) or [])
            cc = set((c.get("metadata") or {}).get("c4_changes", []) or [])
            shared = ca & cc
            if not shared:
                continue
            ai, ci = a["id"], c["id"]
            linked = ci in reach.get(ai, set()) or ai in reach.get(ci, set())
            if not linked:
                rep.info(ai, f"shares c4_changes {sorted(shared)} with '{ci}' and neither blocks the other — keep them in different waves")


def _closest(word, vocab):
    best, bestd = None, 99
    for v in vocab:
        d = _lev(word, v)
        if d < bestd:
            best, bestd = v, d
    return best if bestd <= 3 else None


def _lev(a, b):
    if a == b:
        return 0
    prev = list(range(len(b) + 1))
    for i, ca in enumerate(a, 1):
        cur = [i]
        for j, cb in enumerate(b, 1):
            cur.append(min(prev[j] + 1, cur[-1] + 1, prev[j - 1] + (ca != cb)))
        prev = cur
    return prev[-1]


def _find_cycle(adj):
    WHITE, GRAY, BLACK = 0, 1, 2
    color = {n: WHITE for n in adj}
    stack = []
    def dfs(n):
        color[n] = GRAY
        stack.append(n)
        for m in adj.get(n, ()):
            if color.get(m) == GRAY:
                return stack[stack.index(m):] + [m]
            if color.get(m) == WHITE:
                r = dfs(m)
                if r:
                    return r
        stack.pop()
        color[n] = BLACK
        return None
    for n in adj:
        if color[n] == WHITE:
            r = dfs(n)
            if r:
                return r
    return None


def _transitive(adj):
    reach = {}
    def walk(n, seen):
        for m in adj.get(n, ()):
            if m not in seen:
                seen.add(m)
                walk(m, seen)
        return seen
    for n in adj:
        reach[n] = walk(n, set())
    return reach


def main():
    ap = argparse.ArgumentParser(description="Validate a groomed-bead JSONL file.")
    ap.add_argument("jsonl", help="path to the emitted .jsonl file")
    ap.add_argument("--c4-docs", nargs="*", default=[], help="C4 doc file(s) for the component vocabulary")
    ap.add_argument("--schema", default=os.path.join(os.path.dirname(__file__), "..", "references", "bead.schema.json"))
    ap.add_argument("--strict", action="store_true", help="treat warnings as errors")
    args = ap.parse_args()

    if not os.path.exists(args.jsonl):
        print(f"validate_beads: file not found: {args.jsonl}", file=sys.stderr)
        return 2

    beads, bad = load_beads(args.jsonl)
    rep = Report()
    for lineno, msg in bad:
        rep.err(f"line {lineno}", f"invalid JSON: {msg}")

    vocab = parse_c4_vocab(args.c4_docs)
    have_c4 = bool(args.c4_docs)
    if have_c4 and not vocab:
        rep.warn("*", "no component names parsed from C4 docs — vocabulary check skipped")
        have_c4 = False

    for b in beads:
        structural_check(b, rep, os.path.abspath(args.schema))
    semantic_checks(beads, rep, vocab, have_c4)

    # duplicate ids
    seen = {}
    for b in beads:
        i = b.get("id")
        if i:
            seen[i] = seen.get(i, 0) + 1
    for i, n in seen.items():
        if n > 1:
            rep.err(i, f"duplicate id appears {n} times")

    print(f"=== validate_beads: {len(beads)} beads from {args.jsonl} ===")
    def show(items, tag):
        for bead, msg in items:
            print(f"  {tag} [{bead}] {msg}")
    show(rep.errors, "ERROR")
    show(rep.warnings, "WARN ")
    show(rep.infos, "info ")
    n_err, n_warn = len(rep.errors), len(rep.warnings)
    print(f"--- {n_err} error(s), {n_warn} warning(s), {len(rep.infos)} note(s) ---")
    if n_err or (args.strict and n_warn):
        print("RESULT: FAIL — fix the errors above and re-run.")
        return 1
    print("RESULT: PASS" + (" (warnings present)" if n_warn else ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
