// Tests for the Frontmatter component.
//
// No test here carries an `AC:` citation, and that is deliberate rather
// than an omission. This component proves no acceptance criterion on its
// own: it participates in AC-SPEC-003.2, which the Schema Registry proves
// by checking a raw value against a format constraint, and in
// AC-SPEC-002.4, which Corpus proves by indexing the rest of a corpus
// around a document that would not parse. Both cite it from their own
// tests once they land. What is verified here is the behavior those two
// depend on: that the region comes from the shared scanner, that the raw
// text survives, and that malformed input yields a located error rather
// than a panic or an abandoned walk.
package frontmatter_test

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

const valid = "---\n" +
	"title: \"A Blueprint: Deeper\"\n" +
	"slug: adr-011\n" +
	"updated_at: 2026-07-29\n" +
	"tags:\n" +
	"  - one\n" +
	"  - two\n" +
	"nested:\n" +
	"  key: value\n" +
	"---\n" +
	"\n# Body\n"

func TestValidFrontmatterYieldsEveryProperty(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte(valid))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}
	if doc == nil {
		t.Fatal("valid frontmatter parsed to no document")
	}

	var names []string
	for _, property := range doc.Properties {
		names = append(names, property.Name)
	}
	want := []string{"title", "slug", "updated_at", "tags", "nested"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("properties = %v, want %v in declaration order", names, want)
	}

	values := doc.Values()
	if values["title"] != "A Blueprint: Deeper" {
		t.Errorf("title = %#v, want the resolved string", values["title"])
	}
	if !reflect.DeepEqual(values["tags"], []any{"one", "two"}) {
		t.Errorf("tags = %#v, want a two-element sequence", values["tags"])
	}
	if !reflect.DeepEqual(values["nested"], map[string]any{"key": "value"}) {
		t.Errorf("nested = %#v, want a nested mapping", values["nested"])
	}
}

// The generic model converts to JSON, which is how a block is decoded
// into one of the generated records: they carry json tags and no yaml
// ones.
func TestTheValueTreeConvertsToJSON(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte(valid))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	encoded, err := doc.JSON()
	if err != nil {
		t.Fatalf("JSON: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("the emitted JSON does not decode: %v", err)
	}
	if decoded["slug"] != "adr-011" {
		t.Errorf("decoded slug = %#v, want %q", decoded["slug"], "adr-011")
	}
}

// Every property carries the source text of its value as written, which
// is what lets a format constraint be checked against the author's text
// rather than against whatever the YAML parser resolved it to. The
// unquoted date is the case that matters: a parser that coerced it to a
// timestamp would leave no way to notice that it is not an RFC 3339
// date-time.
func TestRawTextIsRetainedForAValueYAMLWouldCoerce(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte(valid))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	cases := map[string]string{
		"title":      `"A Blueprint: Deeper"`,
		"slug":       "adr-011",
		"updated_at": "2026-07-29",
		"tags":       "- one\n  - two",
		"nested":     "key: value",
	}
	for name, want := range cases {
		property, ok := doc.Property(name)
		if !ok {
			t.Errorf("no property %q", name)
			continue
		}
		if property.Raw != want {
			t.Errorf("%s raw = %q, want %q", name, property.Raw, want)
		}
	}

	// The quoted title is the proof that raw text is the source and not a
	// re-rendering of the parsed value: the quotes are part of what the
	// author wrote and are absent from the value.
	title, _ := doc.Property("title")
	if title.Value == title.Raw {
		t.Errorf("raw text and parsed value are identical (%q); the quoting was lost", title.Raw)
	}
}

// A property's span addresses the value inside its line, which is what a
// finding needs in order to point at the offending text rather than at
// the line containing it.
func TestPropertySpansAddressTheValueWithinItsLine(t *testing.T) {
	source := []byte(valid)
	doc, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	for _, property := range doc.Properties {
		if got := string(span.Slice(source, property.Position)); got != property.Raw {
			t.Errorf("%s: the span covers %q but the raw text is %q", property.Name, got, property.Raw)
		}
		if got := string(span.Slice(source, property.Key)); got != property.Name {
			t.Errorf("%s: the key span covers %q", property.Name, got)
		}
	}

	updated, _ := doc.Property("updated_at")
	if updated.Position.StartLine != 4 || updated.Position.StartColumn != 13 {
		t.Errorf("updated_at value at %d:%d, want 4:13",
			updated.Position.StartLine, updated.Position.StartColumn)
	}
}

// A multi-byte character earlier in the block must not shift the
// position or the raw text reported for a later property.
//
// The underlying YAML scanner (goccy/go-yaml) advances its own token
// position by one unit per Unicode codepoint rather than per byte, so a
// naive translation of that position into a byte offset lands short by
// the accumulated byte/rune width difference of everything earlier in
// the block — an em dash (three bytes, one codepoint) understates the
// true byte offset of every property after it by two. This went
// undetected by every fixture in this file, which is pure ASCII;
// running the profile self-application gate against this repository's
// own specification corpus (several of whose titles and descriptions
// use em dashes) surfaced it as spurious `format` violations on
// `updated_at`, because the corrupted raw text — a few trailing
// characters of the property's own key, glued onto a truncated value —
// is not a valid RFC 3339 timestamp.
func TestAMultiByteCharacterEarlierInTheBlockDoesNotShiftALaterProperty(t *testing.T) {
	source := []byte("---\n" +
		"title: \"Boundary — Unified — Separated\"\n" +
		"updated_at: 2026-07-29T00:00:00Z\n" +
		"---\n")
	doc, perr := frontmatter.Parse(source)
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	updated, ok := doc.Property("updated_at")
	if !ok {
		t.Fatal("no updated_at property")
	}
	if updated.Raw != "2026-07-29T00:00:00Z" {
		t.Errorf("updated_at Raw = %q, want %q", updated.Raw, "2026-07-29T00:00:00Z")
	}
	if got := string(span.Slice(source, updated.Position)); got != updated.Raw {
		t.Errorf("updated_at span covers %q, want it to match Raw %q", got, updated.Raw)
	}

	title, ok := doc.Property("title")
	if !ok {
		t.Fatal("no title property")
	}
	if got := string(span.Slice(source, title.Position)); got != title.Raw {
		t.Errorf("title span covers %q, want it to match Raw %q", got, title.Raw)
	}
}

// A document with no frontmatter is not a defect here. It yields neither
// a document nor an error, and the caller records it as excluded.
func TestAbsentFrontmatterIsNeitherADocumentNorAnError(t *testing.T) {
	cases := map[string]string{
		"plain body":        "# Title\n\nSome prose.\n",
		"empty file":        "",
		"rule further down": "# Title\n\n---\n\nMore prose.\n",
		"indented opener":   "  ---\ntitle: x\n---\n",
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			doc, perr := frontmatter.Parse([]byte(source))
			if perr != nil {
				t.Errorf("parse error %q for a document with no frontmatter", perr.Message)
			}
			if doc != nil {
				t.Errorf("a document with no frontmatter parsed to %d properties", len(doc.Properties))
			}
		})
	}
}

// A block that opens and never closes is a defect, not an absence: the
// author meant to write frontmatter. The scanner reports no region for
// it, so the error is located at the dangling delimiter.
func TestAnUnterminatedBlockIsALocatedError(t *testing.T) {
	source := []byte("---\ntitle: x\nslug: y\n")

	doc, perr := frontmatter.Parse(source)
	if doc != nil {
		t.Error("an unterminated block parsed to a document")
	}
	if perr == nil {
		t.Fatal("an unterminated block produced no error")
	}
	if !strings.Contains(perr.Message, "never closed") {
		t.Errorf("message = %q, want it to say the block is never closed", perr.Message)
	}
	if perr.Span.Offset != 0 || perr.Span.StartLine != 1 {
		t.Errorf("position = %+v, want the opening delimiter", perr.Span)
	}
}

// Malformed YAML is reported with its position and never raised, so the
// caller records the failure against this document and keeps indexing.
func TestMalformedYAMLIsALocatedErrorRatherThanAPanic(t *testing.T) {
	cases := map[string]struct {
		source string
		line   uint64
	}{
		"unterminated string": {source: "---\ntitle: \"x\nslug: y\n---\n", line: 2},
		"tab indentation":     {source: "---\ntitle:\n\tslug: y\n---\n", line: 3},
		"duplicate key":       {source: "---\nslug: a\nslug: b\n---\n", line: 3},
		"not a mapping":       {source: "---\n- one\n- two\n---\n", line: 1},
		"bare scalar":         {source: "---\njust prose\n---\n", line: 1},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			doc, perr := frontmatter.Parse([]byte(c.source))
			if doc != nil {
				t.Error("malformed frontmatter parsed to a document")
			}
			if perr == nil {
				t.Fatal("malformed frontmatter produced no error")
			}
			if perr.Message == "" {
				t.Error("the parse error carries no message")
			}
			if perr.Span.StartLine != c.line {
				t.Errorf("error at line %d, want %d (%q)", perr.Span.StartLine, c.line, perr.Message)
			}
			if perr.Span.Offset+perr.Span.Length > uint64(len(c.source)) {
				t.Errorf("position %+v runs past the document", perr.Span)
			}
		})
	}
}

// An empty block is a document that declares nothing, which is different
// from a document with no block at all: the first is a member with no
// properties, the second is not a member.
func TestAnEmptyBlockIsADocumentWithNoProperties(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte("---\n---\n\n# Body\n"))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}
	if doc == nil {
		t.Fatal("an empty block parsed to no document")
	}
	if len(doc.Properties) != 0 {
		t.Errorf("properties = %v, want none", doc.Properties)
	}
	if len(doc.Values()) != 0 {
		t.Errorf("values = %v, want none", doc.Values())
	}
}

// A block declaring exactly one property parses to a different YAML node
// shape than one declaring several, and both must be read the same way.
func TestASinglePropertyBlockIsReadLikeAnyOther(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte("---\nslug: adr-011\n---\n"))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}
	property, ok := doc.Property("slug")
	if !ok {
		t.Fatalf("properties = %v, want slug", doc.Properties)
	}
	if property.Raw != "adr-011" || property.Value != "adr-011" {
		t.Errorf("property = %+v, want the slug as both raw text and value", property)
	}
}

// The region a document reports is exactly the one the shared scanner
// finds. This is the behavioral half of the obligation that new parsing
// capability builds on the Spans scanner rather than searching for the
// delimiters itself; the source check below is the other half.
func TestTheRegionComesFromTheSharedScanner(t *testing.T) {
	for name, source := range map[string]string{
		"typical": valid,
		"empty":   "---\n---\n",
		"no body": "---\nslug: a\n---",
	} {
		t.Run(name, func(t *testing.T) {
			doc, perr := frontmatter.Parse([]byte(source))
			if perr != nil || doc == nil {
				t.Fatalf("parse: doc=%v err=%v", doc != nil, perr)
			}
			want, found := span.Find([]byte(source), "---", "\n---", 0)
			if !found {
				t.Fatal("the scanner finds no region in a document that parsed")
			}
			if doc.Region != want {
				t.Errorf("region = %+v, want the scanner's %+v", doc.Region, want)
			}
		})
	}
}

// The scanner is reached for rather than reimplemented. A future edit
// that added its own search for the delimiters would keep every test
// above passing while quietly reintroducing the divergence the shared
// scanner exists to prevent, so the source itself is asserted.
func TestTheImplementationCallsTheSharedScanner(t *testing.T) {
	source, err := os.ReadFile("frontmatter.go")
	if err != nil {
		t.Fatalf("read frontmatter.go: %v", err)
	}
	if !strings.Contains(string(source), "span.Find(") {
		t.Error("frontmatter.go does not call span.Find")
	}
	for _, forbidden := range []string{"bytes.Index(", "strings.Index(", "bytes.Split(", "strings.Split("} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("frontmatter.go uses %s; the region is located with the shared scanner", forbidden)
		}
	}
}

// A sequence-valued property carries one item per element, each at its
// own span. The whole value's span is the right extent for a finding
// about the property and the wrong one for a finding about one of the
// slugs inside it, so both are kept.
func TestASequenceValuedPropertyCarriesPerItemSpans(t *testing.T) {
	src := []byte("---\n" +
		"block:\n" +
		"  - \"quoted-one\"\n" +
		"  - bare-two\n" +
		"flow: [alpha, \"beta gamma\"]\n" +
		"empty: []\n" +
		"scalar: just-one\n" +
		"composite:\n" +
		"  - key: value\n" +
		"  - other: value\n" +
		"---\n\nBody\n")

	doc, perr := frontmatter.Parse(src)
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	for _, tc := range []struct {
		property string
		raws     []string
		values   []any
	}{
		{"block", []string{`"quoted-one"`, "bare-two"}, []any{"quoted-one", "bare-two"}},
		{"flow", []string{"alpha", `"beta gamma"`}, []any{"alpha", "beta gamma"}},
		{"empty", nil, nil},
		{"composite", []string{"key: value", "other: value"}, nil},
	} {
		property, ok := doc.Property(tc.property)
		if !ok {
			t.Fatalf("no property %q", tc.property)
		}
		if property.Items == nil {
			t.Errorf("%s: Items is nil, want a sequence's items", tc.property)
			continue
		}
		if len(property.Items) != len(tc.raws) {
			t.Errorf("%s: %d items, want %d", tc.property, len(property.Items), len(tc.raws))
			continue
		}
		for i, item := range property.Items {
			if item.Raw != tc.raws[i] {
				t.Errorf("%s[%d]: raw = %q, want %q", tc.property, i, item.Raw, tc.raws[i])
			}
			if covered := string(span.Slice(src, item.Position)); covered != tc.raws[i] {
				t.Errorf("%s[%d]: span covers %q, want %q", tc.property, i, covered, tc.raws[i])
			}
			if tc.values != nil && item.Value != tc.values[i] {
				t.Errorf("%s[%d]: value = %#v, want %#v", tc.property, i, item.Value, tc.values[i])
			}
		}
	}

	// Two items of the same list sit at two different positions, which
	// is the whole point of recording them.
	block, _ := doc.Property("block")
	if block.Items[0].Position == block.Items[1].Position {
		t.Error("both items of a list carry the same span")
	}
	if block.Items[0].Position.StartLine != 3 || block.Items[1].Position.StartLine != 4 {
		t.Errorf("items are on lines %d and %d, want 3 and 4",
			block.Items[0].Position.StartLine, block.Items[1].Position.StartLine)
	}
}

// Spans are derived from a token's line and column, never from
// tok.Position.Offset, because the YAML scanner's offset drifts one unit
// per comment line and one per multi-byte codepoint (see absolute).
// Deriving from the offset shifts every later span left and corrupts the
// raw source text a format check reads: with three comments above the
// list, item 0's Raw became " - contract-plaint".
func TestSpansSurviveCommentsAndMultiByteText(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before string
	}{
		{"nothing before the list", ""},
		{"one comment", "# a comment\n"},
		{"three comments", "# one\n# two\n# three\n"},
		{"a multi-byte value", "title: alpha — beta\n"},
		{"comments around a multi-byte value", "# one\ntitle: alpha — beta\n# two\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("---\n" + tc.before +
				"refs:\n" +
				"  - contract-plaintext\n" +
				"  - adr-018\n" +
				"---\n\nBody\n")

			doc, perr := frontmatter.Parse(src)
			if perr != nil {
				t.Fatalf("parse: %v", perr.Message)
			}
			property, ok := doc.Property("refs")
			if !ok {
				t.Fatal("no property \"refs\"")
			}

			want := []string{"contract-plaintext", "adr-018"}
			if len(property.Items) != len(want) {
				t.Fatalf("%d items, want %d", len(property.Items), len(want))
			}
			for i, item := range property.Items {
				if item.Raw != want[i] {
					t.Errorf("item[%d]: raw = %q, want %q", i, item.Raw, want[i])
				}
				if covered := string(span.Slice(src, item.Position)); covered != want[i] {
					t.Errorf("item[%d]: span covers %q, want %q", i, covered, want[i])
				}
				// Every element starts at the same column whatever
				// precedes the block; a drifting offset moves it.
				if item.Position.StartColumn != 5 {
					t.Errorf("item[%d]: starts at column %d, want 5",
						i, item.Position.StartColumn)
				}
			}
		})
	}
}

// A property whose value is not a sequence carries no items at all, so a
// caller can tell "not a list" from "an empty list".
func TestANonSequencePropertyCarriesNoItems(t *testing.T) {
	doc, perr := frontmatter.Parse([]byte(valid))
	if perr != nil {
		t.Fatalf("parse: %v", perr.Message)
	}

	for _, name := range []string{"title", "slug", "nested"} {
		property, ok := doc.Property(name)
		if !ok {
			t.Fatalf("no property %q", name)
		}
		if property.Items != nil {
			t.Errorf("%s: Items = %v, want nil", name, property.Items)
		}
	}
	if tags, _ := doc.Property("tags"); tags.Items == nil {
		t.Error("tags: Items is nil, want a sequence's items")
	}
}

// Locate reports where a block sits without reading it, so a caller that
// needs only the body's offset — the corpus walk, for a document whose
// frontmatter did not parse — asks the component that owns the format
// rather than re-deriving the delimiters.
func TestLocateReportsTheBlockWithoutParsingIt(t *testing.T) {
	// A run of three hyphens inside a value is content: only a "---"
	// beginning a line closes the block.
	withHyphens := []byte("---\ntitle: a---b\nstatus: [unclosed\n---\n\n# Real heading\n")
	region, ok := frontmatter.Locate(withHyphens)
	if !ok {
		t.Fatalf("Locate found no block in %q", withHyphens)
	}
	if got, want := string(span.Slice(withHyphens, region.Outer)), "---\ntitle: a---b\nstatus: [unclosed\n---"; got != want {
		t.Errorf("Outer slices to %q, want %q", got, want)
	}
	if got, want := string(span.Slice(withHyphens, region.Inner)), "\ntitle: a---b\nstatus: [unclosed"; got != want {
		t.Errorf("Inner slices to %q, want %q", got, want)
	}

	// A source with no block reports not-found.
	if _, ok := frontmatter.Locate([]byte("# Just a heading\n\nbody\n")); ok {
		t.Error("Locate found a block in a source that carries none")
	}

	// A block that does not begin at the first byte is not frontmatter.
	if _, ok := frontmatter.Locate([]byte("intro\n\n---\ntitle: One\n---\n")); ok {
		t.Error("Locate found a block that does not open at the first byte")
	}

	// An opened block that never closes is not located.
	if _, ok := frontmatter.Locate([]byte("---\ntitle: One\n")); ok {
		t.Error("Locate found a block that is never closed")
	}
}

// Parse and Locate share one implementation, so a block Parse reads and
// one Locate reports occupy the same region.
func TestParseAndLocateAgreeOnTheRegion(t *testing.T) {
	src := []byte("---\ntitle: One\nslug: adr-001\n---\n\nbody\n")

	doc, err := frontmatter.Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if doc == nil {
		t.Fatal("Parse returned no document")
	}
	region, ok := frontmatter.Locate(src)
	if !ok {
		t.Fatal("Locate found no block")
	}
	if region != doc.Region {
		t.Errorf("Locate = %+v, Parse recorded %+v", region, doc.Region)
	}
}

// AC: AC-SPEC-003.7
func TestTopLevelScalarReadsALegibleValueFromABlockThatDoesNotParse(t *testing.T) {
	const broken = "title: [unclosed\n"
	cases := []struct {
		name  string
		src   string
		want  string
		found bool
	}{
		{"plain", "---\n" + broken + "slug: adr-012\n---\n", "adr-012", true},
		{"plain with a comment", "---\n" + broken + "slug: adr-012 # the decision\n---\n", "adr-012", true},
		{"double-quoted", "---\n" + broken + "slug: \"adr-012\"\n---\n", "adr-012", true},
		{"single-quoted", "---\n" + broken + "slug: 'adr-012'\n---\n", "adr-012", true},
		{"single-quoted with an escaped quote", "---\nslug: 'it''s'\n---\n", "it's", true},
		{"CRLF", "---\r\n" + "slug: adr-012\r\n---\r\n", "adr-012", true},
		{"indented key ignored", "---\n" + broken + "  slug: adr-012\n---\n", "", false},
		{"absent key", "---\n" + broken + "type: adr\n---\n", "", false},
		{"a longer key sharing the prefix", "---\nslugs: adr-012\n---\n", "", false},
		{"empty value", "---\nslug: \"\"\n---\n", "", false},
		{"no value", "---\nslug:\n---\n", "", false},
		{"unclosed quote", "---\nslug: \"adr-012\n---\n", "", false},
		{"a collection", "---\nslug: [adr-012]\n---\n", "", false},
		{"unterminated block", "---\n" + broken + "slug: adr-012\n", "", false},
		{"no block", "slug: adr-012\n", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := frontmatter.TopLevelScalar([]byte(tc.src), "slug")
			if got != tc.want || found != tc.found {
				t.Errorf("TopLevelScalar = (%q, %v), want (%q, %v)", got, found, tc.want, tc.found)
			}
		})
	}
}
