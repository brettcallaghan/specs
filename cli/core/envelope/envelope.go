// Package envelope reads a document's typed frontmatter without naming a
// document type.
//
// A document's frontmatter is a [specmodel.TypedFrontmatter]: the union
// variant its `type` selects. [Decode] builds one from a parsed block,
// and [Of] reads from one the properties every document type shares — the
// envelope — which is all Document Core needs of most documents.
package envelope

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Envelope is what every document's frontmatter shares, whatever its type:
// the declared type and the [specmodel.DocumentFrontmatter] envelope.
type Envelope struct {
	// Type is the value the document declares in its `type` property —
	// the discriminator that selected its [specmodel.TypedFrontmatter]
	// variant.
	Type specmodel.DocumentType

	specmodel.DocumentFrontmatter
}

// envelopes memoizes [Of] per frontmatter value. An index entry's
// frontmatter is read many times — by every rule and query that needs a
// slug, a status or a type — and never changes once parsed.
var envelopes sync.Map // *specmodel.TypedFrontmatter -> Envelope

// Of returns the shared envelope of a document's typed
// frontmatter: its declared type and the properties every document type
// carries.
//
// Every variant carries the envelope's properties under the same names,
// so they are copied by name and this function names no document type.
// It does not go through the value's JSON encoding, which the generated
// encoder refuses for a document missing a required enum such as
// `status` — the very documents the checks must still describe. It
// returns the zero Envelope for a nil frontmatter.
func Of(f *specmodel.TypedFrontmatter) Envelope {
	if f == nil || f.Value == nil {
		return Envelope{}
	}
	if cached, ok := envelopes.Load(f); ok {
		return cached.(Envelope)
	}
	value := reflect.ValueOf(f.Value)
	envelope := Envelope{Type: specmodel.DocumentType(kebab(value.Type().Name()))}
	shared := reflect.ValueOf(&envelope.DocumentFrontmatter).Elem()
	for i := 0; i < shared.NumField(); i++ {
		field := value.FieldByName(shared.Type().Field(i).Name)
		if field.IsValid() && field.Type() == shared.Field(i).Type() {
			shared.Field(i).Set(field)
		}
	}
	envelopes.Store(f, envelope)
	return envelope
}

// variant is one [specmodel.TypedFrontmatter] variant as
// [Decode] needs it: the generated struct, and a struct type
// with the same fields and no methods to decode through.
type variant struct {
	generated reflect.Type
	relaxed   reflect.Type
}

// variants maps each union tag — the `type` value a document declares —
// to its variant, discovered once from the generated union itself.
var variants = sync.OnceValue(discoverVariants)

// discoverVariants reads the variants from the generated union rather than
// from a list: every variant has an As<Variant> accessor whose first
// result is the variant's type, and the variant's tag is its name in
// kebab case (`SystemBlueprint` is `system-blueprint`), the convention
// the schema's union follows. A variant added to the schema is therefore
// decodable the moment it is generated, and this package names no
// document type.
//
// The tag is not read by encoding a zero value of the variant: the
// generated encoder refuses an empty enum, and every variant carries one.
func discoverVariants() map[string]variant {
	found := map[string]variant{}
	union := reflect.TypeOf(specmodel.TypedFrontmatter{})
	variantType := reflect.TypeOf((*specmodel.TypedFrontmatterVariant)(nil)).Elem()
	for i := 0; i < union.NumMethod(); i++ {
		method := union.Method(i)
		if !strings.HasPrefix(method.Name, "As") || method.Type.NumOut() != 2 {
			continue
		}
		generated := method.Type.Out(0)
		if generated.Kind() != reflect.Struct || !generated.Implements(variantType) {
			continue
		}
		fields := make([]reflect.StructField, generated.NumField())
		for f := range fields {
			fields[f] = generated.Field(f)
		}
		found[kebab(generated.Name())] = variant{generated: generated, relaxed: reflect.StructOf(fields)}
	}
	return found
}

// kebab converts a Go type name to kebab case: `ModuleBlueprint` becomes
// `module-blueprint`.
func kebab(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			r += 'a' - 'A'
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Decode decodes a document's frontmatter, given as JSON, into
// the [specmodel.TypedFrontmatter] variant its `type` selects.
//
// It does not enforce required properties, which the generated decoder
// does. A document missing a required property is the subject of a check
// (`frontmatter/missing-property`), and must still be read so the check
// can report it; enforcing it here would turn every such document into a
// parse failure. Each property is still decoded through its generated
// type, so a status outside the vocabulary or a malformed timestamp is
// still an error, and a property the document's type does not define is
// left to the schema check.
func Decode(raw []byte) (specmodel.TypedFrontmatter, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return specmodel.TypedFrontmatter{}, err
	}
	v, ok := variants()[probe.Type]
	if !ok {
		return specmodel.TypedFrontmatter{}, fmt.Errorf("unknown document type %q", probe.Type)
	}
	relaxed := reflect.New(v.relaxed)
	if err := json.Unmarshal(raw, relaxed.Interface()); err != nil {
		return specmodel.TypedFrontmatter{}, err
	}
	value := relaxed.Elem().Convert(v.generated).Interface().(specmodel.TypedFrontmatterVariant)
	return specmodel.TypedFrontmatter{Value: value}, nil
}

// Encode encodes a document's typed frontmatter as the generated
// [specmodel.TypedFrontmatter] encoder does — the `type` discriminator
// first, then the variant's properties in declaration order — except that
// an enum-typed property holding the empty value is omitted rather than
// refused.
//
// [Decode] leaves a required enum the document does not declare at its
// empty value, so the document can be read and the omission reported by a
// check (`frontmatter/missing-property`). The generated encoder refuses
// that value, because it is not a member of the enum, which would make the
// index unwritable over a corpus that merely fails a check. Omitting the
// property records what the author wrote, and nothing they did not.
//
// For a frontmatter with no empty enum the bytes are exactly those
// [json.Marshal] produces for the generated union.
func Encode(f specmodel.TypedFrontmatter) ([]byte, error) {
	if f.Value == nil {
		return nil, fmt.Errorf("envelope: encoding frontmatter: nil variant")
	}
	value := reflect.ValueOf(f.Value)
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("envelope: encoding frontmatter: variant %s is not a struct", value.Type())
	}
	tag, err := json.Marshal(kebab(value.Type().Name()))
	if err != nil {
		return nil, err
	}

	out := append([]byte(`{"type":`), tag...)
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		if !field.IsExported() || field.Anonymous {
			return nil, fmt.Errorf("envelope: encoding frontmatter: %s.%s is not a plain exported field", value.Type(), field.Name)
		}
		name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "-" && options == "" {
			continue
		}
		if name == "" {
			name = field.Name
		}
		v := value.Field(i)
		if hasOption(options, "string") {
			return nil, fmt.Errorf("envelope: encoding frontmatter: %s.%s uses the unsupported \",string\" option", value.Type(), field.Name)
		}
		if hasOption(options, "omitempty") && isEmpty(v) {
			continue
		}
		if isEmptyEnum(v) {
			continue
		}
		key, err := json.Marshal(name)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(v.Interface())
		if err != nil {
			return nil, fmt.Errorf("envelope: encoding frontmatter property %s: %w", name, err)
		}
		out = append(out, ',')
		out = append(out, key...)
		out = append(out, ':')
		out = append(out, encoded...)
	}
	return append(out, '}'), nil
}

// hasOption reports whether a comma-separated json tag option list holds
// option.
func hasOption(options, option string) bool {
	for options != "" {
		var next string
		next, options, _ = strings.Cut(options, ",")
		if next == option {
			return true
		}
	}
	return false
}

// isEmpty is encoding/json's omitempty test: false, 0, a nil pointer or
// interface, and an empty array, slice, map or string.
func isEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

// marshaler is the generated enums' encoding method.
var marshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()

// isEmptyEnum reports whether v is a generated enum — a string type that
// encodes itself — holding the empty value, which no generated enum
// admits.
func isEmptyEnum(v reflect.Value) bool {
	return v.Kind() == reflect.String && v.Len() == 0 && v.Type().Implements(marshaler)
}

// DeclaredSlug returns the slug an index entry declares: its envelope's
// slug where its frontmatter parsed, the slug recovered from a block that
// did not parse where one was legible (AC-SPEC-003.7), and the empty
// string otherwise. Every slug lookup reads an entry's slug through this
// one function, so a member whose frontmatter failed is found by its
// slug wherever a member that parsed would be.
func DeclaredSlug(entry specmodel.IndexEntry) string {
	if entry.Frontmatter != nil {
		return string(Of(entry.Frontmatter).Slug)
	}
	if entry.RecoveredSlug != nil {
		return *entry.RecoveredSlug
	}
	return ""
}
