package schemaregistry_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/schemaregistry"
)

// Declares reads the branch Validate would select, and follows $ref and
// allOf — a property required in full — but not an alternative.
func TestDeclaresReadsTheSelectedBranch(t *testing.T) {
	schema := []byte(`{
	  "definitions": {
	    "Shared": {"properties": {"inherited": {}}},
	    "Any": {
	      "anyOf": [
	        {"properties": {"type": {"const": "widget"}, "own": {}},
	         "allOf": [{"$ref": "#/definitions/Shared"}],
	         "oneOf": [{"properties": {"optional": {}}}]},
	        {"properties": {"type": {"const": "gadget"}, "other": {}}}
	      ]
	    }
	  }
	}`)
	registry, err := schemaregistry.New(schema)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	for _, tc := range []struct {
		documentType, property string
		want                   bool
	}{
		{"widget", "own", true},
		{"widget", "inherited", true},
		{"widget", "optional", false},
		{"widget", "other", false},
		{"gadget", "other", true},
		{"gadget", "own", false},
	} {
		got, err := registry.Declares(tc.documentType, tc.property)
		if err != nil {
			t.Fatalf("Declares(%q, %q): %v", tc.documentType, tc.property, err)
		}
		if got != tc.want {
			t.Errorf("Declares(%q, %q) = %v, want %v", tc.documentType, tc.property, got, tc.want)
		}
	}

	if _, err := registry.Declares("absent", "own"); err == nil {
		t.Error("Declares for a type with no branch returned no error")
	}
}
