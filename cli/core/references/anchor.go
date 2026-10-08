package references

import (
	"fmt"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// resolveAnchors records what the anchor half of every reference in refs
// resolved to, against the headings the same index carries.
//
// Like [Resolve], this is a lookup and nothing else: it opens no
// document and re-parses no frontmatter, so the same references resolved
// against the same index resolve identically on every platform and every
// run. It is kept in its own file because it answers a narrower
// question — what one anchor matched — while [Resolve] answers what a
// slug reached; the two are composed by [Resolve] rather than merged,
// so a caller that wants slug resolution alone still gets it without
// the heading lookups this needs.
func resolveAnchors(refs []specmodel.Reference, index specmodel.SpecIndex) {
	headings := outlinesByPath(index)
	for i := range refs {
		refs[i].AnchorResolution = resolveAnchor(refs[i], headings)
	}
}

// outlined is one member's recorded headings with the heading path of
// each, derived once however many references target the member.
type outlined struct {
	headings []specmodel.Heading
	paths    [][]string
}

// lazyOutlines derives a member's heading paths the first time a
// reference targets it.
type lazyOutlines struct {
	headings map[string][]specmodel.Heading
	derived  map[string]outlined
}

func outlinesByPath(index specmodel.SpecIndex) *lazyOutlines {
	return &lazyOutlines{headings: headingsByPath(index), derived: map[string]outlined{}}
}

// lookup returns the member at path's headings and paths, and whether
// the index recorded headings for that path at all.
func (l *lazyOutlines) lookup(path string) (outlined, bool) {
	if found, ok := l.derived[path]; ok {
		return found, true
	}
	headings, known := l.headings[path]
	if !known {
		return outlined{}, false
	}
	found := outlined{headings: headings, paths: HeadingPaths(headings)}
	l.derived[path] = found
	return found, true
}

// HeadingPaths returns the heading path of each of headings, as
// [markdown.HeadingPaths] derives it from their text and level.
//
// Exported because every consumer that selects a heading by path —
// anchor resolution here, and the headings and refs queries — must
// derive the same path from the same recorded headings, and one
// conversion keeps them from disagreeing.
func HeadingPaths(headings []specmodel.Heading) [][]string {
	outline := make([]markdown.Outline, len(headings))
	for i, heading := range headings {
		outline[i] = markdown.Outline{Text: heading.Text, Level: int(heading.Level)}
	}
	return markdown.HeadingPaths(outline)
}

// headingsByPath maps every corpus member's path to its recorded
// headings, so a target's headings can be looked up without a linear
// scan of the index per reference.
//
// The map holds only what index.Entries actually recorded. A path this
// index never indexed as a member — a `.feature` file's own path, most
// concretely — is absent from it rather than mapped to an empty slice,
// which is what lets [resolveAnchor] tell "no headings recorded here at
// all" from "recorded, and there are none" (AC-SPECANCH-002.4).
func headingsByPath(index specmodel.SpecIndex) map[string][]specmodel.Heading {
	byPath := make(map[string][]specmodel.Heading, len(index.Entries))
	for _, entry := range index.Entries {
		byPath[entry.Path] = entry.Headings
	}
	return byPath
}

// AnchorTarget returns the path of the document ref's anchor is checked
// against, and whether there is one to check it against at all.
//
// It is the document ref.Path names when the reference carries no slug
// of its own (AC-SPECANCH-002.6), and otherwise wherever the slug half
// already resolved. A reference carrying no anchor, and one whose slug
// is present and did not resolve, have no target: the parent feature
// already reports the second as broken, and an anchor finding on top of
// it would be a second finding for one mistake (AC-SPECANCH-002.8).
//
// Exported because the check that REPORTS a broken anchor names the
// target document in its message, and deriving that name a second time
// where the finding is built would be free to disagree with the
// derivation that decided what the anchor was matched against.
func AnchorTarget(ref specmodel.Reference) (string, bool) {
	if ref.Anchor == nil {
		return "", false
	}
	if ref.Slug == "" {
		return ref.Path, true
	}
	resolved, ok := ref.Resolution.AsResolved()
	if !ok {
		return "", false
	}
	return resolved.TargetPath, true
}

// ResolvedSelfAnchor reports whether ref is the anchor-only form —
// `[[#Heading]]`, the way a document references a section of itself
// (AC-SPECANCH-002.6) — with its anchor resolved against the document
// holding it.
//
// The slug half of such a reference is empty and matches no declared
// slug, so resolution records it unresolved. A consumer keying off that
// half alone therefore calls a reference which reaches exactly the
// heading it names broken — reporting it, dropping it from the
// reference graph, or leaving it unrewritten. The three that would are
// the check, the graph and the rewrite, and each asks this instead
// (AC-SPECREF-002.12, AC-SPECREF-001.8, AC-SPECMD-001.8), so the set
// they exempt, record and target is one set rather than three
// derivations free to disagree.
//
// The conjunction is what keeps it narrow. An anchor-only reference
// whose anchor names no heading of its own document reaches nothing and
// is not one of these, so an empty slug alone is never waved through.
func ResolvedSelfAnchor(ref specmodel.Reference) bool {
	if ref.Slug != "" || ref.Anchor == nil || ref.AnchorResolution == nil {
		return false
	}
	_, resolved := ref.AnchorResolution.AsResolved()
	return resolved
}

// resolveAnchor returns what ref's anchor resolved to, or nil where an
// anchor resolution does not apply to this reference at all.
//
// A nil result covers two cases the caller does not need to tell apart at
// the call site: ref carries no anchor, and ref carries an anchor whose
// slug half is present and did not resolve — the parent feature already
// reports that reference as broken, and an anchor finding on top of it
// would be a second finding for one mistake (AC-SPECANCH-002.8).
func resolveAnchor(ref specmodel.Reference, headings *lazyOutlines) *specmodel.AnchorResolution {
	targetPath, checkable := AnchorTarget(ref)
	if !checkable {
		return nil
	}

	target, known := headings.lookup(targetPath)
	if !known {
		result := specmodel.NewAnchorResolutionUnverified(specmodel.AnchorUnverified{
			Reason: fmt.Sprintf("no headings are recorded for %s", targetPath),
		})
		return &result
	}

	// An anchor is a heading path query: one segment is matched on its
	// identifier, and several select the heading named by the last
	// beneath each earlier one (AC-SPECANCH-002.10) — or, where they
	// spell some heading's full path, that heading alone.
	var matches []specmodel.AnchorMatch
	selected := markdown.SelectPaths(*ref.Anchor, target.paths)
	for i, heading := range target.headings {
		if selected[i] {
			matches = append(matches, specmodel.AnchorMatch{
				Text:        heading.Text,
				HeadingPath: markdown.JoinPath(target.paths[i]),
				Span:        heading.Span,
			})
		}
	}

	switch len(matches) {
	case 0:
		result := specmodel.NewAnchorResolutionUnresolved(specmodel.AnchorUnresolved{})
		return &result
	case 1:
		result := specmodel.NewAnchorResolutionResolved(specmodel.AnchorResolved{Heading: matches[0]})
		return &result
	default:
		result := specmodel.NewAnchorResolutionAmbiguous(specmodel.AnchorAmbiguous{Matches: matches})
		return &result
	}
}
