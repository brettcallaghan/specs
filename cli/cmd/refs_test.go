// Tests for `refs`'s two views that the snapshot cases under
// cli/testdata/cases/ cannot express: a case compares one run's bytes,
// and AC-SPECREF-003.8 is a claim relating two runs — the compact answer
// leaves out no reference the full answer carries.
package cmd_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeRefsFixture materializes a corpus holding, between them, every
// kind of reference the compact view treats differently: a body
// reference, a frontmatter reference, an outbound reference that does not
// resolve, and an anchor matching more than one heading.
func writeRefsFixture(t *testing.T, dir string) {
	t.Helper()

	adr := func(slug, title, status, extra, body string) string {
		return fmt.Sprintf("---\ntitle: %q\ndescription: A decision used by the refs view test.\nslug: %s\n"+
			"type: adr\nstatus: %s\n%s---\n\n# %s\n\n%s",
			title, slug, status, extra, title, body)
	}

	mustWriteFile(t, filepath.Join(dir, "specs.toml"), `[workspace]
name = "acme"
root = "docs"

[profile]
name = "specs"
`)
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-900-old-approach.md"), adr(
		"adr-900-old-approach", "ADR-900: Old Approach", "retired",
		"superseded_by: adr-901-new-approach\n",
		"Replaced by [[adr-901-new-approach]].\n\n## Decision\n\nThe approach.\n\n### Decision\n\nThe narrower one.\n"))
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-901-new-approach.md"), adr(
		"adr-901-new-approach", "ADR-901: New Approach", "approved", "",
		"This refines [[adr-900-old-approach#Decision]].\n\nIt also leans on [[adr-404-never-written]].\n"))
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-800-current.md"), adr(
		"adr-800-current", "ADR-800: Current", "approved", "",
		"Still cites [[adr-900-old-approach]].\n\nAnd again, [[adr-900-old-approach]].\n"))
}

// refsTriple is what AC-SPECREF-003.8 compares: which document, which
// line, and what was written.
type refsTriple struct {
	Path string
	Line uint64
	Text string
}

// fullTriples reads the triples out of a --full --json answer.
func fullTriples(t *testing.T, out string) []refsTriple {
	t.Helper()
	var answer struct {
		References []struct {
			Path string `json:"path"`
			Text string `json:"text"`
			Span struct {
				StartLine uint64 `json:"start_line"`
			} `json:"span"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(out), &answer); err != nil {
		t.Fatalf("decoding the full answer: %v\n%s", err, out)
	}
	triples := []refsTriple{}
	for _, r := range answer.References {
		triples = append(triples, refsTriple{r.Path, r.Span.StartLine, r.Text})
	}
	return triples
}

// compactAnswer is the compact view's JSON shape.
type compactAnswer struct {
	References []struct {
		Path       string                   `json:"path"`
		References []map[string]interface{} `json:"references"`
	} `json:"references"`
}

// compactTriples reads the triples out of a compact --json answer.
func compactTriples(t *testing.T, answer compactAnswer) []refsTriple {
	t.Helper()
	triples := []refsTriple{}
	for _, document := range answer.References {
		for _, r := range document.References {
			line, _ := r["line"].(float64)
			text, _ := r["text"].(string)
			triples = append(triples, refsTriple{document.Path, uint64(line), text})
		}
	}
	return triples
}

// AC: AC-SPECREF-003.6, AC-SPECREF-003.7, AC-SPECREF-003.8, AC-SPECANCH-003.4
//
// The compact answer to each query carries the same references as the
// full answer — the same count and the same (path, line, text) triples —
// and keeps only the fields the standard's compact-view table keeps:
// origin only for a frontmatter reference, resolution only, as a bare
// state, for a reference that did not resolve, and ambiguous_anchor where
// an anchor matched more than one heading.
func TestTheCompactRefsAnswerLeavesOutNoReferenceTheFullAnswerCarries(t *testing.T) {
	dir := t.TempDir()
	writeRefsFixture(t, dir)

	queries := [][]string{
		{"adr-900-old-approach"},
		{"adr-900-old-approach#Decision"},
		{"adr-901-new-approach"},
	}
	var sawFrontmatter, sawAmbiguous bool
	for _, query := range queries {
		t.Run(strings.Join(query, " "), func(t *testing.T) {
			args := append([]string{"refs"}, query...)

			code, full, stderr := runInFixture(t, dir, append(append([]string{}, args...), "--full", "--json"))
			if code != 0 {
				t.Fatalf("--full exited %d: %s", code, stderr)
			}
			code, compact, stderr := runInFixture(t, dir, append(append([]string{}, args...), "--json"))
			if code != 0 {
				t.Fatalf("compact exited %d: %s", code, stderr)
			}

			var answer compactAnswer
			if err := json.Unmarshal([]byte(compact), &answer); err != nil {
				t.Fatalf("decoding the compact answer: %v\n%s", err, compact)
			}
			want, got := fullTriples(t, full), compactTriples(t, answer)
			if len(want) == 0 {
				t.Fatal("the full answer carries no reference; the fixture no longer exercises this query")
			}
			// Grouping by document reorders references across documents
			// but never within one, and the full answer's order is the
			// corpus's, so the triples compare as sets of per-document
			// sequences, grouped by document.
			if !reflect.DeepEqual(byDocument(want), byDocument(got)) {
				t.Errorf("compact triples differ from full:\n full    %v\n compact %v", want, got)
			}

			seen := map[string]bool{}
			for _, document := range answer.References {
				if seen[document.Path] {
					t.Errorf("document %s has more than one entry", document.Path)
				}
				seen[document.Path] = true
				for _, r := range document.References {
					for key := range r {
						switch key {
						case "line", "text":
						case "origin":
							sawFrontmatter = true
							if kind := r[key].(map[string]interface{})["kind"]; kind == "body" {
								t.Errorf("origin written for a body reference: %v", r)
							}
						case "ambiguous_anchor":
							sawAmbiguous = true
							if r[key] != true {
								t.Errorf("ambiguous_anchor written where false: %v", r)
							}
						default:
							t.Errorf("compact reference carries %q, which the view drops: %v", key, r)
						}
					}
				}
			}
		})
	}
	if !sawFrontmatter || !sawAmbiguous {
		t.Errorf("fixture coverage: frontmatter=%v ambiguous=%v, want both true",
			sawFrontmatter, sawAmbiguous)
	}
}

// AC: AC-SPECREF-003.8, AC-SPECREF-003.9
//
// The outbound compact answer groups by cited slug: one entry per slug,
// in first-cited order, carrying the resolved path (null where it
// resolved to nothing) and every line the full answer carries for it, and
// a resolution state only for the slug that did not resolve.
func TestTheOutboundCompactRefsAnswerGroupsByCitedSlug(t *testing.T) {
	dir := t.TempDir()
	writeRefsFixture(t, dir)

	args := []string{"refs", "adr-901-new-approach", "--outbound", "--json"}
	code, full, stderr := runInFixture(t, dir, append(append([]string{}, args...), "--full"))
	if code != 0 {
		t.Fatalf("--full exited %d: %s", code, stderr)
	}
	code, compact, stderr := runInFixture(t, dir, args)
	if code != 0 {
		t.Fatalf("compact exited %d: %s", code, stderr)
	}

	var answer struct {
		References []struct {
			Slug       string                   `json:"slug"`
			Path       *string                  `json:"path"`
			Resolution map[string]string        `json:"resolution"`
			Lines      []map[string]interface{} `json:"lines"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(compact), &answer); err != nil {
		t.Fatalf("decoding the compact answer: %v\n%s", err, compact)
	}
	if !strings.Contains(compact, `"path":null`) {
		t.Errorf("an unresolved slug must render path as an explicit null: %s", compact)
	}

	lines := 0
	seen := map[string]bool{}
	for _, entry := range answer.References {
		if seen[entry.Slug] {
			t.Errorf("slug %s has more than one entry", entry.Slug)
		}
		seen[entry.Slug] = true
		lines += len(entry.Lines)
		unresolved := entry.Slug == "adr-404-never-written"
		if unresolved != (entry.Path == nil) {
			t.Errorf("slug %s: path %v, want null exactly for the unresolved slug", entry.Slug, entry.Path)
		}
		if unresolved != (entry.Resolution != nil) {
			t.Errorf("slug %s: resolution %v, want it exactly for the unresolved slug", entry.Slug, entry.Resolution)
		}
	}
	if want := len(fullTriples(t, full)); lines != want || want == 0 {
		t.Errorf("compact carries %d lines, the full answer %d", lines, want)
	}
	if len(answer.References) != 2 {
		t.Errorf("got %d entries, want 2 (one per cited slug)", len(answer.References))
	}
}

// AC: AC-SPECREF-003.9
//
// Slugs resolve case-insensitively, so two spellings of one cited slug
// are one entry, spelled as first cited, carrying both lines.
func TestTheOutboundCompactRefsAnswerFoldsCaseVariantsOfOneSlug(t *testing.T) {
	dir := t.TempDir()
	writeRefsFixture(t, dir)
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-902-case-variants.md"),
		"---\ntitle: \"ADR-902: Case Variants\"\ndescription: A decision used by the refs view test.\n"+
			"slug: adr-902-case-variants\ntype: adr\nstatus: approved\n---\n\n"+
			"# ADR-902: Case Variants\n\nBuilds on [[adr-901-new-approach]].\n\nAs does [[ADR-901-New-Approach]].\n")

	code, compact, stderr := runInFixture(t, dir, []string{"refs", "adr-902-case-variants", "--outbound", "--json"})
	if code != 0 {
		t.Fatalf("compact exited %d: %s", code, stderr)
	}
	var answer struct {
		References []struct {
			Slug  string                   `json:"slug"`
			Path  *string                  `json:"path"`
			Lines []map[string]interface{} `json:"lines"`
		} `json:"references"`
	}
	if err := json.Unmarshal([]byte(compact), &answer); err != nil {
		t.Fatalf("decoding the compact answer: %v\n%s", err, compact)
	}
	if len(answer.References) != 1 {
		t.Fatalf("got %d entries, want 1 for two spellings of one slug: %s", len(answer.References), compact)
	}
	entry := answer.References[0]
	if entry.Slug != "adr-901-new-approach" || entry.Path == nil || len(entry.Lines) != 2 {
		t.Errorf("entry %s path %v with %d lines, want adr-901-new-approach resolved with 2 lines",
			entry.Slug, entry.Path, len(entry.Lines))
	}
}

// byDocument groups triples by path, each group in its original order.
func byDocument(triples []refsTriple) map[string][]refsTriple {
	grouped := map[string][]refsTriple{}
	for _, triple := range triples {
		grouped[triple.Path] = append(grouped[triple.Path], triple)
	}
	return grouped
}

// AC: AC-SPECREF-003.7
//
// --json changes the syntax of whichever view was chosen and never the
// view: the YAML and JSON renderings of one view carry the same top-level
// keys, and the default is the compact view under either.
func TestJSONChangesTheSyntaxOfARefsViewAndNeverTheView(t *testing.T) {
	dir := t.TempDir()
	writeRefsFixture(t, dir)

	for _, view := range [][]string{{}, {"--full"}} {
		args := append([]string{"refs", "adr-900-old-approach"}, view...)
		_, yaml, _ := runInFixture(t, dir, args)
		_, asJSON, _ := runInFixture(t, dir, append(append([]string{}, args...), "--json"))

		hasSpanYAML := strings.Contains(yaml, "start_line:")
		hasSpanJSON := strings.Contains(asJSON, `"start_line"`)
		wantSpan := len(view) == 1
		if hasSpanYAML != wantSpan || hasSpanJSON != wantSpan {
			t.Errorf("view %v: span in YAML=%v JSON=%v, want %v", view, hasSpanYAML, hasSpanJSON, wantSpan)
		}
	}
}
