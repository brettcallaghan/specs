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
