// Tests for substituting resolved targets into a document's bytes.
package mdlinks_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/mdlinks"
	"github.com/brettcallaghan/agent-plugins/cli/core/references"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// bodyRef builds a body-origin reference over src, locating text's own
// occurrence in src to derive Position the way references.Extract would:
// the span between the brackets, not the brackets themselves.
func bodyRef(t *testing.T, src, text string) specmodel.Reference {
	t.Helper()
	needle := "[[" + text + "]]"
	at := strings.Index(src, needle)
	if at < 0 {
		t.Fatalf("fixture %q does not contain %q", src, needle)
	}
	slug, anchor, anchored := strings.Cut(text, "#")

	ref := specmodel.Reference{
		Path:   "specs/adrs/adr-011.md",
		Text:   text,
		Slug:   slug,
		Origin: specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{}),
		Span:   span.Of([]byte(src), at+2, len(text)),
	}
	if anchored {
		ref.Anchor = &anchor
	}
	return ref
}

// frontmatterRef builds a frontmatter-origin reference, for tests proving
// Substitute leaves it alone.
func frontmatterRef(t *testing.T, src, text string) specmodel.Reference {
	t.Helper()
	at := strings.Index(src, text)
	if at < 0 {
		t.Fatalf("fixture %q does not contain %q", src, text)
	}
	return specmodel.Reference{
		Path:   "specs/adrs/adr-011.md",
		Text:   text,
		Slug:   text,
		Origin: specmodel.NewReferenceOriginFrontmatterProperty(specmodel.ReferenceFrontmatterOrigin{Property: "sources"}),
		Span:   span.Of([]byte(src), at, len(text)),
	}
}

func ptr(s string) *string { return &s }

// AC: AC-SPECMD-001.1
func TestATargetableWikilinkBecomesAMarkdownLink(t *testing.T) {
	src := "See [[adr-012]] for context."
	refs := []specmodel.Reference{bodyRef(t, src, "adr-012")}
	targets := []*string{ptr("adr-012.md")}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "See [adr-012](adr-012.md) for context."
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

// AC: AC-SPECMD-001.2
func TestAnAnchorIsKeptInTheEmittedLinksVisibleText(t *testing.T) {
	src := "See [[container-api#Request Router]] for context."
	refs := []specmodel.Reference{bodyRef(t, src, "container-api#Request Router")}
	targets := []*string{ptr("container-api.md#request-router")}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "See [container-api#Request Router](container-api.md#request-router) for context."
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

// AC: AC-SPECMD-001.4
//
// This exercises the real extractor rather than a hand-built reference
// list: the Design prohibits reimplementing code-region exclusion here,
// so the property to prove is that References never hands Substitute a
// reference for a wikilink written inside a code span, an indented block,
// or a fenced block — leaving nothing for Substitute to do there.
func TestAWikilinkInsideACodeRegionIsNeverRewritten(t *testing.T) {
	src := "Inline: `[[adr-012]]`.\n\n" +
		"Fenced:\n\n```\n[[adr-013]]\n```\n\n" +
		"Indented:\n\n    [[adr-014]]\n"

	refs := references.Extract(references.Options{Path: "specs/adrs/adr-011.md", Source: []byte(src)})
	if len(refs) != 0 {
		t.Fatalf("Extract found %d references in code regions, want 0: %#v", len(refs), refs)
	}

	out, count := mdlinks.Substitute([]byte(src), refs, nil)

	if string(out) != src {
		t.Fatalf("out = %q, want the input unchanged", out)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

// AC: AC-SPECMD-001.5
func TestNothingButTheWikilinksIsAltered(t *testing.T) {
	src := "Intro prose with an ém dash — and a résumé.\n\n" +
		"See [[adr-012]] and also [[no-such-slug]] here.\n\n" +
		"Trailing prose, unchanged: 你好."
	refs := []specmodel.Reference{
		bodyRef(t, src, "adr-012"),
		bodyRef(t, src, "no-such-slug"),
	}
	targets := []*string{ptr("adr-012.md"), nil}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "Intro prose with an ém dash — and a résumé.\n\n" +
		"See [adr-012](adr-012.md) and also [[no-such-slug]] here.\n\n" +
		"Trailing prose, unchanged: 你好."
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

// AC: AC-SPECMD-001.6
func TestNoRewritableWikilinkEmitsTheDocumentUnchanged(t *testing.T) {
	src := "Nothing to rewrite here, just prose."

	out, count := mdlinks.Substitute([]byte(src), nil, nil)

	if string(out) != src {
		t.Fatalf("out = %q, want the input unchanged", out)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

// AC: AC-SPECMD-002.7
func TestAnUnresolvableWikilinkIsLeftExactlyAsWritten(t *testing.T) {
	src := "See [[no-such-slug]] for context."
	refs := []specmodel.Reference{bodyRef(t, src, "no-such-slug")}
	targets := []*string{nil}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	if string(out) != src {
		t.Fatalf("out = %q, want the input unchanged", out)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

// AC: AC-SPECMD-004.1
//
// Extraction, not Substitute, is what makes this a no-op: a document
// whose wikilinks have already become markdown links contains no `[[...]]`
// construct for the real extractor to find, so Substitute is called with
// no references at all — nothing here special-cases a second run.
func TestAnAlreadyRewrittenDocumentComesOutUnchanged(t *testing.T) {
	src := "See [adr-012](adr-012.md) for context."

	refs := references.Extract(references.Options{Path: "specs/adrs/adr-011.md", Source: []byte(src)})
	out, count := mdlinks.Substitute([]byte(src), refs, nil)

	if string(out) != src {
		t.Fatalf("out = %q, want the input unchanged", out)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0", count)
	}
}

// AC: AC-SPECMD-004.2
func TestAnExistingMarkdownLinkIsNeverRewritten(t *testing.T) {
	src := "Mixed: [adr-012](adr-012.md) and [[adr-013]]."

	refs := references.Extract(references.Options{Path: "specs/adrs/adr-011.md", Source: []byte(src)})
	if len(refs) != 1 {
		t.Fatalf("Extract found %d references, want exactly 1 (the wikilink only): %#v", len(refs), refs)
	}
	targets := []*string{ptr("adr-013.md")}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "Mixed: [adr-012](adr-012.md) and [adr-013](adr-013.md)."
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

// A frontmatter-origin reference has no `[[...]]` construct — its
// Position covers the property's raw YAML value — so Substitute must
// leave it untouched even when it settled a target, rather than treating
// its Position as bracket-delimited and corrupting the frontmatter block.
func TestAFrontmatterOriginReferenceIsLeftUntouched(t *testing.T) {
	src := "---\nsources: adr-012\n---\n\nSee [[adr-013]] for context.\n"
	refs := []specmodel.Reference{
		frontmatterRef(t, src, "adr-012"),
		bodyRef(t, src, "adr-013"),
	}
	targets := []*string{ptr("adr-012.md"), ptr("adr-013.md")}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "---\nsources: adr-012\n---\n\nSee [adr-013](adr-013.md) for context.\n"
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 — the frontmatter reference must not be counted", count)
	}
}

// The count reported spans a mix of replaced, unresolved and
// frontmatter-origin references in one call, proving the count is exactly
// the number of markdown links actually emitted rather than the number of
// references seen.
func TestReplacedCountReflectsOnlyActualSubstitutions(t *testing.T) {
	src := "[[adr-012]], [[adr-013]], [[no-such-slug]]."
	refs := []specmodel.Reference{
		bodyRef(t, src, "adr-012"),
		bodyRef(t, src, "adr-013"),
		bodyRef(t, src, "no-such-slug"),
	}
	targets := []*string{ptr("adr-012.md"), ptr("adr-013.md"), nil}

	out, count := mdlinks.Substitute([]byte(src), refs, targets)

	want := "[adr-012](adr-012.md), [adr-013](adr-013.md), [[no-such-slug]]."
	if string(out) != want {
		t.Fatalf("out = %q, want %q", out, want)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}
