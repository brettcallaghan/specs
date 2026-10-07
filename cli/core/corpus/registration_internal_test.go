package corpus

import (
	"encoding/json"
	"testing"
)

// resolvePointer implements the URI-fragment form of a JSON Pointer.
func TestResolvePointer(t *testing.T) {
	var document any
	const source = `{"definitions":{"W":{},"a/b":{},"m~n":{},"sp ace":{}},"allOf":[{},{"x":1}]}`
	if err := json.Unmarshal([]byte(source), &document); err != nil {
		t.Fatal(err)
	}

	for _, row := range []struct {
		fragment string
		want     bool
	}{
		{"#/definitions/W", true},
		{"#definitions/W", false},
		{"#/allOf/0", true},
		{"#/allOf/1/x", true},
		{"#/allOf/2", false},
		{"#/allOf/01", false},
		{"#/allOf/-", false},
		{"#/allOf/x", false},
		{"#/allOf/1/x/y", false},
		{"#/definitions/a~1b", true},
		{"#/definitions/m~0n", true},
		{"#/definitions/sp%20ace", true},
		{"#/definitions/%ZZ", false},
		{"#/definitions/Missing", false},
		{"#", true},
		{"definitions/W", false},
		{"", false},
	} {
		if got := resolvePointer(document, row.fragment); got != row.want {
			t.Errorf("resolvePointer(%q) = %t, want %t", row.fragment, got, row.want)
		}
	}
}
