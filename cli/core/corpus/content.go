package corpus

import (
	"strings"

	"github.com/brettcallaghan/specs/cli/core/finding"
	"github.com/brettcallaghan/specs/cli/core/gherkin"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// gherkinContentFindings reports every VCON content and diagnostic
// defect over features and failures — [Corpus.recomputeGherkin]'s two
// results, shared with VTAG's identity checks by the caller (check.go)
// so one Check or CheckFor call parses the corpus's Gherkin once. rules
// is read only for the criterion pattern gherkin/scenario-prose quotes,
// so an agent fixing the finding knows the form a real identifier takes.
func gherkinContentFindings(rules FeatureTagRules, features []*gherkin.Feature, failures []SourceFailure) []specmodel.Finding {
	var findings []specmodel.Finding
	findings = append(findings, gherkinParseFindings(failures)...)
	findings = append(findings, criterionEmptyFindings(features)...)
	findings = append(findings, scenarioProseFindings(rules, features)...)
	findings = append(findings, partiallyTaggedOutlineFindings(features)...)
	return findings
}

// gherkinParseFindings reports gherkin/parse: a Gherkin source that
// could not be parsed, at the line and column of the offending text in
// the containing document, with the parser's own expectation in the
// message (AC-GHK-006.1).
//
// This is the one VCON rule that cannot be found from features alone —
// a source that fails to parse produces no feature at all, so it is
// invisible to every walk but the full recompute that already produced
// failures.
func gherkinParseFindings(failures []SourceFailure) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, failure := range failures {
		for _, err := range failure.Errors {
			findings = append(findings, finding.GherkinParse(failure.Path, err.Span, err.Message))
		}
	}
	return findings
}

// criterionEmptyFindings reports gherkin/criterion-empty: a criterion
// with no steps (AC-GHK-006.2).
func criterionEmptyFindings(features []*gherkin.Feature) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, feature := range features {
		for _, criterion := range feature.Criteria {
			if criterion.Record.Text != "" {
				continue
			}
			findings = append(findings, finding.CriterionEmpty(feature.Record.Path, criterion.Record.Span))
		}
	}
	return findings
}

// scenarioProseFindings reports gherkin/scenario-prose: a non-empty
// Scenario description, as a probable untagged criterion.
//
// Only [gherkin.Criterion.Record].Description is read here — a
// Feature's own description is never visited by this function, which is
// what makes AC-GHK-006.4's exemption a property of this check's shape
// rather than a condition it tests and could get wrong: a Feature
// description carries the user story, and nothing about a Feature
// record resembles a criterion's steps for a normative statement to be
// mistaken for.
func scenarioProseFindings(rules FeatureTagRules, features []*gherkin.Feature) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, feature := range features {
		for _, criterion := range feature.Criteria {
			if criterion.Record.Description == nil {
				continue
			}
			if strings.TrimSpace(*criterion.Record.Description) == "" {
				continue
			}
			findings = append(findings, finding.ScenarioProse(
				feature.Record.Path, criterion.Record.Span, rules.CriterionPattern))
		}
	}
	return findings
}

// partiallyTaggedOutlineFindings reports gherkin/examples-partially-tagged:
// a Scenario Outline whose Examples blocks carry an identifier on some
// and not others, naming the untagged blocks (AC-GHK-002.7, AC-GHK-006.2).
//
// One finding per outline, from [gherkin.Feature.PartialOutlines] — the
// Gherkin component's own record of the mixture, populated by `criteria`
// alongside the criterion records themselves so this check and VTAG's
// untagged-criterion exclusion (identity.go's [untaggedSpans]) read one
// shared fact rather than two independently derived guesses about it.
func partiallyTaggedOutlineFindings(features []*gherkin.Feature) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, feature := range features {
		for _, outline := range feature.PartialOutlines {
			findings = append(findings, finding.ExamplesPartiallyTagged(feature.Record.Path, outline.Untagged))
		}
	}
	return findings
}
