package cmd

import (
	"fmt"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// The rule identifiers this package's checks raise.
//
// They are not [finding.RuleSchemaViolation] and its siblings: those name
// a frontmatter defect, and reusing them for a specs.toml defect would
// mislabel a configuration problem as a document one. The vocabulary
// itself — the Finding envelope and its severities — is
// [[contract-specs-configuration]]'s "report configuration defects
// through the same finding machinery as corpus defects"; only the rule
// identifiers and messages are this package's own, namespaced under
// config/ the way finding.go namespaces its own under frontmatter/ and
// slug/.
const (
	// RuleConfigSchema covers a specs.toml value that violates the
	// generated configuration schema for a reason other than a missing or
	// undeclared property, which are modelled separately below because a
	// consumer acting on either needs the property name as a value rather
	// than as a substring of a message.
	RuleConfigSchema specmodel.RuleId = "config/schema"
	// RuleConfigMissingProperty covers a specs.toml table missing a
	// property the configuration schema requires.
	RuleConfigMissingProperty specmodel.RuleId = "config/missing-property"
	// RuleConfigUndeclaredKey covers a key specs.toml declares that the
	// configuration schema does not define (AC-SPEC-008.4) — most often a
	// typo, which is exactly why an undeclared key is checked rather than
	// silently ignored.
	RuleConfigUndeclaredKey specmodel.RuleId = "config/undeclared-key"
	// RuleConfigUnknownProfile covers a [profile].name naming no profile
	// in the registry.
	RuleConfigUnknownProfile specmodel.RuleId = "config/unknown-profile"
	// RuleConfigInvalidRegistration covers a defect in the active
	// profile's own registration — a pairing or typed reference naming a
	// property never registered as reference-bearing, a typed reference
	// naming a target type the profile does not define, or a lifecycle
	// status outside the profile's own vocabulary — reported per
	// [corpus.ValidateRegistration]. A registration is compiled in and so
	// has no position of its own; every finding this rule covers is
	// positioned at [profile].name, the configuration value that selected
	// it (AC-SPECREF-004.3).
	RuleConfigInvalidRegistration specmodel.RuleId = "config/invalid-registration"
)

// rootPointer is what an empty JSON Pointer renders as in a message, so a
// defect about the configuration object itself does not read as
// beginning with a bare colon — the same substitution finding.go makes
// for a frontmatter-level defect.
const rootPointer = "/"

// renderPointer renders a JSON Pointer for a message.
func renderPointer(pointer string) string {
	if pointer == "" {
		return rootPointer
	}
	return pointer
}

// configSchemaViolation reports a specs.toml value that fails the
// generated configuration schema for a reason other than a missing or
// undeclared property.
//
// value is the exact source text the author wrote — quotes and all,
// where the scanner located one — rather than a re-rendering of the
// decoded value, the same raw-text discipline
// [[schemaregistry.SchemaViolation]] applies to a frontmatter defect.
func configSchemaViolation(path string, at specmodel.Span, pointer, constraint, value string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleConfigSchema,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message: fmt.Sprintf("%s: %q does not satisfy %s",
			renderPointer(pointer), value, constraint),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    pointer,
			Constraint: constraint,
			Value:      value,
		}),
	}
}

// configMissingProperty reports a specs.toml table missing a property the
// configuration schema requires.
func configMissingProperty(path string, at specmodel.Span, pointer, property string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleConfigMissingProperty,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message: fmt.Sprintf("%s: required property %q is not declared",
			renderPointer(pointer), property),
		Detail: specmodel.NewFindingDetailMissingProperty(specmodel.FindingMissingProperty{
			Pointer:  pointer,
			Property: property,
		}),
	}
}

// configUndeclaredKey reports a key specs.toml declares that the
// configuration schema does not define.
//
// It reuses [specmodel.FindingSchemaViolation] rather than a shape of its
// own: this package adds no new generated Finding detail variant (the
// union is closed by schemas/finding.strata and regenerating it is out of
// this change's scope), and "additionalProperties" is the schema
// vocabulary's own name for exactly this constraint. Value carries the
// offending key's own name — the "value" that violates the constraint is
// the key's mere presence, not a value it holds.
func configUndeclaredKey(path string, at specmodel.Span, pointer, key string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleConfigUndeclaredKey,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message: fmt.Sprintf("%s: %q is not a key this configuration defines",
			renderPointer(pointer), key),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    pointer,
			Constraint: "additionalProperties",
			Value:      key,
		}),
	}
}

// configSyntaxError reports a specs.toml file BurntSushi/toml could not
// parse at all. It carries [RuleConfigSchema] rather than a rule of its
// own: a syntax error and a schema violation are both "this file does not
// satisfy the configuration contract", and a consumer branching on rule
// identifier has no reason to tell them apart.
func configSyntaxError(path string, at specmodel.Span, message string) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleConfigSchema,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message:  message,
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "",
			Constraint: "toml-syntax",
			Value:      message,
		}),
	}
}

// configUnknownProfile reports a [profile].name naming no profile in the
// registry, listing the registered names
// (contract-specs-configuration.md).
func configUnknownProfile(path string, at specmodel.Span, name string, registered []string) specmodel.Finding {
	list := "none are registered"
	if len(registered) > 0 {
		list = strings.Join(registered, ", ")
	}
	return specmodel.Finding{
		Rule:     RuleConfigUnknownProfile,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message:  fmt.Sprintf("%q is not a registered profile; registered: %s", name, list),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "/profile/name",
			Constraint: "known-profile",
			Value:      name,
		}),
	}
}

// configRegistrationDefect reports one defect [corpus.ValidateRegistration]
// found in the active profile's own registration, positioned at
// [profile].name — the only file position a compiled-in registration has
// (AC-SPECREF-004.3). key names the offending property or status,
// qualified with its document type where the defect belongs to one, so a
// reader sees exactly what [corpus.RegistrationDefect.Key] alone would
// leave ambiguous between two document types registering the same
// property name.
func configRegistrationDefect(path string, at specmodel.Span, defect corpus.RegistrationDefect) specmodel.Finding {
	key := defect.Key
	if defect.DocumentType != "" {
		key = fmt.Sprintf("%s.%s", defect.DocumentType, defect.Key)
	}
	return specmodel.Finding{
		Rule:     RuleConfigInvalidRegistration,
		Severity: finding.SeverityError,
		Path:     path,
		Position: at,
		Message:  fmt.Sprintf("profile registration: %q %s", key, defect.Detail),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "/profile/name",
			Constraint: "valid-registration",
			Value:      key,
		}),
	}
}
