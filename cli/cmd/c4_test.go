// Tests for `c4` that the snapshot cases under cli/testdata/cases/ cannot
// express: the harness there compares what ONE run produced, and
// AC-C4-004.4 is a claim about a SECOND run. The first run builds and
// persists the index; the second reads it warm.
package cmd_test

import (
	"os"
	"path/filepath"
	"testing"
)

// AC: AC-C4-004.4
func TestTwoC4RunsOverAnUnchangedCorpusEmitTheSameBytes(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "specs.toml"), `[workspace]
name = "acme"
root = "docs"

[profile]
name = "spec-driven-dev"
`)
	for _, doc := range []struct{ path, slug, kind, extra string }{
		{"docs/systems/two.md", "system-two", "system-blueprint", ""},
		{"docs/systems/one.md", "system-one", "system-blueprint", ""},
		{"docs/containers/beta.md", "container-beta", "container-blueprint", "system: system-one\n"},
		{"docs/containers/alpha.md", "container-alpha", "container-blueprint", "system: system-one\n"},
	} {
		mustWriteFile(t, filepath.Join(dir, filepath.FromSlash(doc.path)), "---\ntitle: \""+doc.slug+"\"\n"+
			"description: A fixture for the c4 determinism test.\nslug: "+doc.slug+"\ntype: "+doc.kind+
			"\nstatus: approved\nupdated_at: 2026-10-07T00:00:00Z\nentity_name: "+doc.slug+"\n"+doc.extra+"---\n\n# "+doc.slug+"\n")
	}

	for _, args := range [][]string{{"c4"}, {"c4", "--json"}} {
		_ = os.Remove(filepath.Join(dir, "docs", "specs.json"))
		firstCode, first, firstErr := runInFixture(t, dir, args)
		if firstCode != 0 {
			t.Fatalf("%v: the first run exited %d\n--- stderr ---\n%s", args, firstCode, firstErr)
		}
		if _, err := os.Stat(filepath.Join(dir, "docs", "specs.json")); err != nil {
			t.Fatalf("%v: the first run persisted no index, so the second is not the warm path: %v", args, err)
		}
		secondCode, second, secondErr := runInFixture(t, dir, args)
		if secondCode != 0 {
			t.Fatalf("%v: the second run exited %d\n--- stderr ---\n%s", args, secondCode, secondErr)
		}
		if first == "" || first != second {
			t.Errorf("%v: two runs over an unchanged corpus differ.\n--- first ---\n%s\n--- second ---\n%s", args, first, second)
		}
	}
}
