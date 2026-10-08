// Tests for slug recovery: a member whose frontmatter fails to parse but
// whose slug is still legible keeps that slug, so one mistyped value
// produces one finding rather than one per citation of the document.
package corpus_test

import (
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
)

// citingBody renders a conforming FRD whose body cites slug.
func citingBody(slug string) string {
	return strings.TrimSuffix(document("frd", "frd-citing", "Citing"), "\n") +
		"\n\nSee [[" + slug + "]] for the decision.\n"
}

// slugRecoveryCases are the two ways a member's frontmatter fails while
// its slug stays legible: the YAML parses and the typed decode rejects a
// value, and the YAML itself fails inside a delimited block.
func slugRecoveryCases() map[string]string {
	return map[string]string{
		"typed decode fails": strings.Replace(
			document("adr", "adr-012", "Reference conventions"), "status: approved", "status: actve", 1),
		"yaml fails inside a delimited block": "---\n" +
			"title: [unclosed\n" +
			"slug: adr-012\n" +
			"type: adr\n" +
			"---\n\n# Reference conventions\n",
	}
}

// AC: AC-SPEC-003.7
func TestALegibleSlugIsRecoveredFromAMemberThatFailsToParse(t *testing.T) {
	for name, broken := range slugRecoveryCases() {
		t.Run(name, func(t *testing.T) {
			repo := t.TempDir()
			write(t, repo, "specs/adrs/adr-012-reference-conventions.md", broken)
			write(t, repo, "specs/frds/frd-citing.md", citingBody("adr-012"))

			built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))

			entry, found := built.Entry("specs/adrs/adr-012-reference-conventions.md")
			if !found {
				t.Fatal("the broken member is absent from the index")
			}
			if entry.ParseError == nil {
				t.Fatal("the broken member carries no parse error")
			}
			if entry.RecoveredSlug == nil || *entry.RecoveredSlug != "adr-012" {
				t.Errorf("RecoveredSlug = %v, want adr-012", entry.RecoveredSlug)
			}

			declared, found := built.Slug("adr-012")
			if !found {
				t.Fatal("slug adr-012 is legible on the broken member and the slug query does not find it")
			}
			if declared.Path != "specs/adrs/adr-012-reference-conventions.md" {
				t.Errorf("slug adr-012 resolved to %q", declared.Path)
			}

			for _, ref := range built.Index().References {
				if ref.Slug != "adr-012" {
					continue
				}
				resolved, ok := ref.Resolution.AsResolved()
				if !ok {
					t.Errorf("[[adr-012]] in %s did not resolve: %+v", ref.Path, ref.Resolution)
					continue
				}
				if resolved.TargetPath != "specs/adrs/adr-012-reference-conventions.md" {
					t.Errorf("[[adr-012]] in %s resolved to %q", ref.Path, resolved.TargetPath)
				}
			}

			findings, err := built.Check()
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			var parse int
			for _, f := range findings {
				switch f.Rule {
				case finding.RuleUnresolvedReference:
					t.Errorf("unexpected %s at %s: %s", f.Rule, f.Path, f.Message)
				case finding.RuleFrontmatterParse:
					parse++
				}
			}
			if parse != 1 {
				t.Errorf("%d frontmatter/parse findings, want exactly 1", parse)
			}
		})
	}
}

// A member whose frontmatter fails with no legible slug recovers nothing:
// an indented slug is not a top-level property, and an unterminated block
// has no extent to read one from.
//
// AC: AC-SPEC-003.7
func TestNoSlugIsRecoveredWhereNoneIsLegible(t *testing.T) {
	cases := map[string]string{
		"indented key": "---\ntitle: [unclosed\n  slug: adr-012\ntype: adr\n---\n\n# Body\n",
		"no slug key":  "---\ntitle: [unclosed\ntype: adr\n---\n\n# Body\n",
		"empty slug": strings.NewReplacer("slug: adr-012", `slug: ""`, "status: approved", "status: actve").
			Replace(document("adr", "adr-012", "T")),
		"unterminated block": "---\ntitle: [unclosed\nslug: adr-012\ntype: adr\n\n# Body\n",
	}
	for name, src := range cases {
		t.Run(name, func(t *testing.T) {
			repo := t.TempDir()
			write(t, repo, "specs/adrs/adr-012.md", src)

			built := corpusAt(t, repo)
			if entry, found := built.Entry("specs/adrs/adr-012.md"); found && entry.RecoveredSlug != nil {
				t.Errorf("RecoveredSlug = %q, want none", *entry.RecoveredSlug)
			}
			if _, found := built.Slug("adr-012"); found {
				t.Error("slug adr-012 is not legible and the slug query found it")
			}
		})
	}
}

// A member whose frontmatter fails with no legible slug is reported once,
// by its parse failure. A citation that reaches it — by slug, which its
// file name begins with, or anchored — raises no finding against the
// citing document, and no anchor finding stands in for the suppressed
// one; a citation matching nothing is still unresolved.
//
// AC: AC-SPEC-003.8, AC-SPECREF-002.7
func TestACitationOfAnUnreadableMemberIsAttributedToItsParseFailure(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-012-reference-conventions.md",
		"---\ntitle: [unclosed\nslug: adr-012\ntype: adr\n\n# Reference conventions\n\n## Decision\n\nText.\n")
	write(t, repo, "specs/frds/frd-citing.md", strings.TrimSuffix(citingBody("adr-012"), "\n")+
		"\nAnchored: [[adr-012#Decision]]. Missing anchor: [[adr-012#Nowhere]]. Broken: [[adr-404]].\n")

	built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	var parse, unresolved int
	for _, f := range findings {
		switch {
		case f.Rule == finding.RuleFrontmatterParse && f.Path == "specs/adrs/adr-012-reference-conventions.md":
			parse++
		case f.Rule == finding.RuleUnresolvedReference && strings.Contains(f.Message, "adr-404"):
			unresolved++
		case f.Path == "specs/frds/frd-citing.md":
			t.Errorf("unexpected %s against the citing document: %s", f.Rule, f.Message)
		}
	}
	if parse != 1 {
		t.Errorf("%d frontmatter/parse findings on the unreadable member, want 1", parse)
	}
	if unresolved != 1 {
		t.Errorf("%d reference/unresolved findings for [[adr-404]], want 1: a citation matching nothing is still broken", unresolved)
	}
}
