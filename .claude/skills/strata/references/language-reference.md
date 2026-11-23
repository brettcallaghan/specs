<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: strata-lang/strata site/content/docs/reference/language.mdx
     Vendored from release: v0.9.0
     Regenerate with: scripts/sync-strata-docs.sh
     Upstream owns this text. Fix it there, then re-run that script. -->

import { Callout } from 'nextra/components'

# Language reference

Look up any construct in the Strata language here while you write a schema. Each entry gives the construct's syntax, the types it accepts, and the rules the compiler enforces.

<Callout type="info">
New to Strata? The [quick start](/docs/quick-start) walks you through your first schema. For step-by-step recipes, such as modeling polymorphic data, porting nullable types from GraphQL, and sharing types between packages, see [Model data](/docs/guides/model-data). For the reasoning behind the rules on this page, see [Language design](/docs/concepts/language).
</Callout>

## Source files

A Strata source file has the extension `.strata` and contains any number of top-level declarations in any order. A declaration can reference a type declared later in the file, or in another file of the same [package](#packages).

| Declaration | Keyword | Defines |
| :--- | :--- | :--- |
| [Struct](#structs) | `struct` | A named group of typed properties |
| [Enum](#enums) | `enum` | A scalar type with a fixed set of named values |
| [Union](#unions) | `union` | A value that is exactly one of several structs |
| [Type alias](#type-aliases) | `alias` | A new name for a type expression |

Line breaks separate the members of a declaration (properties, enum entries, and union members). There's no comma or semicolon separator.

### Source file names

Below the package root, every file and directory name must be a portable name. A portable name starts with an ASCII letter and contains only ASCII letters, digits, `_`, and `-`. The `.strata` extension isn't part of the name.

| Name | Portable |
| :--- | :--- |
| `user.strata`, `user_profile.strata`, `user-profile.strata`, `v2/api.strata` | Yes |
| `1st.strata`, `_drafts/user.strata`, `my file.strata`, `user.v2.strata`, `café.strata` | No |

The build doesn't check directories above the package root, so a package can live under `~/My Projects/`. It checks a dependency's names below the dependency's own package root. Without a manifest, it checks names below the input directory, or below a file input's own directory.

A name that isn't portable fails the build before it writes any file. The error names the segment and the file:

```text
Error: Source Name Error: "1st-order" in "1st-order.strata" is not a portable name. A portable name starts with an ASCII letter and contains only ASCII letters, digits, "_", and "-".
```

Two source paths in one package can't differ only in letter case, or only in `-` against `_`. `user-profile.strata` and `user_profile.strata` fail the build together:

```text
Error: Source Name Error: "user-profile.strata" and "user_profile.strata" differ only in letter case or in "-" against "_", so a target would write them to one output file. Rename one of them.
```

The same name in two directories (`shared/types.strata` and `api/types.strata`), or in a package and its dependency, builds. A target that reserves a name renames that one output file and keeps every type in it. For example, Rust writes `self.strata` as `self_.rs`, and Go writes `user_test.strata` as `user_test_gen.go`.

## Structs

```strata filename="user.strata"
struct User {
  name: String
  email: String
  age: Int32?
}
```

A struct is a named group of properties. Each property is written `name: Type`. `Type` is any [scalar](#scalars), [collection](#collections), enum, union, struct, alias, or [template](#templates) instantiation, and it can be followed by `?` ([optionals](#optionals)) and a [constraint](#constraints).

- Property names are unique within a struct, inherited properties included.
- A struct can have no properties: `struct Empty {}`.

### Inheritance

```strata filename="marker.strata"
struct Positioned {
  x: Float
  y: Float
}

struct Colored {
  color: String
}

struct Marker: Positioned, Colored {
  label: String
}
```

A struct inherits every property of each struct listed after the colon. `Marker` above has the properties `x`, `y`, `color`, and `label`.

- A struct can list any number of parents, separated by commas.
- Every parent is a struct. An alias is rejected as a parent, even an alias of a struct.
- A [template](#templates) instantiation can be a parent: `struct LabeledValue<T>: Value<T>`.
- A struct can't redeclare a property it inherits. The compiler reports `Invalid redeclaration of ancestor property 'Positioned.x'`.

Inheritance doesn't appear on the wire: a `Marker` is one JSON object with four properties. Each [target page](/docs/reference/targets) shows how that target generates an inheriting struct. [How inheritance survives languages without it](/docs/concepts/language#how-inheritance-survives-languages-without-it).

## Scalars

| Type | Values |
| :--- | :--- |
| `String` | A sequence of Unicode characters. |
| `Bool` | `true` or `false`. |
| `Int` | Signed integer, −9,007,199,254,740,991 to 9,007,199,254,740,991. |
| `Int8` | −128 to 127. |
| `Int16` | −32,768 to 32,767. |
| `Int32` | −2,147,483,648 to 2,147,483,647. |
| `Int64` | −9,223,372,036,854,775,808 to 9,223,372,036,854,775,807. |
| `UInt` | Unsigned integer, 0 to 9,007,199,254,740,991. |
| `UInt8` | 0 to 255. |
| `UInt16` | 0 to 65,535. |
| `UInt32` | 0 to 4,294,967,295. |
| `UInt64` | 0 to 18,446,744,073,709,551,615. |
| `Float` | Floating-point number. Every target represents it as a 64-bit double. |
| `Float32` | 32-bit floating-point number, single precision (about 7 decimal digits). |
| `Float64` | 64-bit floating-point number, double precision (about 15–17 decimal digits). |
| `DateTime` | A calendar date and time with an offset from UTC, encoded as an [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) string: `YYYY-MM-DDTHH:MM:SS[.fraction]` followed by `Z` or `±HH:MM`. See [DateTime](#datetime). |
| `Date` | A calendar date, encoded as an [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) `full-date` string: `YYYY-MM-DD`. |

`Int` and `UInt` are limited to ±(2^53 − 1). `Int64` and `UInt64` cover the full 64-bit range. Every language target generates a 64-bit integer type for `Int` and `UInt`. The *Safe integers* row of the [feature matrix](/docs/reference/targets#feature-matrix) shows which targets reject an out-of-range value on decode. [Why 53 bits?](/docs/concepts/wire-formats/json#why-int-stops-at-53-bits-and-why-64-bit-integers-are-still-open)

### DateTime

A `DateTime` value is a calendar clock that carries the offset it was written with.

- **Offset.** Decoding `2024-06-15T08:00:00+02:00` and re-encoding it produces `+02:00`, not `Z` and not a shifted wall-clock time.
- **Zero offset.** `Z`, `+00:00`, and `-00:00` denote the same value.
- **Precision.** Fractional seconds are held to nanosecond precision, and `.1` and `.100` are the same value.
- **Equality.** Two values are equal when they denote the same instant with the same offset. Comparing encoded strings byte-for-byte isn't a valid equality test.

The [JSON wire format](/docs/reference/wire-formats/json#datetime) shows how a `DateTime` is written, including the zero-offset spelling and fraction truncation. [DateTime across targets](/docs/reference/targets#datetime-across-targets) compares how each target implements equality and what each decoder accepts beyond the strict grammar. [Why a calendar clock and not an instant?](/docs/concepts/wire-formats/json#a-datetime-is-the-value-not-the-text)

## Collections

### Arrays

```strata
struct Transform {
  tags: [String]
  matrix: [[Float]]
}
```

`[T]` is an ordered list of values of type `T`. `T` can be any type, including another array.

### Maps

```strata
struct Inventory {
  countsBySku: [String: Int32]
  namesById: [UInt32: String]
}
```

`[K: V]` maps keys of type `K` to values of type `V`.

- `K` is `String`, `Bool`, or an integer type (`Int`, `Int8`, `Int16`, `Int32`, `Int64`, `UInt`, `UInt8`, `UInt16`, `UInt32`, or `UInt64`). Any other key type is rejected, including a float, `Date`, `DateTime`, an enum, a struct, an alias, and an optional.
- `V` can be any type.
- Entry order isn't preserved.

The [feature matrix](/docs/reference/targets#feature-matrix) shows per-target support for `Bool` and integer keys, and [Arrays and maps](/docs/reference/wire-formats/json#arrays-and-maps) shows how keys are written.

## Optionals

```strata
struct User {
  name: String
  email: String
  age: Int32?
}
```

Every property is required unless its type ends in `?`. An optional property can be absent from the payload or present as `null`, and Strata doesn't distinguish the two. [Optional properties](/docs/reference/wire-formats/json#optional-properties) gives the encoding. [Why?](/docs/concepts/wire-formats/json#absent-and-null-are-one-state)

### Where optionality can be declared

A `?` is accepted at these positions:

| Position | Example |
| :--- | :--- |
| Property type | `age: Int32?` |
| Array element | `[User?]` |
| Map value | `[String: User?]` |
| Template argument | `Paginated<User?>` |

Two rules limit where a `?` can appear. [Why?](/docs/concepts/language#why-optionality-lives-at-the-use-site)

**An alias is never optional.** `alias OptionalColor: RgbColor?` is rejected with `Alias 'OptionalColor' is optional; write 'OptionalColor?' at each use site instead`. An alias whose optional is nested inside a collection, such as `alias Names: [String?]`, is accepted.

**An optional of an optional is rejected.** `String??` doesn't parse. A template instantiation that would produce one is rejected too:

```strata
struct Box<T> {
  value: T?
}

struct Root {
  b: Box<String?>
}
```

```
Error: Optional of Optional: Template 'Box' places parameter 'T' in an optional position, so instantiating it with 'String?' would give an optional of an optional
```

`Box<String>` already yields an optional `value`. An optional argument is accepted wherever the parameter isn't already in an optional position. For example, `Paginated<User?>` with `items: [T]` gives `items: [User?]`.

## Literals

Literals are the constant values written in a schema.

| Literal | Examples | Used in |
| :--- | :--- | :--- |
| String | `"pending"`, `"🥭"`, `"\n"` | Enum values, union tags, the discriminator name, equality constraints |
| Integer | `42`, `-1`, `1_000_000`, `0b1010`, `0o755`, `0xCAFE` | Enum values, union tags, numeric and length constraints |
| Float | `0.5`, `-1.25`, `1e3`, `2.5E-3` | Numeric constraints on float types |
| Boolean | `true`, `false` | Equality constraints on `Bool` |
| Regular expression | `/^#[0-9A-Fa-f]{6}$/` | [Regex constraints](#regex-constraints) |

- **Strings** accept the escapes `\n`, `\t`, `\r`, `\\`, `\'`, `\"`, `\0`, and `\u{...}`, a Unicode code point of one to eight hex digits. Any other escape is an error.
- **Integers** take an optional leading `-`, an optional `0b` (binary), `0o` (octal), or `0x` (hexadecimal) prefix, and `_` between digits.
- **Floats** have a fractional part, an exponent (`e` or `E`, optionally signed), or both.

## Enums

```strata filename="status.strata"
enum Status {
  Inactive
  Pending
  Active
}
```

An enum is a scalar type whose value is one of a fixed set of named entries.

- **Implicit values.** An entry without a value takes its own name as a `String` value, so `Pending` above is the string `"Pending"` on the wire.
- **Explicit values.** `Name: value` gives an entry a value of its own: `Pending: "pending"`, `Pending: 2`.
- **Value types.** Values are strings or integers. `Bool` and float values are rejected, as is a `Bool` enum type.
- **Type.** An enum's value type is `String` by default, inferred from the first explicit value when one is given, or declared after a colon. A declared type sets the integer width:

  ```strata filename="status.strata"
  enum Status: UInt8 {
    Inactive: 0
    Pending: 2
    Active: 3
  }
  ```

- **Consistency.** Every value has the same type, and no two entries share a value. In an integer enum, every entry has an explicit value, because an implicit value is a string. The compiler reports a mismatch as `Expected value of type 'Int', found 'String'` and a repeat as `Enum 'Status' has duplicate value '2'`.
- **Size.** An enum has at least one entry.

### Enum spread

```strata
enum Status {
  Inactive
  Pending
  Active
}

enum DetailedStatus {
  ...Status
  Canceled
}
```

`...Name` copies every entry of another enum, with its value, into this one.

- Spreads are transitive: an enum that spreads `DetailedStatus` also receives `Status`'s entries.
- Both enums have the same value type. Spreading a `UInt8` enum into a `String` enum is rejected.
- An entry can't redeclare a spread entry.

## Unions

```strata filename="shape.strata"
struct Rectangle {
  width: Float
  height: Float
}

struct Circle {
  radius: Float
}

union Shape {
  Rectangle
  Circle
}
```

A union value is exactly one of its member structs.

- Every member is a struct, or an alias of one. Any other member is rejected. For example, an enum member is reported as `Type 'E' is not a struct. Union 'Shape' may only contain struct types`.
- A union has at least one member.

A union value carries one extra property, the **discriminator**, whose value, the **tag**, names the member. The discriminator is `__type` unless [`@discriminator`](#custom-discriminators) renames it, and the tag is the member's name unless [`@tag`](#custom-tags) sets it. [Unions](/docs/reference/wire-formats/json#unions) gives the JSON encoding.

### Custom tags

```strata
union Shape {
  @tag(1) Rectangle
  @tag(2) Circle
}
```

`@tag(...)` sets a member's tag value.

- A tag is a string or integer literal.
- Every tag in a union has the same type, the default name tags included: once one member has an integer tag, every member needs one. A mix is reported as `Union 'Shape' tags must be of the same type`.
- No two members share a tag.

### Custom discriminators

```strata
@discriminator("shape")
union Shape {
  Rectangle
  Circle
}
```

`@discriminator(...)` takes a string literal and renames the discriminator property. With it, a `Circle` encodes as `{ "shape": "Circle", "radius": 3.14 }`. The name can't equal a property of any member: `@discriminator("radius")` is rejected.

### Named members

```strata
union Outcome {
  Ok: Success
  Err: Failure
}
```

`Name: Type` gives a member a name of its own, which becomes its default tag. `@tag` applies to a named member too: `@tag("ok") Ok: Success`.

A member must be named when its type is a [template](#templates) instantiation or a [qualified reference](#qualified-references):

```strata
union ScalarValue {
  string: Value<String>
  int: Value<Int>
}
```

Member names are unique within a union, and so are member types. Two members that resolve to the same type, such as `Width: Length` and `Height: Length`, are rejected. `Value<String>` and `Value<Int>` are different types and can share a union.

### Union spread

```strata
union Shape2D {
  Rectangle
  Circle
}

union Shape3D {
  ...Shape2D
  Sphere
}
```

`...Name` copies every member of another union, with its tag, into this one.

## Documentation comments

Documentation comments are Strata's only comment form. There are no line comments: `#` and `//` are syntax errors.

```strata
"""
A circle: every point in a plane at a fixed distance from a center point.
"""
struct Circle {
  """
  The distance from the center to any point on the circumference.
  """
  radius: Float
}
```

A block of Markdown between triple quotes documents the declaration or member that follows it. Documentation comments are accepted before:

- a struct, enum, union, or alias
- a struct property
- an enum entry
- a union member

A comment before a spread (`...Status`) is a syntax error. The one-line form, `"""The value stored in this node."""`, is also accepted. Each language target carries the text into its native documentation format, and the Markdown and [doc-generation](/docs/reference/doc-generation) outputs render it.

## Constraints

A constraint follows a type and narrows the values it accepts.

| Constraint | Syntax | Applies to |
| :--- | :--- | :--- |
| [Regex](#regex-constraints) | `~= /pattern/` | `String` |
| [Comparison](#comparisons) | `> n`, `>= n`, `< n`, `<= n` | Integer and float types |
| [Range](#ranges) | `in a .. b` (and three open forms) | Integer and float types |
| [Equality](#equality-constraints) | `== value` | `String`, `Bool`, and integer types |
| [Length](#length-constraints) | `length` followed by a comparison, range, or equality | `String`, arrays, and maps |

A constraint on any other type is rejected, for example with `Equality constraints are not supported for type 'Float'`. Those types include an enum, `Date`, `DateTime`, a struct, and an alias. An alias carries only the constraint written in its declaration. A template parameter (`v: T > 0`) can't be constrained either.

<Callout type="warning">
Not every target checks constraints when data arrives. The *Runtime constraint validation* row of the [feature matrix](/docs/reference/targets#feature-matrix) lists which do.
</Callout>

### Placement

One constraint applies per type position. `Int > 0 < 10` is a syntax error, and a range expresses both bounds. A container and its element type each have their own position, and the `?` of an optional goes before the constraint:

```strata
struct Scores {
  scores: [Int > 0] length in 1 .. 10
  labels: [String: String length in 1 .. 50] length <= 20
  bonus: Int? in 0 .. 100
  page: Paginated<Int32 > 0>
}
```

Every bound fits the constrained type. An integer type takes integer bounds within its range, so `UInt8 > 300` and `Int in 0.5 .. 10` are rejected. A float type takes integer or float bounds.

### Regex constraints

```strata
struct Layer {
  color: String ~= /^#[0-9A-Fa-f]{6}$/
}
```

`~=` requires a `String` to match a regular expression written between `/` delimiters. A pattern can't contain `/`, escaped or not. Write a slash as `\x2F`.

Each target runs patterns on its own engine:

| Target | Engine | Lookaround | Backreferences |
| :--- | :--- | :---: | :---: |
| JSON Schema | ECMA 262 | ● | ● |
| TypeScript | ECMA 262 | ● | ● |
| Swift | `NSRegularExpression` (ICU) | ● | ● |
| Kotlin | `java.util.regex` | ● | ● |
| Rust | [`regex`](https://docs.rs/regex/) | – | – |
| Go | RE2 | – | – |

<SupportLegend yes na="Not supported by the engine" />

Strata accepts only the dialect every engine runs the same way. `strata build` checks every pattern and reports two kinds of error:

- **Invalid Regular Expression**: the pattern doesn't parse, or it uses a backreference (`\1`) or lookaround (`(?=...)`, `(?!...)`, `(?<=...)`, `(?<!...)`).
- **Unsupported Regular Expression**: the pattern parses but is outside the shared dialect. The message names the engine that disagrees and, where one exists, the portable spelling.

| Supported | Rejected as unsupported |
| :--- | :--- |
| Character classes `[...]`, negated classes `[^...]` | Nested classes `[[abc]d]`, class intersection `[a-z&&[b-d]]`, POSIX classes `[[:alpha:]]` |
| Quantifiers `*`, `+`, `?`, `{n}`, `{n,}`, `{n,m}`, and their lazy forms | Inline flags `(?i)` and scoped flags `(?i:...)` |
| Alternation `\|`, groups `(...)`, non-capturing groups `(?:...)`, named groups `(?<name>...)` | The named-group spelling `(?P<name>...)` |
| Anchors `^`, `$`, word boundaries `\b`, `\B` | Text anchors `\A`, `\z` |
| Escapes `\d`, `\w`, `\s` (and `\D`, `\W`, `\S`), `\n`, `\t`, `\r`, `\xHH`, `\u{...}`, and escaped metacharacters | An escaped character that isn't a metacharacter (`a\-b` outside a class), `\a`, `\x{...}` |
| Unicode general categories `\p{L}`, `\p{Nd}`, `\P{L}` | Long category names `\p{Letter}`, scripts `\p{Greek}`, binary properties `\p{Alphabetic}`, the brace-less `\pL` |
| | An unescaped `]` or `}` outside a class |

[Write a regex every target accepts](/docs/guides/model-data#write-a-regex-every-target-accepts) shows the common rewrites. [Why the dialect is restricted](/docs/concepts/language#why-regex-patterns-are-restricted).

### Comparisons

| Operator | Meaning |
| :--- | :--- |
| `>` | Greater than |
| `>=` or `≥` | Greater than or equal to |
| `<` | Less than |
| `<=` or `≤` | Less than or equal to |

```strata
struct Limits {
  positive: Int > 0
  atMost: Int <= 42
  ratio: Float ≥ 0.5
}
```

### Ranges

| Operator | Name | Equivalent to |
| :--- | :--- | :--- |
| `a .. b` | Closed | `x >= a && x <= b` |
| `a >. b` | Left-open | `x > a && x <= b` |
| `a .< b` | Right-open | `x >= a && x < b` |
| `a >< b` | Open | `x > a && x < b` |

```strata
struct Ranges {
  closed: Int in 1 .. 10
  leftOpen: Int in 1 >. 10
  rightOpen: Int in 1 .< 10
  open: Int in 1 >< 10
  byte: UInt8 in 0x00 .. 0xFF
}
```

The lower bound is less than the upper bound. `in 10 .. 1` is rejected.

### Equality constraints

```strata
struct ClickEvent {
  kind: String == "click"
  version: Int == 2
  trusted: Bool == true
}
```

`==` fixes a property to a single value. It applies to `String`, `Bool`, and integer types.

### Length constraints

```strata
struct Team {
  name: String length in 3 .. 20
  members: [String] length >= 1
  settings: [String: String] length <= 20
}
```

`length` constrains the number of characters in a `String`, elements in an array, or entries in a map. It's followed by a comparison, range, or equality, and its bounds aren't negative.

## Type aliases

```strata
alias RgbColor: String ~= /^#[0-9A-Fa-f]{6}$/
alias PositiveInt: Int > 0
```

`alias Name: Type` declares `Name` as another name for a type expression, including any constraint on it. The type can be any type: a scalar, collection, struct, enum, union, template instantiation, another alias, or a [qualified reference](#qualified-references).

- An alias is used wherever a type is accepted, except as a map key or a struct's parent.
- An alias doesn't end in `?`. See [Where optionality can be declared](#where-optionality-can-be-declared).
- A use of an alias takes no further constraint.
- An alias has no type parameters: `alias List<T>: [T]` is a syntax error.
- Aliases can refer to other aliases, but not in a cycle.

## Templates

```strata filename="paginated.strata"
struct Paginated<T> {
  items: [T]
  page: Int32
  total: Int32
}
```

A template is a struct or union with type parameters, listed in angle brackets after its name and separated by commas: `struct Pair<T, U>`. Enums and aliases take no type parameters.

- An instantiation such as `Paginated<User>` supplies exactly one type per parameter. A bare `Paginated` is rejected with `Template 'Paginated' requires 1 type argument, but none were provided`.
- An instantiation is accepted anywhere a type is: properties, collection elements, map values, optionals, other template arguments, parent lists, and union members (which must be [named](#named-members)).
- A parameter takes no constraint inside the template body.
- A union template's parameter, when used as a member, must be instantiated with a struct.

```strata
struct Value<T> {
  value: T
}

struct LabeledValue<T>: Value<T> {
  label: String
}

union Result<T> {
  Ok: T
  Err: Failure
}

struct Demo {
  page: Paginated<User>
  nested: Value<Paginated<User>>
  labeled: [LabeledValue<Int>]
  result: Result<User>
}
```

## Recursive types

```strata filename="tree.strata"
struct TreeNode {
  value: String
  children: [TreeNode]
}
```

A type can refer to itself, directly or through a cycle of other types, as long as every reference in the cycle passes through one of these:

| Indirection | Example |
| :--- | :--- |
| Array | `children: [TreeNode]` |
| Optional | `next: ListNode?` |
| Map | `children: [String: FileNode]` |
| Union | `left: Expr`, inside `BinaryOp`, where `BinaryOp` is a member of the union `Expr` |

A cycle of plain struct or alias references is rejected:

```strata
struct A {
  b: B
}

struct B {
  a: A
}
```

```
Error: Cyclic Dependency: Cycle detected, struct 'A' cannot reference itself: 'A' -> 'B' -> 'A'.
```

Templates follow the same rule, so `children: [Node<T>]` inside `struct Node<T>` is accepted. [Model a tree](/docs/guides/model-data#model-a-tree) has a worked example, and [Break a cycle the compiler rejects](/docs/guides/model-data#break-a-cycle-the-compiler-rejects) shows the fix for this error.

## Packages

A package is a directory tree of `.strata` files that share one namespace. Any file can reference a type declared in any other file of the package, with no import statement:

```strata filename="address.strata"
struct Address {
  street: String
  city: String
}
```

```strata filename="user.strata"
struct User {
  name: String
  email: String
  address: Address
}
```

Type names are unique across the whole package. If you declare `Color` in both `theme.strata` and `paint.strata`, the compiler reports `A type named 'Color' already exists`, the same as for declaring it twice in one file.

A package's [manifest](/docs/reference/manifest#package-roots) defines which files belong to the package, including [nested packages](/docs/reference/manifest#nested-packages). [Why packages work this way](/docs/concepts/packages).

### Package names

A package's manifest sets the package's name in [`packageName`](/docs/reference/manifest#packagename). Each target cases it for its output, as [Output file names](/docs/reference/cli#output-file-names) shows. The `package` layout also uses it to name the project it writes, such as the Cargo crate or the npm package ([The package layout](/docs/reference/cli#the-package-layout)).

Package names are PascalCase by convention (`DesignTokens`), matching the type names they qualify.

### Qualified references

```strata filename="api/account.strata"
struct Account {
  id: String
  owner: Shared.User
}
```

A dependency's types are referenced as `Qualifier.Type`, where the qualifier is the key the package's manifest [declares the dependency](/docs/reference/manifest#dependencies) under.

- A qualified reference is accepted wherever a type name is, except as a map key: properties, collection elements, map values, parents, template names and arguments (`Shared.Box<Shared.User>`), and alias targets.
- A union member whose type is qualified must be [named](#named-members): `User: Shared.User`.
- A bare name always refers to a type of the current package. [Why no imports?](/docs/concepts/packages#dependencies-are-named-never-ambient)
- References have exactly two segments:

  ```
  Error: Invalid Type Reference: Invalid type reference 'A.B.C': only '<Package>.<Type>' references are supported.
  ```

An [alias](#type-aliases) of an external type is a type of the declaring package. So `alias Owner: Shared.User` is referenced as `Owner` locally and as `Api.Owner` by `Api`'s consumers, who don't need to declare `Shared` themselves.

### Visibility

A package can reference only the packages its own manifest declares. If `Api` depends on `Shared` and `Shared` depends on `Core`, then `Core` is compiled as part of building `Api`, but `Api` can't reference `Core.Identifier`:

```
Error: Unknown Package Qualifier: 'Core' is not a declared dependency of this package. Declared dependencies: 'Shared'.
```

### Standalone files

A `.strata` file with no `strata.config.json` in any ancestor directory has no package and can't declare dependencies. What it can see depends on how it's built, as [Input resolution](/docs/reference/cli#input-resolution) describes.
