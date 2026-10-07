// Tests for the C4 Model package's reading of reserved relationship tags.
package c4model_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// tagged is a corpus in which container-a draws, to container-b and
// container-c, each statement given; container-a consumes contract-one
// and contract-two, both provided by container-b, and contract-three,
// provided by container-c.
func tagged(statements ...string) map[string]string {
	lines := append([]string{
		`Container(a, "A", "Go")`, `Container(b, "B", "Go")`, `Container(c, "C", "Go")`,
	}, statements...)
	return map[string]string{
		"a.md": head("container-a", "container-blueprint", "A",
			"contracts_consumed:", `  - "contract-one"`, `  - "contract-two"`, `  - "contract-three"`) +
			"# A\n\n" + mermaid("C4Container", lines...),
		"b.md":     head("container-b", "container-blueprint", "B"),
		"c.md":     head("container-c", "container-blueprint", "C"),
		"one.md":   head("contract-one", "contract", "", "provider: container-b"),
		"two.md":   head("contract-two", "contract", "", "provider: container-b"),
		"three.md": head("contract-three", "contract", "", "provider: container-c"),
	}
}

// AC: AC-C4-003.1
func TestTheInteractionTagSetsTheInteraction(t *testing.T) {
	for _, example := range []struct {
		tags, interaction string
	}{
		{"sync", "sync"},
		{"async", "async"},
		{"sync+async", "<absent>"},
		{"audited", "<absent>"},
	} {
		t.Run(example.tags, func(t *testing.T) {
			model := build(t, tagged(`Rel(a, b, "Uses", $tags="`+example.tags+`")`))
			r := relationship(t, model, "container-a", "container-b", "Uses")
			got := "<absent>"
			if r.Interaction != nil {
				got = string(*r.Interaction)
			}
			if got != example.interaction {
				t.Errorf("interaction = %s, want %s", got, example.interaction)
			}
		})
	}
}

// AC: AC-C4-003.2
func TestAContractTagNamesTheGoverningContract(t *testing.T) {
	model := build(t, tagged(
		`Rel(a, b, "Uses", $tags="contract:contract-orders+contract:contract-one")`,
	))

	r := relationship(t, model, "container-a", "container-b", "Uses")
	if r.Contract == nil || r.Contract.Slug != "contract-orders" || r.Contract.Source != specmodel.C4ContractSourceTag {
		t.Errorf("contract = %+v, want contract-orders, the first as written, with source tag", r.Contract)
	}
}

// AC: AC-C4-003.3
func TestAnUntaggedRelationshipDerivesItsContractWhenExactlyOneFits(t *testing.T) {
	model := build(t, tagged(`Rel(a, c, "Uses")`))

	r := relationship(t, model, "container-a", "container-c", "Uses")
	if r.Contract == nil || r.Contract.Slug != "contract-three" || r.Contract.Source != specmodel.C4ContractSourceDerived {
		t.Errorf("contract = %+v, want contract-three with source derived", r.Contract)
	}
}

// AC: AC-C4-003.4
func TestNoContractIsDerivedWhenMoreThanOneFits(t *testing.T) {
	model := build(t, tagged(`Rel(a, b, "Uses")`, `Rel(c, b, "Uses")`))

	if r := relationship(t, model, "container-a", "container-b", "Uses"); r.Contract != nil {
		t.Errorf("contract = %+v, want none: two consumed contracts name container-b", r.Contract)
	}
	// container-c consumes nothing, so none fits.
	if r := relationship(t, model, "container-c", "container-b", "Uses"); r.Contract != nil {
		t.Errorf("contract = %+v, want none: container-c consumes no contract", r.Contract)
	}
}

// AC: AC-C4-003.5
func TestUnreservedTagsPassThrough(t *testing.T) {
	model := build(t, tagged(`Rel(a, b, "Uses", $tags="async+audited")`, `Rel(a, c, "Uses", $tags="sync")`))

	r := relationship(t, model, "container-a", "container-b", "Uses")
	if r.Tags == nil || strings.Join(*r.Tags, ",") != "audited" {
		t.Errorf("tags = %v, want exactly audited", r.Tags)
	}
	if r := relationship(t, model, "container-a", "container-c", "Uses"); r.Tags != nil {
		t.Errorf("tags = %v, want absent when only reserved tags are written", *r.Tags)
	}
}
