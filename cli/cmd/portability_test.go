// Cross-platform reproducibility: the end-to-end evidence for
// AC-SPEC-009.7 ("WHEN the CLI runs against identical corpus content on
// Linux, macOS, and Windows, the CLI SHALL report the same findings in
// the same order"), including its worked example — "a corpus checked out
// with CRLF line endings on Windows yields byte-identical JSON output to
// the same corpus on Linux".
//
// The claim has two halves, and this file is only one of them:
//
//   - *Across platforms.* The committed expected artifacts under
//     cli/testdata/cases/ are one set of bytes, authored once. The
//     .github/workflows/cli.yml matrix runs `go test ./...` on
//     ubuntu-latest, macos-latest and windows-latest against that one
//     set. Byte-identical output across the three is therefore asserted
//     by the ordinary snapshot suite, on each platform, rather than by
//     anything written here — there is no way to observe three operating
//     systems from inside one test process.
//
//   - *Across checkouts.* That matrix alone does not cover the example,
//     because .gitattributes pins the working tree to LF on every
//     platform, so the Windows leg reads the same bytes the Linux leg
//     does. Deliberately so: a repository that lets its own committed
//     expected artifacts arrive as CRLF cannot compare them to anything.
//     The example is instead reproduced here, synthetically and on every
//     platform: every snapshot case is replayed a second time from a
//     working tree whose files have had every LF rewritten to CRLF —
//     exactly what `git config core.autocrlf true` produces on Windows —
//     and held to the same committed artifacts.
//
// The two together are the claim: one set of expected bytes, three
// operating systems, both line-ending conventions.
package cmd_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestCRLFCheckoutProducesIdenticalOutput replays every snapshot case
// from a CRLF working tree and requires the same exit status and the
// same expected artifacts, byte for byte, as the LF run in
// TestSnapshotCases.
//
// What it is really testing is corpus.normalizeLineEndings: the CLI
// normalizes CRLF to LF at the single point where it reads a candidate
// file, so a checkout's line-ending convention reaches neither the byte
// offsets in a span nor the content hash the warm-cache path compares.
// Before that normalization existed this test failed on the spans and
// the input_hash of every index-producing case while the check reports
// passed — which is the shape of the bug it now guards: a divergence
// that a findings-only assertion would not have seen.
//
// AC: AC-SPEC-009.7
func TestCRLFCheckoutProducesIdenticalOutput(t *testing.T) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", casesDir, err)
	}

	ran := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ran++
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			runSnapshotCaseWith(t, filepath.Join(casesDir, name), copyTreeCRLF)
		})
	}

	// A zero-case run would report success while asserting nothing, and
	// this file's whole value is that it asserts something.
	if ran == 0 {
		t.Fatalf("no snapshot cases found under %s", casesDir)
	}
}

// TestCopyTreeCRLFRewritesLineEndings checks the fixture machinery
// itself: a CRLF run that quietly copied LF files through would pass for
// the wrong reason and retire the guarantee above without anyone
// noticing.
func TestCopyTreeCRLFRewritesLineEndings(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	const content = "---\ntitle: x\n---\n\n# x\n"
	if err := os.WriteFile(filepath.Join(src, "doc.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	if err := copyTreeCRLF(src, dst); err != nil {
		t.Fatalf("copyTreeCRLF: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "doc.md"))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if want := "---\r\ntitle: x\r\n---\r\n\r\n# x\r\n"; string(got) != want {
		t.Errorf("copyTreeCRLF produced %q, want %q", got, want)
	}
}

// TestCopyTreeCRLFIsIdempotent checks that a source already holding CRLF
// is copied unchanged rather than gaining a second CR per line. Nothing
// under cli/testdata/cases/ holds CRLF today, but a fixture that later
// did would otherwise be silently corrupted into a corpus no checkout
// ever produces.
func TestCopyTreeCRLFIsIdempotent(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	const content = "a\r\nb\n"
	if err := os.WriteFile(filepath.Join(src, "doc.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}
	if err := copyTreeCRLF(src, dst); err != nil {
		t.Fatalf("copyTreeCRLF: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "doc.md"))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if want := "a\r\nb\r\n"; string(got) != want {
		t.Errorf("copyTreeCRLF produced %q, want %q", got, want)
	}
}

// copyTreeCRLF is copyTree with every LF in every copied file rewritten
// to CRLF, reproducing what a Windows checkout under
// `core.autocrlf=true` puts on disk.
//
// It rewrites a lone LF only, so a source that already holds CRLF is
// copied unchanged rather than growing a second carriage return.
func copyTreeCRLF(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, toCRLF(data), 0o644)
	})
}

// toCRLF rewrites every line ending in src to CRLF, leaving a CRLF that
// is already there alone.
func toCRLF(src []byte) []byte {
	return bytes.ReplaceAll(normalizeToLF(src), []byte("\n"), []byte("\r\n"))
}

// normalizeToLF collapses CRLF to LF. It is the test's own copy of what
// the CLI does on read, kept here rather than imported so that a
// regression in the CLI's normalization cannot also silently change what
// this fixture produces.
func normalizeToLF(src []byte) []byte {
	return bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
}
