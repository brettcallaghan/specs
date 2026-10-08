package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// readErrorText renders a failure to open a file in wording that does not
// vary by operating system.
//
// [os.Open] returns a [fs.PathError] whose text is the platform's own
// syscall message — "no such file or directory" on Unix, "The system
// cannot find the file specified." on Windows — so interpolating it
// verbatim makes a finding, and every fixture that captures one,
// platform-specific. The path is already carried in the finding's Path
// field and rendered as its header, so nothing is lost by reporting only
// the category here.
//
// An error matching neither category keeps its own text: an unanticipated
// failure is worth surfacing in full, and it is rare enough not to be
// worth a fixture.
func readErrorText(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "no such file"
	case errors.Is(err, fs.ErrPermission):
		return "permission denied"
	default:
		return err.Error()
	}
}

// writeErrorText renders why a file could not be written, in wording that
// does not vary by operating system: the operation and path a
// [*fs.PathError] carries are dropped, since the caller's message already
// names the destination, and the cause is rendered as [readErrorText]
// renders it.
func writeErrorText(err error) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return readErrorText(pathErr.Err)
	}
	return readErrorText(err)
}

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
	// RuleLinkMapInvalid covers a link map that cannot be read at all, or
	// that go-yaml cannot parse — including a slug repeated as a
	// top-level key, which go-yaml itself rejects as a duplicate mapping
	// key. A record missing its path property is modelled separately as
	// [RuleLinkMapMissingPath], the same missing-property split
	// [RuleConfigMissingProperty] makes from [RuleConfigSchema].
	RuleLinkMapInvalid specmodel.RuleId = "linkmap/invalid"
	// RuleLinkMapMissingPath covers a link-map record that declares no
	// path property.
	RuleLinkMapMissingPath specmodel.RuleId = "linkmap/missing-path"
	// RuleDocumentUnreadable covers the `links` operand itself failing to
	// open — a mistyped path, a permissions error — distinct from
	// [finding.RuleFrontmatterParse], which covers a document that opened
	// but whose frontmatter block did not parse.
	RuleDocumentUnreadable specmodel.RuleId = "document/unreadable"
)

// pointerPrefix is the prefix a message about the value at pointer
// begins with: the pointer and a colon, or nothing for the root pointer
// "", since a message about the object itself needs no location and a
// bare "/" says nothing to a reader. The finding's detail keeps the
// pointer as it is.
func pointerPrefix(pointer string) string {
	if pointer == "" {
		return ""
	}
	return pointer + ": "
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
		Span:     at,
		Message: fmt.Sprintf("%s%q does not satisfy %s",
			pointerPrefix(pointer), value, constraint),
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
		Span:     at,
		Message: fmt.Sprintf("%srequired property %q is not declared",
			pointerPrefix(pointer), property),
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
		Span:     at,
		Message: fmt.Sprintf("%s%q is not a key this configuration defines",
			pointerPrefix(pointer), key),
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
		Span:     at,
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
// (specs-configuration-contract.md).
func configUnknownProfile(path string, at specmodel.Span, name string, registered []string) specmodel.Finding {
	list := "none are registered"
	if len(registered) > 0 {
		list = strings.Join(registered, ", ")
	}
	return specmodel.Finding{
		Rule:     RuleConfigUnknownProfile,
		Severity: finding.SeverityError,
		Path:     path,
		Span:     at,
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
		Span:     at,
		Message:  fmt.Sprintf("profile registration: %q %s", key, defect.Detail),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "/profile/name",
			Constraint: "valid-registration",
			Value:      key,
		}),
	}
}

// linkMapUnreadable reports a link map [LoadLinkMap] could not open at
// all — most plausibly a path naming nothing, per
// [[container-spec-cli#Command Surface]]'s AC-SPECMD-005.8. There is no
// source to derive a span from, so the finding carries the position an
// empty file would.
func linkMapUnreadable(path string, err error) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleLinkMapInvalid,
		Severity: finding.SeverityError,
		Path:     path,
		Span:     wholeFile(nil),
		Message:  fmt.Sprintf("cannot be read: %s", readErrorText(err)),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "",
			Constraint: "readable",
			Value:      readErrorText(err),
		}),
	}
}

// linkMapUnparseable reports a link map go-yaml could not decode —
// invalid YAML syntax, or a slug repeated as a top-level key, which
// go-yaml itself rejects as a duplicate mapping key rather than letting
// one silently win (AC-SPECMD-005.8).
func linkMapUnparseable(path string, src []byte, err error) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleLinkMapInvalid,
		Severity: finding.SeverityError,
		Path:     path,
		Span:     wholeFile(src),
		Message:  fmt.Sprintf("cannot be parsed as YAML: %s", err),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "",
			Constraint: "yaml-syntax",
			Value:      err.Error(),
		}),
	}
}

// documentUnreadable reports the `links` operand itself failing to open,
// per the delta's Behavior ("IF the named document cannot be read... the
// links subcommand SHALL report an error finding and write no rewritten
// document"). There is no source to derive a span from, so the finding
// carries the position an empty file would — the same fallback
// [linkMapUnreadable] uses for its own no-source case.
func documentUnreadable(path string, err error) specmodel.Finding {
	return specmodel.Finding{
		Rule:     RuleDocumentUnreadable,
		Severity: finding.SeverityError,
		Path:     path,
		Span:     wholeFile(nil),
		Message:  fmt.Sprintf("cannot be read: %s", readErrorText(err)),
		Detail: specmodel.NewFindingDetailSchemaViolation(specmodel.FindingSchemaViolation{
			Pointer:    "",
			Constraint: "readable",
			Value:      readErrorText(err),
		}),
	}
}

// linkMapMissingPath reports a link-map record for slug that declares no
// path property (AC-SPECMD-005.8). The pointer names the slug the same
// way a JSON Pointer would, though the source is YAML rather than the
// JSON a schema violation elsewhere in this package points into.
func linkMapMissingPath(path string, src []byte, slug string) specmodel.Finding {
	pointer := "/" + slug
	return specmodel.Finding{
		Rule:     RuleLinkMapMissingPath,
		Severity: finding.SeverityError,
		Path:     path,
		Span:     wholeFile(src),
		Message:  fmt.Sprintf("%s: required property %q is not declared", pointer, "path"),
		Detail: specmodel.NewFindingDetailMissingProperty(specmodel.FindingMissingProperty{
			Pointer:  pointer,
			Property: "path",
		}),
	}
}
