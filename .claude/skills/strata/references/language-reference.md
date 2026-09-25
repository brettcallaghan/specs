<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: strata-lang/strata site/content/docs/lang.mdx
     Vendored from release: v0.8.0
     Regenerate with: scripts/sync-strata-docs.sh
     Upstream owns this text. Fix it there, then re-run that script. -->

import { Callout } from 'nextra/components'

# Language

## Structs
A `struct` is Strata's main building block: a named group of typed _properties_ (aka fields).

Each property has a name and a type. Available types are:
- [scalars](#scalars) — numbers, strings, bools, and dates
- collections — [maps](#maps) and [arrays](#arrays)
- other user-defined types — [enums](#enums), [unions](#unions), or other structs

```strata
"""
A post on a bulletin board.
"""
struct Post {
  """
  Unique identifier of the `Post`.
  """
  id: String

  """
  Unique identifier of the `User` who created the `Post`.
  """
  createdBy: String

  """
  Date and time at which the `Post` was created.
  """
  createdAt: DateTime

  """
  Date and time at which the `Post` was last updated.
  """
  updatedAt: DateTime

  """
  Title of the `Post` displayed in the UI.
  """
  title: String

  """
  The content of the `Post` written in Markdown.
  """
  body: String

  """
  An array of tags assigned to the `Post` by the author.
  """
  tags: [String]

  """
  An array of reactions left by readers.
  """
  reactions: [Reaction]
}
```

### Inheritance
Structs can inherit from one or more other structs. Common properties get defined once and reused.

Here `Post` and `Comment` both inherit from `Persisted` and `ReactionSubject`:
```strata
struct Persisted {
  id: String
  createdBy: String
  createdAt: UInt
  updatedAt: UInt
}

struct ReactionSubject {
  reactions: [Reaction]
}

struct Post: Persisted, ReactionSubject {
  title: String
  body: String
  tags: [String]
}

struct Comment: Persisted, ReactionSubject {
  subjectId: String
  message: String
}
```

> [!NOTE]
> 
> Many languages do not support inheritance or multiple inheritance.
> For these languages, the Strata transpiler will flatten properties into a single definition.
>
> When interfaces are supported by the target language, the transpiler may preserve the type hierarchy by
> defining interfaces for inherited types which merged definitions conform to.

## Scalars
Scalars are Strata's primitive types — single values like numbers, strings, bools, and dates.

> [!NOTE]
>
> `Int` and `UInt` transpile to 64-bit integers on most targets but are constrained to **53 bits of precision** so they round-trip safely through JavaScript, which stores all numbers as IEEE-754 doubles. For full 64-bit width, use `Int64` or `UInt64`.

| Type | Description |
| :--- | :---------- |
| `String` | A sequence of characters. Strings can contain letters, numbers, symbols, emojis, and whitespace. |
| `Bool` | `true` or `false`. |
| `Int` | Signed integer in the range −9,007,199,254,740,991 to 9,007,199,254,740,991 (53-bit JS-safe). |
| `Int8` | 8-bit signed integer (−128 to 127). |
| `Int16` | 16-bit signed integer (−32,768 to 32,767). |
| `Int32` | 32-bit signed integer (−2,147,483,648 to 2,147,483,647). |
| `Int64` | 64-bit signed integer (−9,223,372,036,854,775,808 to 9,223,372,036,854,775,807). |
| `UInt` | Unsigned integer in the range 0 to 9,007,199,254,740,991 (53-bit JS-safe). |
| `UInt8` | 8-bit unsigned integer (0 to 255). |
| `UInt16` | 16-bit unsigned integer (0 to 65,535). |
| `UInt32` | 32-bit unsigned integer (0 to 4,294,967,295). |
| `UInt64` | 64-bit unsigned integer (0 to 18,446,744,073,709,551,615). |
| `Float` | Floating-point number with platform-dependent precision (typically 32 or 64 bits). |
| `Float32` | 32-bit float, single precision (~7 decimal digits). |
| `Float64` | 64-bit float, double precision (~15–17 decimal digits). |
| `DateTime` | Date and time encoded as an [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) string: `YYYY-MM-DDTHH:MM:SS.sssZ`. |
| `Date` | Calendar date encoded as an [RFC 3339](https://datatracker.ietf.org/doc/html/rfc3339) string: `YYYY-MM-DD`. |

## Collections

### Arrays
An array is an ordered list of elements of the same type.

```strata
struct Post {
  tags: [String]
}
```

Elements can be any supported type, including nested arrays:
```strata
struct Transform {
  matrix: [[Float]]
}
```

### Maps
A map stores values keyed by another value.

- **Keys** must be `String` or an integer type (`Int`, `Int8`/`Int16`/`Int32`/`Int64`, `UInt`, `UInt8`/`UInt16`/`UInt32`/`UInt64`).
- **Values** can be any supported type.

```strata
struct BulletinBoard {
  postsById: [String: Post]
}
```

> [!NOTE]
>
> Treat map entries as **unordered**. Most target languages don't guarantee insertion order on iteration.

## Optionals
Every property is **required by default**. Mark a type with `?` to make the property optional — it may be absent or `null`.

`firstName` and `lastName` are required; `middleName` is optional:
```strata
struct Person {
  firstName: String
  middleName: String?
  lastName: String
}
```

Collections may also contain optional values.
```strata
struct OptionalElements {
  arrayOfOptionals: [Person?]
  mapOfOptionals: [String: Person?]
}
```

> [!NOTE]
>
> Strata has no distinct `undefined` state — most target languages don't model one, so the absent and `null` cases are unified.

### Where optionality may be declared

A `?` goes at a *use* site — a property, an array element, a map value, a template argument. It may not be baked into a type's name, and it may not be applied twice to the same type. Two rules follow from that, and both are enforced by the compiler.

**An alias is never optional.** `alias OptionalColor: RgbColor?` is rejected. Optionality belongs where the type is used, so a property's declaration always tells you whether it may be absent — you should never have to open another file to find out. Write the alias without the `?` and mark each use:

```strata
alias RgbColor: String ~= /^#[0-9A-Fa-f]{6}$/

struct Layer {
  backgroundColor: RgbColor
  foregroundColor: RgbColor?
}
```

An alias whose type merely *contains* an optional below the top level is fine — `alias Names: [String?]` and `alias Lookup: [String: RgbColor?]` both resolve to a collection, and only the element and value types are optional.

**An optional of an optional is rejected.** `String??` has never parsed. A template argument can reach the same shape less obviously, and that is rejected too: given `struct Box<T> { value: T? }`, the instantiation `Box<String?>` would substitute `String?` into a position that already carries a `?`, so the compiler refuses it rather than quietly reducing it to one level. Drop the `?` from the argument, or drop it from the template body and write it at each use:

```strata
struct Box<T> {
  value: T?
}

struct Root {
  b: Box<String>     # value is String?
}
```

The rule is about the *parameter's position in the template body*, not about the argument being optional. An optional argument landing anywhere else stays legal, because those shapes are representable on the wire:

```strata
struct Paginated<T> {
  items: [T]
}

struct Root {
  page: Paginated<User?>   # items is [User?] — fine
}
```

`[]`, `[null]` and `[{...}]` are three distinct payloads, so an optional inside an array carries real information. A doubled optional does not: JSON has one representation for "no value", so the second `?` would describe a state no payload could ever express and no target could round-trip.

> [!NOTE]
>
> **Converting from GraphQL or OpenAPI?** Both let you name a nullable type — a GraphQL nullable field type, an OpenAPI schema with `nullable: true` — and Strata does not, so those types have no one-to-one alias. Push the optionality down instead: name the base type with an alias, then mark each property that uses it with `?`. A type that is nullable in *every* position becomes an alias plus a `?` at each of its use sites, which is more edits but leaves each declaration self-describing.

## Enums
An `enum` is a scalar type with a fixed set of named values — use it when a property's value must come from a known list.

```strata
enum Fruit {
  MANGO
  BANANA
  PINEAPPLE
}
```

Each entry takes its name as its value, so the shorthand above is equivalent to:

```strata
enum Fruit: String {
  MANGO: "MANGO"
  BANANA: "BANANA"
  PINEAPPLE: "PINEAPPLE"
}
```

The explicit form is useful when you want the on-the-wire values to differ from the entry names:

```strata
enum Fruit {
  MANGO: "🥭"
  BANANA: "🍌"
  PINEAPPLE: "🍍"
}
```

The default value type is `String`, but `Bool` and any integer type work too. When values are given without an explicit type annotation, the enum's type is inferred from the first entry. All values must be the same type and unique.

```strata
enum Fruit {
  MANGO: 1
  BANANA: 2
  PINEAPPLE: 3
}
```

Specify the type explicitly only when the default inferred type isn't what you want — for example, to pick a specific integer width:

```strata
enum Fruit: UInt8 {
  MANGO: 1
  BANANA: 2
  PINEAPPLE: 3
}
```

## Unions
A `union` is a type whose value can take one of several different struct shapes. Useful for polymorphic data — "this is either an X or a Y". Every member of a union must be a `struct`.

```strata
struct Circle {
  radius: Float
}

struct Rectangle {
  width: Float
  height: Float
}

union Shape {
  Circle
  Rectangle
}
```

Strata unions are **tagged** (a.k.a. discriminated) — each member carries a unique tag value identifying which struct shape is present. By default the tag lives in a property called `__type` and its value is the member's name:

```json filename="Circle.json"
{
  "__type": "Circle",
  "radius": 3.14
}
```

For the `Shape` union above, `__type` is `"Circle"` or `"Rectangle"`.

Override the tag value per member with `@tag(...)`. String, integer, and boolean tags are supported:

```strata
union Shape {
  @tag(1) Circle
  @tag(2) Rectangle
}
```

### Custom discriminators
Override the discriminator property name with `@discriminator`:

```strata
@discriminator(shape)
union Shape {
  Circle
  Rectangle
}
```

A `Circle` now serializes as:
```json filename="Circle.json"
{
  "shape": "Circle",
  "radius": 3.14
}
```

## Documentation
Documentation is a first-class feature. Wrap a Markdown block in triple quotes (`"""`) above any type definition, property, union variant, or enum entry to document it.

```strata
"""
A circle is a geometric shape consisting of all points in a plane that are 
equidistant from a fixed center point.
"""
struct Circle {
  """
  The distance from the center of a circle to any point on its circumference.
  """
  radius: Float
}
```

## Constraints
Constraints narrow the set of legal values for a type.

### Regex constraints
Constrain a `String` with `~=` followed by a regular expression between `/` delimiters:
```strata
struct Fruit {
  """
  Color of the fruit provided as a hex RGB string.
  """
  color: String ~= /^#[0-9A-Fa-f]{6}$/
}
```

#### Portable regex syntax
Strata stores the pattern verbatim and delegates execution to each target's regex engine. The engines differ in what they support:

| Target | Engine | Lookaround | Backreferences |
| :--- | :--- | :---: | :---: |
| JSON Schema | ECMA 262 | ✅ | ✅ |
| TypeScript | ECMA 262 | ✅ | ✅ |
| Swift | `NSRegularExpression` (ICU) | ✅ | ✅ |
| Kotlin | `java.util.regex` | ✅ | ✅ |
| Rust | [`regex`](https://docs.rs/regex/) | ❌ | ❌ |
| Go | RE2 | ❌ | ❌ |

To keep validation consistent across all targets, stick to the portable subset supported everywhere:

- **Safe to use:** character classes (`[...]`), quantifiers (`*`, `+`, `?`, `{n,m}`), alternation (`|`), groups (`(...)`), non-capturing groups (`(?:...)`), anchors (`^`, `$`), standard escapes (`\d`, `\w`, `\s`, `\b`).
- **Avoid:** lookaround (`(?=...)`, `(?!...)`, `(?<=...)`, `(?<!...)`), backreferences (`\1`), named captures, possessive quantifiers, inline flags (`(?i)`). These compile fine for JSON Schema, TypeScript, Swift, and Kotlin but will fail at the Rust and Go runtime layers.

> [!NOTE]
>
> Strata does not validate regex syntax at parse time. Invalid or non-portable patterns surface at the target's compile or runtime layer, not when you run `strata build`.

### Numeric constraints
Numeric constraints apply to integer and floating-point types.

#### Comparisons
| Operator | Meaning |
| :--- | :---------- |
| `>` | Greater than |
| `>=` or `≥` | Greater than or equal to |
| `<` | Less than |
| `<=` or `≤` | Less than or equal to |

```strata
struct ComparisonOperatorDemo {
  greaterThan: Int > 42
  greaterThanOrEqualTo: Int >= 42
  greaterThanOrEqualToUnicode: Int ≥ 42
  lessThan: Int < 42
  lessThanOrEqualTo: Int <= 42
  lessThanOrEqualToUnicode: Int ≤ 42
}
```

#### Ranges
Range operators constrain a numeric type to values between a lower and upper bound.

| Operator | Meaning |
| :--- | :---------- |
| `..` | **Closed** — both bounds included. `x in l .. r` ≡ `x ≥ l && x ≤ r` |
| `>.` | **Left-open** — only the upper bound included. `x in l >. r` ≡ `x > l && x ≤ r` |
| `.<` | **Right-open** — only the lower bound included. `x in l .< r` ≡ `x ≥ l && x < r` |
| `><` | **Open** — neither bound included. `x in l >< r` ≡ `x > l && x < r` |

```strata
struct RangeOperatorDemo {
  """
  Integers between `1` and `10`, including `1` and `10`
  """
  closed: Int in 1 .. 10

  """
  Integers greater than `1` but less than or equal to `10`
  """
  leftOpen: Int in 1 >. 10

  """
  Integers greater than or equal to `1` but less than `10`
  """
  rightOpen: Int in 1 .< 10

  """
  Integers greater than `1` and less than `10`
  """
  open: Int in 1 >< 10
}
```

### Equality constraints
Pin a `String`, `Bool`, or integer type to a single value with `==`. The property becomes a constant.

```strata
struct Demo {
  code: Int == 42
}
```

Most useful in polymorphic scenarios where a struct variant is identified by a constant field:

```strata
struct Fruit {
  name: String
  color: String
  isJuicy: Bool
}

struct JuicyFruit: Fruit {
  """
  Juicy fruits are always juicy.
  """
  isJuicy: Bool == true

  stainsShirts: Bool
}
```

### Length constraints
Constrain the size of a `String` (characters), array (elements), or map (entries) with the `length` keyword followed by a comparison, range, or equality.

#### String length

```strata
struct User {
  """
  Username must be between 3 and 20 characters.
  """
  username: String length in 3 .. 20

  """
  Bio must be at most 500 characters.
  """
  bio: String length <= 500
}
```

#### Array length

```strata
struct Team {
  """
  A team must have at least 1 member.
  """
  members: [String] length >= 1

  """
  Tags must have between 1 and 10 elements.
  """
  tags: [String] length in 1 .. 10
}
```

#### Map length

```strata
struct Config {
  """
  Configuration must have at least 1 property.
  """
  settings: [String: String] length >= 1

  """
  Metadata can have at most 20 properties.
  """
  metadata: [String: Int] length <= 20
}
```

## Type aliases
A type alias gives a name to a type expression. Most useful for naming constrained types you reuse — regex-matched strings, formatted identifiers, and so on.

An alias names a type; it does not decide whether a property may be absent. `alias OptionalColor: RgbColor?` is a compile error — see [Where optionality may be declared](#where-optionality-may-be-declared) for the rewrite.

**Colors**

```strata
"""
An RGB hexadecimal color string (e.g. `"#f38811"`)
"""
alias RgbColor: String ~= /^#[0-9A-Fa-f]{6}$/

struct Layer {
  backgroundColor: RgbColor
  foregroundColor: RgbColor?
}
```

**Form Validation**
```strata
alias Email: String ~= /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/
alias Username: String ~= /^[a-zA-Z0-9_]{3,20}$/

struct UserRegistration {
  username: Username
  email: Email
}
```

**Identifiers**
```strata
alias UUID: String ~= /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/
alias Slug: String ~= /^[a-z0-9]+(?:-[a-z0-9]+)*$/

struct Article {
  id: UUID
  slug: Slug
  title: String
}
```

## Templates
Templates are Strata's generics — reusable structures parameterized by one or more types.

### Defining templates
Add type parameters in angle brackets after the name:

```strata
"""
A container that can hold any type of value.
"""
struct Value<T> {
  value: T
}
```

`T` is a placeholder that gets bound to a concrete type each time the template is instantiated. Multiple parameters are comma-separated:

```strata
"""
A pair holding two values of potentially different types.
"""
struct Pair<T, U> {
  first: T
  second: U
}
```

### Using templates
A template instantiation works anywhere a regular struct type does — properties, collections, optionals, unions, nested inside other templates:

```strata
struct Demo {
  basic: Value<String>
  items: [Value<String>]
  lookup: [String: Value<Int>]
  optional: Value<String>?
  nested: Value<Pair<String, Int>>
}
```

An optional type argument is allowed — `Value<String?>` gives `value: String?`, exactly as if you had written the `?` inline. The one instantiation Strata rejects is the one that would apply a second `?` to the same type, which happens when the template body already marks that parameter optional. See [Where optionality may be declared](#where-optionality-may-be-declared).

Templates can show up as union variants and can themselves be inherited:

```strata
union ScalarValue {
  Value<String>: "string"
  Value<Int>: "int"
  Value<Float>: "float"
  Value<Bool>: "bool"
}

"""
Adds a label to a contained value.
"""
struct LabeledValue<T>: Container<T> {
  label: String
}
```

Unions may also be templated:

```strata
struct Error {
  code: Int
  message: String
}

union Result<T> {
  """Success."""
  Ok: T

  """Failure."""
  Err: Error
}
```

## Recursive type references
Strata lets types reference themselves — useful for trees, linked lists, ASTs, and other self-referential data.

A type can reference itself **indirectly** through:
- **Arrays** — `children: [TreeNode]`
- **Optionals** — `next: LinkedListNode?`
- **Maps** — `children: [String: FileNode]`
- **Unions** — when a struct references a union that lists it as a variant

Each form provides indirection so the type still has a finite, known size in memory.

**Tree Structures (Array Recursion)**

Each node contains zero or more children of the same type.

```strata
"""
A tree node that can have any number of children.
"""
struct TreeNode {
  """The value stored in this node."""
  value: String
  """Child nodes forming the tree structure."""
  children: [TreeNode]
}
```

**Linked Lists (Optional Recursion)**

Each node optionally points to a successor.

```strata
"""
A singly-linked list node.
"""
struct LinkedListNode {
  """The value stored in this node."""
  value: Int
  """Reference to the next node, if any."""
  next: LinkedListNode?
}
```

**File Systems (Map Recursion)**

Children are indexed by name.

```strata
"""
A file system node representing either a file or directory.
"""
struct FileNode {
  """The file or directory name."""
  name: String
  """Child nodes indexed by name."""
  children: [String: FileNode]
}
```

**Abstract Syntax Trees (Union Recursion)**

A union variant can reference the parent union.

```strata
"""
A binary operation with two operands.
"""
struct BinaryOp {
  left: Expr
  right: Expr
  operator: String
}

"""
A literal value.
"""
struct Literal {
  value: String
}

"""
An expression that can be either a literal or a binary operation.
"""
union Expr {
  Literal
  BinaryOp
}
```

**Mutual Recursion**

Two or more types can reference each other in a cycle, as long as every step in the cycle goes through an indirection (array, optional, map, or union).

```strata
"""
A person who can own pets.
"""
struct Person {
  name: String
  pets: [Pet]
}

"""
A pet that may have an owner.
"""
struct Pet {
  name: String
  owner: Person?
}
```

### Generic recursive types
Templates support the same recursive patterns as concrete types.

```strata
"""
A generic tree node that can hold any type of value.
"""
struct Node<T> {
  """The value stored in this node."""
  value: T
  """Child nodes."""
  children: [Node<T>]
}

"""
A generic linked list.
"""
struct LinkedList<T> {
  value: T
  next: LinkedList<T>?
}
```

### When recursion is forbidden
A type cannot directly contain itself — that would describe an infinite-size value that can't be laid out in memory. Every recursive reference must pass through an array, optional, map, or union.

```strata
"""This is NOT allowed - creates an infinite-size type."""
struct BadNode {
  value: String
  self: BadNode
}

"""Mutual direct recursion is also forbidden."""
struct A {
  b: B
}

struct B {
  a: A
}
```
## Packages

A package is a directory tree of `.strata` files that share one namespace. Any member file can reference types defined in any other member file — no import statements, no include lists:

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

Because the namespace is shared, type names must be unique across the whole package — defining `Color` in both `theme.strata` and `paint.strata` is a duplicate-type error, the same as defining it twice in one file.

### The manifest

A directory becomes a package root by containing a file named exactly `strata.config.json`. The manifest holds the package's compiler options, and every field is optional — `{}` is a perfectly good manifest that says "everything under here is one package, defaults for everything":

```json filename="strata.config.json"
{
  "packageName": "DesignTokens"
}
```

The manifest is validated against a JSON Schema generated from Strata's own options schema, so mistakes are caught at the source rather than surfacing as confusing build behavior. A failed validation names the manifest path, the offending property, and the violated constraint:

```
Error: invalid manifest /repo/design-tokens/strata.config.json:
  /packageName: 42 is not of types "null", "string"
```

The available options are the same ones documented on each [transpiler page](./transpilers) — the manifest is where they live when they describe the package itself rather than a one-off build. See the [CLI page](./cli#packages-and-manifest-discovery) for how `--config` layers over the manifest.

### Package roots and the nearest-ancestor rule

To find the package a file belongs to, walk up from the file's own directory: the first directory containing a `strata.config.json` is the package root — the same rule `git` uses to locate its repository root.

```
repo/
  strata.config.json      ← package root for everything below…
  common.strata
  models/
    user.strata           ← belongs to repo/ (nearest manifest is repo's)
```

Membership follows from the root: every `.strata` file under the package root is a member, however deeply nested.

### Nested packages

A subdirectory with its own manifest is its own package, and its subtree is carved out of the outer package:

```
repo/
  strata.config.json      ← outer package
  common.strata           ← member of repo/
  examples/
    strata.config.json    ← inner package root
    demo.strata           ← member of examples/, NOT of repo/
```

The two packages are fully isolated by default: `demo.strata` cannot reference types from `common.strata`, and vice versa. Carving is triggered by the nested manifest's presence, not its validity — a broken `examples/strata.config.json` still keeps `demo.strata` out of the outer package. Being nested buys no special access in either direction; if the outer package wants the inner one's types, it declares it as a [dependency](#dependencies) exactly as it would a package anywhere else on disk.

### Package names

The package name is the manifest's `packageName`; when it's omitted, the name derives from the package root's directory name. The name feeds each transpiler's package mode — Cargo crate names, npm package names, and so on.

PascalCase is the package naming convention: `DesignTokens`, not `design-tokens` or `design_tokens`. Package names appear in source positions (`DesignTokens.Color`), so they follow the same style as type names. Transpilers apply per-target casing automatically — `DesignTokens` becomes `design-tokens` as an npm package name.

### Dependencies

A package reaches types in another package by declaring it as a dependency. Declarations live in the manifest, keyed by the **qualifier** those types will be referenced through:

```json filename="api/strata.config.json"
{
  "packageName": "Api",
  "dependencies": {
    "Shared": { "path": "../shared" }
  }
}
```

`path` points at a directory that is itself a package root — it must contain a `strata.config.json`. Relative paths resolve against the directory holding the declaring manifest; absolute paths are used as written. That is the whole mechanism: there is no registry, no version range, and no lockfile. A dependency is a path on disk, which is what makes a monorepo of sibling packages work with no infrastructure at all.

#### Qualified references

Once declared, the dependency's types are written `Qualifier.Type`:

```strata filename="shared/types.strata"
"""A user shared across packages."""
struct User {
  name: String
  email: String
  age: Int32?
}
```

```strata filename="api/api.strata"
"""An account owned by a shared user."""
struct Account {
  id: String
  owner: Shared.User
}
```

A qualified reference is accepted everywhere a type name is: property types, supertypes, union members, template heads and arguments, map values, and alias targets. Map *keys* are the one exception, and not because of qualification — they are restricted to `String`, `Bool`, and integer types, so a named type is rejected there whether it is qualified or local.

**A bare name always means a type of the package you are writing in.** No exceptions, no import form that lets you drop the qualifier. That is a deliberate trade: you type six more characters, and in exchange a name never changes meaning because of something a manifest did. Two packages in one build can both declare `User` without shadowing, and adding a type to a dependency can never capture a name you were already using.

References are exactly two segments. `A.B.C` is rejected with a message that says so rather than a generic parse error:

```
Invalid type reference 'A.B.C': only '<Package>.<Type>' references are supported.
```

#### Aliasing an external type

When a qualifier gets repetitive, or when you want a name of your own for someone else's type, alias it:

```strata filename="api/api.strata"
"""The party responsible for an account."""
alias Owner: Shared.User

struct Account {
  id: String
  owner: Owner
}
```

The alias is a type of *your* package, so you reference it bare — and so can your own consumers, as `Api.Owner`. The underlying `Shared.User` travels with it into their build without them declaring `Shared` themselves.

#### Only direct dependencies are visible

You can name the packages your own manifest declares, and nothing else. If `Api` depends on `Shared` and `Shared` depends on `Core`, then `Api` cannot write `Core.Identifier` — even though `Core` is resolved and compiled as part of building `Api`:

```
'Core' is not a declared dependency of this package. Declared dependencies: 'Shared'.
```

Declare `Core` in `Api`'s manifest if `Api` wants it. The rule keeps a package's visible surface equal to what it asked for: a dependency can add dependencies of its own without silently widening what its consumers can reach.

#### Graph rules

Two constraints apply to the closure a build resolves, and both fail the build with the paths involved:

```
Error: dependency cycle: /repo/a -> /repo/b -> /repo/a
Error: two packages in one dependency closure are both named 'Shared': /repo/vendor/shared and /repo/lib/shared
```

Package names are the namespace bundled output is organized by, so two packages sharing a name in one closure is ambiguous rather than merely confusing.

#### Qualifier naming

Qualifiers are PascalCase by convention, matching package names — `Shared`, `DesignTokens`. A qualifier must be spelled as a Strata type identifier (it appears in source, after all), must not collide with a type declared in the same package, and must not equal the declaring package's own name:

```
Error: /repo/api/strata.config.json declares dependency 'Api' under its own package name;
qualified references through it would read as self-references
```

The qualifier is the consumer's choice, not the dependency's property: nothing stops you from declaring the package at `../shared` under the qualifier `Common` if that reads better in your source. See the [CLI page](./cli#dependency-builds) for how a dependency closure is built and emitted.

### Standalone files

A `.strata` file with no manifest in any ancestor directory is an implicit single-file package: it compiles alone, with defaults, exactly as it did before packages existed. Sibling files in the same directory are **not** merged in — without a manifest, each file is its own island. Drop a `strata.config.json` next to them when you want them to share a namespace.
