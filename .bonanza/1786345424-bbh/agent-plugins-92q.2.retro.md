`scripts/check-skills.sh` warns when a skill body exceeds **500 lines**:

```python
    body = text.split('---', 2)[2]
    n = len(body.split('\n'))
    if n > 500: warn(p, f"body is {n} lines (guidance: under 500)")
```
(`scripts/check-skills.sh:117-118`)

Lines are the wrong unit, and that is why this gate has been firing for weeks
while telling nobody anything useful. The real constraint is the **5,000-token
skill re-injection cap** after compaction, which is a size in tokens, not a
count of newlines.

Measured on `plugin/skills/bead-bonanza/SKILL.md` — 869 lines, 45,114 body
chars, mean line length 51.9:

| Unit | Value | Bearing on the real constraint |
|---|---|---|
| lines | 869 | none — an artefact of the 80-column wrap convention |
| body bytes | ~45,100 | proportional to the tokens the cap counts |
| body tokens | ~11,500 | this is what the 5,000-token cap actually limits |

At 52 chars/line the warning fires at ~26 KB, while the *same content*
rewrapped at 200 chars reads as 230 lines and passes silently at the same
size. So the gate can be satisfied by rewrapping, which changes nothing about
whether the skill survives compaction. Conversely the 80-column convention is
good for diffs and review and should be kept — it is not the problem.

Consequence today: `scripts/check-skills.sh --strict` exits 1 with
`0 error(s), 1 warning(s)` on a clean tree, `bead-bonanza` the sole finding.
That red is carried as a hand-written exception in `AGENTS.md`'s
environmental-failure ledger and is re-copied into every bead-bonanza dispatch
prompt.

**This is a recurrence, not a new finding.** RFC-0004's Resolution deferred
"record the gate profile's baseline verdict" to bead `agent-plugins-1a9`
(theme `gate-baseline`) explicitly blocked on this units question: "the
standard states the body budget in CHARACTERS while `check-skills.sh` measures
LINES." That bead no longer exists in the database and never reached
`.beads/issues.jsonl` on any branch, so the deferral was lost — see the
sibling bead that re-files it. The blocking question now has an answer.


## Retro — implementation

### Findings
- The bead's Design step 4 asks which skills newly trip the threshold. One does: `plugin/skills/spec/SKILL.md` at 21,419 body bytes (~5,354 tokens) — 319 lines, so it passed the line gate silently while being over the real cap. The gate baseline therefore moves from 1 warning to 2, and the second one is a genuine new finding about a skill nobody was watching. No bead owns trimming it.
- The Design's measured figures (45,114 body chars) were counted as `len(str)`; the implementation counts `len(str.encode('utf-8'))`, so bead-bonanza reports 45,317. Not a contradiction, but the bead's numbers do not reproduce exactly against the shipped gate.

### What I tried that did not work
- Nothing abandoned. First reading was the implemented one.

### What I decided silently
- Byte count is of the UTF-8 encoding, not of the Python `str`. The two differ by ~300 bytes on bead-bonanza because of em dashes and arrows. UTF-8 is the right unit: the cap counts what is on the wire, and it errs high, which is the safe direction.
- The threshold is a named module constant `BODY_BYTE_MAX` rather than an inline literal, so the number and its derivation live in one place and the warning string cannot drift from the test.
- The reported token figure is `bytes // 4` (integer division), inlined rather than a second constant — it is an order-of-magnitude aid in the message, not a measurement, and pretending to more precision would invite treating it as one.

### Observed pre-fix failure
    scripts/check-skills.sh --strict
      WARN  plugin/skills/bead-bonanza/SKILL.md: body is 859 lines (guidance: under 500)
      --- 0 error(s), 1 warning(s) ---
      RESULT: FAIL
The unit is the defect: the same body rewrapped at 200 columns reads as ~230 lines and passes, at identical size.
