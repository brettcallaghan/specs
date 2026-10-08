package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// checkRules is every rule identifier a corpus check run can raise today.
//
// This is not new domain vocabulary: each value is one of the constants
// cli/core/finding already exports (RuleSchemaViolation and its
// siblings). Naming them here is what lets [runCheck] declare every
// known check up front — including the ones a run's findings happen not
// to mention — which is what makes "not requested" distinguishable from
// "requested and passed" (AC-SPEC-006.8's report distinction, carried by
// [finding.Report] itself; see buildCheckReport below).
//
// RuleUnresolvedAnchor, RuleAmbiguousAnchor, RuleRetiredTarget,
// RuleIncompletePairing and RuleTypeMismatch are
// declared here rather than left to [finding.Report.Add]'s own
// auto-registration of an unnamed rule: without a declaration, a run
// that raises none of them would leave them out of the report's checks
// accounting entirely instead of showing "passed", and --only could not
// name one to request it on a run that never happens to raise it. This
// is what lets --only name each of the reference-integrity checks and
// what lets --strict reach the reference/retired-target warning even on
// a corpus that never triggers it.
// The ten gherkin/* identity rules (VTAG) and four gherkin/* content
// rules (VCON) are declared for the same reason: without it, a corpus
// with no Gherkin defect at all would leave them out of the accounting
// entirely, rather than showing "passed".
// The ten c4/* rules are declared for the same reason.
var checkRules = []specmodel.RuleId{
	finding.RuleFrontmatterParse,
	finding.RuleSchemaViolation,
	finding.RuleMissingProperty,
	finding.RuleSlugPattern,
	finding.RuleDuplicateSlug,
	finding.RuleUnresolvedReference,
	finding.RuleUnresolvedAnchor,
	finding.RuleAmbiguousAnchor,
	finding.RuleRetiredTarget,
	finding.RuleIncompletePairing,
	finding.RuleTypeMismatch,
	finding.RuleCorpusEmpty,
	finding.RuleCriterionUntagged,
	finding.RuleFeatureUntagged,
	finding.RuleIdentifierMalformed,
	finding.RuleIdentifierMultiple,
	finding.RuleIdentifierDuplicate,
	finding.RuleCriterionCodeMismatch,
	finding.RuleOwnerMissing,
	finding.RuleOwnerMismatch,
	finding.RuleOwnerUnresolved,
	finding.RuleTagUnregistered,
	finding.RuleGherkinParse,
	finding.RuleCriterionEmpty,
	finding.RuleScenarioProse,
	finding.RuleExamplesPartiallyTagged,
	finding.RuleC4Parse,
	finding.RuleC4UndeclaredNode,
	finding.RuleC4AmbiguousNode,
	finding.RuleC4Bidirectional,
	finding.RuleC4DuplicateEdge,
	finding.RuleC4DrawingConflict,
	finding.RuleC4InteractionConflict,
	finding.RuleC4ContractUnresolved,
	finding.RuleC4ContractMismatch,
	finding.RuleC4TagSeparator,
}

// newCheckCmd builds the `check` subcommand: run conformance checks over
// the corpus, or over one document, per FEAT-SPEC-006 (AC-SPEC-011.2).
func newCheckCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var jsonOut, strict bool
	var only []string

	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Check the corpus, or a single document, for conformance",
		Long: "check validates every corpus member's frontmatter against the active " +
			"profile's schema and slug patterns, and reports duplicate slugs and " +
			"unresolved references, either across the whole corpus or, given a " +
			"path, for one document.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var path string
			if len(args) == 1 {
				path = args[0]
			}
			return runCheck(streams, *configPath, path, jsonOut, strict, only)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the answer as JSON instead of YAML")
	cmd.Flags().BoolVar(&strict, "strict", false,
		"exit non-zero when the run produces only warning or unverified findings")
	cmd.Flags().StringSliceVar(&only, "only", nil,
		"run only the named checks (by rule identifier); repeatable or comma-separated")
	return cmd
}

// runCheck runs a check over the corpus or over one document, renders the
// result, and sets the process exit status.
//
// path empty means the whole corpus (corpus.Corpus.Check); path set
// narrows to that one document (corpus.Corpus.CheckFor) — both proving
// AC-SPEC-011.2. It opens the corpus warm (corpus.Open) rather than
// building fresh, because a check is exactly the query FEAT-SPEC-009's
// latency budgets are measured against, and a persisted index at the
// default location answers it without re-walking the corpus.
func runCheck(streams *ioStreams, configPath, path string, jsonOut, strict bool, only []string) error {
	// An --only value naming no known check is a malformed question, not
	// a request the corpus happens to satisfy: answered with a report, a
	// misspelt rule would read as "every requested check passed" and exit
	// zero, silently disabling whatever gate passed it (AC-SPEC-006.10).
	// It is refused before the corpus is opened, so no report is written.
	if err := validateOnly(only); err != nil {
		return invocationError("%w", err)
	}

	cfg, configPath, findings, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	if findings != nil {
		return reportConfigDefects(streams, findings)
	}

	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		return err
	}

	c, err := openCorpus(opts)
	if err != nil {
		return err
	}

	var raw []specmodel.Finding
	if path == "" {
		raw, err = c.Check()
	} else {
		raw, err = c.CheckFor(path)
	}
	if err != nil {
		// A path naming no corpus member is the caller's mistake, not a
		// defect in the corpus, so it is an error rather than a finding —
		// and it must not render as a report, because a report with no
		// findings is what "this document passed" looks like
		// (`agent-plugins-ltd`). corpus already distinguishes a file the
		// index excluded, with its reason, from one discovery never saw.
		//
		// It is an invocation error (status 2) and never a pass: a gate
		// that cannot inspect its subject has decided nothing
		// (AC-SPEC-011.6).
		if errors.Is(err, corpus.ErrNotAMember) {
			return invocationError("%w", withoutPrefix(err, "corpus: "))
		}
		return fmt.Errorf("checking: %w", err)
	}

	report := buildCheckReport(raw, only)
	if err := renderReport(streams, report, jsonOut); err != nil {
		return err
	}
	streams.exitCode = exitCodeFor(report, strict)
	return nil
}

// buildCheckReport folds a check run's findings into a [finding.Report]
// that accounts for every rule in checkRules, honoring --only.
//
// This is the "report must distinguish a check that PASSED from one that
// was NOT REQUESTED" requirement from this bead's Design section, and it
// is answered entirely by [finding.Report]'s own accounting rather than
// by anything invented here (AC-SPEC-006.8): a rule named by only (or
// every rule, when only is empty) is marked requested via Report.Request
// and then Report.Ran once its matching findings — if any — have been
// added; a rule left out of only is simply never requested, so it stays
// [specmodel.CheckOutcomeNotRequested] in the rendered report.
//
// Every corpus check already runs regardless of only (corpus.Corpus.Check
// has no notion of a subset); what --only narrows is which of those
// results this report holds anyone accountable for, which is what
// "run only the named checks" means from the report a caller reads.
func buildCheckReport(findings []specmodel.Finding, only []string) *finding.Report {
	requested := requestedRules(only)

	report := finding.NewReport(checkRules...)
	for _, rule := range checkRules {
		if requested == nil || requested[rule] {
			report.Request(rule)
		}
	}
	for _, f := range findings {
		if requested == nil || requested[f.Rule] {
			report.Add(f)
		}
	}
	for _, rule := range checkRules {
		if requested == nil || requested[rule] {
			report.Ran(rule)
		}
	}
	return report
}

// requestedRules converts --only's raw values into a lookup set, or nil
// when --only was not given — nil is the sentinel [buildCheckReport]
// reads as "every check was requested", rather than an empty set meaning
// "no check was requested".
func requestedRules(only []string) map[specmodel.RuleId]bool {
	if len(only) == 0 {
		return nil
	}
	set := make(map[specmodel.RuleId]bool, len(only))
	for _, name := range only {
		set[specmodel.RuleId(name)] = true
	}
	return set
}

// ErrUnknownCheck reports an --only value naming no rule identifier in
// checkRules.
var ErrUnknownCheck = errors.New("is not a known check")

// validateOnly refuses any --only value that names no rule in checkRules,
// naming every unknown value and every known rule so the caller can
// correct a typo without a second lookup — the shape list's
// unregistered --type error already takes (AC-SPECLIST-002.4).
func validateOnly(only []string) error {
	known := make(map[specmodel.RuleId]bool, len(checkRules))
	for _, rule := range checkRules {
		known[rule] = true
	}
	var unknown []string
	for _, name := range only {
		if !known[specmodel.RuleId(name)] {
			unknown = append(unknown, fmt.Sprintf("%q", name))
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	names := make([]string, 0, len(checkRules))
	for _, rule := range checkRules {
		names = append(names, string(rule))
	}
	sort.Strings(names)
	return fmt.Errorf("%s %w for --only; known: %s",
		strings.Join(unknown, ", "), ErrUnknownCheck, strings.Join(names, ", "))
}

// exitCodeFor derives the process exit status from a rendered report, per
// AC-SPEC-006.6 and AC-SPEC-006.7: at least one error exits non-zero
// unconditionally; a run with only warning or unverified findings exits
// zero unless strict is set.
func exitCodeFor(report *finding.Report, strict bool) int {
	if report.HasErrors() {
		return statusGateFailed
	}
	if strict && hasWarningOrUnverified(report) {
		return statusGateFailed
	}
	return statusOK
}

// hasWarningOrUnverified reports whether report carries a finding of
// severity warning or unverified. [finding.Report.HasErrors] answers the
// error half of this question; there is no equivalent for the other two
// severities because nothing before this bead needed one.
func hasWarningOrUnverified(report *finding.Report) bool {
	for _, f := range report.Model().Findings {
		if f.Severity == finding.SeverityWarning || f.Severity == finding.SeverityUnverified {
			return true
		}
	}
	return false
}
