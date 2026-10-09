// Tests for `list` that the snapshot cases under cli/testdata/cases/
// cannot express: the harness there compares what ONE run produced, and
// AC-SPECLIST-004.3 is a claim about a SECOND run — that two invocations
// over an unchanged corpus emit the same bytes in the index's own order.
//
// The two runs are not equivalent behind the scenes, which is what makes
// the claim worth proving here. The first has no persisted index and
// builds one, and `openCorpus` writes it; the second reads that index
// warm. A summary assembled differently on the two paths, or an order
// that came from a map rather than from the index, would pass every
// snapshot case in the suite and fail this.
package cmd_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// writeListFixture materializes a corpus whose members are written in an
// order that is NOT their path order, so that an implementation reporting
// discovery order rather than the index's own would be visible.
func writeListFixture(t *testing.T, dir string) {
	t.Helper()

	mustWriteFile(t, filepath.Join(dir, "specs.toml"), `[workspace]
name = "acme"
root = "docs"

[profile]
name = "specs"
`)
	for _, member := range []struct{ path, slug, docType, title string }{
		{"docs/standards/naming-standards.md", "naming-standards", "standards", "Naming standards"},
		{"docs/adrs/adr-002-second.md", "adr-002-second", "adr", "ADR-002: Second"},
		{"docs/adrs/adr-001-first.md", "adr-001-first", "adr", "ADR-001: First"},
		{"docs/frds/frd-checkout.md", "frd-checkout", "frd", "Feature Requirements: Checkout"},
	} {
		mustWriteFile(t, filepath.Join(dir, filepath.FromSlash(member.path)), `---
title: `+quoted(member.title)+`
description: A fixture document used by the listing determinism test.
slug: `+member.slug+`
type: `+member.docType+`
status: active
---

# `+member.title+`

## Overview

What this document is about.
`)
	}
}

// quoted double-quotes a title for YAML frontmatter. Some of these
// titles carry a colon and some do not; quoting all of them uniformly
// keeps the fixture about ordering rather than about frontmatter parsing.
func quoted(title string) string {
	return `"` + title + `"`
}

// AC: AC-SPECLIST-004.3
func TestTwoListRunsOverAnUnchangedCorpusEmitTheSameBytesInIndexOrder(t *testing.T) {
	dir := t.TempDir()
	writeListFixture(t, dir)

	firstCode, first, firstErr := runInFixture(t, dir, []string{"list"})
	if firstCode != 0 {
		t.Fatalf("the first run exited %d\n--- stdout ---\n%s--- stderr ---\n%s", firstCode, first, firstErr)
	}
	if _, err := os.Stat(filepath.Join(dir, "docs", "specs.json")); err != nil {
		t.Fatalf("the first run persisted no index, so the second is not the warm path this test is about: %v", err)
	}

	secondCode, second, secondErr := runInFixture(t, dir, []string{"list"})
	if secondCode != 0 {
		t.Fatalf("the second run exited %d\n--- stderr ---\n%s", secondCode, secondErr)
	}

	if first != second {
		t.Errorf("two runs over an unchanged corpus differ.\n--- first ---\n%s\n--- second ---\n%s", first, second)
	}

	paths := listedPaths(first)
	if len(paths) != 4 {
		t.Fatalf("listed %v, want all four members", paths)
	}
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	if !equalStrings(paths, sorted) {
		t.Errorf("paths = %v, want them in the index's own order %v, which the emitter fixes by repository-relative path",
			paths, sorted)
	}
}

// listedPaths reads the `path:` value of every entry the YAML rendering
// carries, in the order it wrote them.
func listedPaths(rendered string) []string {
	var paths []string
	for _, line := range strings.Split(rendered, "\n") {
		trimmed := strings.TrimSpace(line)
		trimmed = strings.TrimPrefix(trimmed, "- ")
		if value, found := strings.CutPrefix(trimmed, "path: "); found {
			paths = append(paths, value)
		}
	}
	return paths
}

// equalStrings reports whether two string slices carry the same values in
// the same order.
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// AC: AC-SPECLIST-002.4
//
// The snapshot case list-type-empty pins the single-occurrence form,
// `--type ""`. This covers the variant a snapshot case would only
// duplicate a corpus for: an empty occurrence standing BESIDE a valid
// one.
//
// It is the case a fix keyed off "the slice is empty" or off
// cmd.Flags().Changed("type") would leave defective — pflag's CSV reader
// drops the empty occurrence before either could see it, so
// `--type "" --type adr` arrived as `["adr"]` and answered with the adr
// members rather than reporting the malformed value.
func TestAnEmptyTypeBesideARegisteredOneIsStillRejected(t *testing.T) {
	dir := t.TempDir()
	writeListFixture(t, dir)

	code, stdout, stderr := runInFixture(t, dir, []string{"list", "--type", "", "--type", "adr"})

	if code == 0 {
		t.Errorf("exit code = 0, want non-zero: an empty --type value is not a registered type and must be reported rather than dropped\n--- stdout ---\n%s", stdout)
	}
	if !strings.Contains(stderr, `"" is not a registered document type`) {
		t.Errorf("stderr = %q, want it to report the empty value itself", stderr)
	}
	if !strings.Contains(stderr, "registered: adr,") {
		t.Errorf("stderr = %q, want it to name every registered type so the caller can correct the question", stderr)
	}
	if strings.Contains(stdout, "path:") {
		t.Errorf("stdout reported corpus members, so the empty value was answered rather than rejected:\n%s", stdout)
	}
}
