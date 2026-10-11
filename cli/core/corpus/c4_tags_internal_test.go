package corpus

import (
	"testing"

	"github.com/brettcallaghan/specs/cli/core/c4diagram"
)

func TestSameTagsComparesContractSlugsWithoutRegardToCase(t *testing.T) {
	tags := func(slug string, unreserved ...string) c4diagram.Tags {
		return c4diagram.Tags{
			Reserved:   []c4diagram.ReservedTag{{Name: c4diagram.TagSync}, {Name: c4diagram.TagContract, Slug: slug}},
			Unreserved: unreserved,
		}
	}
	cases := []struct {
		name string
		a, b c4diagram.Tags
		want bool
	}{
		{"contract slugs differing in case", tags("Spec-Index"), tags("spec-index"), true},
		{"different contract slugs", tags("spec-index"), tags("spec-other"), false},
		{"unreserved tags differing in case", tags("spec-index", "Foo"), tags("spec-index", "foo"), false},
		{"reserved tag order", tags("spec-index"), c4diagram.Tags{Reserved: []c4diagram.ReservedTag{
			{Name: c4diagram.TagContract, Slug: "spec-index"}, {Name: c4diagram.TagSync}}}, false},
	}
	for _, c := range cases {
		if got := sameTags(c.a, c.b); got != c.want {
			t.Errorf("%s: sameTags = %v, want %v", c.name, got, c.want)
		}
	}
}
