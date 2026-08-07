<!-- GENERATED COPY — DO NOT EDIT.
     Source of truth: strata-lang/strata strata/options.strata
     Vendored from commit: ce18d66f188a (2026-08-07)
     Regenerate with: scripts/sync-strata-docs.sh
     Upstream owns this text. Fix it there, then re-run that script. -->

# Compiler options — the schema of `strata.config.json`

Verbatim `options.strata` from strata-lang/strata. Every field states its
default in its doc block.

```strata
struct CompilerOptions {
    """
    Semantic package name shared across all language targets.

    Each transpiler applies its own casing conventions:
    - Rust: kebab-case (e.g., "my-models")
    - Swift: PascalCase (e.g., "MyModels")
    - Kotlin: project name in kebab-case, fully-qualified package derived from `groupId` + lowercase name
    - Go: sanitized to a valid Go package identifier (lowercased, hyphens and underscores stripped, e.g., "mymodels")
    - TypeScript: kebab-case (e.g., "my-models")

    When not set, the package name is derived from the input path.

    Defaults to `null`.
    """
    packageName: String?

    """
    Package version shared across all language targets.

    Each transpiler writes this version into its native manifest:
    - Rust: `version` in `Cargo.toml`
    - Kotlin: `version` in `build.gradle.kts`
    - TypeScript: `version` in `package.json`
    - Swift: (reserved for future use in `Package.swift`)
    - Go: not applicable — `go.mod` has no version field

    Defaults to `"0.0.0"`.
    """
    version: String?

    """
    Packages this package depends on, keyed by the qualifier its types are referenced through.

    A qualified reference names its source package by qualifier: with
    `"dependencies": {"Shared": {"path": "../shared"}}`, the type `User` declared in the
    package at `../shared` is referenced as `Shared.User`. Bare names always mean types of
    the declaring package, so a qualifier is required for every cross-package reference.

    Qualifiers are conventionally PascalCase, matching package names. The key must be a
    valid Strata type identifier, must not collide with a type declared in this package,
    and must not equal this package's own name.

    Dependencies are declared only in a package manifest (`strata.config.json`) — never in
    a `--config` build overlay.

    Defaults to `null` (no dependencies).
    """
    dependencies: [String: DependencySpec]?

    """
    When `true`, `null` properties are omitted from serialized output instead of being written as explicit `null`s.

    Benefits:
    - Reduces payload size by omitting optional fields that are not set
    - Produces the same output whether an optional property is unset or explicitly `null` — helpful for hashing and diffing workflows

    Defaults to `true`.
    """
    skipNullSerialization: Bool?

    """ Options for transpiling Strata's raw value enums. """
    enums: EnumOptions?

    """ Options for transpiling Strata's discriminated unions. """
    unions: UnionOptions?

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

Only local path dependencies are supported today; the shape is open to additional
location fields (registries, git sources) being added later without changing how
existing manifests are read.
"""
struct DependencySpec {
    """
    Path to the dependency's package root — the directory that directly contains its
    `strata.config.json`.

    A relative path resolves against the directory containing the declaring manifest
    (`"../shared"` in `repo/api/strata.config.json` resolves to `repo/shared/`). An
    absolute path is used as-is. Nested, sibling, and unrelated locations all behave
    identically — containment grants no extra visibility.
    """
    path: String
}

"""
Options for transpiling Strata's raw value enums.
"""
struct EnumOptions {
    """
    Controls whether generated enums include a fallback variant for values not known at generation time.

    When set, deserialization no longer fails on values introduced after the code was generated — useful for
    client applications that need forward compatibility with newer servers, and especially for native apps
    where shipping a client update is slow.

    Defaults to `null`: unknown values are rejected at deserialization time.
    """
    unknownVariant: EnumUnknownVariantOption?
}

"""
Specifies how unknown enum variants should be handled during deserialization.
"""
enum EnumUnknownVariantOption {
    """ The unknown variant will contain the unrecognized raw value: `__Unknown(V)` or `case __unknown(RawValueType)`. """
    Value
}

"""
Options for transpiling Strata's discriminated unions.
"""
struct UnionOptions {
    """
    Controls whether generated unions include a fallback variant for tags not known at generation time.

    When set, deserialization no longer fails on variants introduced after the code was generated — useful for
    client applications that need forward compatibility with newer servers, and especially for native apps
    where shipping a client update is slow.

    Defaults to `null`: unknown tags are rejected at deserialization time.
    """
    unknownVariant: UnionUnknownVariantOption?
}

"""
Specifies how unknown union variants should be handled during deserialization.
"""
enum UnionUnknownVariantOption {
    """ The unknown variant will contain the unrecognized tag: `__Unknown(T)` or `case __unknown(TagType)`. """
    Tag
}

struct RustOptions {
    """
    Indentation string to use for generated code.

    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    When `true`, generates a trait for each struct with readonly accessor methods.
    Structs implement their own trait and the traits of all supertypes.

    Defaults to `true`.
    """
    generateInterfaces: Bool?

    """
    Package generation options for Rust.

    When present and `--package` is used, the transpiler emits a full Cargo crate
    instead of a single file. These options configure the generated `Cargo.toml`.

    Defaults to `null`.
    """
    package: RustPackageOptions?
}

"""
Options for generating a Rust package (Cargo crate).
"""
struct RustPackageOptions {
    """
    Rust edition. Defaults to "2021".
    """
    edition: String?
}

struct SwiftOptions {
    """
    Indentation string to use for generated code.
    
    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    When `true`, generates classes instead of structs for data models.
    
    Structs are recommended for most cases due to value semantics, but classes
    may be needed for scenarios requiring reference semantics or inheritance.
    
    Defaults to `false` (generates structs).
    """
    useClasses: Bool?

    """
    When `true`, generates properties with `var` instead of `let`, making them mutable.
    
    By default, properties are immutable (`let`) which aligns with Swift best practices
    for value types and encourages safer, more predictable code. Enable this option
    when you need mutable properties for incremental updates or integration with
    frameworks that require mutability.
    
    Defaults to `false` (generates `let`).
    """
    mutableProperties: Bool?

    """
    When `true`, generates a protocol for each struct with readonly property accessors.
    Structs conform to their own protocol and the protocols of all supertypes.

    Defaults to `true`.
    """
    generateInterfaces: Bool?

    """
    Package generation options for Swift.

    When present and `--package` is used, the transpiler emits a full SPM package
    instead of a single file. These options configure the generated `Package.swift`.

    Defaults to `null`.
    """
    package: SwiftPackageOptions?
}

"""
Options for generating a Swift package (SPM).
"""
struct SwiftPackageOptions {
    """
    swift-tools-version. Defaults to "5.9".
    """
    toolsVersion: String?

    """
    Casing convention for generated directory names.

    Defaults to `Pascal` (PascalCase).
    """
    directoryCasing: CasingStyle?
}

struct KotlinOptions {
    """
    Indentation string to use for generated code.

    Defaults to four spaces: "    ".
    """
    indentation: String?

    """
    When `true`, generates properties with `var` instead of `val`, making them mutable.

    By default, properties are immutable (`val`) which aligns with Kotlin best practices
    for data classes. Enable this option when you need mutable properties.

    Defaults to `false` (generates `val`).
    """
    mutableProperties: Bool?

    """
    When `true`, generates an interface for each data class with readonly property declarations.
    Data classes implement their own interface and the interfaces of all supertypes.

    Defaults to `true`.
    """
    generateInterfaces: Bool?

    """
    Package generation options for Kotlin.

    When present and `--package` is used, the transpiler emits a full Gradle project
    instead of a single file. These options configure the generated build files.

    Defaults to `null`.
    """
    package: KotlinPackageOptions?
}

"""
Options for generating a Kotlin package (Gradle project).
"""
struct KotlinPackageOptions {
    """
    Maven group ID (e.g., "com.example"). Required for Kotlin packages.
    This prefixes all generated package declarations and determines directory structure.
    """
    groupId: String

    """
    Kotlin version for build.gradle.kts. Defaults to "1.9.22".
    """
    kotlinVersion: String?

    """
    JVM toolchain version for build.gradle.kts. Defaults to "17".
    """
    jvmToolchain: String?

    """
    Casing convention for generated file names.

    Defaults to `Pascal` (PascalCase).
    """
    fileCasing: CasingStyle?
}

struct TypeScriptOptions {
    """
    Indentation string to use for generated code.

    Defaults to two spaces: "  ".
    """
    indentation: String?

    """
    When `true`, generates properties without `readonly`.

    Defaults to `false` (generates `readonly`).
    """
    mutableProperties: Bool?

    """
    When `true`, generates the serde code — property tables, `deserialize`
    and `serialize` functions, and the constraint checks they run — alongside
    the type definitions.

    When `false`, only the type definitions are emitted, and the generated
    module has no dependency on `@strata-lang/serde` at all. Useful when the
    data source is trusted (e.g. compiler-generated AST) and the validation
    overhead is unnecessary.

    Serialization is not a separable half of this: both directions run off one
    property table, and that table names the `deserialize` function of every
    type it references, so a module that kept its serializers would still pull
    in the whole runtime.

    Defaults to `true`.
    """
    generateValidation: Bool?

    """
    Package generation options for TypeScript.

    When present and `--package` is used, the transpiler emits an npm package
    instead of a single file. These options configure the generated `package.json` and `tsconfig.json`.

    Defaults to `null`.
    """
    package: TypeScriptPackageOptions?
}

"""
Options for generating a TypeScript package (npm).
"""
struct TypeScriptPackageOptions {
    """
    TypeScript version for devDependencies.

    Defaults to "^5.4".
    """
    typescriptVersion: String?

    """
    ECMAScript target for tsconfig.

    Defaults to "ES2020".
    """
    target: String?

    """
    Casing convention for generated file names.

    Defaults to `Pascal` (PascalCase).
    """
    fileCasing: CasingStyle?
}

struct GoOptions {
    """
    Indentation string to use for generated code.

    Go convention is to use tabs (`gofmt` enforces this). Defaults to a single tab.
    """
    indentation: String?

    """
    Package generation options for Go.

    When present and `--package` is used, the transpiler emits a full Go module
    instead of a single file. These options configure the generated `go.mod`.

    Defaults to `null`.
    """
    package: GoPackageOptions?
}

"""
Options for generating a Go module.
"""
struct GoPackageOptions {
    """
    Go module path written to `go.mod`. Required when `--package` is used with the
    Go target (e.g., "github.com/example/my-models").
    """
    modulePath: String

    """
    Go language version for `go.mod`. Defaults to "1.24".
    """
    goVersion: String?
}

"""
Options for the JSON Schema transpiler.
"""
struct JSONSchemaOptions {
    """
    When `true`, generates a single bundled schema file for the entire package with all definitions included.
    When `false`, generates separate schema files for each source file with cross-file references.
    
    Bundling reduces file count and ensures all references are local (no external `$ref`).
    
    Defaults to `false` for backward compatibility.
    """
    bundle: Bool?

    """
    When `true`, excludes `description` fields from the generated JSON Schema.
    This reduces schema size by 10-30% for production builds where documentation is not needed.
    
    Defaults to `false` (documentation is included).
    """
    excludeDoc: Bool?

    """
    Specifies which JSON Schema draft version to target.
    
    - `Draft04`: JSON Schema Draft 4 (older tooling, widest compatibility)
    - `Draft07`: JSON Schema Draft-07 (default, widely supported)
    - `Draft201909`: JSON Schema Draft 2019-09 (adds unevaluatedProperties, recursive refs)
    - `Draft202012`: JSON Schema Draft 2020-12 (modern features)
    - `OpenAi`: Unofficial draft targeting OpenAI's Structured Outputs API requirements
    
    Defaults to `Draft07` for backward compatibility.
    """
    schemaVersion: JsonSchemaDraft?

    """
    Specifies the root type to use for filtering JSON Schema output.
    
    When set, only the specified type and its transitive dependencies will be included in the output,
    and a top-level `$ref` will point to the root type definition.
    This is useful for generating focused schemas for specific types (e.g., per-endpoint OpenAPI schemas).
    
    - The root type must exist in the compiled sources
    - All types referenced by the root (directly or transitively) will be included
    - Works with templates, inheritance, and unions
    
    Defaults to `null` (all types are included, no top-level $ref).
    """
    root: String?

    """
    When `false`, adds `additionalProperties: false` to every generated object schema,
    so documents containing keys not declared in the schema fail validation.

    Disabling it is useful for schemas that gate configuration files, where a misspelled
    key should be reported instead of silently ignored. The `OpenAi` schema version
    always emits `additionalProperties: false` regardless of this option.

    A closed shape and `allOf` composition cannot coexist in JSON Schema: a derived
    type's `allOf` carries a `$ref` to its base, and that referenced subschema
    independently rejects every property the derived type adds. Setting this to
    `false` therefore makes inheriting types emit their ancestors' properties inline
    instead of as a `$ref`. Validation semantics are unchanged — the same documents
    are accepted and rejected — but the emitted schema is larger and shares no
    definitions between a type and its supertypes. `unevaluatedProperties` does not
    avoid this: it can close a derived type, but a base type still has to be closed
    for its own direct use, and a closed base cannot be composed no matter which
    keyword closed it.

    Defaults to `true` (undeclared keys are allowed).
    """
    allowAdditionalProperties: Bool?
}

"""
Specifies which JSON Schema draft version to target.

JSON Schema uses the term "draft" rather than "version" to refer to different iterations of the specification.
"""
enum JsonSchemaDraft {
    """ JSON Schema Draft 4 (widely supported, older tooling) """
    Draft04

    """ JSON Schema Draft-07 (widely supported) """
    Draft07

    """ JSON Schema Draft 2019-09 (adds advanced features like unevaluatedProperties) """
    Draft201909

    """ JSON Schema Draft 2020-12 (modern standard) """
    Draft202012

    """
    Unofficial draft targeting OpenAI's Structured Outputs API requirements.

    This mode enforces the requirements of OpenAI's structured model outputs:
    - Adds `additionalProperties: false` to all object types
    - Makes all properties required (optional ones use `anyOf: [{"type": "null"}, <schema>]`)
    - Represents discriminated unions using `anyOf` instead of `if`/`then`/`allOf`
    - Flattens inherited struct properties into the child struct schema
    - Omits the `$schema` header
    - Uses `$defs` for shared type definitions
    - Implies bundled mode (all `$ref` are local to the document)

    Note: Some Strata features produce output that OpenAI cannot enforce at generation time
    (e.g., numeric range constraints, string length/pattern constraints, date formats). These
    annotations are retained in the schema as documentation for the model.
    """
    OpenAi
}

"""
Casing convention used for generated identifiers such as file names, directory
names, and package parts.
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
