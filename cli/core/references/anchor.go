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
	headings := headingsByPath(index)
	for i := range refs {
		refs[i].AnchorResolution = resolveAnchor(refs[i], headings)
	}
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

// resolveAnchor returns what ref's anchor resolved to, or nil where an
// anchor resolution does not apply to this reference at all.
//
// A nil result covers two cases the caller does not need to tell apart at
// the call site: ref carries no anchor, and ref carries an anchor whose
// slug half is present and did not resolve — the parent feature already
// reports that reference as broken, and an anchor finding on top of it
// would be a second finding for one mistake (AC-SPECANCH-002.8).
func resolveAnchor(ref specmodel.Reference, headings map[string][]specmodel.Heading) *specmodel.AnchorResolution {
	targetPath, checkable := AnchorTarget(ref)
	if !checkable {
		return nil
	}

	targetHeadings, known := headings[targetPath]
	if !known {
		result := specmodel.NewAnchorResolutionUnverified(specmodel.AnchorUnverified{
			Reason: fmt.Sprintf("no headings are recorded for %s", targetPath),
		})
		return &result
	}

	wanted := markdown.Identifier(*ref.Anchor)
	var matches []specmodel.AnchorMatch
	for _, heading := range targetHeadings {
		if markdown.Identifier(heading.Text) == wanted {
			matches = append(matches, specmodel.AnchorMatch{
				Text:     heading.Text,
				Position: heading.Position,
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
