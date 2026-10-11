package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/gherkin"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// newFeatureCmd builds the `feature` subcommand: emit a feature as a
// standalone Gherkin file (EMIT's --emit and --out, AC-GHK-009.1), and
// query the corpus by feature identifier or code (QRY's --code and
// --json, AC-GHK-007.1, AC-GHK-007.3).
//
// --emit and the query surface are mutually exclusive: --emit always
// takes exactly one identifier, the query surface takes zero or one plus
// the flags below, and the RunE closure branches between them before
// either flag-conversion function is called.
func newFeatureCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var emit bool
	var out string
	var code string
	var jsonOut bool
	var full bool

	cmd := &cobra.Command{
		Use:   "feature [<id>]",
		Short: "Query a feature, or emit one as a standalone Gherkin file",
		Long: "feature <id> returns that feature's compact summary: its id, code, name, path and " +
			"criteria_count. The answer is compact by default and --full returns every " +
			"field, the description and criteria included; --json changes the syntax, " +
			"never the view. " +
			"feature --code <CODE> returns every feature grouped under that code, and " +
			"feature with no argument every feature in the corpus. The answer is a " +
			"YAML mapping whose one key, features, holds a list of the matching " +
			"features in every case, one identifier included; --json writes the same " +
			"values as JSON. Given an identifier or --code, the answer also carries " +
			"declared, false with an empty list where the corpus declares no such " +
			"feature or code; that is an answer, and exits 0. " +
			"feature <id> --emit writes the named feature as a valid Gherkin .feature " +
			"file, carrying its tags, description and criteria, to standard output " +
			"by default and to the path --out names.",
		Args: func(c *cobra.Command, args []string) error {
			if emit {
				return cobra.ExactArgs(1)(c, args)
			}
			return cobra.MaximumNArgs(1)(c, args)
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if emit {
				if code != "" || jsonOut {
					return invocationError("--code and --json cannot be combined with --emit")
				}
				if full {
					return invocationError("--full cannot be combined with --emit")
				}
				return runFeatureEmit(streams, *configPath, args[0], out)
			}
			if code != "" && len(args) == 1 {
				return invocationError("an identifier and --code cannot both be given")
			}
			return runFeatureQuery(streams, *configPath, args, code, full, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&emit, "emit", false,
		"write the feature as a standalone .feature file instead of querying it")
	cmd.Flags().StringVar(&out, "out", "",
		"where to write the emitted file: a path, or omitted for standard output")
	cmd.Flags().StringVar(&code, "code", "",
		"list every feature grouped under this code")
	cmd.Flags().BoolVar(&full, "full", false,
		"return every field of every feature, its description and criteria included")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the answer as JSON instead of YAML")
	return cmd
}

// runFeatureEmit renders the feature named id as a standalone `.feature`
// file and writes it to out, or to streams.out where out is empty
// (AC-GHK-009.1).
//
// The corpus is opened warm (corpus.Open), the same as check,
// slug-exists, next-id and refs, because looking a feature up by
// identifier is exactly the kind of corpus query those subcommands'
// shared latency budget is measured against.
func runFeatureEmit(streams *ioStreams, configPath, id, out string) error {
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

	// An identifier no feature declares is an operand this producer
	// cannot act on (exit 2): unlike the query, --emit has no answer that
	// could carry "not declared", so writing nothing and saying why is the
	// whole of what it can do.
	feature, found := c.Feature(specmodel.FeatureId(id))
	if !found {
		return invocationError("%s: not found", id)
	}

	rules := opts.Profile.FeatureTags()
	data := gherkin.Render(toRenderFeature(feature, rules))

	if out == "" {
		_, err := streams.out.Write(data)
		if err != nil {
			return fmt.Errorf("writing the emitted feature to standard output: %w", err)
		}
		return nil
	}

	if dir := filepath.Dir(out); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return environmentError("%s cannot be written: %s", filepath.ToSlash(out), writeErrorText(out, err))
		}
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return environmentError("%s cannot be written: %s", filepath.ToSlash(out), writeErrorText(out, err))
	}
	// A file was written: name it, the same convention `index` follows
	// for the same reason — a human running this interactively sees
	// where the artifact landed, and forward slashes keep the one line
	// this command writes byte-identical across platforms.
	fmt.Fprintln(streams.out, filepath.ToSlash(out))
	return nil
}

// runFeatureQuery converts args and the --code and --json flags into one
// corpus lookup and renders the result — the whole of what this command's
// query surface does (AC-GHK-007.1, AC-GHK-007.3, AC-GHK-007.5).
//
// The corpus is opened warm (corpus.Open), the same as check, slug-exists,
// next-id and refs, so a query answers from the persisted index without
// re-walking the corpus when its inputs are unchanged (AC-GHK-007.6).
//
// An identifier argument selects one feature by [corpus.Corpus.Feature],
// --code every feature [corpus.Corpus.FeaturesByCode] groups under it,
// and neither every feature in the corpus. An identifier or code the
// corpus does not declare is answered with the empty selection and
// declared false, exiting zero (AC-GHK-007.4), so "nothing matches" and
// "nothing by that name exists" read apart.
func runFeatureQuery(streams *ioStreams, configPath string, args []string, code string, full, jsonOut bool) error {
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

	if len(args) == 1 {
		feature, found := c.Feature(specmodel.FeatureId(args[0]))
		if !found {
			return renderFeatures(streams, nil, declared(false), full, jsonOut)
		}
		return renderFeatures(streams, []specmodel.Feature{feature}, declared(true), full, jsonOut)
	}

	if code != "" {
		features := c.FeaturesByCode(specmodel.FeatureCode(code))
		return renderFeatures(streams, features, declared(len(features) > 0), full, jsonOut)
	}
	return renderFeatures(streams, c.Features(), nil, full, jsonOut)
}

// featuresResult is the feature query's answer: the matching records
// under the one key `features`, whatever the arguments selected — one
// identifier, a code, or every feature — so a caller never branches on
// what it passed ([[cli-output-standards]]'s "One shape per command").
// Features is never nil, so an empty selection renders as `features: []`.
//
// Declared states whether the subject the invocation named — an
// identifier, or a code, which is declared exactly when some feature
// carries it — is one the corpus declares (AC-GHK-007.4). It is present
// exactly when a subject is named, true or false, and absent for the
// listing of every feature, which names none; it precedes the collection
// it qualifies, as refs' declared does.
type featuresResult struct {
	Declared *bool               `json:"declared,omitempty"`
	Features []specmodel.Feature `json:"features"`
}

// featuresCompactResult is the default answer: the properties the query
// established, exactly as [featuresResult] carries them, over the
// features in the compact view of [[cli-output-standards]]
// (AC-GHK-007.10).
type featuresCompactResult struct {
	Declared *bool                   `json:"declared,omitempty"`
	Features []featureCompactSummary `json:"features"`
}

// featureCompactSummary keeps what lets a caller choose a feature to
// fetch in full: which it is and where, and how many criteria it holds.
// CriteriaCount is derived, and is named apart from the full record's
// `criteria` so that one key never holds two types.
type featureCompactSummary struct {
	Id            specmodel.FeatureId   `json:"id"`
	Code          specmodel.FeatureCode `json:"code"`
	Name          string                `json:"name"`
	Path          string                `json:"path"`
	CriteriaCount int                   `json:"criteria_count"`
}

// compactFeatures derives the compact view from the full answer, so the
// two list the same features in the same order by construction
// (AC-GHK-007.10).
func compactFeatures(result featuresResult) featuresCompactResult {
	compact := featuresCompactResult{
		Declared: result.Declared,
		Features: make([]featureCompactSummary, 0, len(result.Features)),
	}
	for _, feature := range result.Features {
		compact.Features = append(compact.Features, featureCompactSummary{
			Id:            feature.Id,
			Code:          feature.Code,
			Name:          feature.Name,
			Path:          feature.Path,
			CriteriaCount: len(feature.Criteria),
		})
	}
	return compact
}

// declared is the not-declared property's value for a query that names a
// subject: a pointer, so that a query naming none can leave it absent.
func declared(found bool) *bool { return &found }

// renderFeatures writes the answer as YAML, or as JSON under --json, both
// carrying the same values in the same order (AC-GHK-007.1, AC-GHK-007.5).
// A subject the corpus does not declare is an answer too — the empty
// selection with declared false, exiting zero (AC-GHK-007.4).
func renderFeatures(streams *ioStreams, features []specmodel.Feature, isDeclared *bool, full, asJSON bool) error {
	if features == nil {
		features = []specmodel.Feature{}
	}
	var result any = featuresResult{Declared: isDeclared, Features: features}
	if !full {
		result = compactFeatures(featuresResult{Declared: isDeclared, Features: features})
	}
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("rendering the features as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

// toRenderFeature converts an index record into what [gherkin.Render]
// needs, deriving every tag Render writes from rules — the caller's own
// vocabulary, per [gherkin.RenderFeature]'s promise that the renderer
// holds none of its own.
//
// The owning-document tag is written whenever rules registers one
// (OwnerPrefix set) and the feature carries a document slug, regardless
// of the feature's own source form: an emitted file is always a
// standalone feature source, and [[adr-021]] constraint 11 requires that
// form to name its owner even where the feature was read from a fenced
// block that did not.
func toRenderFeature(f specmodel.Feature, rules corpus.FeatureTagRules) gherkin.RenderFeature {
	var description string
	if f.Description != nil {
		description = *f.Description
	}

	criteria := make([]gherkin.RenderCriterion, 0, len(f.Criteria))
	for _, criterion := range f.Criteria {
		criteria = append(criteria, toRenderCriterion(criterion, rules))
	}

	return gherkin.RenderFeature{
		Tags:        featureTags(f, rules),
		Name:        f.Name,
		Description: description,
		Criteria:    criteria,
	}
}

// featureTags returns the tags Render writes above `Feature:`: the
// owning-document tag where rules registers one, then the feature's own
// identifier.
func featureTags(f specmodel.Feature, rules corpus.FeatureTagRules) []string {
	var tags []string
	if rules.OwnerPrefix != "" && f.DocumentSlug != "" {
		tags = append(tags, string(f.DocumentSlug))
	}
	if f.Id != "" {
		tags = append(tags, string(f.Id))
	}
	return tags
}

// toRenderCriterion converts one nested criterion record into what
// [gherkin.Render] needs.
func toRenderCriterion(c specmodel.Criterion, rules corpus.FeatureTagRules) gherkin.RenderCriterion {
	var name, description, rule string
	if c.Name != nil {
		name = *c.Name
	}
	if c.Description != nil {
		description = *c.Description
	}
	if c.Rule != nil {
		rule = *c.Rule
	}

	return gherkin.RenderCriterion{
		Tags:        criterionTags(c, rules),
		Name:        name,
		Description: description,
		Rule:        rule,
		Text:        c.Text,
	}
}

// criterionTags returns the tags Render writes on a criterion's
// `Scenario` or `Scenario Outline` line: its identifier, followed by
// whichever of rules' registered qualifier tags the criterion's
// properties call for, in rules' own registration order.
func criterionTags(c specmodel.Criterion, rules corpus.FeatureTagRules) []string {
	var tags []string
	if c.Id != "" {
		tags = append(tags, string(c.Id))
	}
	for _, qualifier := range rules.Qualifiers {
		switch qualifier.Meaning {
		case specmodel.QualifierMeaningNegation:
			if c.Negative {
				tags = append(tags, qualifier.Tag)
			}
		case specmodel.QualifierMeaningManualVerification:
			if c.Manual {
				tags = append(tags, qualifier.Tag)
			}
		case specmodel.QualifierMeaningVerificationLevel:
			if c.Level != nil && *c.Level == qualifier.Level {
				tags = append(tags, qualifier.Tag)
			}
		}
	}
	return tags
}
