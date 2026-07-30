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

errors, warnings = [], []
def err(w, m):  errors.append(f"{w}: {m}")
def warn(w, m): warnings.append(f"{w}: {m}")

def is_generated(p):
    # .claude/skills/ holds both hand-authored project skills and copies
    # materialized from plugin/skills/ by dogfood-install.sh. Check sources
    # only; a generated copy is a symptom, and its source is the real target.
    return os.path.exists(os.path.join(os.path.dirname(p), '.dogfood-generated'))

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

    body = text.split('---', 2)[2]
    n = len(body.split('\n'))
    if n > 500: warn(p, f"body is {n} lines (guidance: under 500)")

# ---- 2. bundled paths resolve from the file that cites them ----------------
IGNORE = {'AGENTS.md','CLAUDE.md','CONTRIBUTING.md','SKILL.md','index.md',
          'decisions.md','dispatch.md','.retro.md','.claude/settings.local.json',
          'plugin/.claude-plugin/plugin.json','.claude-plugin/marketplace.json'}
TEACHING = {'template-authoring-standards.md','cross-document-references.md'}
for p in glob.glob('plugin/skills/**/*.md', recursive=True) + glob.glob('.claude/skills/**/*.md', recursive=True):
    if '/references/shared/' in p or is_generated(p): continue
    d = os.path.dirname(p)
    for m in re.finditer(r'`((?:\.\./)*[A-Za-z0-9_./-]+\.(?:md|json|sh|py))`', open(p).read()):
        t = m.group(1)
        if t in IGNORE or t.startswith(('http','docs/','rfcs/','plans/','word/','ooxml/','scripts/')): continue
        if os.path.basename(p) in TEACHING and t == 'references/c4-guide.md': continue
        if not os.path.exists(os.path.join(d, t)):
            err(p, f"backticked path does not resolve from this file: {t}")

# ---- 3. no skill reaches outside its own directory ------------------------
for p in glob.glob('plugin/skills/**/*.md', recursive=True):
    if '/references/shared/' in p: continue
    for m in re.finditer(r'\.\./[a-z-]+/references/', open(p).read()):
        err(p, "cites a sibling skill's references directly; vendor it into "
               "references/shared/ via scripts/sync-shared-references.sh")

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

print("=== check-skills ===")
for e in errors:   print(f"  ERROR {e}")
for w in warnings: print(f"  WARN  {w}")
print(f"--- {len(errors)} error(s), {len(warnings)} warning(s) ---")
strict = os.environ.get('STRICT') == '1'
if errors or (strict and warnings):
    print("RESULT: FAIL"); sys.exit(1)
print("RESULT: PASS" + (" (warnings present)" if warnings else ""))
PY
