// Internal tests for the properties of the walk that are not observable
// from outside the package.
package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

// Every candidate is offered to the builder exactly once, with its bytes.
//
// This is what "no document is read twice" rests on. [scanCandidates] is
// the only walk in the package, [Build] and [Fingerprint] both consume
// it, and reference extraction is fed by the same offer that carries a
// file's bytes to classification — so a second read of any document would
// have to be a second walk, and a second walk would show up here as a
// second offer.
func TestTheWalkOffersEachCandidateExactlyOnce(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"adrs/adr-001.md":        "---\ntitle: One\n---\n\n[[adr-002]]\n",
		"adrs/nested/adr-002.md": "---\ntitle: Two\n---\n\n[[adr-001]]\n",
		"notes/loose.md":         "# No frontmatter\n",
		"notes/not-markdown.txt": "passed over silently\n",
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
	err := scanCandidates(root, root, func(reported string, src []byte, readErr error) {
		offers[reported]++
		if readErr != nil {
			t.Errorf("%s could not be read: %v", reported, readErr)
			return
		}
		if string(src) != files[reported] {
			t.Errorf("%s was offered %q", reported, src)
		}
	})
	if err != nil {
		t.Fatalf("scanCandidates: %v", err)
	}

	want := []string{"adrs/adr-001.md", "adrs/nested/adr-002.md", "notes/loose.md"}
	if len(offers) != len(want) {
		t.Fatalf("offers = %v, want one per markdown candidate: %v", offers, want)
	}
	for _, reported := range want {
		if offers[reported] != 1 {
			t.Errorf("%s was offered %d times, want exactly once", reported, offers[reported])
		}
	}
}
