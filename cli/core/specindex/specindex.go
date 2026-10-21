// Package specindex holds the hand-written half of the index model: the
// schema version the emitter declares, and the constructor that pins it.
//
// The shape of the index — [specmodel.SpecIndex] and everything it
// contains — is generated from schemas/index.strata, because the index
// crosses a boundary and its JSON Schema is the interim machine-readable
// specification of that boundary. Strata has no notion of a constant, so
// the one value the shape cannot express lives here instead, next to the
// only function that writes it.
//
// This package exists as a separate directory rather than as a file
// beside the generated structs because scripts/generate-schemas.sh
// rebuilds cli/core/specmodel/ from scratch on every run: a hand-written
// file placed there, test file or not, is deleted by the next
// regeneration.
package specindex

import "github.com/brettcallaghan/specs/cli/core/specmodel"

// SchemaVersion is the value every emitted index declares in its
// schema_version property.
//
// It is pinned to "2.2" for as long as the index's shape only grows:
// later features add properties to the index and to its entries under
// this same version. Changing the shape of what is already there —
// rather than adding alongside it — updates this constant, the Strata
// source that constrains the property, and the emitted index in one
// change, per the Change guidance in
// specs/modules/document-core/document-core-module.md.
const SchemaVersion = "2.2"

// New returns an empty index for the given workspace, with the schema
// version already declared.
//
// Every collection is non-nil so that an index containing no members
// marshals its collections as [] rather than as null. That matters
// because the emitted index is compared byte for byte across platforms,
// and a consumer that iterates a collection should not have to
// distinguish "empty" from "absent".
func New(workspace string) specmodel.SpecIndex {
	return specmodel.SpecIndex{
		SchemaVersion: SchemaVersion,
		Workspace:     workspace,
		Entries:       []specmodel.IndexEntry{},
		Exclusions:    []specmodel.IndexExclusion{},
		References:    []specmodel.Reference{},
		Features:      []specmodel.Feature{},
	}
}
