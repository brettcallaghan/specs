// Package numbering derives the next identifier for a numbered document
// type from the built [specmodel.SpecIndex], per
// specs/architecture/module-document-core.md's Numbering component.
//
// Numbering is global per type: the next number is one greater than the
// highest number in use for that type anywhere in the corpus, regardless
// of which directory a member with that number sits in. This package
// answers the query by reading the index the Corpus component already
// built rather than walking the filesystem, so the query costs a lookup.
package numbering

import (
	"errors"
	"strconv"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// ErrNotNumbered is returned by [Next] when the active profile does not
// register docType as numbered — either because docType is absent from
// rules altogether, or because it is present with Numbered set to
// false.
//
// A type the profile has not opted in is a different situation from a
// numbered type with no members yet: the former is a query the profile
// does not support, the latter is answered as 1.
var ErrNotNumbered = errors.New("numbering: document type is not registered as numbered")

// Next returns one greater than the highest number in use for docType
// anywhere in index, and errors ErrNotNumbered if rules does not
// register docType as numbered.
//
// The search considers every entry in index whose parsed frontmatter
// declares type docType, regardless of the entry's path — directory
// layout never affects the answer, which is the reason this package
// reads the index rather than the filesystem. An index with no member of
// that type — including an empty corpus — has a highest-in-use of zero,
// so Next returns 1.
//
// The number itself is read from each matching member's slug: the run
// of digits immediately following the "<docType>-" prefix every
// registered slug pattern requires (see cli/profile's documentTypes
// table). A member whose slug carries no such run — which a document
// conforming to its type's slug pattern cannot produce — is skipped
// rather than treated as a parse error, since Next only ever answers
// from members that already validate.
func Next(index specmodel.SpecIndex, docType string, rules map[string]corpus.DocumentTypeRules) (int, error) {
	typeRules, ok := rules[docType]
	if !ok || !typeRules.Numbered {
		return 0, ErrNotNumbered
	}

	prefix := docType + "-"
	highest := 0
	for _, entry := range index.Entries {
		if entry.Frontmatter == nil || string(entry.Frontmatter.Type) != docType {
			continue
		}
		n, ok := leadingNumber(string(entry.Frontmatter.Slug), prefix)
		if !ok {
			continue
		}
		if n > highest {
			highest = n
		}
	}
	return highest + 1, nil
}

// leadingNumber reports the integer value of the run of ASCII digits
// immediately following prefix in slug, and whether such a run exists.
//
// Leading zeros are accepted and do not affect the value returned
// (adr-001 and adr-1 both report 1) — the identifiers this package
// compares differ only in display width, a formatting concern for the
// caller that renders a next identifier back into a slug, not in the
// number they name.
func leadingNumber(slug, prefix string) (int, bool) {
	rest, ok := strings.CutPrefix(slug, prefix)
	if !ok {
		return 0, false
	}

	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}

	n, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0, false
	}
	return n, true
}
