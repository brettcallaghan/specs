package corpus

import (
	"fmt"
	"sort"
)

// RegistrationDefect is one fault found in an active profile's
// registration.
//
// It is data, not a rendered message and not an error: a registration is
// compiled in, so it has no position of its own, and the position a
// reader needs is the one in the configuration that selected the profile.
// Only the caller knows that position, so [ValidateRegistration] hands
// the defect back and the caller turns it into a finding.
type RegistrationDefect struct {
	// DocumentType is the document type whose rules carry the defect,
	// keyed as [Profile.DocumentTypes] keys it. Empty where the defect is
	// in the profile-level lifecycle registration, which belongs to no
	// one type.
	DocumentType string

	// Key is the offending key exactly as the profile registered it: the
	// name of a property, or a status.
	//
	// It is reported so that the caller's finding names what to fix. A
	// registration defect reported without its key sends a reader
	// searching a table for the entry that caused it.
	Key string

	// Detail says what is wrong with Key, in one sentence and without
	// repeating it.
	Detail string
}

// ValidateRegistration reports every defect in profile's registration.
//
// A profile is data, and data can contradict itself in ways no compiler
// catches: a pairing over a property nothing registered as
// reference-bearing checks a value that was never extracted, and reports
// success. That is the vacuous pass this validation exists to turn into a
// finding, and it is why every defect is reported rather than the first,
// and why nothing here substitutes a default for an offending key — a
// registration the caller was told is wrong is better than a registration
// silently repaired into one that checks less than it says.
//
// Three defects are found, and they are the whole of what a registration
// can get wrong on its own terms:
//
//   - a pairing or a typed reference naming a property the same document
//     type does not register as reference-bearing;
//   - a typed reference naming a required target type the profile does
//     not define;
//   - a lifecycle registration naming a status outside the profile's own
//     status vocabulary.
//
// Everything else a registration can get wrong — a slug pattern that does
// not compile, a schema that is not JSON — is found by the component that
// uses it, against the document it was used on.
//
// The result is empty for a registration with no defect, and ordered
// deterministically: by document type, then by key, with the
// profile-level lifecycle defects last. Callers render it, so its order
// is output.
func ValidateRegistration(profile Profile) []RegistrationDefect {
	if profile == nil {
		return nil
	}

	var defects []RegistrationDefect
	types := profile.DocumentTypes()
	for _, name := range sortedKeys(types) {
		defects = append(defects, documentTypeDefects(name, types[name], types)...)
	}
	return append(defects, lifecycleDefects(profile.Lifecycle())...)
}

// documentTypeDefects reports the defects in one document type's rules.
// types is the whole registration, because a typed reference names a type
// and whether the profile defines it is a question about all of them.
func documentTypeDefects(name string, rules DocumentTypeRules, types map[string]DocumentTypeRules) []RegistrationDefect {
	bearing := make(map[string]bool, len(rules.ReferenceProperties))
	for _, property := range rules.ReferenceProperties {
		bearing[property] = true
	}

	var defects []RegistrationDefect
	for _, property := range sortedKeys(rules.TypedReferenceProperties) {
		if !bearing[property] {
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          property,
				Detail: "is registered as a typed reference but not as a reference-bearing " +
					"property of this type, so nothing would resolve for its target type to be checked against",
			})
		}
		target := rules.TypedReferenceProperties[property]
		if _, defined := types[target]; !defined {
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          property,
				Detail: fmt.Sprintf("requires its target to declare the document type %q, "+
					"which this profile does not define", target),
			})
		}
	}

	for _, pairing := range rules.Pairings {
		if bearing[pairing.Property] {
			continue
		}
		defects = append(defects, RegistrationDefect{
			DocumentType: name,
			Key:          pairing.Property,
			Detail: fmt.Sprintf("is paired with %q but is not registered as a reference-bearing "+
				"property of this type, so the pairing would compare a reference nothing extracted",
				pairing.Counterpart),
		})
	}
	return defects
}

// lifecycleDefects reports the statuses a lifecycle registration names
// that the profile's own vocabulary does not contain.
//
// A profile registering no lifecycle at all has no defect: the zero value
// is a corpus whose documents have no lifecycle. A profile that registers
// statuses but no vocabulary has one defect per registered status, which
// is the honest reading — every one of them is outside a vocabulary that
// is empty — and it is what keeps this check from passing vacuously on
// the registration most likely to be half-written.
func lifecycleDefects(lifecycle LifecycleRules) []RegistrationDefect {
	vocabulary := make(map[string]bool, len(lifecycle.Statuses))
	for _, status := range lifecycle.Statuses {
		vocabulary[status] = true
	}

	var defects []RegistrationDefect
	for _, registered := range []struct {
		statuses []string
		meaning  string
	}{
		{lifecycle.LiveStatuses, "at which a document is in effect"},
		{lifecycle.RetiredStatuses, "at which a document is no longer in effect"},
	} {
		for _, status := range registered.statuses {
			if vocabulary[status] {
				continue
			}
			defects = append(defects, RegistrationDefect{
				Key: status,
				Detail: "is registered as a status " + registered.meaning +
					", but is not one of the statuses this profile defines",
			})
		}
	}
	return defects
}

// sortedKeys returns m's keys in ascending order, so that a defect list
// built by walking a registration is the same list on every run.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
