// Tests for the run record's model: the declared key set read off the
// generated declaration, and the rules a write is checked against.
package manifest_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

func raw(t *testing.T, values map[string]string) map[string]json.RawMessage {
	t.Helper()

	out := make(map[string]json.RawMessage, len(values))
	for k, v := range values {
		out[k] = json.RawMessage(v)
	}
	return out
}

// AC: AC-SWARM-CORE-002.19
//
// The key set this package validates against is read from the generated
// model, not restated here. A key added to schemas/swarm/manifest.strata
// is therefore known to reading, requiring and validating at once, and
// this test is what shows the derivation actually reaches the declaration
// rather than a copy of it.
func TestTheDeclaredKeySetIsReadFromTheGeneratedDeclaration(t *testing.T) {
	keys := manifest.Keys()
	if len(keys) < 20 {
		t.Fatalf("the declared set has %d keys, which is too few to be the manifest's", len(keys))
	}

	for name, want := range map[string]swarmmodel.RunStage{
		"checks_baseline": swarmmodel.RunStagePreFlight,
		"admitted":        swarmmodel.RunStageScope,
		"collision_edges": swarmmodel.RunStagePlan,
		"base_sha":        swarmmodel.RunStageDispatch,
		"findings_path":   swarmmodel.RunStageRetro,
	} {
		key, ok := manifest.Lookup(name)
		if !ok {
			t.Errorf("%q is not in the declared set", name)
			continue
		}
		if key.Stage != want {
			t.Errorf("%q becomes required at %q, want %q", name, key.Stage, want)
		}
	}

	// Each key appears once, whatever the inheritance chain repeats.
	seen := map[string]bool{}
	for _, k := range keys {
		if seen[k.Name] {
			t.Errorf("%q is in the declared set twice", k.Name)
		}
		seen[k.Name] = true
		if k.Type == nil {
			t.Errorf("%q has no declared type", k.Name)
		}
	}

	if _, ok := manifest.Lookup("circuit_breaker"); ok {
		t.Error("a key nothing declares is in the set")
	}
}

// AC: AC-SWARM-CORE-002.2
//
// A key outside the declared set is refused rather than stored, and the
// refusal names it — a run whose record grew a key nobody declared is a
// run whose reader cannot tell a typo from a new convention.
func TestAKeyOutsideTheDeclaredSetIsRefused(t *testing.T) {
	err := manifest.Check(raw(t, map[string]string{"circuit_breaker": `3`}))
	if err == nil {
		t.Fatal("Check accepted an undeclared key")
	}
	if !strings.Contains(err.Error(), "circuit_breaker") {
		t.Errorf("the refusal does not name the key: %v", err)
	}
}

// AC: AC-SWARM-CORE-002.9
//
// A value whose type is not its key's is refused. The declared type comes
// from the generated model, so this cannot disagree with what a consumer
// validating against the schema would decide.
func TestAValueOfTheWrongTypeForItsKeyIsRefused(t *testing.T) {
	for name, change := range map[string]map[string]string{
		"text where a count belongs": {"review_passes": `"two"`},
		"a count where text belongs": {"branch": `7`},
		"text where a list belongs":  {"admitted": `"run-1.1"`},
		"a wrong row type inside an object": {
			"standing_auth": `{"review_budget": {"value": "two", "source": "default"}}`,
		},
		"an unknown enum value": {"effort": `"frantic"`},
	} {
		if err := manifest.Check(raw(t, change)); err == nil {
			t.Errorf("%s: Check accepted it", name)
		}
	}

	// The negative control: the same keys with the right types pass, so
	// the rule cannot be satisfied by refusing everything.
	ok := raw(t, map[string]string{
		"review_passes": `2`,
		"branch":        `"claude/branch-x"`,
		"admitted":      `["run-1.1", "run-1.2"]`,
		"effort":        `"unrecorded"`,
	})
	if err := manifest.Check(ok); err != nil {
		t.Errorf("Check refused well-typed values: %v", err)
	}
}

// AC: AC-SWARM-CORE-002.14
//
// A multi-key write that fails on one key stores none of them, and the
// refusal names every reason rather than the first — a caller that fixed
// one and re-ran to find the next would be doing the work this is for.
func TestAMultiKeyWriteReportsEveryRefusalAndStoresNone(t *testing.T) {
	changes := raw(t, map[string]string{
		"branch":          `"claude/branch-x"`,
		"review_passes":   `"two"`,
		"circuit_breaker": `3`,
		"admitted":        `["run-1.1"]`,
	})

	err := manifest.Check(changes)
	if err == nil {
		t.Fatal("Check accepted a write with two bad keys")
	}
	for _, want := range []string{"review_passes", "circuit_breaker"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if !strings.Contains(err.Error(), "storing none of it") {
		t.Errorf("the refusal does not say that nothing was stored: %v", err)
	}
}

// AC: AC-SWARM-CORE-002.3
//
// An absent key is reported as absent rather than as an empty value of
// its declared type. "Not yet written" and "written as nothing" are
// different states of a run, and a reader must not have to guess which.
func TestAnAbsentKeyIsReportedAsAbsentRatherThanEmpty(t *testing.T) {
	doc := manifest.NewDocument(raw(t, map[string]string{"branch": `"claude/branch-x"`}))

	if _, present := doc.Get("checks_baseline"); present {
		t.Error("an absent key reports as present")
	}

	var baseline string
	present, err := doc.Decode("checks_baseline", &baseline)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if present {
		t.Error("Decode reports an absent key as present")
	}
	if baseline != "" {
		t.Errorf("Decode wrote %q into the target for an absent key", baseline)
	}

	// And a key that IS present, so absence is not the only answer.
	var branch string
	present, err = doc.Decode("branch", &branch)
	if err != nil || !present || branch != "claude/branch-x" {
		t.Errorf("Decode of a present key = %q, %v, %v", branch, present, err)
	}
}

// AC: AC-SWARM-CORE-002.5
//
// A list-valued key reads back as a list rather than as one string. The
// measured hazard behind this is that the tracker's per-key write stores
// an array as a string, so a record written that way would read back
// wrong; the declared type is what makes it detectable.
func TestAListValuedKeyReadsBackAsAList(t *testing.T) {
	doc := manifest.NewDocument(raw(t, map[string]string{
		"admitted": `["run-1.1", "run-1.2", "run-1.3"]`,
	}))

	var admitted []string
	present, err := doc.Decode("admitted", &admitted)
	if err != nil || !present {
		t.Fatalf("Decode: %v, present=%v", err, present)
	}
	if len(admitted) != 3 || admitted[0] != "run-1.1" {
		t.Errorf("admitted = %v, want three bead ids", admitted)
	}

	// The shape the per-key write would have produced is refused, so a
	// record written that way is a failure rather than a list of one.
	if err := manifest.Check(raw(t, map[string]string{"admitted": `"[\"run-1.1\"]"`})); err == nil {
		t.Error("Check accepted a list stored as a string")
	}
}

// Merging applies at the top level and leaves the receiver alone, which
// is what makes a write that names one key unable to disturb another.
func TestMergingAppliesAtTheTopLevelAndDoesNotMutateWhatItMerges(t *testing.T) {
	base := manifest.NewDocument(raw(t, map[string]string{
		"branch":        `"claude/branch-x"`,
		"review_passes": `1`,
	}))

	merged := base.Merged(raw(t, map[string]string{"review_passes": `2`}))

	if got, _ := merged.Get("branch"); string(got) != `"claude/branch-x"` {
		t.Errorf("the untouched key became %s", got)
	}
	if got, _ := merged.Get("review_passes"); string(got) != `2` {
		t.Errorf("the written key became %s", got)
	}
	if got, _ := base.Get("review_passes"); string(got) != `1` {
		t.Errorf("Merged mutated the document it was called on: %s", got)
	}
}

// A key this version of the CLI does not know is preserved rather than
// dropped, because the record is not exclusively owned: several
// components write into it, and a reader that rewrote only what it
// understood would delete the rest.
func TestAKeyThisVersionDoesNotKnowSurvivesBeingRead(t *testing.T) {
	doc := manifest.NewDocument(raw(t, map[string]string{
		"branch":        `"claude/branch-x"`,
		"something_new": `{"a": 1}`,
		"review_passes": `1`,
	}))

	names := doc.Names()
	if len(names) != 3 {
		t.Fatalf("Names() = %v, want all three", names)
	}
	if got, ok := doc.Get("something_new"); !ok || string(got) != `{"a": 1}` {
		t.Errorf("the unknown key was not preserved: %s, %v", got, ok)
	}
}

// The guard for a conditional write is the values the covered keys held
// when the caller read the record — and a key that was absent then must
// still be absent, because a key another writer added is as much a
// concurrent change as one they altered.
func TestTheGuardCoversOnlyTheKeysAWriteTouchesIncludingTheirAbsence(t *testing.T) {
	base := manifest.NewDocument(raw(t, map[string]string{
		"branch":        `"claude/branch-x"`,
		"review_passes": `1`,
	}))
	changes := raw(t, map[string]string{"review_passes": `2`, "findings_path": `"docs/retros/x.md"`})

	expect := manifest.Expect(base, changes)
	if _, ok := expect.Get("branch"); ok {
		t.Error("the guard covers a key the write does not touch")
	}
	if got, ok := expect.Get("review_passes"); !ok || string(got) != `1` {
		t.Errorf("the guard does not carry the value the key held: %s, %v", got, ok)
	}
	if _, ok := expect.Get("findings_path"); ok {
		t.Error("the guard carries a value for a key that was absent")
	}

	// Nothing moved.
	if moved := manifest.Moved(expect, base, changes); len(moved) != 0 {
		t.Errorf("Moved reported %v against an unchanged record", moved)
	}

	// A key altered by someone else.
	altered := base.Merged(raw(t, map[string]string{"review_passes": `2`}))
	if moved := manifest.Moved(expect, altered, changes); len(moved) != 1 || moved[0] != "review_passes" {
		t.Errorf("Moved = %v, want [review_passes]", moved)
	}

	// A key that appeared after the read.
	appeared := base.Merged(raw(t, map[string]string{"findings_path": `"docs/retros/other.md"`}))
	if moved := manifest.Moved(expect, appeared, changes); len(moved) != 1 || moved[0] != "findings_path" {
		t.Errorf("Moved = %v, want [findings_path]", moved)
	}

	// A key the write does not touch changing is not a conflict: another
	// component writing a different key is the normal case, not a race.
	elsewhere := base.Merged(raw(t, map[string]string{"branch": `"claude/other"`}))
	if moved := manifest.Moved(expect, elsewhere, changes); len(moved) != 0 {
		t.Errorf("Moved = %v for a change outside the write", moved)
	}
}

// Whitespace and key order are not changes. Without this a re-read whose
// encoder spaced an object differently would read as a concurrent write.
func TestAReEncodingIsNotAConcurrentChange(t *testing.T) {
	changes := raw(t, map[string]string{"models": `{"light":"haiku"}`})
	expect := manifest.NewDocument(raw(t, map[string]string{
		"models": `{"light": "haiku", "standard": "sonnet"}`,
	}))
	current := manifest.NewDocument(raw(t, map[string]string{
		"models": `{"standard":"sonnet","light":"haiku"}`,
	}))

	if moved := manifest.Moved(expect, current, changes); len(moved) != 0 {
		t.Errorf("Moved = %v for the same object written differently", moved)
	}
}

// A metadata object of the shape a tracker hands back becomes a document
// whose values are the JSON they were.
func TestAMetadataObjectBecomesADocument(t *testing.T) {
	doc, err := manifest.ParseDocument(map[string]any{
		"branch":        "claude/branch-x",
		"review_passes": float64(2),
		"admitted":      []any{"run-1.1"},
	})
	if err != nil {
		t.Fatalf("ParseDocument: %v", err)
	}
	if got, _ := doc.Get("branch"); string(got) != `"claude/branch-x"` {
		t.Errorf("branch = %s", got)
	}
	if got, _ := doc.Get("review_passes"); string(got) != `2` {
		t.Errorf("review_passes = %s", got)
	}
	if got, _ := doc.Get("admitted"); string(got) != `["run-1.1"]` {
		t.Errorf("admitted = %s", got)
	}

	// An epic with no manifest yet reads as an empty record, not an error.
	empty, err := manifest.ParseDocument(nil)
	if err != nil {
		t.Fatalf("ParseDocument(nil): %v", err)
	}
	if len(empty.Names()) != 0 {
		t.Errorf("an empty metadata object read as %v", empty.Names())
	}
}
