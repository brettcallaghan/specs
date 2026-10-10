package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"

	"github.com/brettcallaghan/specs/cli/core/schema"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// configSchemaURL is the in-memory location the generated configuration
// schema is registered under, the same convention
// [[schemaregistry.resourceURL]] uses: never a real file or network
// location, just a stable identifier a $ref inside the schema resolves
// against.
const configSchemaURL = "mem://cmd/specs-config.json"

// configSchemaLocation is compiled directly, bypassing the schema
// document's own top-level "$ref": "#/definitions/SpecsConfig" — compiling
// the ref indirectly wraps every violation naming a nested property in a
// *kind.Reference in the flattened Basic output, hiding which constraint
// actually failed. Compiling the definition location directly is what
// [[schemaregistry]] does for its own per-type branches, for the same
// reason.
const configSchemaLocation = configSchemaURL + "#/definitions/SpecsConfig"

var (
	configSchemaOnce sync.Once
	configSchema     *jsonschema.Schema
	configSchemaErr  error
)

// compiledConfigSchema compiles the generated configuration schema once
// and reuses the compilation afterward, the same lazy-once discipline
// [[corpus.Corpus.schemaRegistry]] applies for the union document schema.
func compiledConfigSchema() (*jsonschema.Schema, error) {
	configSchemaOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema.SpecsConfig))
		if err != nil {
			configSchemaErr = fmt.Errorf("the generated config schema is not valid JSON: %w", err)
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(configSchemaURL, doc); err != nil {
			configSchemaErr = fmt.Errorf("registering the generated config schema: %w", err)
			return
		}
		configSchema, configSchemaErr = compiler.Compile(configSchemaLocation)
		if configSchemaErr != nil {
			configSchemaErr = fmt.Errorf("compiling the generated config schema: %w", configSchemaErr)
		}
	})
	return configSchema, configSchemaErr
}

// checkConfigSchema validates raw — the generic value tree
// [decodeConfig] produced — against the generated configuration schema,
// reporting every violation in one pass (AC-SPEC-008.4) rather than
// stopping at the first.
func checkConfigSchema(path string, src []byte, raw map[string]any, loc locations) ([]specmodel.Finding, error) {
	compiled, err := compiledConfigSchema()
	if err != nil {
		return nil, err
	}

	verr := compiled.Validate(raw)
	if verr == nil {
		return nil, nil
	}
	validationErr, ok := verr.(*jsonschema.ValidationError)
	if !ok {
		return nil, fmt.Errorf("checking %s: %w", path, verr)
	}

	// DetailedOutput, not schemaregistry's flat BasicOutput, because this
	// schema's own top-level properties (workspace, profile, conventions)
	// are themselves $ref to struct definitions: a violation nested one
	// level inside one of them — WorkspaceConfig.root's type, say — sits
	// behind two stacked $ref hops, and BasicOutput's single-step
	// Reference-unwrap does not reach through both, reporting a bare
	// "Reference" kind with none of the underlying constraint. Walking
	// DetailedOutput's tree to its leaves reaches the real kind
	// regardless of how many $ref hops sit above it. schemaregistry.go's
	// own branches never nest this way — a document type's properties are
	// validated directly against the branch's own top-level schema, with
	// at most one $ref hop for a scalar alias like Slug — which is why its
	// flat walk has been sufficient there.
	var units []jsonschema.OutputUnit
	collectLeaves(*validationErr.DetailedOutput(), &units)

	var findings []specmodel.Finding
	for _, unit := range units {
		findings = append(findings, findingsForConfigUnit(path, src, loc, unit)...)
	}
	return findings, nil
}

// collectLeaves appends every leaf of a DetailedOutput tree — a unit with
// no further nested errors — to out, in the tree's own order.
func collectLeaves(unit jsonschema.OutputUnit, out *[]jsonschema.OutputUnit) {
	if len(unit.Errors) == 0 {
		if unit.Error != nil {
			*out = append(*out, unit)
		}
		return
	}
	for _, sub := range unit.Errors {
		collectLeaves(sub, out)
	}
}

// findingsForConfigUnit converts one leaf validation failure into the
// Finding(s) it represents: one missing-property finding per property a
// "required" failure names, one undeclared-key finding per name an
// "additionalProperties" failure names — each located at its own key, not
// at the enclosing object all of them were found on — or one general
// schema-violation finding for any other constraint.
func findingsForConfigUnit(path string, src []byte, loc locations, unit jsonschema.OutputUnit) []specmodel.Finding {
	if unit.Error == nil {
		return nil
	}
	pointer := unit.InstanceLocation

	switch k := unit.Error.Kind.(type) {
	case *kind.Required:
		at := loc.keySpan(pointer, src)
		findings := make([]specmodel.Finding, 0, len(k.Missing))
		for _, property := range k.Missing {
			findings = append(findings, configMissingProperty(path, at, pointer, property))
		}
		return findings

	case *kind.AdditionalProperties:
		findings := make([]specmodel.Finding, 0, len(k.Properties))
		for _, name := range k.Properties {
			keyPointer := pointer + "/" + name
			findings = append(findings, configUndeclaredKey(path, loc.keySpan(keyPointer, src), keyPointer, name))
		}
		return findings

	default:
		at := loc.valueSpan(pointer, src)
		constraint := strings.Join(unit.Error.Kind.KeywordPath(), "/")
		if constraint == "" {
			constraint = "schema"
		}
		value := loc.text(pointer, src)
		return []specmodel.Finding{configSchemaViolation(path, at, pointer, constraint, value)}
	}
}
