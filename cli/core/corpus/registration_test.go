// Tests for the widened half of the Profile extension point: the
// reference registration a document type carries, the lifecycle
// registration a profile carries, and the validation of both.
//
// Every profile here registers a vocabulary of its own invention —
// "thing" types, "current" and "sunset" statuses — because a test that
// reached for this framework's vocabulary would be asserting that the
// host works for one profile rather than for any.
package corpus_test

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
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

func TestValidateRegistrationAcceptsACompleteRegistration(t *testing.T) {
	if defects := corpus.ValidateRegistration(registeringProfile()); len(defects) != 0 {
		t.Errorf("a registration with no defect reports %+v", defects)
	}
}

// All three defect classes, in one registration, so the test also proves
// that validation reports every defect it finds rather than the first.
func TestValidateRegistrationReportsEveryDefectClass(t *testing.T) {
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

	defects := corpus.ValidateRegistration(profile)

	// "supports" produces two defects of its own: it is neither
	// reference-bearing nor does it name a defined type.
	wantKeys := []string{"realizes", "sponsor", "supports", "supports", "provisional", "archived"}
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
		case "provisional", "archived":
			if defect.DocumentType != "" {
				t.Errorf("lifecycle defect %+v is attributed to a document type", defect)
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
