package cmd

import (
	"os"
	"sort"

	"github.com/goccy/go-yaml"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// LinkMap resolves a corpus document's slug to the path a link to it is
// to use, exactly as a link-map file's author wrote it — see
// [LoadLinkMap]. A path value is never interpreted, resolved or
// normalized on the way into this map, and it is a caller's, not this
// package's, discipline to keep it that way on the way out.
type LinkMap map[string]string

// linkMapRecord is one link-map entry's decoded shape. A record may
// carry any other property — a copying script's own per-document data,
// for instance — and this package reads Path and ignores the rest
// (AC-SPECMD-005.6); go-yaml's default decoding into a struct already
// does that without any option needing to be set.
type linkMapRecord struct {
	Path string `yaml:"path"`
}

// LoadLinkMap reads path as YAML — a mapping whose keys are corpus
// document slugs, each holding a record that carries a path property —
// mirroring [Load]'s specs.toml handling in this package: the map is
// caller-supplied file input, so it is read by the Command Surface, and
// every defect it carries is reported as a finding in one run rather
// than resolving a map with any known defect (AC-SPECMD-005.4).
//
// Unlike [Load], there is no ErrConfigNotFound distinction here: the
// link map is a per-invocation input a caller names explicitly, not
// repository configuration with its own absence contract, so an
// unreadable file is reported the same way an unparseable one is — a
// finding naming the file and the fault, and no map resolved
// (AC-SPECMD-005.8).
//
// A slug appearing twice as a top-level key is a parse fault, not a
// silent overwrite: go-yaml rejects a duplicate mapping key while
// decoding unless told otherwise, and LoadLinkMap tells it nothing
// otherwise.
//
// The file is decoded with a YAML parser, so a map written in JSON is
// accepted unchanged — JSON is valid YAML (AC-SPECMD-005.7).
//
// LoadLinkMap reports every record missing a path in one run rather than
// stopping at the first, the same "report every defect" discipline
// [Load] applies to specs.toml; findings are ordered by slug so the
// result is deterministic despite Go's randomized map iteration.
func LoadLinkMap(path string) (LinkMap, []specmodel.Finding) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, []specmodel.Finding{linkMapUnreadable(path, err)}
	}

	// Decoded untyped first, so a syntax fault — the file is not YAML —
	// is told apart from a shape fault — it is YAML, but not a mapping of
	// slugs to records declaring a path — and the shape fault is reported
	// with the shape expected rather than with the decoder's account of
	// the mismatch (AC-SPECMD-005.13).
	var untyped any
	if err := yaml.Unmarshal(src, &untyped); err != nil {
		return nil, []specmodel.Finding{linkMapUnparseable(path, src, err)}
	}
	if shape := linkMapShapeFaults(path, src, untyped); len(shape) > 0 {
		return nil, shape
	}

	var raw map[string]linkMapRecord
	if err := yaml.Unmarshal(src, &raw); err != nil {
		return nil, []specmodel.Finding{linkMapUnparseable(path, src, err)}
	}

	slugs := make([]string, 0, len(raw))
	for slug := range raw {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	var findings []specmodel.Finding
	resolved := make(LinkMap, len(raw))
	for _, slug := range slugs {
		record := raw[slug]
		if record.Path == "" {
			findings = append(findings, linkMapMissingPath(path, src, slug))
			continue
		}
		resolved[slug] = record.Path
	}
	if len(findings) > 0 {
		return nil, findings
	}
	return resolved, nil
}

// linkMapShapeFaults reports every way the decoded document departs from
// a link map's shape: a top level that is not a mapping, and, ordered by
// slug, each entry that is not a record or whose path is not a string. An
// empty document is an empty map, not a fault.
func linkMapShapeFaults(path string, src []byte, document any) []specmodel.Finding {
	if document == nil {
		return nil
	}
	entries, ok := document.(map[string]any)
	if !ok {
		return []specmodel.Finding{linkMapWrongShape(path, src, "", describeYAML(document))}
	}

	slugs := make([]string, 0, len(entries))
	for slug := range entries {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	var findings []specmodel.Finding
	for _, slug := range slugs {
		record, isRecord := entries[slug].(map[string]any)
		if !isRecord {
			findings = append(findings, linkMapWrongShape(path, src, slug, describeYAML(entries[slug])))
			continue
		}
		if value, declared := record["path"]; declared {
			if _, isString := value.(string); !isString {
				findings = append(findings, linkMapWrongShape(path, src, slug, "a path that is "+describeYAML(value)))
			}
		}
	}
	return findings
}

// describeYAML names the kind of a decoded YAML value, as a finding
// states what it found.
func describeYAML(value any) string {
	switch value.(type) {
	case nil:
		return "an empty value"
	case string:
		return "a string"
	case bool:
		return "a boolean"
	case int, int64, uint64, float64:
		return "a number"
	case []any:
		return "a list"
	case map[string]any:
		return "a mapping"
	}
	return "a scalar"
}

// wholeFile is the span a link-map finding carries when the fault has no
// narrower location: the file could not be opened at all, or the fault
// is a decode failure go-yaml reports without a token position this
// package translates.
func wholeFile(src []byte) specmodel.Span {
	return span.Of(src, 0, len(src))
}
