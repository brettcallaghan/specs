// Tests for slug recovery: a member whose frontmatter fails to parse but
// whose slug is still legible keeps that slug, so one mistyped value
// produces one finding rather than one per citation of the document.
package corpus_test

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/finding"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
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
// one. A citation matching nothing is still unresolved, and so is one the
// file name merely begins with when it is no numbered slug — [[adr]],
// [[adr-012-reference]] — since it does not depend on the member's slug.
//
// AC: AC-SPEC-003.8, AC-SPECREF-002.7
func TestACitationOfAnUnreadableMemberIsAttributedToItsParseFailure(t *testing.T) {
	repo := t.TempDir()
	write(t, repo, "specs/adrs/adr-012-reference-conventions.md",
		"---\ntitle: [unclosed\nslug: adr-012\ntype: adr\n\n# Reference conventions\n\n## Decision\n\nText.\n")
	write(t, repo, "specs/frds/frd-citing.md", strings.TrimSuffix(citingBody("adr-012"), "\n")+
		"\nAnchored: [[adr-012#Decision]]. Missing anchor: [[adr-012#Nowhere]]. Broken: [[adr-404]], [[adr]], [[adr-012-reference]].\n")

	built := buildChecked(t, repo, checkProfile(permissiveCheckSchema))
	findings, err := built.Check()
	if err != nil {
		t.Fatalf("Check: %v", err)
	}

	var parse int
	unresolved := map[string]int{}
	for _, f := range findings {
		switch {
		case f.Rule == finding.RuleFrontmatterParse && f.Path == "specs/adrs/adr-012-reference-conventions.md":
			parse++
		case f.Rule == finding.RuleUnresolvedReference && strings.Contains(f.Message, "[[adr-404]]"):
			unresolved["adr-404"]++
		case f.Rule == finding.RuleUnresolvedReference && strings.Contains(f.Message, "[[adr]]"):
			unresolved["adr"]++
		case f.Rule == finding.RuleUnresolvedReference && strings.Contains(f.Message, "[[adr-012-reference]]"):
			unresolved["adr-012-reference"]++
		case f.Path == "specs/frds/frd-citing.md":
			t.Errorf("unexpected %s against the citing document: %s", f.Rule, f.Message)
		}
	}
	if parse != 1 {
		t.Errorf("%d frontmatter/parse findings on the unreadable member, want 1", parse)
	}
	for _, slug := range []string{"adr-404", "adr", "adr-012-reference"} {
		if unresolved[slug] != 1 {
			t.Errorf("%d reference/unresolved findings for [[%s]], want 1: a citation of no member's slug is still broken", unresolved[slug], slug)
		}
	}
}

// recoveredDuplicateCases pair a parsed member declaring adr-001 with a
// member whose frontmatter fails while it still legibly declares the same
// slug, in each way a block can fail and in a second letter case. Each
// case names the broken file's source and where its slug value sits.
func recoveredDuplicateCases() map[string]struct {
	broken    string
	slug      string
	line, col uint64
} {
	good := document("adr", "adr-001", "Example")
	return map[string]struct {
		broken    string
		slug      string
		line, col uint64
	}{
		"yaml fails": {
			broken: strings.Replace(good, "status: approved", "status: [approved", 1),
			slug:   "adr-001", line: 4, col: 7,
		},
		"typed decode fails": {
			broken: strings.Replace(good, "status: approved", "status: actve", 1),
			slug:   "adr-001", line: 4, col: 7,
		},
		"slugs differ in case": {
			broken: strings.NewReplacer("status: approved", "status: [approved", "slug: adr-001", "slug: ADR-001").
				Replace(good),
			slug: "ADR-001", line: 4, col: 7,
		},
	}
}

// A slug recovered from a member whose frontmatter fails takes part in
// duplicate detection exactly as a parsed one does, cold and warm, and on
// a shared slug the parsed member answers lookups even though the broken
// one sorts first in path order.
//
// AC: AC-SPEC-003.3, AC-SPEC-003.4, AC-SPEC-003.7
func TestARecoveredSlugCollidesWithAParsedMemberDeclaringIt(t *testing.T) {
	const (
		broken = "specs/adrs/a-broken.md"
		good   = "specs/adrs/b-good.md"
	)
	for name, tc := range recoveredDuplicateCases() {
		t.Run(name, func(t *testing.T) {
			repo := t.TempDir()
			write(t, repo, broken, tc.broken)
			write(t, repo, good, document("adr", "adr-001", "Example"))
			write(t, repo, "specs/frds/frd-citing.md", citingBody("adr-001"))

			opts := checkingOptions(repo)
			dest := corpus.IndexPath(opts, "")
			cold, warm, err := corpus.Open(opts, dest)
			if err != nil || warm {
				t.Fatalf("Open cold: warm=%v err=%v", warm, err)
			}
			if err := corpus.Write(cold.Index(), dest, nil); err != nil {
				t.Fatalf("Write: %v", err)
			}
			served, warm, err := corpus.Open(opts, dest)
			if err != nil || !warm {
				t.Fatalf("Open warm: warm=%v err=%v", warm, err)
			}

			for label, c := range map[string]*corpus.Corpus{"cold": cold, "warm": served} {
				duplicates := map[string]specmodel.Finding{}
				for _, f := range c.Findings() {
					if f.Rule == finding.RuleDuplicateSlug {
						duplicates[f.Path] = f
					}
				}
				if len(duplicates) != 2 {
					t.Fatalf("%s: slug/duplicate findings on %v, want one on each of %s and %s",
						label, keysOf(duplicates), broken, good)
				}
				if f := duplicates[broken]; !strings.Contains(f.Message, good) ||
					f.Span.StartLine != tc.line || f.Span.StartColumn != tc.col {
					t.Errorf("%s: finding on the broken member = %q at %d:%d, want naming %s at %d:%d",
						label, f.Message, f.Span.StartLine, f.Span.StartColumn, good, tc.line, tc.col)
				}
				if f := duplicates[good]; !strings.Contains(f.Message, broken) {
					t.Errorf("%s: finding on the parsed member = %q, want naming %s", label, f.Message, broken)
				}

				declared, found := c.Slug("adr-001")
				if !found || declared.Path != good {
					t.Errorf("%s: slug adr-001 answers %q (found %v), want the parsed member %s",
						label, declared.Path, found, good)
				}
				for _, ref := range c.Index().References {
					if ref.Slug != "adr-001" {
						continue
					}
					resolved, ok := ref.Resolution.AsResolved()
					if !ok || resolved.TargetPath != good {
						t.Errorf("%s: [[adr-001]] resolved to %+v, want %s", label, ref.Resolution, good)
					}
				}
			}
			if !reflect.DeepEqual(cold.Findings(), served.Findings()) {
				t.Errorf("findings differ:\n cold %+v\n warm %+v", cold.Findings(), served.Findings())
			}
		})
	}
}

// keysOf returns m's keys, sorted.
func keysOf[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
