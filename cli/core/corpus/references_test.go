// Tests for the reference graph the index carries: what the walk
// extracts, what resolution records, and the two directions a built or
// warm corpus answers from.
//
// The fixture profile and document renderer are corpus_test.go's, so
// these tests read the same corpus the discovery tests do. Where a test
// needs a reference-bearing frontmatter property, it registers one on a
// copy of that profile rather than on the shared one: what a property
// holds is a per-type registration, and a fixture that changed it for
// every test in the package would be changing what the other tests are
// about.
package corpus_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// citing renders a conforming document whose body is the given text, for
// the tests that need a document to reference something.
func citing(docType, slug, title, body string) string {
	return fmt.Sprintf(`---
title: %q
description: "A fixture document."
slug: %s
type: %s
status: approved
updated_at: "2026-08-05T00:00:00Z"
---

# %s

%s
`, title, slug, docType, title, body)
}

// referencingProfile is the fixture profile with one frontmatter property
// registered as reference-bearing for both of its types.
func referencingProfile(property string) corpus.Profile {
	types := map[string]corpus.DocumentTypeRules{}
	for name, rules := range indexProfile().DocumentTypes() {
		rules.ReferenceProperties = []string{property}
		types[name] = rules
	}
	return stubProfile{schema: []byte(`{"definitions":{}}`), types: types}
}

// referenceSlugs returns the slug half of each reference, in order.
func referenceSlugs(found []specmodel.Reference) []string {
	out := make([]string, 0, len(found))
	for _, reference := range found {
		out = append(out, reference.Slug)
	}
	return out
}

// targetPath returns the path one reference resolved to, or "" where it
// resolved to something else.
func targetPath(reference specmodel.Reference) string {
	if resolved, ok := reference.Resolution.AsResolved(); ok {
		return resolved.TargetPath
	}
	return ""
}

// The index carries every reference its members make, resolved, ordered
// by the path they occur in and then by their position within it — one
// rule that serves both a document's own ordering and byte-identical
// output across runs.
//
// AC: AC-SPECREF-001.5
func TestTheIndexCarriesEveryReferenceOrderedByPathThenPosition(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/zeta.md", citing("frd", "frd-zeta", "Zeta",
		"Refines [[adr-002]], then [[adr-001]], then [[adr-002]] again."))
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Cites [[frd-zeta]]."))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	built := corpusAt(t, repo)
	index := built.Index()

	type row struct{ path, slug string }
	var got []row
	for _, reference := range index.References {
		got = append(got, row{reference.Path, reference.Slug})
	}
	want := []row{
		{"specs/adrs/adr-001.md", "frd-zeta"},
		{"specs/zeta.md", "adr-002"},
		{"specs/zeta.md", "adr-001"},
		{"specs/zeta.md", "adr-002"},
	}
	if len(got) != len(want) {
		t.Fatalf("references = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("reference %d = %+v, want %+v", i, got[i], want[i])
		}
	}

	// Every row resolved, and each to the document declaring the slug it
	// names rather than to whichever file happened to be walked first.
	for _, reference := range index.References {
		if target := targetPath(reference); !strings.HasSuffix(target, reference.Slug+".md") &&
			!strings.HasSuffix(target, "zeta.md") {
			t.Errorf("[[%s]] in %s resolved to %q", reference.Slug, reference.Path, target)
		}
	}

	// The emitted bytes are a function of corpus content alone: a second
	// build of the same files renders identically.
	first, second := rendered(t, index), rendered(t, corpusAt(t, repo).Index())
	if string(first) != string(second) {
		t.Errorf("two builds of one corpus emit different indexes:\n%s\n%s", first, second)
	}
}

// Inbound references are derived for every document from the same
// collection the outbound ones are read from, so an edge is stored once
// and read from either end.
//
// AC: AC-SPECREF-001.3
func TestInboundReferencesAreDerivedFromTheSameCollection(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-003]]."))
	write(t, repo, "specs/adrs/adr-002.md", citing("adr", "adr-002", "Second", "Also refines [[adr-003]]."))
	write(t, repo, "specs/adrs/adr-003.md", document("adr", "adr-003", "Third"))

	built := corpusAt(t, repo)

	inbound := built.InboundReferences("specs/adrs/adr-003.md")
	if len(inbound) != 2 {
		t.Fatalf("inbound = %+v, want two", inbound)
	}
	// Each inbound reference is cited at its position in the *referring*
	// document, because that is the file a reader opens to change it.
	wantReferrers := []string{"specs/adrs/adr-001.md", "specs/adrs/adr-002.md"}
	for i, reference := range inbound {
		if reference.Path != wantReferrers[i] {
			t.Errorf("inbound %d comes from %q, want %q", i, reference.Path, wantReferrers[i])
		}
		if reference.Position.Length == 0 {
			t.Errorf("inbound %d carries an empty span", i)
		}
	}

	if got := referenceSlugs(built.OutboundReferences("specs/adrs/adr-001.md")); !equal(got, []string{"adr-003"}) {
		t.Errorf("outbound of adr-001 = %v, want [adr-003]", got)
	}
	if got := built.OutboundReferences("specs/adrs/adr-003.md"); len(got) != 0 {
		t.Errorf("outbound of adr-003 = %+v, want none: it references nothing", got)
	}
	if got := built.InboundReferences("specs/adrs/adr-001.md"); len(got) != 0 {
		t.Errorf("inbound of adr-001 = %+v, want none: nothing references it", got)
	}

	// No edge is stored twice: the index carries one row per occurrence,
	// and both directions read those same rows.
	if got := len(built.Index().References); got != 2 {
		t.Fatalf("the index carries %d reference rows, want 2 — the reverse edges are stored, not derived", got)
	}
	outbound := built.OutboundReferences("specs/adrs/adr-002.md")
	if len(outbound) != 1 || outbound[0] != inbound[1] {
		t.Errorf("the outbound and inbound views of one edge differ:\n%+v\n%+v", outbound, inbound)
	}
}

// A reference that resolves to no corpus member is recorded with its
// unresolved state rather than dropped, and the index is complete: the
// indexer describes the corpus it found, including the broken parts of
// it, and reporting them is a check's job.
func TestAnUnresolvedReferenceIsRecordedAndTheIndexIsComplete(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First",
		"Refines [[adr-404]] and [[adr-002]]."))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))
	write(t, repo, "specs/notes/loose.md", "# No frontmatter\n")

	built := corpusAt(t, repo)
	index := built.Index()

	if len(index.Entries) != 2 || len(index.Exclusions) != 1 {
		t.Fatalf("the index is incomplete: %d entries, %d exclusions", len(index.Entries), len(index.Exclusions))
	}
	if len(index.References) != 2 {
		t.Fatalf("references = %+v, want both recorded", index.References)
	}
	if _, ok := index.References[0].Resolution.AsUnresolved(); !ok {
		t.Errorf("[[adr-404]] resolved to %+v, want unresolved", index.References[0].Resolution.Value)
	}
	if got := targetPath(index.References[1]); got != "specs/adrs/adr-002.md" {
		t.Errorf("[[adr-002]] resolved to %q", got)
	}

	// An unresolved reference has one end, so it appears in the outbound
	// direction and in no inbound one.
	if got := len(built.OutboundReferences("specs/adrs/adr-001.md")); got != 2 {
		t.Errorf("outbound = %d, want both references", got)
	}
	if got := built.InboundReferences("specs/notes/loose.md"); len(got) != 0 {
		t.Errorf("an excluded file has inbound references: %+v", got)
	}

	// Nothing about a broken link is a finding at this stage: the index
	// records it, and no check reports it yet.
	if len(built.Findings()) != 0 {
		t.Errorf("findings = %+v, want none — indexing never fails", built.Findings())
	}
}

// A value in a frontmatter property the active profile registers as
// reference-bearing is an outbound reference of that document, resolved
// exactly as a body reference is and distinguishable from one.
func TestARegisteredFrontmatterPropertyContributesAResolvedReference(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", strings.Replace(
		document("adr", "adr-001", "First"),
		"updated_at: \"2026-08-05T00:00:00Z\"",
		"updated_at: \"2026-08-05T00:00:00Z\"\nsuperseded_by: adr-002",
		1))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	built, err := corpus.Build(corpus.Options{
		RepoRoot:  repo,
		Root:      filepath.Join(repo, "specs"),
		Workspace: "fixture-workspace",
		Profile:   referencingProfile("superseded_by"),
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	outbound := built.OutboundReferences("specs/adrs/adr-001.md")
	if len(outbound) != 1 {
		t.Fatalf("outbound = %+v, want the frontmatter reference", outbound)
	}
	origin, ok := outbound[0].Origin.AsFrontmatterProperty()
	if !ok {
		t.Fatalf("origin = %+v, want a frontmatter property", outbound[0].Origin.Value)
	}
	if origin.Property != "superseded_by" {
		t.Errorf("origin names %q", origin.Property)
	}
	if got := targetPath(outbound[0]); got != "specs/adrs/adr-002.md" {
		t.Errorf("resolved to %q", got)
	}
	if got := len(built.InboundReferences("specs/adrs/adr-002.md")); got != 1 {
		t.Errorf("inbound of adr-002 = %d, want the frontmatter edge", got)
	}

	// The same corpus under a profile registering nothing carries no
	// reference at all: what holds one is registered, never guessed.
	if got := corpusAt(t, repo).Index().References; len(got) != 0 {
		t.Errorf("references = %+v under a profile registering none", got)
	}
}

// A slug matching a file the index recorded as an exclusion or a parse
// failure records that outcome, with the file's path, rather than
// "matches nothing" — the document is right there, and its frontmatter is
// what is wrong.
//
// AC: AC-SPECREF-002.7
func TestAReferenceToANonMemberRecordsTheReasonAndPath(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First",
		"See [[loose]], [[broken]] and [[absent]]."))
	write(t, repo, "specs/notes/loose.md", "# No frontmatter at all\n")
	write(t, repo, "specs/notes/broken.md", "---\ntitle: \"unterminated\n---\n\n# Broken\n")

	built := corpusAt(t, repo)
	byslug := map[string]specmodel.Reference{}
	for _, reference := range built.Index().References {
		byslug[reference.Slug] = reference
	}

	loose, ok := byslug["loose"].Resolution.AsNonMember()
	if !ok {
		t.Fatalf("[[loose]] resolved to %+v, want a non-member", byslug["loose"].Resolution.Value)
	}
	if loose.Path != "specs/notes/loose.md" {
		t.Errorf("path = %q", loose.Path)
	}
	if loose.Reason == nil || *loose.Reason != specmodel.ExclusionReasonNoFrontmatter {
		t.Errorf("reason = %v, want no-frontmatter", loose.Reason)
	}

	broken, ok := byslug["broken"].Resolution.AsNonMember()
	if !ok {
		t.Fatalf("[[broken]] resolved to %+v, want a non-member", byslug["broken"].Resolution.Value)
	}
	if broken.Path != "specs/notes/broken.md" {
		t.Errorf("path = %q", broken.Path)
	}
	if broken.ParseError == nil {
		t.Error("the non-member carries no parse error, so the reader is told nothing about why")
	}

	if _, ok := byslug["absent"].Resolution.AsUnresolved(); !ok {
		t.Errorf("[[absent]] resolved to %+v, want unresolved", byslug["absent"].Resolution.Value)
	}
}

// A corpus served warm answers both directions from the persisted index
// without parsing a document, which is what keeps a reference query and
// the inbound half of a single-document check off the corpus walk.
//
// The proof is a row planted in the persisted index that no document on
// disk carries: a warm corpus that serves it cannot have derived it from
// the files.
//
// AC: AC-SPECREF-003.4
func TestAWarmCorpusServesTheGraphFromThePersistedIndex(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002]]."))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	opts := indexOptions(repo)
	dest := buildAndWrite(t, repo)

	persisted := readIndex(t, dest)
	if len(persisted.References) != 1 {
		t.Fatalf("the built index carries %d references, want 1", len(persisted.References))
	}
	persisted.References = append(persisted.References, specmodel.Reference{
		Path:     "specs/adrs/adr-002.md",
		Text:     "adr-001",
		Slug:     "adr-001",
		Origin:   specmodel.NewReferenceOriginBody(specmodel.ReferenceBodyOrigin{}),
		Position: persisted.References[0].Position,
		Resolution: specmodel.NewReferenceResolutionResolved(
			specmodel.ReferenceResolved{TargetPath: "specs/adrs/adr-001.md"}),
	})
	if err := corpus.Write(persisted, dest, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}

	opened, warm, err := corpus.Open(opts, dest)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !warm {
		t.Fatal("Open rebuilt an index whose inputs are unchanged")
	}

	outbound := opened.OutboundReferences("specs/adrs/adr-002.md")
	if len(outbound) != 1 || outbound[0].Slug != "adr-001" {
		t.Fatalf("outbound = %+v; the planted row is absent, so the graph was re-derived", outbound)
	}
	if got := len(opened.InboundReferences("specs/adrs/adr-001.md")); got != 1 {
		t.Errorf("inbound of adr-001 = %d, want the planted edge", got)
	}
	if got := len(opened.InboundReferences("specs/adrs/adr-002.md")); got != 1 {
		t.Errorf("inbound of adr-002 = %d, want the edge the build recorded", got)
	}
}

// The warm and cold paths are interchangeable: a corpus served warm
// answers both directions exactly as the build that wrote the index did.
func TestTheGraphIsIdenticalWarmAndCold(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002]], [[adr-404]]."))
	write(t, repo, "specs/adrs/adr-002.md", citing("adr", "adr-002", "Second", "Answers [[adr-001]]."))

	opts := indexOptions(repo)
	dest := buildAndWrite(t, repo)

	cold, err := corpus.Build(opts)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	warmed, warm, err := corpus.Open(opts, dest)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !warm {
		t.Fatal("Open rebuilt an index whose inputs are unchanged")
	}

	for _, path := range []string{"specs/adrs/adr-001.md", "specs/adrs/adr-002.md"} {
		for _, direction := range []struct {
			name string
			read func(*corpus.Corpus, string) []specmodel.Reference
		}{
			{"outbound", (*corpus.Corpus).OutboundReferences},
			{"inbound", (*corpus.Corpus).InboundReferences},
		} {
			a, err := json.Marshal(direction.read(cold, path))
			if err != nil {
				t.Fatalf("marshalling: %v", err)
			}
			b, err := json.Marshal(direction.read(warmed, path))
			if err != nil {
				t.Fatalf("marshalling: %v", err)
			}
			if string(a) != string(b) {
				t.Errorf("%s references of %s differ warm and cold:\n%s\n%s", direction.name, path, a, b)
			}
		}
	}
}

// Extraction and resolution ride the walk that already reads every
// candidate's bytes, so a corpus whose documents all reference one
// another is indexed inside the same budget as one that references
// nothing.
//
// The complement to this measurement is
// TestTheWalkOffersEachCandidateExactlyOnce, which holds the read side
// directly; together they are what "no document is read twice" means.
//
// AC: AC-SPEC-009.3, AC-SPEC-009.9
func TestAColdBuildWithAFullGraphMeetsTheSameBudget(t *testing.T) {
	const (
		documents = 1000
		target    = time.Second
	)

	repo := t.TempDir()
	for i := range documents {
		rel := fmt.Sprintf("specs/section-%02d/adr-%04d.md", i%20, i)
		body := fmt.Sprintf("Refines [[adr-%04d]] and [[adr-%04d]], and see [[adr-%04d]].",
			(i+1)%documents, (i+7)%documents, (i+13)%documents)
		write(t, repo, rel, citing("adr", fmt.Sprintf("adr-%04d", i), fmt.Sprintf("Document %d", i), body))
	}

	started := time.Now()
	built := corpusAt(t, repo)
	elapsed := time.Since(started)

	if got := len(built.Index().References); got != 3*documents {
		t.Fatalf("references = %d, want %d", got, 3*documents)
	}
	for _, reference := range built.Index().References {
		if _, ok := reference.Resolution.AsResolved(); !ok {
			t.Fatalf("[[%s]] in %s did not resolve", reference.Slug, reference.Path)
		}
	}
	reportLatency(t, fmt.Sprintf("cold build of %d documents with %d references", documents, 3*documents),
		elapsed, target)
}

// A path a caller names any of the three ways the other lookups accept
// reaches the same references, so the graph is queried the way Entry and
// Slug are rather than by a path spelled exactly as the index holds it.
func TestReferenceLookupsAcceptTheSamePathFormsTheOtherQueriesDo(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-001.md", citing("adr", "adr-001", "First", "Refines [[adr-002]]."))
	write(t, repo, "specs/adrs/adr-002.md", document("adr", "adr-002", "Second"))

	built := corpusAt(t, repo)
	absolute := filepath.Join(repo, "specs", "adrs", "adr-001.md")

	for _, form := range []string{"specs/adrs/adr-001.md", absolute} {
		if got := len(built.OutboundReferences(form)); got != 1 {
			t.Errorf("outbound for %q = %d, want 1", form, got)
		}
	}
	if _, found := built.Entry("specs/adrs/adr-001.md"); !found {
		t.Error("the fixture document is not an index entry")
	}
	if _, found := built.Slug("adr-001"); !found {
		t.Error("the fixture document declares no findable slug")
	}
}
