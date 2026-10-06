package envelope

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Every tag the generated union decodes must be one Decode discovers,
// selecting the same variant. Discovery derives a tag from the variant's
// name, so a variant whose tag broke that convention would be
// undecodable; this pins the convention against the generated decoder.
// The generated encoder cannot serve: it refuses a variant's zero value.
func TestEveryUnionVariantIsDiscoveredUnderItsGeneratedTag(t *testing.T) {
	found := variants()
	accessors := 0
	union := reflect.TypeOf(specmodel.TypedFrontmatter{})
	for i := 0; i < union.NumMethod(); i++ {
		if strings.HasPrefix(union.Method(i).Name, "As") {
			accessors++
		}
	}
	if accessors == 0 || len(found) != accessors {
		t.Fatalf("union has %d variants, %d discovered", accessors, len(found))
	}
	for tag, v := range found {
		raw := []byte(`{"type":"` + tag + `"}`)
		// The generated decoder enforces required fields, so it refuses
		// this value — naming the variant the tag selected.
		var generated specmodel.TypedFrontmatter
		err := json.Unmarshal(raw, &generated)
		if err == nil || !strings.HasPrefix(err.Error(), v.generated.Name()+":") {
			t.Errorf("tag %q: generated decoder says %v, want a %s error", tag, err, v.generated.Name())
		}
		decoded, err := Decode(raw)
		if err != nil {
			t.Fatalf("decode %q: %v", tag, err)
		}
		if got := reflect.TypeOf(decoded.Value); got != v.generated {
			t.Errorf("decode %q produced %v, want %v", tag, got, v.generated)
		}
		if got := Of(&decoded).Type; string(got) != tag {
			t.Errorf("envelope type = %q, want %q", got, tag)
		}
	}
}

// A document declaring a type the union does not define is refused with
// the type named.
func TestDecodeFrontmatterRefusesAnUnknownType(t *testing.T) {
	_, err := Decode([]byte(`{"type":"memo"}`))
	if err == nil || !strings.Contains(err.Error(), `"memo"`) {
		t.Fatalf("err = %v, want unknown document type \"memo\"", err)
	}
}

// A document missing a required property still decodes: the omission is
// the missing-property check's finding, not a parse failure.
func TestDecodeFrontmatterToleratesAMissingRequiredProperty(t *testing.T) {
	value, err := Decode([]byte(`{"type":"adr","slug":"adr-001"}`))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got := Of(&value); got.Type != "adr" || got.Slug != "adr-001" || got.Status != "" {
		t.Errorf("envelope = %+v", got)
	}
}
