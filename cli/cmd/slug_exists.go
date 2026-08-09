package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newSlugExistsCmd builds the `slug-exists` subcommand: answer whether a
// corpus member declares a given slug, per AC-SPEC-007.2 (AC-SPEC-011.3).
func newSlugExistsCmd(streams *ioStreams, configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slug-exists <slug>",
		Short: "Report whether a corpus member declares the given slug",
		Long: "slug-exists reports whether any corpus member declares the given " +
			"slug and, if one does, that member's path. The comparison ignores " +
			"letter case.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runSlugExists(streams, *configPath, args[0])
		},
	}
	return cmd
}

// runSlugExists answers [corpus.Corpus.Slug] for slug and renders the
// result.
//
// It always exits zero: whether a slug is in use is a fact about the
// corpus, not a failure of the command that asked. The corpus is opened
// warm (corpus.Open), the same as check, because this is exactly the
// corpus query REQ-SPEC-009.1's latency budget is measured against.
func runSlugExists(streams *ioStreams, configPath, slug string) error {
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

	entry, found := c.Slug(slug)
	if !found {
		fmt.Fprintf(streams.out, "%s: not found\n", slug)
		return nil
	}
	fmt.Fprintf(streams.out, "%s: found at %s\n", slug, entry.Path)
	return nil
}
