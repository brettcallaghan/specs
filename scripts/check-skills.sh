#!/usr/bin/env bash
# check-skills.sh — verify this repo's own invariants about its skills.
#
# Why this exists
#   This framework's thesis is that quality comes from gates, not diligence.
#   It applies that to its users' work orders and blueprints, but every
#   invariant it asserts about ITSELF was enforced by memory — and several
#   had already silently broken.
#
#   The check that prompted this script: an 80-column reflow wrapped a
#   SKILL.md's YAML `description` without continuation indentation. The
#   frontmatter stopped parsing, so the skill's discovery description was
#   gone and it would essentially never trigger — invisible in review,
#   caught only when the skill was loaded for real.
#
# Usage
#   scripts/check-skills.sh            # report problems
#   scripts/check-skills.sh --strict   # treat warnings as errors too
#
# Exit status
#   0  clean (warnings allowed unless --strict)
#   1  errors found
#   2  bad invocation

set -uo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

case "${1-}" in
  --strict) STRICT=1 ;;
  "")       STRICT=0 ;;
  *)        echo "usage: $(basename "$0") [--strict]" >&2; exit 2 ;;
esac

command -v python3 >/dev/null 2>&1 || { echo "check-skills: python3 required" >&2; exit 2; }

STRICT=$STRICT python3 - <<'PY'
import glob, os, re, sys

# The skill re-injection cap: after a compaction Claude re-reads at most
# ~5,000 tokens of a loaded skill, and anything past that is simply gone.
# Expressed in bytes at ~4 bytes/token so the gate needs no tokenizer.
BODY_BYTE_MAX = 20000

errors, warnings = [], []
def err(w, m):  errors.append(f"{w}: {m}")
def warn(w, m): warnings.append(f"{w}: {m}")

def is_generated(p):
    # .claude/skills/ holds hand-authored project skills (triage-issue, strata)
    # and committed symlinks to the plugin/skills/ sources. A symlinked skill
    # is already checked via the plugin/skills glob, so skip it here — checking
    # it twice would double-report and could look like a duplicate skill.
    if os.path.islink(os.path.dirname(p)):
        return True
    # Vendored references (sync-shared-references.sh, sync-strata-docs.sh) all
    # open with this header. Checking the header rather than a path pattern
    # keeps this correct as vendoring destinations grow — and a file whose
    # source of truth is another repo cannot be fixed by editing it here.
    try:
        with open(p) as f:
            return f.readline().startswith('<!-- GENERATED COPY')
    except (OSError, UnicodeDecodeError):
        return False

SKILLS = [p for p in sorted(glob.glob('plugin/skills/*/SKILL.md'))
                  + sorted(glob.glob('.claude/skills/*/SKILL.md'))
          if not is_generated(p)]

# ---- 1. frontmatter parses, and the two required fields are valid ----------
try:
    import yaml
    HAVE_YAML = True
except ImportError:
    HAVE_YAML = False
    warnings.append("frontmatter: pyyaml not installed — using a reduced structural check")

for p in SKILLS:
    text = open(p).read()
    if not text.startswith('---\n'):
        err(p, "does not begin with a YAML frontmatter block"); continue
    fm = text.split('---', 2)[1]
    if HAVE_YAML:
        try:
            d = yaml.safe_load(fm)
        except Exception as e:
            # This is the failure mode that motivated the script: a wrapped
            # scalar whose continuation lines lost their indentation.
            err(p, f"frontmatter does not parse ({type(e).__name__}). A wrapped "
                   f"`description:` needs its continuation lines indented.")
            continue
        if not isinstance(d, dict):
            err(p, "frontmatter is not a mapping"); continue
    else:
        d = {}
        for m in re.finditer(r'^([a-z_]+):[ ]*(.*)$', fm, re.M):
            d[m.group(1)] = m.group(2).strip().strip('"')
        for ln in fm.strip().split('\n'):
            if ln and not ln.startswith(' ') and not re.match(r'^[a-z_]+:', ln):
                err(p, f"frontmatter line is neither a key nor an indented "
                       f"continuation: {ln[:60]!r}")

    name, desc = d.get('name'), d.get('description')
    if not name:
        err(p, "missing `name`")
    else:
        if len(name) > 64:                     err(p, f"name is {len(name)} chars (max 64)")
        if not re.fullmatch(r'[a-z0-9-]+', name): err(p, f"name '{name}' must be lowercase letters, numbers, hyphens")
        for w in ('anthropic', 'claude'):
            if w in name.lower():              err(p, f"name contains the reserved word '{w}'")
        if os.path.basename(os.path.dirname(p)) != name:
            err(p, f"name '{name}' does not match its directory")
    if not desc:
        err(p, "missing `description` — without it the skill will not be discovered")
    else:
        if len(desc) > 1024: err(p, f"description is {len(desc)} chars (max 1024)")
        if len(desc) < 40:   warn(p, f"description is only {len(desc)} chars — likely too thin to trigger on")
        for pov in (' I ', 'I can ', 'you can ', 'You can '):
            if pov in desc: warn(p, f"description may not be third person (found {pov!r})")

    # Size the body against the 5,000-token skill re-injection cap — what
    # Claude re-reads after a compaction. Lines are the wrong unit: the same
    # content rewrapped from 80 to 200 columns loses two thirds of its lines
    # and not one token, so a line gate is satisfied by reflowing. Bytes are
    # proportional to tokens, need no tokenizer (this must run in CI and on a
    # laptop with nothing installed), and at the ~4 bytes/token rule of thumb
    # for English prose plus markdown, 5,000 tokens is ~20,000 bytes.
    body = text.split('---', 2)[2]
    n = len(body.encode('utf-8'))
    if n > BODY_BYTE_MAX:
        warn(p, f"body is {n} bytes (~{n // 4} tokens); above ~{BODY_BYTE_MAX} "
                f"the tail is truncated from context after a compaction")

# ---- 2. bundled paths resolve from the file that cites them ----------------
IGNORE = {'AGENTS.md','CLAUDE.md','CONTRIBUTING.md','SKILL.md','index.md',
          'decisions.md','dispatch.md','.handoff.md','SNAPSHOT.md',
          '.claude/settings.local.json',
          'plugin/.claude-plugin/plugin.json',
          # Filenames a skill *describes* rather than ships: these name files
          # in the user's project, or upstream in strata-lang/strata, so there
          # is nothing here to resolve.
          'strata.config.json','strata-config.json','strata.config.schema.json',
          'options.strata'}
TEACHING = {'spec-authoring-standards.md','cross-document-references.md'}
for p in glob.glob('plugin/skills/**/*.md', recursive=True) + glob.glob('.claude/skills/**/*.md', recursive=True):
    if '/references/shared/' in p or is_generated(p): continue
    d = os.path.dirname(p)
    for m in re.finditer(r'`((?:\.\./)*[A-Za-z0-9_./-]+\.(?:md|json|sh|py))`', open(p).read()):
        t = m.group(1)
        # `hooks/` and `monitors/` are plugin-ROOT components, not skill-bundled
        # files: a skill names them so a reader knows what the plugin installs
        # around it, and the path a reader actually uses is
        # ${CLAUDE_PLUGIN_ROOT}-relative. Nothing resolves from a skill
        # directory and nothing should — same reason `scripts/` is here.
        if t in IGNORE or t.startswith(('http','docs/','plans/','word/','ooxml/','scripts/','.claude/','hooks/','monitors/')): continue
        if os.path.basename(p) in TEACHING and t == 'references/style-guide.md': continue
        if not os.path.exists(os.path.join(d, t)):
            err(p, f"backticked path does not resolve from this file: {t}")

# ---- 3. no skill reaches outside its own directory ------------------------
for p in glob.glob('plugin/skills/**/*.md', recursive=True):
    if '/references/shared/' in p: continue
    for m in re.finditer(r'\.\./[a-z-]+/references/', open(p).read()):
        err(p, "cites a sibling skill's references directly. Two remedies, and "
               "which is right is a rule, not a preference: VENDOR when the "
               "citing skill needs the WHOLE cited file and only APPLIES it — "
               "a procedure, checklist, standard or template — by adding a MAP "
               "line to scripts/sync-shared-references.sh and citing the "
               "generated references/shared/ copy. CITE the owning skill BY "
               "ROLE in prose, with no path, when you need one fact out of a "
               "larger document (the generator vendors whole files, never "
               "sections), or when the file is a REGISTRY that a second copy "
               "would give a second place to extend. Both tests and the two "
               "worked examples: AGENTS.md, \"Citing another skill's "
               "reference\".")

# ---- 4. long reference files carry a table of contents --------------------
for p in glob.glob('plugin/skills/*/references/*.md'):
    if '/shared/' in p: continue
    t = open(p).read()
    if len(t.split('\n')) > 100 and not re.search(r'^#+ Contents|^Sections \(complete', t, re.M):
        err(p, "over 100 lines with no table of contents")

# ---- 5. one routing vocabulary, defined in one place ----------------------
TIERS = ['light','standard','advanced','frontier']
schema = open('plugin/skills/groom-beads/references/bead.schema.json').read()
m = re.search(r'"tier":\s*\{\s*"enum":\s*\[([^\]]+)\]', schema)
if not m:
    err('bead.schema.json', "no `tier` enum found")
elif [x.strip().strip('"') for x in m.group(1).split(',')] != TIERS:
    err('bead.schema.json', f"tier enum does not match {TIERS}")
val = open('plugin/skills/groom-beads/scripts/validate_beads.py').read()
m = re.search(r'^TIERS\s*=\s*\[([^\]]+)\]', val, re.M)
if not m or [x.strip().strip('"') for x in m.group(1).split(',')] != TIERS:
    err('validate_beads.py', f"TIERS does not match the schema enum {TIERS}")
for p in SKILLS + glob.glob('plugin/skills/*/references/*.md'):
    if '/shared/' in p: continue
    body = open(p).read()
    if re.search(r'\bmodel_reason\b', body):
        err(p, "uses `model_reason` — routing is by `tier`/`tier_reason`")

# ---- 6. no `version` in SKILL.md frontmatter ------------------------------
# The Agent Skills upload check (claude.ai, the Skills API) rejects it, and
# Claude Code ignores it. The plugin's release version identifies a skill.
for p in SKILLS:
    if re.search(r'^version:', open(p).read().split('\n---', 1)[0], re.M):
        err(p, "frontmatter has `version`; skills are identified by the plugin's release version")

# ---- 7. references other than templates are leaves -------------------------
# A template's header names every file its type needs, so an agent loads one
# list up front. Any other reference that cites a sibling reference adds a
# second hop, and a reference reached at the second hop tends to get
# previewed rather than read. A leaf names a sibling by role in prose.
# Citations are backticked paths and `[[slug]]`s that resolve to a sibling's
# frontmatter slug; the edges from templates must also form no cycle.
for skill in sorted(glob.glob('plugin/skills/*/')):
    refs = [f for f in glob.glob(os.path.join(skill, 'references', '**', '*.md'), recursive=True)
            if os.path.basename(f) != 'README.md']
    if not refs: continue
    by_name, by_slug = {}, {}
    for f in refs:
        by_name.setdefault(os.path.basename(f), f)
        m = re.search(r'^slug:\s*(\S+)', open(f).read(), re.M)
        if m: by_slug[m.group(1)] = f
    edges = {}
    for f in refs:
        text = open(f).read()
        out = set()
        for m in re.finditer(r'`[^`\s]*?([\w.-]+\.md)(?:#[^`]*)?`', text):
            t = by_name.get(m.group(1))
            if t and t != f: out.add(t)
        for m in re.finditer(r'\[\[([a-z0-9-]+)', text):
            t = by_slug.get(m.group(1))
            if t and t != f: out.add(t)
        edges[f] = out
        if out and not f.endswith('-template.md'):
            err(f, "a reference other than a template cites another reference ("
                   + ", ".join(sorted(os.path.basename(t) for t in out))
                   + "); name it by role in prose, and list it in the templates that need it")
    state = {}
    def visit(n, path):
        state[n] = 1
        for t in sorted(edges.get(n, ())):
            if state.get(t) == 1:
                err(n, "reference cycle: " + " -> ".join(os.path.basename(x) for x in path[path.index(t):] + [t]))
            elif t not in state:
                visit(t, path + [t])
        state[n] = 2
    for n in sorted(edges):
        if n not in state: visit(n, [n])

print("=== check-skills ===")
for e in errors:   print(f"  ERROR {e}")
for w in warnings: print(f"  WARN  {w}")
print(f"--- {len(errors)} error(s), {len(warnings)} warning(s) ---")
strict = os.environ.get('STRICT') == '1'
if errors or (strict and warnings):
    print("RESULT: FAIL"); sys.exit(1)
print("RESULT: PASS" + (" (warnings present)" if warnings else ""))
PY
