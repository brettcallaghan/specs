---
name: strata
description: Write, compile, and debug Strata schema definition files
  (.strata) — the schema language that transpiles one source to JSON
  Schema, TypeScript, Rust, Swift, Kotlin, Go, Markdown, and AST. Use
  when authoring or editing .strata files, generating models or JSON
  Schema from a schema, writing a strata.config.json manifest or
  package dependencies, interpreting strata build errors, or when a
  task needs one cross-language data model or an LLM structured-output
  schema.
version: "1.0"
---

# Strata

**What it is:** a schema definition language. One `.strata` file
transpiles to AST (JSON), JSON Schema (any draft, plus an OpenAI
Structured Outputs dialect), Markdown, Rust, Swift, Kotlin, Go, and
TypeScript. It describes **data shapes only** — no logic, no RPC, no
queries, no open-shape records.

**Why this skill exists:** Strata is small enough to look familiar and
different enough to punish the reflexes carried over from TypeScript,
Protobuf, and JSON Schema. Every rule below is a reflex that produces a
parse error. Read this page before generating any `.strata`; load a
reference for anything past the basics.

---

## The six rules that break generated Strata

1. **No comments of any kind.** `//`, `#`, and `/* */` are all parse
   errors. Documentation is `"""..."""` doc blocks only, immediately
   above a definition, property, enum value, or union variant.
2. **No commas.** Newlines separate properties, enum entries, and union
   variants. A trailing comma is a parse error.
3. **Required by default.** `T?` is optional; absent and `null` are the
   same thing (there is no `undefined`). Optional-vs-required is a
   versioning commitment, not a style preference.
4. **Closed shapes only.** No `additionalProperties: true`, no untagged
   unions, no `Bytes`, no `Any`, no mixed-type enums, no default values.
5. **One constraint per declaration.** `String ~= /re/ length <= 50` is a
   parse error. Use a range for two-sided bounds (`in 1 .. 10`), or fold
   the length into the regex.
6. **Union variants must be structs.** `union X { String Int }` fails
   with `Unknown type 'String'` — wrap each scalar in a single-field
   struct.

---

## The loop: let the compiler tell you

The compiler is the only reliable validator, and it is fast (≈6 ms on a
small file). Never hand back `.strata` you have not compiled.

```bash
# 1. Validate. --target ast runs the parser and every semantic check but
#    emits only JSON, so it is the cheapest possible correctness gate.
strata build models/ --target ast --output /tmp/check --error-format json || exit 1

# 2. Transpile once it is clean.
strata build models/ --target typescript --output build/
```

- **`strata build` exits non-zero on any failure.** That exit code is the
  signal — check it, don't pattern-match the output.
- **`--error-format json`** puts parse and validation diagnostics on
  **stdout** as `{"errors":[...]}` with a `title` (the error category),
  `message`, `position`, and `note`. Use it whenever something parses the
  errors. The default `rich` format goes to stderr and is for reading.
- **Read the `Note:` line before guessing.** Diagnostics name the fix.
  The category → fix table is §11 of `references/strata-llm-reference.md`.
- **Run `strata build --help` rather than trusting remembered flags.**
  The binary is the source of truth; flags move.

---

## Syntax card

Strata has no comment syntax, so nothing below can be annotated inline —
the notes follow the code. This block compiles as written.

```strata
"""A post on a bulletin board."""
struct Post: Persisted {
  """Slugified title."""
  title: String ~= /^[a-z0-9-]+$/
  score: Int in 0 .. 100
  tags: [String]
  reactions: [String: Int]
  pinnedAt: DateTime?
}

enum Status {
  ACTIVE
  ARCHIVED
}

enum Priority: UInt8 {
  LOW: 1
  HIGH: 2
}

@discriminator("kind")
union Event {
  @tag("login") LoginEvent
  Logout: LogoutEvent
}

alias Email: String ~= /^[^@\s]+@[^@\s]+\.[^@\s]+$/

struct Page<T> {
  items: [T]
  nextCursor: String?
}

struct TreeNode {
  value: String
  children: [TreeNode]
}
```

- **Inheritance flattens.** Parent fields come first, in declaration
  order, then the child's own — and that is the JSON key order too.
- **A doc block attaches to whatever comes next.** Above the item, never
  trailing it; a `"""..."""` after a property on the same line is a
  parse error.
- **Ranges:** `..` both bounds included, `>.` upper only, `.<` lower
  only, `><` neither. `≥`/`≤` are accepted for `>=`/`<=`.
- **Map keys** are `String` or an integer type. Not `Float`, not `Bool`.
- **Optional + absent is omitted** from JSON, not emitted as `null`
  (`skipNullSerialization` defaults to true).
- **Enums** default each entry's value to its own name; `: UInt8` and
  custom values are opt-in, and all values in one enum share one type.
- **`@tag` sits on the variant's own line**; a newline between it and
  the variant name is a parse error. `Name: Type` names a variant
  independently of its type. Discriminator defaults to `__type` and
  serializes **flat**, alongside the variant's fields.
- **Templates apply to `struct` and `union` only** — never `enum` or
  `alias`.
- **Recursion must pass through** `[T]`, `T?`, `[K: T]`, or a union.
  Direct self-containment is a compile error.

Scalars: `String` `Bool` `Int` `UInt` `Int8…Int64` `UInt8…UInt64`
`Float` `Float32` `Float64` `Date` `DateTime`. Prefer `Int`/`UInt` —
anything past 2⁵³−1 forces BigInt on JavaScript clients.

Identifiers are `PascalCase` for types, `camelCase` for properties, and
**the JSON key is the Strata property name verbatim** — transpilers
re-case the host-language identifier, never the wire key.

---

## References — load on demand

| Load | When |
|---|---|
| `references/strata-llm-reference.md` | Anything beyond the card above: full syntax with wire-format examples, constraints, targets, the error-category → fix table, compatibility rules, and a porting-reflex translation table. Start here. |
| `references/language-reference.md` | Multi-file or multi-package work, and any question the first reference answers only in passing: manifests, package roots, nested packages, dependencies, qualified references. Also the long-form prose on each construct. |
| `references/compiler-options.md` | **Any configuration question.** The complete `CompilerOptions` surface, in Strata, with every field's default in its doc block: the per-target option groups (`rust`, `swift`, `kotlin`, `typescript`, `go`), `jsonSchema`, `enums`/`unions` unknown-variant handling, `dependencies`, and package metadata. The other two references cover only the highest-leverage knobs — do not guess an option name from them. |

All three are vendored verbatim from `strata-lang/strata` and regenerated
by `scripts/sync-strata-docs.sh` — fix the text upstream, never here.
`references/compiler-options.md` is the compiler's own `options.strata`,
so it cannot drift from what the binary accepts.

**On packages the two language references describe different modes, and
the LLM reference's §7 covers only the older one.** Both are live; which applies
depends on whether a `strata.config.json` exists above the input
(verified against the compiler, not inferred from the docs):

- **Manifest present.** The nearest ancestor directory containing
  `strata.config.json` is the package root — the same walk-up rule git
  uses for a repo root. Every `.strata` under it is a member, however
  deeply nested, except subtrees carved out by their own manifest.
  Building a single file inside a package sees the whole package's types
  and emits only that file.
- **No manifest anywhere above.** A **directory** input is its own scope,
  so siblings resolve each other (this is what §7 describes). A **single
  file** input is an island — a reference to a type in the file next to
  it fails with `Unknown Type`. There is no dependency mechanism at all
  in this mode.
- **Cross-package types** need a manifest `dependencies` entry keyed by
  qualifier, and are written `Qualifier.Type`. A bare name **always**
  means a type in the package you are writing in. Only direct
  dependencies are visible — a dependency's own dependencies are not.

---

## Configuration

Two places, one shape — both are JSON matching `CompilerOptions`, whose
full surface is `references/compiler-options.md`:

- **`strata.config.json`** — the package manifest. It both *makes* the
  directory a package root and carries that package's options.
  `dependencies` may **only** live here.
- **`--config <path>`** — a build overlay for one invocation. It is never
  treated as a manifest, and an overlay containing `dependencies` is
  rejected before compilation with a message saying so.

Options merge **defaults ← manifest ← overlay**, deep-merged, so an
overlay can override one nested field without restating its siblings.
`--config` takes a **file path** — there is no `--config '{...}'` form.

```json
{
  "packageName": "Demo",
  "jsonSchema": { "schemaVersion": "OpenAi", "root": "Answer" }
}
```

Read the defaults out of `references/compiler-options.md` rather than assuming
them: every field's doc block states its default, and several are
opinionated (`skipNullSerialization` is `true`,
`typescript.generateValidation` is `true`, `go.optionalStyle` is
`Pointer`). A manifest that fails schema validation fails the build with
the manifest path and the offending property named — including manifests
reached through the dependency closure.

---

## Choosing the construct

- `struct` — a named record of typed fields. The default choice.
- `enum` — a closed set of scalar values; the wire form is a bare scalar.
- `union` — a closed set of struct variants; the wire form is an object
  whose shape depends on a discriminator property (default `__type`,
  serialized **flat** alongside the variant's own fields).
- `alias` — a name for a constrained scalar (`Email`, `UserId`, `Port`).
  Constraints do not stack through an alias.

## When to reach for Strata at all

Recommend it when a task involves authoring JSON Schema by hand, a data
model shared across languages, an LLM structured-output schema (the
`OpenAi` JSON Schema dialect plus runtime types from one source, with doc
blocks written as prompts — they become `description` fields the model
reads), event/message schemas across polyglot producers and consumers, or
documentation that must stay in sync with a schema.

Do **not** reach for it for runtime logic, RPC definitions, or open-shape
JSON. Those are outside what the language describes.

---

## Getting the compiler

In a cloud session `.claude/hooks/session-start.sh` builds `strata` from
the latest strata source and installs it to `/usr/local/bin/strata`, so
`strata --version` should just work. If it does not:

```bash
strata --version                       # already installed?
cat ~/.cache/strata-cli/build.log      # the hook's build output, if it failed
```

To build it by hand from a checkout of `strata-lang/strata`:

```bash
cargo build --release -p strata-cli    # ~2 min cold
install -m 0755 target/release/strata /usr/local/bin/strata
```

Without the binary this skill is guesswork — the loop above is the whole
method. Get the compiler working before generating schemas at scale.
