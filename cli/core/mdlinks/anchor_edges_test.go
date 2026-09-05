// Edge cases of the two derivations Resolve performs — the anchor
// fragment and the corpus-relative target — beyond the representative
// cases mdlinks_test.go covers.
//
// Both derivations are pure string rules with branches no representative
// case reaches: a rune that is neither ASCII letter nor space, an anchor
// that survives no character at all, and a target whose file name spells
// the same as a directory component of the citing document's own path.
//
// Every anchor case here carries no anchor resolution, so each exercises
// the authored-anchor fallback rather than the resolved-heading path
// mdlinks_test.go covers.
package mdlinks_test

import (
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/mdlinks"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// The heading-anchor rule keeps every letter and digit and drops
// everything else that is not a space, hyphen or underscore — "letter"
// meaning any Unicode letter, not an ASCII one — and no hyphen it
// produces between two surviving characters is collapsed.
//
// The rule is [markdown.Identifier], shared with the matcher that
// decides whether an anchor resolves at all, so an anchor's fragment and
// the identifier it was looked up under cannot disagree. Two of its
// edges follow from the identifier being derived from RENDERED inline
// content rather than from the characters as typed: whitespace around
// the text is not content and is gone before the rule runs, and markup
// that produced content contributes the content and not its own
// punctuation.
//
// AC: AC-SPECMD-001.7
func TestTheAnchorRuleKeepsUnicodeLettersAndDigitsAndDropsTheRest(t *testing.T) {
	cases := []struct {
		name   string
		anchor string
		want   string
	}{
		{name: "digits are kept", anchor: "Section 12 3", want: "section-12-3"},
		// A dot between digits is dropped like any other punctuation, so a
		// numbered heading closes up rather than gaining a separator: 12.4
		// becomes 124, NOT 12-4. Worth its own case beside the spaced digits
		// above, since "drop it" and "hyphenate it" are both plausible
		// readings of a rule written for spaces.
		{name: "a dot between digits is dropped, not hyphenated", anchor: "Section 12.4 Detail", want: "section-124-detail"},
		{name: "accented latin letters are kept and lowercased", anchor: "Café Naïve Ünïcode", want: "café-naïve-ünïcode"},
		{name: "non-latin letters are kept", anchor: "Раздел Один", want: "раздел-один"},
		{name: "ideographs are kept", anchor: "設計 概要", want: "設計-概要"},
		{name: "an emoji is dropped, its spaces are not", anchor: "Hello 🎉 World", want: "hello--world"},
		{name: "a tab is dropped rather than hyphenated", anchor: "Tab\tSeparated", want: "tabseparated"},
		{name: "leading and trailing spaces are not content and are trimmed", anchor: " Leading And Trailing ", want: "leading-and-trailing"},
		{name: "an anchor of only punctuation survives as nothing", anchor: "!!!", want: ""},
		{name: "an anchor of only spaces survives as nothing", anchor: "   ", want: ""},
		// Markup renders: the backticks around Profile are the syntax
		// that produced the word, not characters of it, so the fragment
		// reads as it would had the heading been written without them.
		{name: "inline markup renders to its content", anchor: "The `Profile` interface", want: "the-profile-interface"},
		// A nested heading path is one anchor, per the anchor feature's
		// out-of-scope list: it names no heading, and the fragment keeps
		// the whole of what was written rather than the last segment.
		{name: "a nested heading path stays one anchor", anchor: "Components#Spans", want: "componentsspans"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			anchor := c.anchor
			ref := resolvedRef(t, "specs/adrs/adr-011.md", "adr-012", "specs/adrs/adr-012.md", &anchor)

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			want := "adr-012.md#" + c.want
			if targets[0] == nil || *targets[0] != want {
				t.Fatalf("target = %v, want %q", targets[0], want)
			}
		})
	}
}

// A derived target climbs and descends by whole directory components,
// and a component only ever matches another directory component: the
// cited document's file name is never mistaken for a directory of the
// citing document's path, however the two might spell alike. Getting
// that wrong emits a target one level short of the file it names.
//
// AC: AC-SPECMD-006.1, AC-SPECMD-006.2
func TestADerivedTargetMatchesDirectoriesAgainstDirectoriesOnly(t *testing.T) {
	cases := []struct {
		name       string
		citingPath string
		targetPath string
		want       string
	}{
		{
			name:       "the cited file name spells the same as a directory of the citing path, extension included",
			citingPath: "specs/adrs/adr-012.md/notes.md",
			targetPath: "specs/adrs/adr-012.md",
			want:       "../adr-012.md",
		},
		{
			name:       "the cited file name spells the same as the citing document's own directory",
			citingPath: "specs/a/b/citer.md",
			targetPath: "specs/a/b.md",
			want:       "../b.md",
		},
		{
			name:       "the cited file name spells the same as a directory beside it",
			citingPath: "specs/a/b/citer.md",
			targetPath: "specs/a/b/a.md",
			want:       "a.md",
		},
		{
			name:       "descending several levels",
			citingPath: "specs/adrs/adr-011.md",
			targetPath: "specs/features/alpha/deep/deeper/deep-doc.md",
			want:       "../features/alpha/deep/deeper/deep-doc.md",
		},
		{
			name:       "climbing several levels to the spec root",
			citingPath: "specs/features/alpha/deep/deeper/deep-doc.md",
			targetPath: "specs/root-level.md",
			want:       "../../../../root-level.md",
		},
		{
			name:       "the cited document is at the repository root",
			citingPath: "specs/adrs/adr-011.md",
			targetPath: "README.md",
			want:       "../../README.md",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ref := resolvedRef(t, c.citingPath, "adr-012", c.targetPath, nil)

			targets, findings := mdlinks.Resolve([]specmodel.Reference{ref}, mdlinks.Options{})

			if len(findings) != 0 {
				t.Fatalf("findings = %v, want none", findings)
			}
			if targets[0] == nil || *targets[0] != c.want {
				t.Fatalf("target = %v, want %q", targets[0], c.want)
			}
		})
	}
}
