package references

import (
	"path"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/envelope"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Resolve records what every reference in refs reached, against the
// corpus that index describes — both the slug half every reference
// carries and, where present, the anchor half.
//
// It is the half of this component that [Extract] deliberately leaves
// undone: extraction knows a document, resolution knows a corpus, and
// the two are separated so that a reference cannot reach the index
// without a corpus having been consulted. Every reference is given a
// slug resolution, including the ones that reached nothing — indexing
// never fails, and reporting a broken reference is a check's job. An
// anchor resolution is given only where one applies — see
// [resolveAnchor].
//
// Resolution is a lookup and nothing else. No document is read, no
// frontmatter is parsed, and the answer is a function of index alone, so
// the same references resolved against the same index resolve the same
// way on every platform and on every run. Matching an anchor against a
// heading is a corpus question for the same reason: the headings it
// matches against already live on index, populated by the Markdown
// Structure component, not read from disk again here.
//
// The references are resolved in place, because they are the rows the
// index will carry and copying them to attach these properties to each
// would only give a caller two collections to keep in step.
//
// An embed (`![[…]]`) resolves exactly as the equivalent link: nothing
// on [specmodel.Reference] distinguishes the two, since extraction
// already records an embed the way it records a link.
func Resolve(refs []specmodel.Reference, index specmodel.SpecIndex) {
	targets := newTargets(index)
	for i := range refs {
		refs[i].Resolution = targets.resolve(refs[i].Slug)
	}
	resolveAnchors(refs, index)
}

// targets is what a corpus offers a slug to match against, built once per
// [Resolve] call: the members that declare a slug, and the files beneath
// the spec root that failed to become members.
//
// The two are held apart because they answer differently. A slug that
// matches a member resolves; a slug that matches a non-member reaches a
// document that is right there and whose frontmatter is what is wrong,
// which is a different next step for the reader (AC-SPECREF-002.7).
type targets struct {
	// members maps a lowercased slug to the path of the member declaring
	// it, and nonMembers a lowercased file stem to what the index recorded
	// about that file.
	members    map[string]string
	nonMembers map[string]specmodel.ReferenceNonMember
	// unreadable holds, in path order, the parse failures from which no
	// slug could be recovered: the only non-members a slug also matches
	// as a prefix of their stem (AC-SPEC-003.8).
	unreadable []specmodel.ReferenceNonMember
}

// newTargets indexes one corpus's members and non-members by what a
// reference could match them on.
//
// Where two files answer to the same key the one first in path order
// wins, resolved by comparison rather than by arrival, so the result does
// not depend on the order index happens to hold its collections in.
// Members and non-members are keyed in separate maps because a member
// always wins: a slug that a member declares has resolved, whatever else
// beneath the spec root happens to be named like it.
func newTargets(index specmodel.SpecIndex) targets {
	built := targets{
		members:    make(map[string]string, len(index.Entries)),
		nonMembers: make(map[string]specmodel.ReferenceNonMember, len(index.Entries)),
	}

	for _, entry := range index.Entries {
		// A member whose frontmatter failed to parse but whose slug is
		// still legible declares that slug, and a reference to it resolves
		// as it would to a member that parsed (AC-SPEC-003.7).
		if key := lookupKey(envelope.DeclaredSlug(entry)); key != "" {
			if held, seen := built.members[key]; !seen || entry.Path < held {
				built.members[key] = entry.Path
			}
		}
		if entry.Frontmatter != nil {
			continue
		}
		// An entry with no parsed frontmatter is a file that claimed to be
		// a specification and could not be read. Unless a slug was
		// recovered from it, it declares none — that is what failed — so
		// the only thing a reference can match it on is its name; see
		// [stem]. A recovered member keeps that match too, for a reference
		// naming the file rather than the slug.
		nonMember := specmodel.ReferenceNonMember{
			Path:       entry.Path,
			ParseError: entry.ParseError,
		}
		built.addNonMember(nonMember)
		if entry.RecoveredSlug == nil {
			built.unreadable = append(built.unreadable, nonMember)
		}
	}
	sort.Slice(built.unreadable, func(i, j int) bool { return built.unreadable[i].Path < built.unreadable[j].Path })

	for _, exclusion := range index.Exclusions {
		reason := exclusion.Reason
		built.addNonMember(specmodel.ReferenceNonMember{
			Path:   exclusion.Path,
			Reason: &reason,
		})
	}
	return built
}

// addNonMember records one file that is not a corpus member, keyed by its
// stem, keeping the first in path order where two files share one.
func (t targets) addNonMember(nonMember specmodel.ReferenceNonMember) {
	key := stem(nonMember.Path)
	if key == "" {
		return
	}
	if held, seen := t.nonMembers[key]; !seen || nonMember.Path < held.Path {
		t.nonMembers[key] = nonMember
	}
}

// resolve returns what one reference's slug half reached.
func (t targets) resolve(slug string) specmodel.ReferenceResolution {
	key := lookupKey(slug)
	if key == "" {
		return specmodel.NewReferenceResolutionUnresolved(specmodel.ReferenceUnresolved{})
	}
	if target, ok := t.members[key]; ok {
		return specmodel.NewReferenceResolutionResolved(
			specmodel.ReferenceResolved{TargetPath: target})
	}
	if nonMember, ok := t.nonMembers[key]; ok {
		return specmodel.NewReferenceResolutionNonMember(nonMember)
	}
	// A numbered document is conventionally named <slug>-<title>.md, so a
	// slug whose member could not be read is matched as a prefix of its
	// file name too. Only parse failures are: an excluded file never
	// claimed to be a member, so nothing it is named for depends on it.
	for _, nonMember := range t.unreadable {
		if strings.HasPrefix(stem(nonMember.Path), key+"-") {
			return specmodel.NewReferenceResolutionNonMember(nonMember)
		}
	}
	return specmodel.NewReferenceResolutionUnresolved(specmodel.ReferenceUnresolved{})
}

// lookupKey normalizes a slug for comparison: trimmed of surrounding
// space and lowercased.
//
// Case is folded because slug uniqueness is itself case-insensitive
// (AC-SPEC-003.4), and the two rules have to agree — a corpus that
// refused to hold `adr-012` and `ADR-012` as distinct documents but
// resolved a reference to only one of them would be answering two
// questions about one identifier two ways (AC-SPECREF-002.8).
func lookupKey(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}

// stem returns the lowercased file name of a repository-relative path
// without its extension: `specs/adrs/adr-012.md` becomes `adr-012`.
//
// This is what a reference matches a non-member on, and it is the only
// handle there is. A file the index recorded as a parse failure has no
// readable frontmatter, so it declares no slug that could be compared;
// a file excluded for carrying no frontmatter at all has none either.
// The name is what both of them still have, and it is on the index, so
// the match is reproducible by anything reading the artifact rather than
// resting on knowledge only the build held.
//
// A parse failure no slug could be recovered from is also matched by a
// slug its stem begins with, followed by `-`: `[[adr-012]]` reaches
// `adr-012-reference-conventions.md`, the `<slug>-<title>` naming of
// numbered types (AC-SPEC-003.8). An exclusion is matched on its exact
// stem only.
//
// The consequence, recorded deliberately: a non-member whose intended
// slug matches its file name neither way — `frd-spec-index` for
// `spec-index-frd.md` — matches nothing and reports as an ordinary
// unresolved reference.
func stem(reported string) string {
	name := path.Base(path.Clean(reported))
	return strings.ToLower(strings.TrimSuffix(name, path.Ext(name)))
}
