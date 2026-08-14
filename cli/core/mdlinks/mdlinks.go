// Package mdlinks settles what a wikilink resolves to, raises the two
// findings that decide it, and substitutes the settled targets into a
// document's bytes.
//
// The two halves stay separate functions rather than one merged pass.
// Resolve settles what a reference reaches and needs only the reference's
// already-computed [specmodel.ReferenceResolution]; Substitute consumes
// Resolve's targets alongside the document's bytes and the extents to
// rewrite, which Resolve itself never touches.
// [finding.RuleUnmappedCitation] marks a citation naming a corpus document
// a caller-supplied link map leaves out, and
// [finding.RuleUnresolvedCitation] marks one naming no corpus document at
// all.
//
// Resolve reads a reference's [specmodel.ReferenceResolution], which
// [references.Resolve] already computed against the corpus. This package
// never consults an index directly: whether a slug names a corpus
// document is a question the reference already carries the answer to.
// Substitute likewise never rescans the document for `[[...]]` constructs
// or code regions: it trusts each reference's Position, which
// [references.Extract] already computed by scanning once.
package mdlinks

import (
	"path"
	"strings"
	"unicode"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// LinkMap is a caller-supplied link map: a corpus slug's mapped path, held
// exactly as the caller wrote it.
//
// A nil LinkMap means no map was supplied — including a map that was named
// but failed to load, which the Command Surface treats the same way. A
// non-nil, empty LinkMap still counts as supplied: [Resolve] then raises
// [finding.RuleUnmappedCitation] for every corpus-resolved slug it does not
// cover, exactly as a non-empty map that happens to omit one does.
type LinkMap map[string]string

// Options configures [Resolve].
type Options struct {
	// Map is the link map to consult first. See [LinkMap] for what nil
	// versus a supplied-but-empty map each mean.
	Map LinkMap
	// MapPath names the map file in an unmapped-citation finding's
	// message. Required whenever Map is non-nil; unused otherwise.
	MapPath string
}

// Resolve settles a markdown-link target for every reference in refs, and
// raises both reachability findings this package owns.
//
// refs must already carry a [specmodel.ReferenceResolution] — see
// [references.Resolve] — since this package answers only where a
// corpus-resolved reference should point, not whether it reaches the
// corpus. Every reference is settled in this one pass and every finding is
// returned, never just the first (AC-SPECMD-002.5).
//
// targets is aligned with refs by index: nil at i where refs[i] settled no
// target — an unmapped or unresolved citation — the settled target
// otherwise, anchor fragment included.
func Resolve(refs []specmodel.Reference, options Options) (targets []*string, findings []specmodel.Finding) {
	targets = make([]*string, len(refs))
	for i, ref := range refs {
		target, f := resolveOne(ref, options)
		targets[i] = target
		if f != nil {
			findings = append(findings, *f)
		}
	}
	return targets, findings
}

// resolveOne settles one reference's target, per the delta's Behavior, in
// the order the checks below run:
//
//   - a mapped slug wins outright, whether or not it also names a corpus
//     document (AC-SPECMD-005.5), and raises nothing;
//   - otherwise a slug naming no corpus document raises links/unresolved
//     and settles no target — reached only when the map did not cover it,
//     since a mapped slug has already returned;
//   - otherwise a corpus-resolved slug raises links/unmapped when a map
//     was given and simply omits it (AC-SPECMD-006.3), and derives its
//     target from the corpus when no map was given at all — never both.
func resolveOne(ref specmodel.Reference, options Options) (*string, *specmodel.Finding) {
	fragment := anchorFragment(ref.Anchor)

	if options.Map != nil {
		if mapped, ok := options.Map[ref.Slug]; ok {
			target := mapped + fragment
			return &target, nil
		}
	}

	resolved, isCorpusDocument := ref.Resolution.AsResolved()
	if !isCorpusDocument {
		f := finding.UnresolvedCitation(ref.Path, ref.Position, ref.Slug)
		return nil, &f
	}

	if options.Map != nil {
		f := finding.UnmappedCitation(ref.Path, ref.Position, ref.Slug, options.MapPath)
		return nil, &f
	}

	target := relativeTo(path.Dir(ref.Path), resolved.TargetPath) + fragment
	return &target, nil
}

// anchorFragment derives the `#`-prefixed target fragment for a
// reference's anchor, empty where the reference names none.
func anchorFragment(anchor *string) string {
	if anchor == nil {
		return ""
	}
	return "#" + slugifyAnchor(*anchor)
}

// slugifyAnchor derives a target fragment from anchor text by GitHub's
// heading-anchor rule (AC-SPECMD-001.3): lowercase the text, drop every
// character that is not a letter, digit, space, hyphen or underscore, then
// replace each remaining space with a hyphen.
func slugifyAnchor(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// relativeTo returns target's path relative to fromDir, both taken and
// returned as repository-relative, forward-slashed paths: one `..` per
// directory level climbed above the shared prefix, then target's own
// remaining components unchanged (AC-SPECMD-006.2) — with no further
// normalization, so a target beside fromDir comes back bare rather than
// `./`-prefixed (AC-SPECMD-006.1).
func relativeTo(fromDir, target string) string {
	from := dirParts(fromDir)
	to := strings.Split(path.Clean(target), "/")

	// Only directory components can match; target's final element is
	// always its file name and is never compared against a fromDir
	// component, however the two might coincide in spelling.
	limit := len(from)
	if len(to)-1 < limit {
		limit = len(to) - 1
	}
	common := 0
	for common < limit && from[common] == to[common] {
		common++
	}

	climb := len(from) - common
	parts := make([]string, 0, climb+len(to)-common)
	for i := 0; i < climb; i++ {
		parts = append(parts, "..")
	}
	parts = append(parts, to[common:]...)
	return strings.Join(parts, "/")
}

// dirParts splits a directory path into its components, nil at the
// repository root.
func dirParts(dir string) []string {
	dir = path.Clean(dir)
	if dir == "." || dir == "" {
		return nil
	}
	return strings.Split(dir, "/")
}
