// Package schemaregistry compiles the single union JSON Schema an active
// profile supplies and validates one document's frontmatter against it,
// dispatching to the branch the document's declared type selects.
//
// There is no per-type schema registration here: a profile hands over
// one schema — a discriminated union over every document type it
// defines — and this package compiles it once. The branch a document is
// checked against is found mechanically. Where a top-level definition is
// a union whose every branch pins "type" to a constant, the branch whose
// constant equals the document's declared type is the one used: it is
// the closed shape that admits "type" alongside the type's own
// properties. Otherwise the declared type (kebab-case, as documents
// write it — "system-blueprint") is converted to the PascalCase form the
// top-level "definitions" are keyed by ("SystemBlueprint"). Both are
// read from the schema, never from a list, which is what lets this
// package name no document type of its own: [[module-spec-profile]]'s Change Guidance forbids it, and
// specs/modules/document-core/document-core-module.md states the same rule for
// every package under cli/core.
//
// A slug pattern is a separate concern from the union schema — the
// shared frontmatter envelope declares "slug" once, so a subtype cannot
// narrow it, and the pattern that should govern one document's slug
// comes from the active profile's per-type table instead
// ([[corpus.DocumentTypeRules]]). [CheckSlugPattern] is this package's
// second, independent check for exactly that reason.
package schemaregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// resourceURL is the in-memory location the compiled union schema is
// registered under. It is never a real file or network location — the
// schema always arrives as bytes from the active profile — so an
// arbitrary, stable identifier is all a $ref inside the schema needs to
// resolve against its own document.
const resourceURL = "mem://schemaregistry/union.json"

// Registry compiles one union schema and validates many documents
// against the branches it defines.
//
// A Registry is not safe for concurrent use while a branch is still
// being compiled for the first time: [Registry.Validate] mutates an
// internal cache of compiled branches. Once every document type in use
// has been validated once, further calls only read that cache.
type Registry struct {
	compiler *jsonschema.Compiler
	branches map[string]*jsonschema.Schema

	// variants maps a declared document type to the JSON Pointer of the
	// union branch that pins "type" to it, when the schema has such a
	// union. See [discriminatedVariants].
	variants map[string]string
}

// New compiles schema, the bytes an active profile's Schema method
// returns, into a Registry.
//
// schema must be a JSON Schema document (draft-07, per [[adr-015]])
// whose top-level "definitions" hold one sub-schema per document type,
// each keyed by the PascalCase form of the type value documents
// declare. It is compiled once here; [Registry.Validate] compiles each
// branch it is asked for on first use and reuses that compilation
// afterward.
func New(schema []byte) (*Registry, error) {
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return nil, fmt.Errorf("schemaregistry: schema is not valid JSON: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(resourceURL, doc); err != nil {
		return nil, fmt.Errorf("schemaregistry: registering the union schema: %w", err)
	}

	return &Registry{
		compiler: compiler,
		branches: make(map[string]*jsonschema.Schema),
		variants: discriminatedVariants(schema),
	}, nil
}

// discriminatedVariants finds the top-level definition that is a union
// discriminated by "type" — an anyOf whose every branch declares "type"
// as a string constant — and maps each constant to its branch's JSON
// Pointer. It returns an empty map when the schema has no such union.
//
// The union's branches, not the document structs they are built from,
// are what a document validates against: the structs are closed and do
// not declare "type", because a union's discriminator lives on the union.
func discriminatedVariants(schema []byte) map[string]string {
	var doc struct {
		Definitions map[string]struct {
			AnyOf []struct {
				Properties map[string]struct {
					Const *string `json:"const"`
				} `json:"properties"`
			} `json:"anyOf"`
		} `json:"definitions"`
	}
	variants := map[string]string{}
	if err := json.Unmarshal(schema, &doc); err != nil {
		return variants
	}
	for name, def := range doc.Definitions {
		if len(def.AnyOf) == 0 {
			continue
		}
		found := map[string]string{}
		for i, branch := range def.AnyOf {
			tag := branch.Properties["type"].Const
			if tag == nil {
				found = nil
				break
			}
			found[*tag] = fmt.Sprintf("#/definitions/%s/anyOf/%d", name, i)
		}
		for tag, pointer := range found {
			variants[tag] = pointer
		}
	}
	return variants
}

// Validate validates doc's frontmatter against the union schema's branch
// for documentType, reporting every violation found — a violated
// constraint as a schema-violation Finding located at its JSON Pointer,
// an absent required property as a missing-property Finding — rather
// than stopping at the first (AC-SPEC-003.1).
//
// documentType is the value the document declares in its type property.
// This package never names one: it only transforms the string it is
// given into the branch name the union's "definitions" use.
//
// Where the selected branch constrains a property with a format
// keyword, the property's raw source text is validated in place of its
// YAML-coerced value, so that a value a YAML parser silently
// reinterpreted — a bare date widened to a timestamp, an unquoted
// version number narrowed to a float — is checked as the author wrote
// it (AC-SPEC-003.2).
//
// path identifies the document in every returned Finding. An error is
// returned only when the union schema declares no branch for
// documentType — the caller's Profile.DocumentTypes should make that
// unreachable for a document that was not already excluded as
// unknown-type.
func (r *Registry) Validate(path string, documentType string, doc *frontmatter.Document) ([]specmodel.Finding, error) {
	branch, err := r.branch(documentType)
	if err != nil {
		return nil, err
	}

	instance := instanceValue(branch, doc)
	verr := branch.Validate(instance)
	if verr == nil {
		return nil, nil
	}

	validationErr, ok := verr.(*jsonschema.ValidationError)
	if !ok {
		return nil, fmt.Errorf("schemaregistry: validating %s: %w", path, verr)
	}

	return findingsFor(path, doc, validationErr), nil
}

// branch returns the compiled sub-schema documentType selects, compiling
// and caching it on first use.
func (r *Registry) branch(documentType string) (*jsonschema.Schema, error) {
	if s, ok := r.branches[documentType]; ok {
		return s, nil
	}

	location := resourceURL + "#/definitions/" + pascalCase(documentType)
	if pointer, ok := r.variants[documentType]; ok {
		location = resourceURL + pointer
	}
	schema, err := r.compiler.Compile(location)
	if err != nil {
		return nil, fmt.Errorf("schemaregistry: no schema branch for document type %q: %w", documentType, err)
	}

	r.branches[documentType] = schema
	return schema, nil
}

// Declares reports whether the branch documentType selects declares
// property, either directly or through a schema the branch references
// or requires in full ($ref, allOf).
//
// The branch is selected exactly as [Registry.Validate] selects it, so a
// property this reports as declared is one a document of that type may
// carry. An alternative (anyOf, oneOf) is not followed: a property
// declared on only one alternative is not declared on the type.
//
// An error is returned only when the union schema declares no branch
// for documentType.
func (r *Registry) Declares(documentType, property string) (bool, error) {
	branch, err := r.branch(documentType)
	if err != nil {
		return false, err
	}
	return declares(branch, property, make(map[*jsonschema.Schema]bool)), nil
}

// Enum returns the values the schema of property allows on the branch
// documentType selects, in the schema's own order, and whether that
// schema is an enum at all. The property is found as [Registry.Declares]
// finds it, and its schema is followed through $ref to the enum.
//
// An error is returned only when the union schema declares no branch for
// documentType.
func (r *Registry) Enum(documentType, property string) ([]string, bool, error) {
	branch, err := r.branch(documentType)
	if err != nil {
		return nil, false, err
	}
	s := propertySchema(branch, property, make(map[*jsonschema.Schema]bool))
	for seen := map[*jsonschema.Schema]bool{}; s != nil && !seen[s]; s = s.Ref {
		seen[s] = true
		if s.Enum != nil {
			values := make([]string, 0, len(s.Enum.Values))
			for _, value := range s.Enum.Values {
				values = append(values, enumValue(value))
			}
			return values, true, nil
		}
	}
	return nil, false, nil
}

// propertySchema returns the schema s, or a schema it references or
// requires in full, declares for property, or nil.
func propertySchema(s *jsonschema.Schema, property string, seen map[*jsonschema.Schema]bool) *jsonschema.Schema {
	if s == nil || seen[s] {
		return nil
	}
	seen[s] = true

	if found, ok := s.Properties[property]; ok {
		return found
	}
	if found := propertySchema(s.Ref, property, seen); found != nil {
		return found
	}
	for _, required := range s.AllOf {
		if found := propertySchema(required, property, seen); found != nil {
			return found
		}
	}
	return nil
}

// declares reports whether s, or a schema it references or requires in
// full, declares property. seen guards against a cyclic schema graph.
func declares(s *jsonschema.Schema, property string, seen map[*jsonschema.Schema]bool) bool {
	return propertySchema(s, property, seen) != nil
}

// pascalCase converts a kebab-case document type value to the PascalCase
// form the union schema's "definitions" are keyed by:
// "system-blueprint" becomes "SystemBlueprint". It is a mechanical
// string transform, never a lookup naming a specific type, which is
// what lets this package stay framework-agnostic.
func pascalCase(kebab string) string {
	var b strings.Builder
	for _, word := range strings.Split(kebab, "-") {
		if word == "" {
			continue
		}
		runes := []rune(word)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// instanceValue builds the value schema.Validate checks doc against.
//
// Every property starts from its YAML-coerced Value. A property the
// selected branch constrains with a format keyword is then overwritten
// with its raw source text instead — quotes resolved, but no further
// reinterpretation — so a format check runs against what the author
// wrote (AC-SPEC-003.2).
func instanceValue(branch *jsonschema.Schema, doc *frontmatter.Document) map[string]any {
	instance := make(map[string]any, len(doc.Properties))
	for _, property := range doc.Properties {
		instance[property.Name] = property.Value
	}

	for name, sub := range branch.Properties {
		if !hasFormatConstraint(sub, make(map[*jsonschema.Schema]bool)) {
			continue
		}
		if property, ok := doc.Property(name); ok {
			instance[name] = rawText(property.Raw)
		}
	}

	return instance
}

// hasFormatConstraint reports whether s, or a schema it references
// through $ref/oneOf/anyOf/allOf, declares a format keyword. seen guards
// against a cyclic schema graph.
func hasFormatConstraint(s *jsonschema.Schema, seen map[*jsonschema.Schema]bool) bool {
	if s == nil || seen[s] {
		return false
	}
	seen[s] = true

	if s.Format != nil {
		return true
	}
	if hasFormatConstraint(s.Ref, seen) {
		return true
	}
	for _, alt := range s.OneOf {
		if hasFormatConstraint(alt, seen) {
			return true
		}
	}
	for _, alt := range s.AnyOf {
		if hasFormatConstraint(alt, seen) {
			return true
		}
	}
	for _, alt := range s.AllOf {
		if hasFormatConstraint(alt, seen) {
			return true
		}
	}
	return false
}

// rawText resolves a property's raw source text to the string a format
// or pattern check should see: quotes are removed exactly as a YAML
// scalar's quoting resolves them, but nothing else about the text is
// reinterpreted.
func rawText(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if len(trimmed) < 2 {
		return trimmed
	}

	switch {
	case trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"':
		var s string
		if err := json.Unmarshal([]byte(trimmed), &s); err == nil {
			return s
		}
	case trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'':
		return strings.ReplaceAll(trimmed[1:len(trimmed)-1], "''", "'")
	}
	return trimmed
}

// findingsFor converts every leaf of a validation failure into a
// Finding, walking the error's causes down to the errors that have none,
// so that a oneOf's several failed branches, and every sibling constraint
// violation, are all reported rather than only the first.
//
// The walk goes to the leaves rather than reading the library's flattened
// Basic output because that output reports a property constrained through
// $ref as a failure of $ref: the constraint actually violated — a pattern,
// an enum — sits in the cause beneath it, and only the leaf can name it
// (AC-SPEC-006.1).
func findingsFor(path string, doc *frontmatter.Document, verr *jsonschema.ValidationError) []specmodel.Finding {
	var findings []specmodel.Finding
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			findings = append(findings, findingsForLeaf(path, doc, e)...)
			return
		}
		for _, cause := range e.Causes {
			walk(cause)
		}
	}
	walk(verr)
	return findings
}

// findingsForLeaf converts one leaf validation failure to the Finding(s)
// it represents: several missing-property Findings for one required
// failure naming several properties, one schema-violation Finding per
// undeclared property for one additionalProperties failure naming
// several, or one schema-violation Finding for any other constraint.
func findingsForLeaf(path string, doc *frontmatter.Document, leaf *jsonschema.ValidationError) []specmodel.Finding {
	pointer := instancePointer(leaf.InstanceLocation)
	at := spanFor(doc, pointer)

	if required, ok := leaf.ErrorKind.(*kind.Required); ok {
		findings := make([]specmodel.Finding, 0, len(required.Missing))
		for _, property := range required.Missing {
			findings = append(findings, finding.MissingProperty(path, at, pointer, property))
		}
		return findings
	}

	// additionalProperties is reported by the underlying library against
	// the object as a whole (InstanceLocation is the object's own
	// pointer), naming the offending keys in the error's own Properties
	// field rather than in the pointer. Falling through to the generic
	// branch below would therefore report a finding pointing at the
	// whole frontmatter block with no offending value named — every
	// undeclared property in the same document would collapse into one
	// finding indistinguishable from another. Each undeclared property
	// gets its own finding here instead, addressed by its own pointer and
	// span exactly as a document that declares it would be for any other
	// constraint, and naming its raw text as the value — self-contained
	// per AC-SPEC-006.1, the same bar every other schema-violation
	// finding meets.
	if additional, ok := leaf.ErrorKind.(*kind.AdditionalProperties); ok {
		findings := make([]specmodel.Finding, 0, len(additional.Properties))
		for _, property := range additional.Properties {
			propertyPointer := pointer + "/" + escapePointerToken(property)
			propertyAt := spanFor(doc, propertyPointer)
			value := ""
			if prop, ok := doc.Property(property); ok {
				value = prop.Raw
			}
			findings = append(findings,
				finding.SchemaViolation(path, propertyAt, propertyPointer, "additionalProperties", value))
		}
		return findings
	}

	value := ""
	if name, _ := topLevelProperty(pointer); name != "" {
		if property, ok := doc.Property(name); ok {
			value = property.Raw
		}
	}

	return []specmodel.Finding{
		finding.SchemaViolation(path, at, pointer, constraintOf(leaf.ErrorKind), value),
	}
}

// constraintOf names the constraint a leaf failure violated, stating it
// in full where the keyword alone would not tell an author what to write:
// a pattern with its expression, an enum with every allowed value. Any
// other kind is named by its keyword path.
func constraintOf(errorKind jsonschema.ErrorKind) string {
	switch k := errorKind.(type) {
	case *kind.Pattern:
		return fmt.Sprintf("pattern %q", k.Want)
	case *kind.Enum:
		allowed := make([]string, 0, len(k.Want))
		for _, want := range k.Want {
			allowed = append(allowed, enumValue(want))
		}
		return "one of " + strings.Join(allowed, ", ")
	}
	if constraint := strings.Join(errorKind.KeywordPath(), "/"); constraint != "" {
		return constraint
	}
	return "schema"
}

// enumValue renders one allowed enum value: a string as itself, anything
// else as its JSON text.
func enumValue(value any) string {
	if s, ok := value.(string); ok {
		return s
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}

// instancePointer writes an instance location as a JSON Pointer, the
// form the Basic output used: "" for the object itself, and otherwise
// each token escaped per RFC 6901 behind a "/".
func instancePointer(location []string) string {
	var b strings.Builder
	for _, token := range location {
		b.WriteString("/")
		b.WriteString(escapePointerToken(token))
	}
	return b.String()
}

// spanFor locates a JSON Pointer within the document: the position of
// the top-level property it names, or the whole frontmatter block when
// the pointer names the object itself (a missing-property finding at the
// root) or a property this document does not declare.
func spanFor(doc *frontmatter.Document, pointer string) specmodel.Span {
	if name, _ := topLevelProperty(pointer); name != "" {
		if property, ok := doc.Property(name); ok {
			return property.Position
		}
	}
	return doc.Region.Inner
}

// topLevelProperty splits a JSON Pointer into the name of the top-level
// property it starts at and the remainder of the pointer beneath it.
// An empty pointer — the object itself — yields two empty strings.
func topLevelProperty(pointer string) (name, rest string) {
	trimmed := strings.TrimPrefix(pointer, "/")
	if trimmed == "" {
		return "", ""
	}
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		return unescapePointerToken(trimmed[:i]), trimmed[i:]
	}
	return unescapePointerToken(trimmed), ""
}

// unescapePointerToken reverses RFC 6901 escaping of a single pointer
// token.
func unescapePointerToken(token string) string {
	token = strings.ReplaceAll(token, "~1", "/")
	token = strings.ReplaceAll(token, "~0", "~")
	return token
}

// escapePointerToken applies RFC 6901 escaping to a single pointer
// token, the inverse of unescapePointerToken. The library reports an
// instance location as unescaped tokens, so every pointer this package
// writes is escaped here.
func escapePointerToken(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	token = strings.ReplaceAll(token, "/", "~1")
	return token
}

// CheckSlugPattern reports whether slug matches pattern — the regular
// expression the active profile supplies for the document's declared
// type — returning a slug-pattern Finding naming both when it does not
// (AC-SPEC-003.5).
//
// The pattern comes from the profile's per-type table
// ([[corpus.DocumentTypeRules]]) rather than from the union schema: the
// shared frontmatter envelope declares "slug" once, so no subtype can
// narrow it there. A nil Finding and a nil error means slug matched.
func CheckSlugPattern(path string, at specmodel.Span, slug, pattern string) (*specmodel.Finding, error) {
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("schemaregistry: slug pattern %q does not compile: %w", pattern, err)
	}
	if matcher.MatchString(slug) {
		return nil, nil
	}
	f := finding.SlugPattern(path, at, slug, pattern)
	return &f, nil
}
