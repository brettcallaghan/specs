// Tests for the `links` subcommand that the snapshot cases under
// cli/testdata/cases/ cannot express: the harness in
// cli/cmd/snapshot_test.go compares standard output, standard error and
// exit status, but has no notion of "no file on disk changed" —
// AC-SPECMD-005.3's own guarantee.
package cmd_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
)

// writeLinksFixture materializes a minimal corpus — one ADR citing
// another by wikilink — under dir, so a test can run `links` against it
// without depending on any other case's fixture.
func writeLinksFixture(t *testing.T, dir string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(dir, "specs.toml"), `[workspace]
name = "acme"
root = "docs"

[profile]
name = "specs"
`)
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-900-old-approach.md"), `---
title: "ADR-900: Old Approach"
description: A decision later retired in favor of ADR-901.
slug: adr-900-old-approach
type: adr
status: retired
superseded_by: adr-901-new-approach
---

# ADR-900: Old Approach
`)
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-901-new-approach.md"), `---
title: "ADR-901: New Approach"
description: The decision that replaces ADR-900.
slug: adr-901-new-approach
type: adr
status: approved
---

# ADR-901: New Approach
`)
	mustWriteFile(t, filepath.Join(dir, "docs", "adrs", "adr-800-current.md"), `---
title: "ADR-800: Current Decision"
description: A live decision that still cites the retired one.
slug: adr-800-current
type: adr
status: approved
---

# ADR-800: Current Decision

This decision refines [[adr-900-old-approach]], which has since been
replaced.
`)
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// treeDigest walks dir and returns, for every regular file beneath it,
// its path relative to dir mapped to a hash of its content — a
// fingerprint precise enough that a file created, deleted, or modified
// (including in place, at the same size) is detected.
func treeDigest(t *testing.T, dir string) map[string]string {
	t.Helper()
	digest := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(content)
		digest[rel] = hex.EncodeToString(sum[:])
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	return digest
}

// AC: AC-SPECMD-005.3
//
// links opens the corpus warm exactly as check, refs, slug-exists and
// next-id do (this bead's own Design), but unlike them it may not
// persist a freshly built index: those four subcommands' shared
// openCorpus helper writes the index back to disk on a cold cache
// (specs.json, spec-cli-container.md's "Data" section), and doing that
// here would create a file on a run that AC-SPECMD-005.3 requires create
// none. This test runs on a cold cache specifically, so it would catch a
// regression back onto the shared helper.
func TestLinksCreatesModifiesAndDeletesNoFile(t *testing.T) {
	dir := t.TempDir()
	writeLinksFixture(t, dir)

	before := treeDigest(t, dir)

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%s): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("restoring the working directory: %v", err)
		}
	})

	var stdout, stderr bytes.Buffer
	exitCode := cmd.Execute([]string{"links", "docs/adrs/adr-800-current.md"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstderr: %s", exitCode, stderr.String())
	}
	if stdout.Len() == 0 {
		t.Fatal("stdout is empty; the rewrite did not run")
	}

	after := treeDigest(t, dir)

	if len(after) != len(before) {
		t.Fatalf("file count changed: before = %d, after = %d (before = %v, after = %v)",
			len(before), len(after), before, after)
	}
	for path, sum := range before {
		if after[path] != sum {
			t.Errorf("%s changed on disk (before = %s, after = %s)", path, sum, after[path])
		}
	}
}
