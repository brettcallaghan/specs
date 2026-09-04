// Package schemaregistry compiles the single union JSON Schema an active
// profile supplies and validates one document's frontmatter against it,
// dispatching to the branch the document's declared type selects.
//
// There is no per-type schema registration here: a profile hands over
// one schema — a discriminated union over every document type it
// defines — and this package compiles it once. The branch a document is
// checked against is found mechanically, by converting the document's
// declared type (kebab-case, as documents write it — "system-blueprint")
// to the PascalCase form the union's top-level "definitions" are keyed
// by ("SystemBlueprint"). That transform is a string operation, not a
// lookup table, which is what lets this package name no document type
// of its own: [[module-spec-profile]]'s Change Guidance forbids it, and
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
	}, nil
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

	name := pascalCase(documentType)
	location := resourceURL + "#/definitions/" + name
	schema, err := r.compiler.Compile(location)
	if err != nil {
		return nil, fmt.Errorf("schemaregistry: no schema branch for document type %q: %w", documentType, err)
	}

	r.branches[documentType] = schema
	return schema, nil
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
// Finding, walking the library's flattened Basic output so that a
// oneOf's several failed branches, and every sibling constraint
// violation, are all reported rather than only the first.
func findingsFor(path string, doc *frontmatter.Document, verr *jsonschema.ValidationError) []specmodel.Finding {
	output := verr.BasicOutput()

	units := output.Errors
	if len(units) == 0 {
		units = []jsonschema.OutputUnit{*output}
	}

	var findings []specmodel.Finding
	for _, unit := range units {
		findings = append(findings, findingsForUnit(path, doc, unit)...)
	}
	return findings
}

// findingsForUnit converts one leaf validation failure to the Finding(s)
// it represents: several missing-property Findings for one required
// failure naming several properties, one schema-violation Finding per
// undeclared property for one additionalProperties failure naming
// several, or one schema-violation Finding for any other constraint.
func findingsForUnit(path string, doc *frontmatter.Document, unit jsonschema.OutputUnit) []specmodel.Finding {
	if unit.Error == nil {
		return nil
	}

	pointer := unit.InstanceLocation
	at := spanFor(doc, pointer)

	if required, ok := unit.Error.Kind.(*kind.Required); ok {
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
	if additional, ok := unit.Error.Kind.(*kind.AdditionalProperties); ok {
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

	constraint := strings.Join(unit.Error.Kind.KeywordPath(), "/")
	if constraint == "" {
		constraint = "schema"
	}

	value := ""
	if name, _ := topLevelProperty(pointer); name != "" {
		if property, ok := doc.Property(name); ok {
			value = property.Raw
		}
	}

	return []specmodel.Finding{
		finding.SchemaViolation(path, at, pointer, constraint, value),
	}
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
// token, the inverse of unescapePointerToken. Needed only when this
// package constructs a pointer itself — for every pointer the
// underlying library reports, escaping has already happened before this
// package ever sees it.
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
