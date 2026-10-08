package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
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

	cmd := &cobra.Command{
		Use:   "feature [<id>]",
		Short: "Query a feature, or emit one as a standalone Gherkin file",
		Long: "feature <id> returns that feature's record, including its criteria. " +
			"feature --code <CODE> returns every feature grouped under that code, and " +
			"feature with no argument every feature in the corpus. --json writes the " +
			"matching records as JSON to standard output. An identifier no feature " +
			"declares is reported as such and exits non-zero. " +
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
				return runFeatureEmit(streams, *configPath, args[0], out)
			}
			if code != "" && len(args) == 1 {
				return invocationError("an identifier and --code cannot both be given")
			}
			return runFeatureQuery(streams, *configPath, args, code, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&emit, "emit", false,
		"write the feature as a standalone .feature file instead of querying it")
	cmd.Flags().StringVar(&out, "out", "",
		"where to write the emitted file: a path, or omitted for standard output")
	cmd.Flags().StringVar(&code, "code", "",
		"list every feature grouped under this code")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the matching feature records as JSON to standard output")
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
	cfg, findings, err := Load(configPath, BuiltinRegistry())
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
			return environmentError("%s cannot be written: %s", filepath.ToSlash(out), writeErrorText(err))
		}
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return environmentError("%s cannot be written: %s", filepath.ToSlash(out), writeErrorText(err))
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
// An identifier argument selects one feature by [corpus.Corpus.Feature];
// naming no record is reported distinctly and exits non-zero
// (AC-GHK-007.4). Otherwise --code selects every feature
// [corpus.Corpus.FeaturesByCode] grouped under it, or, given neither, every
// feature in the corpus — both are selections, and an empty selection
// exits zero rather than being confused with an unknown identifier.
func runFeatureQuery(streams *ioStreams, configPath string, args []string, code string, jsonOut bool) error {
	cfg, findings, err := Load(configPath, BuiltinRegistry())
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
		id := args[0]
		feature, found := c.Feature(specmodel.FeatureId(id))
		if !found {
			fmt.Fprintf(streams.err, "%s: not found\n", id)
			streams.exitCode = statusGateFailed
			return nil
		}
		return renderFeature(streams, feature, jsonOut)
	}

	var features []specmodel.Feature
	if code != "" {
		features = c.FeaturesByCode(specmodel.FeatureCode(code))
	} else {
		features = c.Features()
	}
	return renderFeatures(streams, features, jsonOut)
}

// renderFeature writes one feature, including its criteria, to
// streams.out: the generated [specmodel.Feature] JSON in --json mode, and
// a short human summary otherwise — the feature's identifier and name, one
// tab-indented `<id>\t<name>` line per criterion beneath it, matching the
// tab-separated shape `criterion --changed` already writes.
func renderFeature(streams *ioStreams, feature specmodel.Feature, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(feature)
		if err != nil {
			return fmt.Errorf("rendering the feature as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}

	fmt.Fprintf(streams.out, "%s\t%s\n", feature.Id, feature.Name)
	for _, criterion := range feature.Criteria {
		fmt.Fprintf(streams.out, "\t%s\t%s\n", criterion.Id, criterionName(criterion))
	}
	return nil
}

// renderFeatures writes a feature selection to streams.out: the generated
// [specmodel.Feature] JSON array in --json mode (AC-GHK-007.5), and one
// `<id>\t<name>` line per feature otherwise. An empty selection writes a
// fixed sentence in text mode and `[]` in JSON mode — it is a valid answer,
// never the non-zero "not found" an unknown identifier produces
// (AC-GHK-007.4).
func renderFeatures(streams *ioStreams, features []specmodel.Feature, asJSON bool) error {
	if features == nil {
		features = []specmodel.Feature{}
	}
	if asJSON {
		data, err := json.Marshal(features)
		if err != nil {
			return fmt.Errorf("rendering the feature selection as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}

	if len(features) == 0 {
		fmt.Fprintln(streams.out, "no matching features")
		return nil
	}
	for _, feature := range features {
		fmt.Fprintf(streams.out, "%s\t%s\n", feature.Id, feature.Name)
	}
	return nil
}

// criterionName returns c's Name, or the empty string where the notation
// gave it none — the same nil-to-empty reading [toRenderCriterion] applies
// for `--emit`.
func criterionName(c specmodel.Criterion) string {
	if c.Name == nil {
		return ""
	}
	return *c.Name
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
