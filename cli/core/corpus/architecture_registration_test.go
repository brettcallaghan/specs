// Tests for the architecture vocabulary a profile registers through the
// Profile extension point, and for the four defects ValidateRegistration
// reports against it.
//
// As in registration_test.go, the vocabulary is of this test's own
// invention — estates, sites, wings and rooms — so that what is asserted
// is that the host works for any profile rather than for one.
package corpus_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
)

// architectureSchema is a discriminated union over the invented types:
// each branch pins "type" to a constant and declares the properties the
// architecture registration below applies to it.
const architectureSchema = `{
  "definitions": {
    "Any": {
      "anyOf": [
        {"properties": {"type": {"const": "estate"}, "label": {}, "leases_held": {}}},
        {"properties": {"type": {"const": "site"}, "label": {}, "leases_held": {}, "estate_of": {}, "wings": {}}},
        {"properties": {"type": {"const": "wing"}, "label": {}, "leases_held": {}, "wings": {}}},
        {"properties": {"type": {"const": "room"}, "label": {}, "leases_held": {}}},
        {"properties": {"type": {"const": "lease"}, "landlord": {}}}
      ]
    }
  }
}`

// architectureProfile is a profile whose architecture registration is
// complete and free of defects.
func architectureProfile() stubProfile {
	return stubProfile{
		schema: []byte(architectureSchema),
		types: map[string]corpus.DocumentTypeRules{
			"estate": {}, "site": {}, "wing": {}, "room": {}, "lease": {},
		},
		recordBlocks: []corpus.RecordBlockTag{{Tag: "room"}},
		architecture: corpus.ArchitectureRules{
			Elements: map[string]corpus.ElementTypeRules{
				"estate": {Level: corpus.LevelSystem},
				"site":   {Level: corpus.LevelContainer, RosterHeading: "Rooms", RosterTag: "room"},
				"wing":   {Level: corpus.LevelModule, RosterHeading: "Rooms", RosterTag: "room"},
				"room":   {Level: corpus.LevelComponent},
			},
			EntityNameProperty:        "label",
			SystemProperty:            "estate_of",
			ComposedModulesProperty:   "wings",
			ConsumedContractsProperty: "leases_held",
			ContractType:              "lease",
			ProviderProperty:          "landlord",
		},
	}
}

func TestProfileSuppliesTheArchitectureRegistration(t *testing.T) {
	var profile corpus.Profile = architectureProfile()

	rules := profile.Architecture()
	if got := rules.Elements["site"]; got.Level != corpus.LevelContainer || got.RosterTag != "room" {
		t.Errorf(`Elements["site"] = %+v`, got)
	}
	if rules.EntityNameProperty != "label" || rules.ProviderProperty != "landlord" {
		t.Errorf("properties = %+v", rules)
	}
}

func TestValidateRegistrationAcceptsACompleteArchitectureRegistration(t *testing.T) {
	if defects := corpus.ValidateRegistration(architectureProfile()); len(defects) != 0 {
		t.Errorf("defects = %+v, want none", defects)
	}
}

// A profile registering no architecture vocabulary describes a corpus
// with no elements, which is not a malformed profile.
func TestProfileMayRegisterNoArchitecture(t *testing.T) {
	profile := architectureProfile()
	profile.architecture = corpus.ArchitectureRules{}

	if got := profile.Architecture(); !reflect.DeepEqual(got, corpus.ArchitectureRules{}) {
		t.Errorf("Architecture = %+v, want the zero value", got)
	}
	if defects := corpus.ValidateRegistration(profile); len(defects) != 0 {
		t.Errorf("defects = %+v, want none", defects)
	}
	if defects := corpus.ValidateRegistration(stubProfile{}); len(defects) != 0 {
		t.Errorf("an empty profile has defects %+v, want none", defects)
	}
}

func TestValidateRegistrationReportsAnUndefinedArchitectureType(t *testing.T) {
	profile := architectureProfile()
	profile.architecture.Elements["annex"] = corpus.ElementTypeRules{Level: corpus.LevelComponent}
	profile.architecture.ContractType = "covenant"

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 2 {
		t.Fatalf("defects = %+v, want two", defects)
	}
	for i, key := range []string{"annex", "covenant"} {
		if defects[i].Key != key || defects[i].DocumentType != "" {
			t.Errorf("defect %d = %+v, want key %q at profile level", i, defects[i], key)
		}
		if !strings.Contains(defects[i].Detail, "not a document type this profile defines") {
			t.Errorf("defect %d detail = %q", i, defects[i].Detail)
		}
	}
}

func TestValidateRegistrationReportsAnUndeclaredArchitectureProperty(t *testing.T) {
	profile := architectureProfile()
	profile.architecture.SystemProperty = "estate_name"
	profile.architecture.ProviderProperty = "tenant"

	defects := corpus.ValidateRegistration(profile)
	want := []corpus.RegistrationDefect{
		{DocumentType: "site", Key: "estate_name"},
		{DocumentType: "lease", Key: "tenant"},
	}
	if len(defects) != len(want) {
		t.Fatalf("defects = %+v, want %d", defects, len(want))
	}
	for i := range want {
		if defects[i].DocumentType != want[i].DocumentType || defects[i].Key != want[i].Key {
			t.Errorf("defect %d = %+v, want type %q key %q",
				i, defects[i], want[i].DocumentType, want[i].Key)
		}
		if !strings.Contains(defects[i].Detail, "does not declare it on this type") {
			t.Errorf("defect %d detail = %q", i, defects[i].Detail)
		}
	}
}

// Each property is checked only against the types it applies to: the
// system property is declared on the container-level type alone, and
// the composed-modules property on the container- and module-level
// types alone, so neither is reported against the others.
func TestValidateRegistrationChecksEachPropertyOnlyWhereItApplies(t *testing.T) {
	profile := architectureProfile()
	profile.architecture.ComposedModulesProperty = "estate_of"

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 1 || defects[0].DocumentType != "wing" || defects[0].Key != "estate_of" {
		t.Errorf("defects = %+v, want one against wing naming estate_of", defects)
	}
}

func TestValidateRegistrationReportsALevelOutsideTheFour(t *testing.T) {
	profile := architectureProfile()
	profile.architecture.Elements["room"] = corpus.ElementTypeRules{Level: "fixture"}

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 1 {
		t.Fatalf("defects = %+v, want one", defects)
	}
	if defects[0].DocumentType != "room" || defects[0].Key != "fixture" {
		t.Errorf("defect = %+v, want type room key fixture", defects[0])
	}
	if !strings.Contains(defects[0].Detail, "not one of the four C4 levels") {
		t.Errorf("detail = %q", defects[0].Detail)
	}
}

func TestValidateRegistrationReportsAnUnregisteredRosterTag(t *testing.T) {
	profile := architectureProfile()
	profile.architecture.Elements["wing"] = corpus.ElementTypeRules{
		Level: corpus.LevelModule, RosterHeading: "Rooms", RosterTag: "chamber",
	}

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 1 {
		t.Fatalf("defects = %+v, want one", defects)
	}
	if defects[0].DocumentType != "wing" || defects[0].Key != "chamber" {
		t.Errorf("defect = %+v, want type wing key chamber", defects[0])
	}
	if !strings.Contains(defects[0].Detail, "not one of this profile's record-block tags") {
		t.Errorf("detail = %q", defects[0].Detail)
	}
}

// A schema that is not JSON can judge no property, so it yields no
// property defect; the other three defects do not read the schema and
// are still reported.
func TestValidateRegistrationRaisesNoPropertyDefectForAnUnparsableSchema(t *testing.T) {
	profile := architectureProfile()
	profile.schema = []byte("not json")
	profile.architecture.EntityNameProperty = "nowhere"
	profile.architecture.Elements["room"] = corpus.ElementTypeRules{Level: "fixture"}

	defects := corpus.ValidateRegistration(profile)
	if len(defects) != 1 || defects[0].Key != "fixture" {
		t.Errorf("defects = %+v, want only the level defect", defects)
	}
}
