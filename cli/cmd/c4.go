package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// newC4Cmd builds the `c4` subcommand: emit the corpus's C4 model — the
// element tree, the relationships and the statements the model does not
// admit — as YAML, or as JSON under --json, per AC-C4-004.1 through
// AC-C4-004.7.
//
// Wired like [newListCmd]: flags in, one Document Core call, one rendering
// out. It decides nothing about what the model holds; Document Core
// assembles it, so that a wrapper other than cobra gets the same answer.
func newC4Cmd(streams *ioStreams, configPath *string) *cobra.Command {
	var root string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "c4",
		Short: "Emit the corpus's C4 model: elements, relationships and excluded statements",
		Long: "c4 answers with the corpus's C4 model — the software systems, containers, " +
			"components and modules the corpus documents, nested by the containment " +
			"edges it declares, the dependency relationships its blueprint diagrams " +
			"state, and every relationship statement the model does not admit, with " +
			"the reason. The model carries its schema version and the workspace name " +
			"from the configuration, and its collections are ordered so that an " +
			"unchanged corpus emits the same bytes. It exits zero whatever the model " +
			"holds: judging the diagrams is check's job. --root narrows the model to " +
			"one element and its descendants, and the model then carries root_declared; " +
			"a slug naming no element is answered with root_declared false and every " +
			"collection empty, and exits zero.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runC4(streams, *configPath, root, jsonOut)
		},
	}
	cmd.Flags().StringVar(&root, "root", "",
		"narrow the model to the element with this slug and its descendants")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the model as JSON instead of YAML")
	return cmd
}

// runC4 opens the corpus warm and asks Document Core for the model.
//
// Warm, through the shared [openCorpus], for the reason `check` and `list`
// are: the model is a projection of what the index already records. An
// unknown --root is an answer Document Core gives, a model holding nothing
// with root_declared false (AC-C4-004.6), so it is rendered like any other.
func runC4(streams *ioStreams, configPath, root string, jsonOut bool) error {
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

	model, err := c.C4Model(cfg.Workspace, root)
	if err != nil {
		return err
	}

	if jsonOut {
		data, err := json.Marshal(model)
		if err != nil {
			return fmt.Errorf("rendering the C4 model as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(model))
	return nil
}
