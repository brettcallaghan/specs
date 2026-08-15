// Internal tests for the properties of the walk that are not observable
// from outside the package.
package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

// Every source is offered to the builder exactly once, with its bytes,
// and each is offered as the kind of source it is.
//
// This is what "no document is read twice" rests on. [scanSources] is
// the only walk in the package, [Build] and [Fingerprint] both consume
// it, and reference and feature extraction are fed by the same offer that
// carries a file's bytes to classification — so a second read of any
// document would have to be a second walk, and a second walk would show
// up here as a second offer. That the feature sources ride this same walk
// is what keeps their bytes in the input hash.
func TestTheWalkOffersEachSourceExactlyOnce(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"adrs/adr-001.md":        "---\ntitle: One\n---\n\n[[adr-002]]\n",
		"adrs/nested/adr-002.md": "---\ntitle: Two\n---\n\n[[adr-001]]\n",
		"notes/loose.md":         "# No frontmatter\n",
		"notes/not-markdown.txt": "passed over silently\n",
		"features/one.feature":   "Feature: One\n",
		"features/UPPER.FEATURE": "Feature: Two\n",
	}
	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", full, err)
		}
	}

	offers := map[string]int{}
	kinds := map[string]sourceKind{}
	err := scanSources(root, root, func(kind sourceKind, reported string, src []byte, readErr error) {
		offers[reported]++
		kinds[reported] = kind
		if readErr != nil {
			t.Errorf("%s could not be read: %v", reported, readErr)
			return
		}
		if string(src) != files[reported] {
			t.Errorf("%s was offered %q", reported, src)
		}
	})
	if err != nil {
		t.Fatalf("scanSources: %v", err)
	}

	want := map[string]sourceKind{
		"adrs/adr-001.md":        markdownCandidate,
		"adrs/nested/adr-002.md": markdownCandidate,
		"notes/loose.md":         markdownCandidate,
		"features/one.feature":   featureSource,
		"features/UPPER.FEATURE": featureSource,
	}
	if len(offers) != len(want) {
		t.Fatalf("offers = %v, want one per source: %v", offers, want)
	}
	for reported, kind := range want {
		if offers[reported] != 1 {
			t.Errorf("%s was offered %d times, want exactly once", reported, offers[reported])
		}
		if kinds[reported] != kind {
			t.Errorf("%s was offered as kind %d, want %d", reported, kinds[reported], kind)
		}
	}
}
