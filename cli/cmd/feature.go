package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/gherkin"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// newFeatureCmd builds the `feature` subcommand: emit a feature as a
// standalone Gherkin file, per the EMIT delta's --emit and --out flags
// (AC-GHK-009.1).
//
// Querying a feature by identifier — the plain `spec feature <id>` the
// Use line advertises, plus --code, --json and the rest of QRY's surface
// — is agent-plugins-wvd.10's addition and is not built here. This file
// carries only what EMIT needs, so QRY extends the same command rather
// than replacing it.
func newFeatureCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var emit bool
	var out string

	cmd := &cobra.Command{
		Use:   "feature <id>",
		Short: "Emit a feature as a standalone Gherkin file",
		Long: "feature --emit writes the named feature as a valid Gherkin .feature " +
			"file, carrying its tags, description and criteria, to standard output " +
			"by default and to the path --out names.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if !emit {
				return errors.New("spec feature: querying without --emit is not yet supported")
			}
			return runFeatureEmit(streams, *configPath, args[0], out)
		},
	}
	cmd.Flags().BoolVar(&emit, "emit", false,
		"write the feature as a standalone .feature file instead of querying it")
	cmd.Flags().StringVar(&out, "out", "",
		"where to write the emitted file: a path, or omitted for standard output")
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
		return reportAndExit(streams, findings, false)
	}

	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		return err
	}

	c, err := openCorpus(opts)
	if err != nil {
		return err
	}

	feature, found := findFeature(c, specmodel.FeatureId(id))
	if !found {
		fmt.Fprintf(streams.err, "%s: not found\n", id)
		streams.exitCode = 1
		return nil
	}

	rules := opts.Profile.FeatureTags()
	data := gherkin.Render(toRenderFeature(feature, rules))

	if out == "" {
		_, err := streams.out.Write(data)
		if err != nil {
			return fmt.Errorf("cmd: writing the emitted feature to standard output: %w", err)
		}
		return nil
	}

	if dir := filepath.Dir(out); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("cmd: creating %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(out, data, 0o644); err != nil {
		return fmt.Errorf("cmd: writing %s: %w", out, err)
	}
	// A file was written: name it, the same convention `index` follows
	// for the same reason — a human running this interactively sees
	// where the artifact landed, and forward slashes keep the one line
	// this command writes byte-identical across platforms.
	fmt.Fprintln(streams.out, filepath.ToSlash(out))
	return nil
}

// findFeature returns the feature id names from c's index.
//
// A linear scan over Index().Features rather than a lookup structure:
// this is the whole of what EMIT needs from a corpus query, and building
// an indexed lookup is QRY's job once agent-plugins-wvd.10 lands.
func findFeature(c *corpus.Corpus, id specmodel.FeatureId) (specmodel.Feature, bool) {
	for _, feature := range c.Index().Features {
		if feature.Id == id {
			return feature, true
		}
	}
	return specmodel.Feature{}, false
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
