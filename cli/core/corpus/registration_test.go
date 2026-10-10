// Tests for the widened half of the Profile extension point: the
// reference registration a document type carries, the lifecycle and tag
// vocabulary a profile carries, and the validation of all three.
//
// Every profile here registers a vocabulary of its own invention —
// "thing" types, "current" and "sunset" statuses, identifiers spelled
// nothing like this framework's — because a test that reached for this
// framework's vocabulary would be asserting that the host works for one
// profile rather than for any.
package corpus_test

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// registeringProfile is a profile whose registration is complete and
// free of defects: two types, one of them carrying all three per-type
// reference registrations, and a lifecycle over its own statuses.
func registeringProfile() stubProfile {
	return stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"plan": {
				SlugPattern:              `^plan-[a-z0-9-]+$`,
				ReferenceProperties:      []string{"realizes", "owner"},
				TypedReferenceProperties: map[string]string{"realizes": "goal"},
				Pairings: []specmodel.PropertyPairing{
					{Property: "owner", Counterpart: "owns"},
				},
			},
			"goal": {SlugPattern: `^goal-[a-z0-9-]+$`},
		},
		lifecycle: corpus.LifecycleRules{
			Statuses:          []string{"drafting", "current", "sunset"},
			LiveStatuses:      []string{"current"},
			RetiredStatuses:   []string{"sunset"},
			SuccessorProperty: "replaced_by",
		},
		featureTags: registeredFeatureTags(),
	}
}

// registeredFeatureTags is a defect-free tag vocabulary of this test's
// own invention: identifiers whose two shared parts are captured under
// names both patterns define, an owner prefix, and one qualifier of each
// meaning.
func registeredFeatureTags() corpus.FeatureTagRules {
	return corpus.FeatureTagRules{
		FeaturePattern:   `^F-(?P<group>[A-Z]+)-(?P<number>[0-9]+)$`,
		CriterionPattern: `^C-(?P<group>[A-Z]+)-(?P<number>[0-9]+)\.[0-9]+$`,
		OwnerPrefix:      "belongs-to",
		SharedCaptures:   []string{"group", "number"},
		Qualifiers: []corpus.QualifierTag{
			{Tag: "forbids", Meaning: specmodel.QualifierMeaningNegation},
			{Tag: "by-hand", Meaning: specmodel.QualifierMeaningManualVerification},
			{
				Tag:     "in-isolation",
				Meaning: specmodel.QualifierMeaningVerificationLevel,
				Level:   specmodel.VerificationLevelUnit,
			},
		},
	}
}

// AC: AC-SPECREF-004.1
func TestDocumentTypeRulesCarryTheReferenceRegistration(t *testing.T) {
	var profile corpus.Profile = registeringProfile()

	rules, ok := profile.DocumentTypes()["plan"]
	if !ok {
		t.Fatal(`DocumentTypes has no entry keyed "plan"`)
	}

	want := []string{"realizes", "owner"}
	if !reflect.DeepEqual(rules.ReferenceProperties, want) {
		t.Errorf("ReferenceProperties = %v, want %v", rules.ReferenceProperties, want)
	}
	if target := rules.TypedReferenceProperties["realizes"]; target != "goal" {
		t.Errorf(`TypedReferenceProperties["realizes"] = %q, want "goal"`, target)
	}
	if len(rules.Pairings) != 1 {
		t.Fatalf("Pairings has %d entries, want 1", len(rules.Pairings))
	}
	if pairing := rules.Pairings[0]; pairing.Property != "owner" || pairing.Counterpart != "owns" {
		t.Errorf("Pairings[0] = %+v, want owner paired with owns", pairing)
	}

	// A type that registers none of the three is not a defect and not a
	// default: it holds no reference-bearing property at all.
	if bare := profile.DocumentTypes()["goal"]; len(bare.ReferenceProperties) != 0 ||
		len(bare.TypedReferenceProperties) != 0 || len(bare.Pairings) != 0 {
		t.Errorf("a type registering no reference properties reports %+v", bare)
	}
}

func TestProfileSuppliesTheLifecycleRegistration(t *testing.T) {
	var profile corpus.Profile = registeringProfile()

	lifecycle := profile.Lifecycle()
	if want := []string{"current"}; !reflect.DeepEqual(lifecycle.LiveStatuses, want) {
		t.Errorf("LiveStatuses = %v, want %v", lifecycle.LiveStatuses, want)
	}
	if want := []string{"sunset"}; !reflect.DeepEqual(lifecycle.RetiredStatuses, want) {
		t.Errorf("RetiredStatuses = %v, want %v", lifecycle.RetiredStatuses, want)
	}
	if lifecycle.SuccessorProperty != "replaced_by" {
		t.Errorf("SuccessorProperty = %q, want %q", lifecycle.SuccessorProperty, "replaced_by")
	}
}

// A profile registering no lifecycle is a profile whose documents have
// no lifecycle, not one the host has to reject or repair.
func TestProfileMayRegisterNoLifecycle(t *testing.T) {
	var profile corpus.Profile = stubProfile{}

	if lifecycle := profile.Lifecycle(); !reflect.DeepEqual(lifecycle, corpus.LifecycleRules{}) {
		t.Errorf("Lifecycle = %+v, want the zero registration", lifecycle)
	}
	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("a profile registering no lifecycle has defects: %+v", defects)
	}
}

func TestProfileSuppliesTheFeatureTagRegistration(t *testing.T) {
	var profile corpus.Profile = registeringProfile()

	tags := profile.FeatureTags()
	if tags.FeaturePattern == "" || tags.CriterionPattern == "" {
		t.Fatalf("FeatureTags carries no identifier pattern: %+v", tags)
	}
	if tags.OwnerPrefix != "belongs-to" {
		t.Errorf("OwnerPrefix = %q, want %q", tags.OwnerPrefix, "belongs-to")
	}
	if want := []string{"group", "number"}; !reflect.DeepEqual(tags.SharedCaptures, want) {
		t.Errorf("SharedCaptures = %v, want %v", tags.SharedCaptures, want)
	}

	// The meaning travels with the tag, so a host reading the
	// registration never has to know what any spelling means.
	meanings := map[string]specmodel.QualifierMeaning{}
	for _, qualifier := range tags.Qualifiers {
		meanings[qualifier.Tag] = qualifier.Meaning
	}
	if meaning := meanings["forbids"]; meaning != specmodel.QualifierMeaningNegation {
		t.Errorf(`Qualifiers["forbids"] means %q, want a negation`, meaning)
	}
	if meaning := meanings["by-hand"]; meaning != specmodel.QualifierMeaningManualVerification {
		t.Errorf(`Qualifiers["by-hand"] means %q, want manual verification`, meaning)
	}
	for _, qualifier := range tags.Qualifiers {
		if qualifier.Tag != "in-isolation" {
			continue
		}
		if qualifier.Level != specmodel.VerificationLevelUnit {
			t.Errorf("a verification-level qualifier records level %q", qualifier.Level)
		}
	}
}

// A profile registering no tag vocabulary is a profile whose corpus has
// no tagged features, not one the host has to reject or repair.
func TestProfileMayRegisterNoFeatureTags(t *testing.T) {
	var profile corpus.Profile = stubProfile{}

	if tags := profile.FeatureTags(); !reflect.DeepEqual(tags, corpus.FeatureTagRules{}) {
		t.Errorf("FeatureTags = %+v, want the zero registration", tags)
	}
	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("a profile registering no feature tags has defects: %+v", defects)
	}
}

// A pattern that does not compile matches nothing, so every check reading
// it would pass by finding no identifier at all.
func TestValidateRegistrationReportsAnUncompilableIdentifierPattern(t *testing.T) {
	const broken = `^C-(?P<group>[A-Z]+$`
	profile := stubProfile{featureTags: corpus.FeatureTagRules{CriterionPattern: broken}}

	defects := corpus.ValidateRegistration(profile)

	if len(defects) != 1 {
		t.Fatalf("defects = %+v, want exactly one", defects)
	}
	if defects[0].Key != broken {
		t.Errorf("defect names %q, want the offending pattern", defects[0].Key)
	}
	if defects[0].DocumentType != "" {
		t.Errorf("a feature-tag defect is attributed to document type %q", defects[0].DocumentType)
	}
	if defects[0].Detail == "" {
		t.Error("the defect carries no detail")
	}
}

// A derivation over a capture group neither pattern defines compares two
// empty values, which is the vacuous pass this check turns into a
// finding. It is reported once per pattern that fails to define it, so
// the reader is told which of the two to fix.
func TestValidateRegistrationReportsASharedCaptureNoPatternDefines(t *testing.T) {
	tags := registeredFeatureTags()
	tags.SharedCaptures = append(tags.SharedCaptures, "edition")
	profile := stubProfile{featureTags: tags}

	defects := corpus.ValidateRegistration(profile)

	if len(defects) != 2 {
		t.Fatalf("defects = %+v, want one per identifier pattern", defects)
	}
	for _, defect := range defects {
		if defect.Key != "edition" {
			t.Errorf("defect names %q, want the offending capture", defect.Key)
		}
	}
	if defects[0].Detail == defects[1].Detail {
		t.Errorf("both defects read %q, so neither says which pattern to fix", defects[0].Detail)
	}
}

// A registration that names shared captures but no pattern at all is the
// half-written case: nothing defines the captures, so nothing is
// compared, and the check must not pass by finding no pattern to check
// against.
func TestValidateRegistrationReportsSharedCapturesWithNoPattern(t *testing.T) {
	profile := stubProfile{featureTags: corpus.FeatureTagRules{SharedCaptures: []string{"group"}}}

	if defects := corpus.ValidateRegistration(profile); len(defects) != 2 {
		t.Errorf("defects = %+v, want one per identifier pattern", defects)
	}
}

func TestValidateRegistrationAcceptsACompleteRegistration(t *testing.T) {
	if defects := corpus.ValidateRegistration(registeringProfile()); len(defects) != 0 {
		t.Errorf("a registration with no defect reports %+v", defects)
	}
}

// All five defect classes, in one registration, so the test also proves
// that validation reports every defect it finds rather than the first.
func TestValidateRegistrationReportsEveryDefectClass(t *testing.T) {
	const brokenPattern = `^C-(?P<group>[A-Z]+$`

	profile := registeringProfile()
	profile.types["plan"] = corpus.DocumentTypeRules{
		SlugPattern: `^plan-[a-z0-9-]+$`,
		// "realizes" is no longer reference-bearing, and "supports"
		// names a type this profile does not define.
		ReferenceProperties: []string{"owner"},
		TypedReferenceProperties: map[string]string{
			"realizes": "goal",
			"supports": "mission",
		},
		Pairings: []specmodel.PropertyPairing{
			{Property: "sponsor", Counterpart: "sponsors"},
		},
	}
	profile.lifecycle.LiveStatuses = []string{"current", "provisional"}
	profile.lifecycle.RetiredStatuses = []string{"archived"}
	// The criterion pattern no longer compiles, and "edition" is a
	// capture the surviving pattern does not define.
	profile.featureTags.CriterionPattern = brokenPattern
	profile.featureTags.SharedCaptures = append(profile.featureTags.SharedCaptures, "edition")

	defects := corpus.ValidateRegistration(profile)

	// "supports" produces two defects of its own: it is neither
	// reference-bearing nor does it name a defined type.
	wantKeys := []string{
		"realizes", "sponsor", "supports", "supports",
		"provisional", "archived",
		brokenPattern, "edition",
	}
	var gotKeys []string
	for _, defect := range defects {
		gotKeys = append(gotKeys, defect.Key)
		if defect.Detail == "" {
			t.Errorf("defect %+v carries no detail", defect)
		}
	}
	if !reflect.DeepEqual(sorted(gotKeys), sorted(wantKeys)) {
		t.Errorf("defect keys = %v, want %v (defects: %+v)", gotKeys, wantKeys, defects)
	}

	for _, defect := range defects {
		switch defect.Key {
		case "provisional", "archived", brokenPattern, "edition":
			if defect.DocumentType != "" {
				t.Errorf("profile-level defect %+v is attributed to a document type", defect)
			}
		default:
			if defect.DocumentType != "plan" {
				t.Errorf(`defect %+v is attributed to %q, want "plan"`, defect, defect.DocumentType)
			}
		}
	}
}

// A typed reference naming an undefined type is reported, and the
// registration keeps it: nothing falls back to a default for the
// offending key, because a registration quietly repaired into one that
// checks less than it says is the failure being reported.
func TestValidateRegistrationSubstitutesNoDefault(t *testing.T) {
	profile := registeringProfile()
	profile.types["plan"] = corpus.DocumentTypeRules{
		SlugPattern:              `^plan-[a-z0-9-]+$`,
		ReferenceProperties:      []string{"realizes"},
		TypedReferenceProperties: map[string]string{"realizes": "mission"},
	}

	if defects := corpus.ValidateRegistration(profile); len(defects) != 1 {
		t.Fatalf("defects = %+v, want exactly one", defects)
	}

	rules := profile.DocumentTypes()["plan"]
	if target := rules.TypedReferenceProperties["realizes"]; target != "mission" {
		t.Errorf(`TypedReferenceProperties["realizes"] = %q after validation, want it untouched`, target)
	}
	if _, defined := profile.DocumentTypes()["mission"]; defined {
		t.Error("validation defined the document type the registration named")
	}
}

// The defects are returned, not rendered: a compiled-in registration has
// no file position of its own, and the position a reader needs belongs to
// the configuration that selected the profile. Only the caller has it.
func TestValidateRegistrationRendersNothing(t *testing.T) {
	profile := registeringProfile()
	profile.lifecycle.LiveStatuses = []string{"provisional"}

	stdout, stderr := os.Stdout, os.Stderr
	t.Cleanup(func() { os.Stdout, os.Stderr = stdout, stderr })

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("opening a pipe: %v", err)
	}
	os.Stdout, os.Stderr = writer, writer

	defects := corpus.ValidateRegistration(profile)

	if err := writer.Close(); err != nil {
		t.Fatalf("closing the pipe: %v", err)
	}
	written, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading the pipe: %v", err)
	}

	if len(written) != 0 {
		t.Errorf("validation wrote %q", written)
	}
	if len(defects) != 1 {
		t.Fatalf("defects = %+v, want exactly one", defects)
	}
	if defects[0].Key != "provisional" {
		t.Errorf("defect names %q, want the offending status", defects[0].Key)
	}
}

// The same registration validates to the same list every time, because a
// caller renders it and rendered output is compared byte for byte.
func TestValidateRegistrationIsOrderedDeterministically(t *testing.T) {
	profile := registeringProfile()
	profile.types["plan"] = corpus.DocumentTypeRules{
		TypedReferenceProperties: map[string]string{
			"realizes": "goal",
			"supports": "goal",
			"tracks":   "goal",
		},
	}
	profile.types["goal"] = corpus.DocumentTypeRules{
		TypedReferenceProperties: map[string]string{"blocks": "goal"},
	}

	first := corpus.ValidateRegistration(profile)
	if len(first) != 4 {
		t.Fatalf("defects = %+v, want four", first)
	}
	for range 20 {
		if again := corpus.ValidateRegistration(profile); !reflect.DeepEqual(again, first) {
			t.Fatalf("validation reported %+v, previously %+v", again, first)
		}
	}
}

// sorted returns a sorted copy, so a comparison of defect keys is about
// which keys were named rather than about the order they arrived in —
// which TestValidateRegistrationIsOrderedDeterministically covers.
func sorted(values []string) []string {
	out := append([]string(nil), values...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// A profile registering no record-block tag describes documents with no
// record blocks, not one the host has to reject or repair.
func TestProfileMayRegisterNoRecordBlocks(t *testing.T) {
	var profile corpus.Profile = stubProfile{}

	if tags := profile.RecordBlocks(); len(tags) != 0 {
		t.Errorf("RecordBlocks = %+v, want none", tags)
	}
	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("a profile registering no record blocks has defects: %+v", defects)
	}
}

// A tag whose pointer addresses a definition in the profile's own schema
// is valid, and so is a tag registered with no pointer at all.
func TestValidateRegistrationAcceptsResolvingAndAbsentSchemaPointers(t *testing.T) {
	profile := stubProfile{
		schema: []byte(`{"definitions":{"Widget":{"type":"object"}}}`),
		recordBlocks: []corpus.RecordBlockTag{
			{Tag: "widget", Schema: "#/definitions/Widget"},
			{Tag: "note"},
		},
	}

	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("a resolving or absent pointer was reported: %+v", defects)
	}
}

// A pointer resolving to nothing would validate every record against
// nothing, so it is reported rather than skipped.
func TestValidateRegistrationReportsAnUnresolvableSchemaPointer(t *testing.T) {
	profile := stubProfile{
		schema: []byte(`{"definitions":{"Widget":{"type":"object"}}}`),
		recordBlocks: []corpus.RecordBlockTag{
			{Tag: "widget", Schema: "#/definitions/Gadget"},
			{Tag: "bare", Schema: "Widget"},
		},
	}

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 2 {
		t.Fatalf("got %d defects, want 2: %+v", len(defects), defects)
	}
	if defects[0].Key != "widget" || defects[1].Key != "bare" {
		t.Errorf("defect keys = %q, %q; want the registered tags", defects[0].Key, defects[1].Key)
	}
}

// A schema that is not JSON is the schema's fault, reported by the schema
// registry; it must not read as every pointer resolving to nothing.
func TestValidateRegistrationRaisesNoPointerDefectForAnUnparsableSchema(t *testing.T) {
	profile := stubProfile{
		schema:       []byte(`{not json`),
		recordBlocks: []corpus.RecordBlockTag{{Tag: "widget", Schema: "#/definitions/Widget"}},
	}

	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("got %d defects for an unparsable schema, want none: %+v", len(defects), defects)
	}
}

// A pointer with no leading slash is not a JSON Pointer, so it resolves
// to nothing.
func TestValidateRegistrationReportsAPointerWithNoLeadingSlash(t *testing.T) {
	profile := stubProfile{
		schema:       []byte(`{"definitions":{"Widget":{"type":"object"}}}`),
		recordBlocks: []corpus.RecordBlockTag{{Tag: "widget", Schema: "#definitions/Widget"}},
	}

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 1 || defects[0].Key != "widget" {
		t.Errorf("got %+v, want one defect keyed widget", defects)
	}
}

// A record-block tag registered more than once has no single schema, so
// each registration after the first is a defect, whether or not the
// pointers agree, and the duplicate check does not depend on the schema.
func TestValidateRegistrationReportsARecordBlockTagRegisteredTwice(t *testing.T) {
	schema := []byte(`{"definitions":{"W":{},"V":{}}}`)

	same := stubProfile{schema: schema, recordBlocks: []corpus.RecordBlockTag{
		{Tag: "c", Schema: "#/definitions/W"}, {Tag: "c", Schema: "#/definitions/W"},
	}}
	if defects := corpus.ValidateRegistration(same); len(defects) != 1 || defects[0].Key != "c" {
		t.Errorf("same pointer twice: got %+v, want one defect keyed c", defects)
	}

	differ := stubProfile{schema: schema, recordBlocks: []corpus.RecordBlockTag{
		{Tag: "c", Schema: "#/definitions/W"}, {Tag: "c", Schema: "#/definitions/V"},
	}}
	defects := corpus.ValidateRegistration(differ)
	if len(defects) != 1 {
		t.Fatalf("two pointers: got %+v, want one defect", defects)
	}
	for _, pointer := range []string{"#/definitions/W", "#/definitions/V"} {
		if !strings.Contains(defects[0].Detail, pointer) {
			t.Errorf("Detail %q does not name %s", defects[0].Detail, pointer)
		}
	}

	thrice := stubProfile{schema: schema, recordBlocks: []corpus.RecordBlockTag{
		{Tag: "c"}, {Tag: "c"}, {Tag: "c"},
	}}
	if defects := corpus.ValidateRegistration(thrice); len(defects) != 2 {
		t.Errorf("three registrations: got %d defects, want 2: %+v", len(defects), defects)
	}

	unparsable := stubProfile{schema: []byte(`{not json`), recordBlocks: []corpus.RecordBlockTag{
		{Tag: "c", Schema: "#/definitions/W"}, {Tag: "c", Schema: "#/definitions/W"},
	}}
	if defects := corpus.ValidateRegistration(unparsable); len(defects) != 1 {
		t.Errorf("unparsable schema: got %+v, want only the duplicate defect", defects)
	}

	badDuplicate := stubProfile{schema: schema, recordBlocks: []corpus.RecordBlockTag{
		{Tag: "c", Schema: "#/definitions/W"}, {Tag: "c", Schema: "#/definitions/Nope"},
	}}
	if defects := corpus.ValidateRegistration(badDuplicate); len(defects) != 2 {
		t.Errorf("duplicate with a bad pointer: got %d defects, want 2: %+v", len(defects), defects)
	}

	distinct := stubProfile{schema: schema, recordBlocks: []corpus.RecordBlockTag{
		{Tag: "a", Schema: "#/definitions/W"}, {Tag: "b", Schema: "#/definitions/V"},
	}}
	if defects := corpus.ValidateRegistration(distinct); len(defects) != 0 {
		t.Errorf("distinct tags: got %+v, want none", defects)
	}
}
