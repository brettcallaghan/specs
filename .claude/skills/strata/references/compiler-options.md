<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: strata-lang/strata strata/options.strata
     Vendored from release: v0.9.0
     Regenerate with: scripts/sync-strata-docs.sh
     Upstream owns this text. Fix it there, then re-run that script. -->

# Compiler options — the schema of `strata.config.json`

Verbatim `options.strata` from strata-lang/strata. Every field states its
default in its doc block.

```strata
struct CompilerOptions {
    """
    Semantic package name shared by every target.

    Each transpiler applies its own casing conventions:
    - Rust: kebab-case (for example, "my-models")
    - Swift: PascalCase (for example, "MyModels")
    - Kotlin: kebab-case for the project name. The fully qualified package is derived
      from `kotlin.packageGroupId` and the lowercased name.
    - Go: a valid Go package identifier, lowercased with hyphens and underscores
      removed (for example, "mymodels")
    - TypeScript: kebab-case (for example, "my-models")

    When unset, the package name is derived from the input path.

    Defaults to `null`.
    """
    packageName: String?

    """
    Package version shared by every target.

    Distinct from the `version` inside a target's block (`go.version`,
    `kotlin.version`, and so on), which names that target's language, toolchain, or
    draft version.

    Each transpiler writes this version into its native manifest:
    - Rust: `version` in `Cargo.toml`
    - Kotlin: `version` in `build.gradle.kts`
    - TypeScript: `version` in `package.json`
    - Swift: not written (reserved for `Package.swift`)
    - Go: not applicable, because `go.mod` has no version field

    Defaults to `"0.0.0"`.
    """
    version: String?

    """
    Packages this package depends on, keyed by the qualifier through which their types
    are referenced.

    A qualified reference names its source package by qualifier. With
    `"dependencies": {"Shared": {"path": "../shared"}}`, the type `User` declared in the
    package at `../shared` is referenced as `Shared.User`. A bare name always means a
    type of the declaring package, so every cross-package reference needs a qualifier.

    Qualifiers are conventionally PascalCase, matching package names. The key must be a
    valid Strata type identifier, must not collide with a type declared in this package,
    and must not equal this package's own name.

    Dependencies are declared only in a package manifest (`strata.config.json`), never
    in a `--config` build overlay.

    Defaults to `null` (no dependencies).
    """
    dependencies: [String: DependencySpec]?

    """
    When `true`, `null` properties are omitted from serialized output instead of being
    written as explicit `null`s.

    Benefits:
    - Smaller payloads, because optional fields that aren't set are omitted
    - The same output whether an optional property is unset or explicitly `null`, which
      helps hashing and diffing workflows

    Defaults to `true`.
    """
    skipNullSerialization: Bool?

    """
    When `true`, every artifact generated from this package rejects keys it doesn't
    declare.

    The policy is declared once, and every target honors it identically. The JSON Schema
    target closes its object schemas, and each code-generation target closes its decoder
    in its own idiom. No target-scoped option restates or overrides it. To vary the
    policy per target, vary this option in a `--config` build overlay, because each
    target is a separate build invocation.

    Maps stay open under either policy, because a map's keys are data, not schema. The
    `OpenAi` JSON Schema version emits closed object schemas regardless of this option.
    That is a requirement of the OpenAI API, not a property of the policy.

    Defaults to `false`, which selects the open policy: undeclared keys are ignored.
    """
    rejectUnknownProperties: Bool?

    """
    When `true`, generated decoders accept an enum value the schema doesn't declare.

    The undeclared raw value decodes into a fallback that carries it and re-encodes it
    unchanged, so a value added after the code was generated doesn't fail
    deserialization. This helps clients that need forward compatibility with newer
    servers, especially native apps, where shipping an update is slow.

    Doesn't apply to the JSON Schema, AST, and Markdown targets. A JSON Schema
    enumerates exactly the declared values.

    Defaults to `null` (`false`): an undeclared enum value fails deserialization.
    """
    allowUnknownEnumValues: Bool?

    """
    When `true`, generated decoders accept a union variant the schema doesn't declare.

    An object whose tag names no declared variant decodes into a fallback that carries
    the tag, so a variant added after the code was generated doesn't fail
    deserialization.

    Doesn't apply to the JSON Schema, AST, and Markdown targets. A JSON Schema accepts
    only the declared tags. The Go target doesn't implement this option.

    Defaults to `null` (`false`): an undeclared tag fails deserialization.
    """
    allowUnknownUnionVariants: Bool?

    """
    When `true`, each struct gets a readonly abstraction that the struct and every
    subtype implement, and that inherits the abstractions of the struct's supertypes.
    The abstraction is a trait in Rust, a protocol in Swift, and an interface in Kotlin
    and Go. In Go, the interface declares a `GetX()` accessor for each property.

    Always in effect for TypeScript. Every generated TypeScript type is an interface that
    declares its supertypes with `extends`, so `false` changes nothing there.

    Doesn't apply to the JSON Schema, AST, and Markdown targets, which emit no code type
    to abstract over.

    Defaults to `true`.
    """
    generateInterfaces: Bool?

    """
    Narrows the output of every target to one type and the types reachable from it.

    Only the named type and the types it refers to, directly or transitively, are
    emitted. A type refers to another through property types, supertypes, union
    variants, alias targets, and template arguments. The JSON Schema target also adds a
    top-level `$ref` to the named type, which suits focused per-endpoint schemas.
    `strata build --root <type>` sets it for one build.

    A name that matches no type fails the build. A qualified name (`Shared.User`) names
    a type of a declared dependency.

    Defaults to `null`: every type is emitted.
    """
    root: String?

    """
    How a build lays out its output:

    - `mirror`: one source artifact per source file, keeping the source directory
      structure.
    - `package`: the `mirror` files inside the target's project scaffold (manifest and
      build files).
    - `single-file`: one combined source artifact that holds every definition.

    A document target (AST, JSON Schema, and Markdown) emits no project scaffold, so it
    builds `package` as `mirror`. For JSON Schema, a `single-file` document has no
    external `$ref`: every reference is local.

    Defaults to `mirror`, or to `single-file` when `root` is set.
    """
    layout: Layout?

    """ Options for the Rust transpiler. """
    rust: RustOptions?

    """ Options for the JSON Schema transpiler. """
    jsonSchema: JSONSchemaOptions?

    """ Options for the Swift transpiler. """
    swift: SwiftOptions?

    """ Options for the Kotlin transpiler. """
    kotlin: KotlinOptions?

    """ Options for the TypeScript transpiler. """
    typescript: TypeScriptOptions?

    """ Options for the Go transpiler. """
    go: GoOptions?
}

"""
Locates a package that another package depends on.

Only local path dependencies are supported. The shape leaves room for more location
fields, such as registries and git sources, without changing how existing manifests
are read.
"""
struct DependencySpec {
    """
    Path to the dependency's package root: the directory that directly contains its
    `strata.config.json`.

    A relative path resolves against the directory that contains the declaring manifest
    (`"../shared"` in `repo/api/strata.config.json` resolves to `repo/shared/`). An
    absolute path is used as is. Nested, sibling, and unrelated locations all behave
    identically, because containment grants no extra visibility.
    """
    path: String
}

struct RustOptions {
    """
    Indentation string for generated code.

    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    Rust edition written to the generated `Cargo.toml`. Read only in the `package` layout.

    Defaults to `"2021"`.
    """
    edition: String?
}

struct SwiftOptions {
    """
    Indentation string for generated code.

    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    When `true`, generates properties with `var` instead of `let`, which makes them
    mutable.

    By default, properties are immutable (`let`), which follows Swift best practices
    for value types and makes code safer and more predictable. Enable this option when
    you need mutable properties for incremental updates or for frameworks that require
    mutability.

    Defaults to `false` (generates `let`).
    """
    mutableProperties: Bool?

    """
    Swift tools version written as the `swift-tools-version` of the generated
    `Package.swift`. Read only in the `package` layout.

    This is the Swift version, not the package's own version, which is the top-level
    `version`.

    Defaults to `"5.9"`.
    """
    version: String?

    """
    Casing convention for generated directory names. Read only in the `mirror` and
    `package` layouts.

    Defaults to `Pascal` (PascalCase).
    """
    directoryCasing: CasingStyle?
}

struct KotlinOptions {
    """
    Indentation string for generated code.

    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    When `true`, generates properties with `var` instead of `val`, which makes them
    mutable.

    By default, properties are immutable (`val`), which follows Kotlin best practices
    for data classes. Enable this option when you need mutable properties.

    Defaults to `false` (generates `val`).
    """
    mutableProperties: Bool?

    """
    Version of `kotlinx-serialization-json` the generated code is written for.

    The generated source targets this version in every layout, and the `package` layout
    also pins it in the `build.gradle.kts` dependency block.

    Defaults to `"1.6.2"`.
    """
    serializationVersion: String?

    """
    Maven group ID (for example, "com.example"). Required in the `package` layout, where a
    build without it fails with an error naming `kotlin.packageGroupId`.

    It prefixes every generated package declaration and, in the `package` layout,
    determines the directory structure. A `single-file` build that sets `packageName`
    derives its package declaration from `packageGroupId` and `packageName`. When
    `packageGroupId` is unset, that build falls back to `com.example`.

    Defaults to `null`.
    """
    packageGroupId: String?

    """
    Kotlin version written to the generated `build.gradle.kts`. Read only in
    the `package` layout.

    This is the Kotlin version, not the package's own version, which is the top-level
    `version`.

    Defaults to `"1.9.25"`.
    """
    version: String?

    """
    JVM toolchain version written to the generated `build.gradle.kts`. Read only in
    the `package` layout. Ignored when `projectType` is `Multiplatform`, where the JVM
    target carries its own toolchain.

    Defaults to `"17"`.
    """
    jvmToolchain: String?

    """
    Shape of the generated Gradle project. Read only in the `package` layout.

    `Jvm` emits a `kotlin("jvm")` project that uses `jvmToolchain`, with sources in
    `src/main/kotlin`. `Multiplatform` emits a `kotlin("multiplatform")` project that
    declares the targets in `kmpTargets`, with sources in `src/commonMain/kotlin`.

    This setting affects only the build file and the source-set layout. The generated
    Kotlin source is identical under either shape, and so is the wire format.

    Changing it after the first package build moves the sources but leaves
    `build.gradle.kts` as it is, because that file is written once and then belongs to
    the developer. Add the new source set to the existing script, or delete the script
    and let the next build scaffold one that matches. Gradle doesn't treat a source set
    with no files as an error, so the mismatch compiles nothing instead of failing.

    Defaults to `Jvm`.
    """
    projectType: KotlinProjectType?

    """
    Kotlin Multiplatform targets to declare. Ignored unless `projectType` is
    `Multiplatform`.

    A bare target name, such as `"jvm"`, `"linuxX64"`, or `"macosArm64"`, is written as
    a call to it. Any other entry is written into the `kotlin { }` block verbatim, so a
    target that needs configuration is spelled out in full (for example,
    `"js(IR) { nodejs() }"`).

    Defaults to `["jvm"]`.
    """
    kmpTargets: [String]?

    """
    Casing convention for generated file names. Read only in the `mirror` and `package`
    layouts.

    Defaults to `Pascal` (PascalCase).
    """
    fileCasing: CasingStyle?
}

"""
Shape of the Gradle project the Kotlin target generates in the `package` layout.
"""
enum KotlinProjectType {
    """ A `kotlin("jvm")` project, with sources in `src/main/kotlin`. """
    Jvm

    """ A `kotlin("multiplatform")` project, with sources in `src/commonMain/kotlin`. """
    Multiplatform
}

struct TypeScriptOptions {
    """
    Indentation string for generated code.

    Defaults to two spaces: "  ".
    """
    indentation: String?

    """
    When `true`, generates properties without `readonly`.

    Defaults to `false` (generates `readonly`).
    """
    mutableProperties: Bool?

    """
    When `true`, generates the serde code (property tables, `deserialize` and
    `serialize` functions, and the constraint checks they run) alongside the type
    definitions.

    When `false`, only the type definitions are emitted, and the generated module
    doesn't depend on `@strata-lang/serde`. Useful when the data source is trusted
    (for example, compiler-generated AST) and the validation overhead is unnecessary.

    Serialization can't be kept on its own. Both directions run off one property
    table, and that table names the `deserialize` function of every type it
    references, so a module that kept its serializers would still pull in the whole
    runtime.

    Defaults to `true`.
    """
    generateValidation: Bool?

    """
    TypeScript version written to the generated `package.json` `devDependencies`.
    Read only in the `package` layout.

    This is the TypeScript version, not the package's own version, which is the
    top-level `version`.

    Defaults to `"^5.4"`.
    """
    version: String?

    """
    ECMAScript target written to `compilerOptions.target` in the generated
    `tsconfig.json`. Read only in the `package` layout.

    Named for what it selects, so it can't be confused with the build target
    (`--target`), which selects the output format.

    Defaults to `"ES2020"`.
    """
    ecmaTarget: String?

    """
    Casing convention for generated file names. Read only in the `mirror` and `package`
    layouts.

    Defaults to `Pascal` (PascalCase).
    """
    fileCasing: CasingStyle?
}

struct GoOptions {
    """
    Indentation string for generated code.

    Go convention is tabs, which `gofmt` enforces. Defaults to a single tab.
    """
    indentation: String?

    """
    Go module path written to `go.mod` (for example, "github.com/example/my-models").
    Required in the `package` layout, where a build without it fails with an error
    naming `go.modulePath`.

    Defaults to `null`.
    """
    modulePath: String?

    """
    Go language version written to the `go` directive of `go.mod`. Read only in
    the `package` layout.

    This is the Go version, not the package's own version, which is the top-level
    `version`.

    Defaults to `"1.24"`.
    """
    version: String?
}

"""
Options for the JSON Schema transpiler.
"""
struct JSONSchemaOptions {
    """
    When `true`, excludes `description` fields from the generated JSON Schema. This
    reduces schema size by 10-30% for production builds that don't need documentation.

    Defaults to `false` (documentation is included).
    """
    excludeDoc: Bool?

    """
    JSON Schema draft to target.

    - `Draft04`: JSON Schema Draft 4 (older tooling, widest compatibility)
    - `Draft07`: JSON Schema Draft-07 (the default, widely supported)
    - `Draft201909`: JSON Schema Draft 2019-09 (adds `unevaluatedProperties` and
      recursive refs)
    - `Draft202012`: JSON Schema Draft 2020-12 (modern features)
    - `OpenAi`: an unofficial draft that targets the requirements of OpenAI's
      Structured Outputs API

    Defaults to `Draft07`.
    """
    version: JsonSchemaDraft?
}

"""
JSON Schema draft to target.

JSON Schema calls each iteration of its specification a "draft", not a "version".
"""
enum JsonSchemaDraft {
    """ JSON Schema Draft 4 (widely supported, older tooling) """
    Draft04

    """ JSON Schema Draft-07 (widely supported) """
    Draft07

    """ JSON Schema Draft 2019-09 (adds `unevaluatedProperties` and other advanced features) """
    Draft201909

    """ JSON Schema Draft 2020-12 (modern standard) """
    Draft202012

    """
    Unofficial draft that targets the requirements of OpenAI's Structured Outputs API.

    This draft enforces the requirements of OpenAI's structured model outputs:
    - Adds `additionalProperties: false` to every object type
    - Makes every property required (optional ones use
      `anyOf: [{"type": "null"}, <schema>]`)
    - Represents discriminated unions with `anyOf` instead of `if`/`then`/`allOf`
    - Flattens inherited struct properties into the child struct schema
    - Omits the `$schema` header
    - Uses `$defs` for shared type definitions
    - Makes every `$ref` local to its document (`layout` still decides whether the build
      writes one document or one per source file)

    Some Strata features produce output that OpenAI can't enforce at generation time,
    such as numeric range constraints, string length and pattern constraints, and date
    formats. The schema keeps these annotations as documentation for the model.
    """
    OpenAi
}

"""
How a build lays out its output.
"""
enum Layout {
    """ One source artifact per source file, keeping the source directory structure. """
    Mirror: "mirror"

    """ The `mirror` files inside the target's project scaffold. """
    Package: "package"

    """ One combined source artifact that holds every definition. """
    SingleFile: "single-file"
}

"""
Casing convention for generated identifiers such as file names, directory names, and
package parts.
"""
enum CasingStyle {
    """ snake_case """
    Snake: "snake_case"

    """ UPPER_SNAKE_CASE (also known as SCREAMING_SNAKE_CASE) """
    UpperSnake: "UPPER_SNAKE_CASE"

    """ camelCase """
    Camel: "camelCase"

    """ PascalCase """
    Pascal: "PascalCase"

    """ kebab-case """
    Kebab: "kebab-case"

    """ UPPER-KEBAB-CASE """
    UpperKebab: "UPPER-KEBAB-CASE"

    """ Title Case """
    Title: "Title Case"

    """ lowercase """
    Lower: "lowercase"

    """ UPPERCASE """
    Upper: "UPPERCASE"
}
```
