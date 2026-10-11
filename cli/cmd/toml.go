package cmd

import (
	"errors"

	"github.com/BurntSushi/toml"

	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// decodeConfig parses src as TOML into a generic value tree suitable for
// [checkConfigSchema], reporting a syntax error as a Finding rather than
// as a bare decode failure — [[contract-specs-configuration]] checks
// specs.toml rather than merely loading it, and a syntax error is a
// defect in the file like any other.
//
// A syntax error is total: BurntSushi/toml has nothing to report a
// partial tree for, so decodeConfig returns exactly one finding and a nil
// tree in that case, distinct from [checkConfigSchema] finding several
// defects in one well-formed-but-invalid file.
func decodeConfig(path string, src []byte) (map[string]any, []specmodel.Finding) {
	var raw map[string]any
	if _, err := toml.Decode(string(src), &raw); err != nil {
		return nil, []specmodel.Finding{configSyntaxError(path, syntaxErrorSpan(src, err), err.Error())}
	}
	return raw, nil
}

// syntaxErrorSpan locates a TOML syntax error, using the byte position
// BurntSushi/toml's structured [toml.ParseError] reports where the
// failure is one, and the whole file otherwise.
func syntaxErrorSpan(src []byte, err error) specmodel.Span {
	var perr toml.ParseError
	if errors.As(err, &perr) {
		length := perr.Position.Len
		if length < 1 {
			length = 1
		}
		return span.Of(src, perr.Position.Start, length)
	}
	return span.Of(src, 0, len(src))
}

// profileNameFrom extracts `[profile].name` from a decoded generic value
// tree, independent of whether any other part of the tree is
// schema-valid — so an unregistered profile name is still checked, and
// reported, in the same run as an unrelated defect elsewhere in the file
// (specs-configuration-contract.md's "report every defect... rather than
// stopping at the first").
func profileNameFrom(raw map[string]any) (string, bool) {
	section, ok := raw["profile"].(map[string]any)
	if !ok {
		return "", false
	}
	name, ok := section["name"].(string)
	return name, ok
}
