package corpus

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/brettcallaghan/specs/cli/core/schemaregistry"
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
// Thirteen defects are found, and they are the whole of what a registration
// can get wrong on its own terms:
//
//   - a pairing or a typed reference naming a property the same document
//     type does not register as reference-bearing;
//   - a typed reference naming a required target type the profile does
//     not define;
//   - a lifecycle registration naming a status outside the profile's own
//     status vocabulary;
//   - a feature-tag registration whose identifier pattern does not
//     compile;
//   - a feature-tag registration whose shared captures name a capture
//     group an identifier pattern does not define;
//   - a record-block registration whose schema pointer resolves to
//     nothing in the profile's schema;
//   - a record-block tag registered more than once;
//   - an architecture registration naming a document type the profile
//     does not define;
//   - an architecture registration naming a property the union schema
//     does not declare on a type the property applies to;
//   - an element type registered at a level outside the four C4 levels;
//   - a roster record-block tag registered with no roster heading;
//   - a roster heading registered with no roster record-block tag;
//   - a roster record-block tag that is not a registered record-block
//     tag.
//
// The feature-tag two are found here rather than where they are used because
// both are defects of the registration against itself, which is what
// this validation is for: a pattern that cannot compile matches nothing,
// and a derivation over a capture no pattern defines compares two empty
// values and passes. Everything else a registration can get wrong — a
// slug pattern that does not compile, a schema that is not JSON — is
// found by the component that uses it, against the document it was used
// on.
//
// The result is empty for a registration with no defect, and ordered
// deterministically: by document type, then by key, with the
// profile-level lifecycle defects, the feature-tag defects, the
// record-block defects and then the architecture defects last.
// Callers render it, so its order is output.
func ValidateRegistration(profile Profile) []RegistrationDefect {
	if profile == nil {
		return nil
	}

	var defects []RegistrationDefect
	types := profile.DocumentTypes()
	for _, name := range sortedKeys(types) {
		defects = append(defects, documentTypeDefects(name, types[name], types)...)
	}
	defects = append(defects, lifecycleDefects(profile.Lifecycle())...)
	defects = append(defects, featureTagDefects(profile.FeatureTags())...)
	defects = append(defects, recordBlockDefects(profile.RecordBlocks(), profile.Schema())...)
	return append(defects, architectureDefects(profile)...)
}

// architectureDefects reports the faults in an architecture registration:
// a registered type the profile does not define, a level outside the
// four, a registered property the union schema does not declare on a
// type it applies to, and a roster registration that reads no child: a
// roster tag with no roster heading, a roster heading with no roster
// tag, or a roster tag that is not a registered record-block tag.
//
// Defects follow the element types in ascending order — for each, an
// undefined type alone, otherwise its level, then its properties in the
// order [ArchitectureRules] declares them, then its roster tag — and then
// the contract type and its provider property. A type the profile does
// not define has no other defect reported against it: its rules would
// be checked against a branch that does not exist.
//
// A profile registering no architecture vocabulary has no defect: the
// zero value is a corpus whose documents declare no elements. A schema
// that is not JSON yields no property defect, for the reason
// [recordBlockDefects] gives for pointers.
func architectureDefects(profile Profile) []RegistrationDefect {
	rules := profile.Architecture()
	if len(rules.Elements) == 0 && rules.ContractType == "" {
		return nil
	}

	types := profile.DocumentTypes()
	// A nil registry is a schema that is not JSON.
	registry, _ := schemaregistry.New(profile.Schema())
	tags := make(map[string]bool)
	for _, tag := range profile.RecordBlocks() {
		tags[tag.Tag] = true
	}

	var defects []RegistrationDefect
	for _, name := range sortedKeys(rules.Elements) {
		element := rules.Elements[name]
		if _, defined := types[name]; !defined {
			defects = append(defects, undefinedTypeDefect(name, "an element type"))
			continue
		}
		if !element.Level.Valid() {
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          string(element.Level),
				Detail: fmt.Sprintf("is registered as this element type's level, but is not one "+
					"of the four C4 levels (%s, %s, %s, %s)",
					LevelSystem, LevelContainer, LevelModule, LevelComponent),
			})
		}
		for _, property := range []struct {
			name    string
			role    string
			applies bool
		}{
			{rules.EntityNameProperty, "the entity-name property", true},
			{rules.SystemProperty, "the system property", element.Level == LevelContainer},
			{rules.ComposedModulesProperty, "the composed-modules property",
				element.Level == LevelContainer || element.Level == LevelModule},
			{rules.ConsumedContractsProperty, "the consumed-contracts property", true},
		} {
			if property.applies {
				defects = append(defects, undeclaredPropertyDefect(registry, name, property.name, property.role)...)
			}
		}
		switch {
		case element.RosterTag != "" && element.RosterHeading == "":
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          element.RosterTag,
				Detail: "is registered as the roster record-block tag, but no roster heading " +
					"is registered, so no child would ever be read",
			})
		case element.RosterHeading != "" && element.RosterTag == "":
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          element.RosterHeading,
				Detail: "is registered as the roster heading, but no roster record-block tag " +
					"is registered, so no child would ever be read under it",
			})
		case element.RosterTag != "" && !tags[element.RosterTag]:
			defects = append(defects, RegistrationDefect{
				DocumentType: name,
				Key:          element.RosterTag,
				Detail: fmt.Sprintf("is registered as the record-block tag declaring a child under "+
					"the roster heading %q, but is not one of this profile's record-block tags, "+
					"so no child would ever be read there", element.RosterHeading),
			})
		}
	}

	if rules.ContractType != "" {
		if _, defined := types[rules.ContractType]; !defined {
			defects = append(defects, undefinedTypeDefect(rules.ContractType, "the contract type"))
		} else {
			defects = append(defects, undeclaredPropertyDefect(
				registry, rules.ContractType, rules.ProviderProperty, "the provider property")...)
		}
	}
	return defects
}

// undefinedTypeDefect reports an architecture registration naming, in
// role, a document type the profile does not define.
func undefinedTypeDefect(name, role string) RegistrationDefect {
	return RegistrationDefect{
		Key: name,
		Detail: "is registered as " + role + " in the architecture vocabulary, " +
			"but is not a document type this profile defines, so no document could declare it",
	}
}

// undeclaredPropertyDefect reports property, registered in role, when
// the union schema does not declare it on documentType. An empty
// property is not registered and has no defect; a nil registry is a
// schema that is not JSON, against which nothing can be judged.
func undeclaredPropertyDefect(registry *schemaregistry.Registry, documentType, property, role string) []RegistrationDefect {
	if property == "" || registry == nil {
		return nil
	}
	declared, err := registry.Declares(documentType, property)
	if err == nil && declared {
		return nil
	}
	return []RegistrationDefect{{
		DocumentType: documentType,
		Key:          property,
		Detail: "is registered as " + role + " of this type, but the profile's schema does not " +
			"declare it on this type, so every document of the type would be read as lacking it",
	}}
}

// recordBlockDefects reports each record-block tag whose schema pointer
// does not resolve to a definition in schema, and each registration of a
// tag already registered earlier. A tag registered with no pointer has no
// pointer defect: its records are simply not validated. Defects follow
// registration order, a registration's duplicate defect before its
// pointer defect.
//
// An unresolvable pointer is reported rather than skipped because
// skipping it would validate every record against nothing and report
// success.
//
// A schema that is not JSON yields no pointer defect: no pointer can be
// judged against it, and the fault is the schema's rather than any tag's.
// It is reported, once, by the schema registry the corpus compiles the
// active profile's schema with. The duplicate check does not read the
// schema and still runs.
func recordBlockDefects(tags []RecordBlockTag, schema []byte) []RegistrationDefect {
	var document any
	parsed := json.Unmarshal(schema, &document) == nil

	var defects []RegistrationDefect
	first := make(map[string]RecordBlockTag, len(tags))
	for _, tag := range tags {
		if earlier, seen := first[tag.Tag]; seen {
			detail := "is registered as a record block more than once, so which " +
				"schema a record is validated against would depend on registration order"
			if earlier.Schema != tag.Schema {
				detail += fmt.Sprintf(" (pointers %q and %q)", earlier.Schema, tag.Schema)
			}
			defects = append(defects, RegistrationDefect{Key: tag.Tag, Detail: detail})
		} else {
			first[tag.Tag] = tag
		}
		if !parsed || tag.Schema == "" || resolvePointer(document, tag.Schema) {
			continue
		}
		defects = append(defects, RegistrationDefect{
			Key: tag.Tag,
			Detail: fmt.Sprintf("is registered as a record block validated against %q, "+
				"which resolves to nothing in this profile's schema, so its records "+
				"would be checked against nothing", tag.Schema),
		})
	}
	return defects
}

// resolvePointer reports whether fragment, a JSON Pointer in its URI
// fragment form (RFC 6901 sections 4 to 6) such as "#/definitions/Name",
// addresses a value in document.
//
// The fragment is percent-decoded, must then be empty (the whole
// document) or start with "/", and is split into reference tokens with
// "~1" read as "/" and "~0" as "~"; any other "~" makes the pointer
// invalid. A token steps into an object by key and into an array by
// index: "0" or a decimal with no leading zero, within range. "-" and any other token on an array, and any step into a
// scalar, resolve to nothing.
func resolvePointer(document any, fragment string) bool {
	encoded, ok := strings.CutPrefix(fragment, "#")
	if !ok {
		return false
	}
	path, err := url.PathUnescape(encoded)
	if err != nil {
		return false
	}
	if path == "" {
		return true
	}
	path, ok = strings.CutPrefix(path, "/")
	if !ok {
		return false
	}
	current := document
	for _, token := range strings.Split(path, "/") {
		if !validEscapes(token) {
			return false
		}
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		switch node := current.(type) {
		case map[string]any:
			next, found := node[token]
			if !found {
				return false
			}
			current = next
		case []any:
			index, valid := arrayIndex(token, len(node))
			if !valid {
				return false
			}
			current = node[index]
		default:
			return false
		}
	}
	return true
}

// validEscapes reports whether every "~" in token begins "~0" or "~1",
// the only escapes RFC 6901 defines.
func validEscapes(token string) bool {
	for i := 0; i < len(token); i++ {
		if token[i] == '~' {
			if i+1 == len(token) || (token[i+1] != '0' && token[i+1] != '1') {
				return false
			}
			i++
		}
	}
	return true
}

// arrayIndex parses token as an RFC 6901 array index below length.
func arrayIndex(token string, length int) (int, bool) {
	if token == "" || (len(token) > 1 && token[0] == '0') {
		return 0, false
	}
	for _, r := range token {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	index, err := strconv.Atoi(token)
	if err != nil || index >= length {
		return 0, false
	}
	return index, true
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

// featureTagDefects reports the faults in a feature-tag registration
// that are visible in the registration alone: an identifier pattern that
// does not compile, and a shared capture that one of those patterns does
// not define.
//
// A profile registering no feature tags has no defect: the zero value is
// a corpus with no tagged features. A profile that registers shared
// captures but no pattern has one defect per capture, which is the
// honest reading — the empty pattern defines no capture group at all —
// and it is what keeps this check from passing vacuously on the
// registration most likely to be half-written.
func featureTagDefects(tags FeatureTagRules) []RegistrationDefect {
	patterns := []struct {
		expression string
		role       string
	}{
		{tags.FeaturePattern, "a feature identifier"},
		{tags.CriterionPattern, "an acceptance criterion identifier"},
	}

	var defects []RegistrationDefect
	captures := make([]map[string]bool, len(patterns))
	for i, pattern := range patterns {
		compiled, err := regexp.Compile(pattern.expression)
		if err != nil {
			defects = append(defects, RegistrationDefect{
				Key: pattern.expression,
				Detail: fmt.Sprintf("is registered as the pattern %s matches, "+
					"but does not compile as a regular expression: %v", pattern.role, err),
			})
			continue
		}
		captures[i] = map[string]bool{}
		for _, name := range compiled.SubexpNames() {
			if name != "" {
				captures[i][name] = true
			}
		}
	}

	for _, capture := range tags.SharedCaptures {
		for i, pattern := range patterns {
			if captures[i] == nil || captures[i][capture] {
				continue
			}
			defects = append(defects, RegistrationDefect{
				Key: capture,
				Detail: fmt.Sprintf("is registered as shared by a criterion identifier and its "+
					"feature's, but the pattern %s matches defines no capture group of that name, "+
					"so the derivation would compare a value nothing captured", pattern.role),
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
