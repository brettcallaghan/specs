// Whitebox test of the claim that every `headings` match carries a ref
// that pastes back. It runs over this repository's own specs/ corpus,
// because the defect lived in documents whose H1 identifier differs from
// the member's slug, and a synthetic fixture would pick the shape that
// hides it.
package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
)

// AC: AC-SPECSTR-003.9, AC-SPECANCH-003.6
func TestEveryHeadingsMatchCarriesARefThatPastesBack(t *testing.T) {
	configPath, err := filepath.Abs("../../specs.toml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, configPath, findings, err := loadConfig(configPath)
	if err != nil || findings != nil {
		t.Fatalf("loadConfig = %v, %v", findings, err)
	}
	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		t.Fatal(err)
	}
	built, err := corpus.Build(opts)
	if err != nil {
		t.Fatal(err)
	}
	index := built.Index()

	matches, direct := selectHeadings(index, parseSelection(""), nil)
	if len(matches) == 0 {
		t.Fatal("the bare selection matched no heading of this repository's corpus")
	}

	var sample *headingMatch
	for i, match := range matches {
		if !direct[i] {
			t.Errorf("%s: the bare selection reported a match it did not name", match.Ref)
		}
		if !strings.HasSuffix(match.Ref, "#"+match.Anchor) {
			t.Errorf("ref %q does not end in #%s", match.Ref, match.Anchor)
		}
		// Either the ref selects this heading alone, or the --emit
		// refusal lists it as reachable only by span.
		if selectsAlone(index, match, nil) {
			if sample == nil && !strings.Contains(match.Ref, "/") {
				m := match
				sample = &m
			}
			continue
		}
		again, _ := selectHeadings(index, parseSelection(match.Ref), nil)
		found := false
		for _, m := range again {
			found = found || (m.Path == match.Path && m.Section == match.Section)
		}
		if !found {
			t.Errorf("ref %q does not even reach its own heading", match.Ref)
		}
	}
	if sample == nil {
		t.Fatal("no slug-addressed match selects its heading alone")
	}

	// refs <ref> resolves the same heading.
	asked := parseSelection(sample.Ref)
	entry, declared := built.Slug(asked.document)
	if !declared || entry.Path != sample.Path {
		t.Fatalf("refs %q: slug resolves to %q (declared %v), want %s", sample.Ref, entry.Path, declared, sample.Path)
	}
	if got := anchorDeclared(declared, asked.heading, entry.Headings); !*got {
		t.Errorf("refs %q: anchor_declared = false, want true", sample.Ref)
	}
}
