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

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/references"
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
//   - an anchor-only citation reaching a heading of the document being
//     rewritten settles on that fragment alone (AC-SPECMD-001.8), ahead
//     of the map because an empty slug names no corpus document a map
//     could cover;
//   - a mapped slug wins outright, whether or not it also names a corpus
//     document (AC-SPECMD-005.5), and raises nothing;
//   - otherwise a slug naming no corpus document raises links/unresolved
//     and settles no target — reached only when the map did not cover it,
//     since a mapped slug has already returned;
//   - otherwise a corpus-resolved slug raises links/unmapped when a map
//     was given and simply omits it (AC-SPECMD-006.3), and derives its
//     target from the corpus when no map was given at all — never both.
func resolveOne(ref specmodel.Reference, options Options) (*string, *specmodel.Finding) {
	fragment := anchorFragment(ref)

	// The one target that is not a path: the heading this reference
	// names is in the document being rewritten, so a fragment on its own
	// is what a reader follows, in the copy and in the source alike. An
	// anchor-only citation whose anchor resolved to no heading falls
	// through to the rules below and is reported and left as written
	// like any other slug naming no corpus document.
	if references.ResolvedSelfAnchor(ref) {
		return &fragment, nil
	}

	if options.Map != nil {
		if mapped, ok := options.Map[ref.Slug]; ok {
			target := mapped + fragment
			return &target, nil
		}
	}

	resolved, isCorpusDocument := ref.Resolution.AsResolved()
	if !isCorpusDocument {
		f := finding.UnresolvedCitation(ref.Path, ref.Span, ref.Slug)
		return nil, &f
	}

	if options.Map != nil {
		f := finding.UnmappedCitation(ref.Path, ref.Span, ref.Slug, options.MapPath)
		return nil, &f
	}

	target := relativeTo(path.Dir(ref.Path), resolved.TargetPath) + fragment
	return &target, nil
}

// anchorFragment derives the `#`-prefixed target fragment for a
// reference's anchor, empty where the reference names none.
//
// The text the fragment is derived FROM is the resolved heading's, not
// the anchor's, wherever the anchor resolved to a heading
// (AC-SPECMD-001.3). The two agree wherever an anchor names a heading in
// full; where they differ it is the heading that a reader following the
// emitted link has to land on, so the anchor as authored would emit a
// fragment matching nothing. Every other anchor resolution — no heading,
// several, or a target whose headings were never recorded — and an
// absent resolution alike fall back to the anchor as the author wrote it
// (AC-SPECMD-001.7), which keeps the rewrite total and the copy no worse
// than the source it was made from.
//
// The derivation itself is [markdown.Identifier], the one the Markdown
// Structure component exposes and the References component matches
// anchors through. This package deliberately holds no second copy of it:
// two derivations can disagree, and every disagreement is a link the
// checks call resolved and a reader cannot follow.
func anchorFragment(ref specmodel.Reference) string {
	if ref.Anchor == nil {
		return ""
	}

	text := *ref.Anchor
	if ref.AnchorResolution != nil {
		if resolved, ok := ref.AnchorResolution.AsResolved(); ok {
			text = resolved.Heading.Text
		}
	}
	return "#" + markdown.Identifier(text)
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
