package profile_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
	"github.com/brettcallaghan/specs/cli/profile"
)

// repositoryRoot locates this repository's root directory from this test
// file's own location on disk, rather than from the process's working
// directory: `go test` sets that to this package's directory, not the
// repository root, and a caller invoking the suite from elsewhere cannot
// be relied on to set it at all.
//
// The two parent steps up from this file (cli/profile/…) land on the
// repository root; that is verified against specs.toml and specs/
// existing there rather than trusted outright, because a silently wrong
// directory would make this gate pass vacuously over zero documents —
// exactly the failure mode this test exists to avoid (see the doc
// comment on [TestProfileSelfApplicationAgainstThisRepositorysCorpus]).
func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not resolve this test file's own path")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(thisFile)))

	if _, err := os.Stat(filepath.Join(root, "specs.toml")); err != nil {
		t.Fatalf("resolved repository root %q has no specs.toml: %v", root, err)
	}
	if info, err := os.Stat(filepath.Join(root, "specs")); err != nil || !info.IsDir() {
		t.Fatalf("resolved repository root %q has no specs/ directory (err=%v)", root, err)
	}
	return root
}

// TestProfileSelfApplicationAgainstThisRepositorysCorpus is the profile
// self-application gate specs/modules/spec-profile/spec-profile-module.md names
// as "the one that catches a registered schema drifting from the
// templates it mirrors": it builds this framework's own specs
// profile and runs Document Core's discovery and schema/slug-pattern
// checks — [corpus.Build] and [corpus.Corpus.Check] — against this
// repository's own specification corpus.
//
// It drives corpus.Build and Corpus.Check directly rather than through
// cli/cmd, because stage 2's command surface (a `check` subcommand) had
// not landed when this gate was wired (agent-plugins-bge.14, stage 2 of
// 3 itself) — those two calls are the whole of what a future CLI
// invocation would make, so nothing about the check itself is a stand-in.
//
// The corpus root is specs/, not docs/. The blueprint's Definition of
// Done originally named docs/; this repository's specs/ reorganization
// left docs/ holding a single file (session-bootstrap.md), which would
// make this gate pass vacuously. The corrected root — specs/, 27
// markdown documents at the time this gate was wired — is recorded on
// agent-plugins-bge.14 (author-confirmed 2026-08-04) and matches
// [workspace].root in this repository's own specs.toml.
//
// The result is asserted exactly, not merely "no error" or "no findings
// above some severity": per bge.14's instruction not to weaken the gate
// to make it green, a corpus defect this run surfaces is filed as its
// own bead rather than fixed here or hidden by loosening this assertion.
// Zero findings is expected here, not a placeholder — a finding
// appearing where this asserts none is exactly the drift this gate
// exists to catch, and this test should fail rather than pass silently.
//
// agent-plugins-tff (spec-index-blueprint.md's `plan: agent-plugins-bge`
// tripped additionalProperties because FeatureBlueprint's schema did not
// declare `plan`) was the one finding this test previously asserted
// exactly. It is resolved: FeatureBlueprint now declares an optional
// `plan: String?`, schemas/frontmatter.strata was regenerated, and this
// document validates.
//
// Two false-positive findings this same run surfaced during development
// were engine bugs, not corpus defects, and were fixed rather than filed
// (both cited from this comment because fixing them is what makes the
// assertion below exact rather than accommodating noise):
//
//   - cli/core/frontmatter's readProperty computed a property's byte
//     offset directly from the YAML scanner's token position, which
//     goccy/go-yaml (v1.19.2) advances one unit per Unicode codepoint
//     rather than per byte. Every document in this corpus whose title or
//     description used an em dash — several do — had every property
//     after it misaligned, corrupting the raw text a format check
//     validates and producing seven spurious `updated_at` "does not
//     satisfy format" findings. See frontmatter.go's absolute and
//     byteOffsetOfRune, and
//     TestAMultiByteCharacterEarlierInTheBlockDoesNotShiftALaterProperty
//     in frontmatter_test.go.
//   - cli/core/schemaregistry's findingsForUnit reported every
//     additionalProperties violation against the whole frontmatter
//     block with an empty value, because the underlying library reports
//     that failure against the object as a whole and the offending key
//     names live in the error's own Properties field rather than in its
//     JSON Pointer. See findingsForUnit's *kind.AdditionalProperties
//     branch and TestValidateReportsOneFindingPerUndeclaredProperty in
//     schemaregistry_test.go.
//
// agent-plugins-6n4 (work-order-template.md is not valid YAML) is NOT
// exercised by this test: that file is
// plugin/skills/spec/references/work-order-template.md, outside the
// specs/ root this gate is scoped to. The blueprint's Tracker
// Reconciliation entry for 6n4 says this gate reports it, which was true
// only under a broader (or pre-reorg) notion of corpus root; that
// staleness is filed as agent-plugins-mvx for whoever verifies the
// Definition of Done to reconcile, rather than resolved here by
// widening this gate's root unreviewed.
func TestProfileSelfApplicationAgainstThisRepositorysCorpus(t *testing.T) {
	root := repositoryRoot(t)
	specRoot := filepath.Join(root, "specs")

	built, err := corpus.Build(corpus.Options{
		RepoRoot:  root,
		Root:      specRoot,
		Workspace: "agent-plugins",
		Profile:   profile.New(),
	})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}

	index := built.Index()
	if len(index.Entries) == 0 {
		t.Fatal("the corpus discovered no members at all; the spec root is likely wrong " +
			"(this is the vacuous-pass failure mode this test exists to avoid)")
	}

	// specs/adrs/README.md is a plain index page with no frontmatter —
	// an expected exclusion, not a defect.
	if len(index.Exclusions) != 1 {
		t.Fatalf("exclusions = %+v, want exactly one", index.Exclusions)
	}
	if got := index.Exclusions[0]; got.Path != "specs/adrs/README.md" ||
		got.Reason != specmodel.ExclusionReasonNoFrontmatter {
		t.Errorf("exclusion = %+v, want {specs/adrs/README.md, no-frontmatter}", got)
	}

	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("Check returned %d findings, want exactly 0 (see this test's doc comment): %+v",
			len(findings), findings)
	}
}

// Over this repository's own corpus, which declares every required
// property, the index corpus.Marshal writes — entry frontmatter encoded
// by the envelope package so that an undeclared required enum is omitted
// rather than refused — is byte-identical to the generated encoder's,
// across every document type the corpus holds.
func TestTheIndexEncodingIsTheGeneratedEncodersOverThisRepositorysCorpus(t *testing.T) {
	root := repositoryRoot(t)
	built, err := corpus.Build(corpus.Options{
		RepoRoot:  root,
		Root:      filepath.Join(root, "specs"),
		Workspace: "agent-plugins",
		Profile:   profile.New(),
	})
	if err != nil {
		t.Fatalf("corpus.Build: %v", err)
	}
	index := built.Index()
	want, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		t.Fatalf("the generated encoder refused this repository's index: %v", err)
	}
	got, err := corpus.Marshal(index)
	if err != nil {
		t.Fatalf("corpus.Marshal: %v", err)
	}
	if string(got) != string(want)+"\n" {
		t.Error("corpus.Marshal differs from the generated encoder over this repository's corpus")
	}
	types := map[string]bool{}
	for _, entry := range index.Entries {
		if entry.Frontmatter != nil {
			types[fmt.Sprintf("%T", entry.Frontmatter.Value)] = true
		}
	}
	if len(types) < 5 {
		t.Errorf("the corpus covers only %d document types; this test means less than it says", len(types))
	}
}
