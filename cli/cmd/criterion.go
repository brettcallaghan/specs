package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/revisions"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// newCriterionCmd builds the `criterion` subcommand: list acceptance
// criteria that changed against a git base (--changed, --since and
// --exit-code, FEAT-GHK-008), and query the corpus by criterion
// identifier, owning feature or qualifier tag (--of, --tag, AC-GHK-007.2,
// AC-GHK-007.3). --json applies to both.
//
// --changed and the query surface are mutually exclusive: --changed takes
// no identifier and answers a comparison against the corpus at a git
// revision, the query surface takes zero or one identifier plus the flags
// below and answers from the corpus directly, and the RunE closure
// branches between them before either flag-conversion function is
// called.
func newCriterionCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var changed bool
	var since string
	var exitCode bool
	var of string
	var tag string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "criterion [<id>]",
		Short: "Query an acceptance criterion, or list ones that changed locally",
		Long: "criterion <id> returns that criterion's record. criterion --of " +
			"<feature-id> returns every criterion beneath one feature, and criterion " +
			"--tag <qualifier> every criterion carrying one qualifier. The answer is a " +
			"YAML mapping whose one key, criteria, holds a list of the matching " +
			"records in every case, one identifier included; --json writes the same " +
			"values as JSON. An identifier no " +
			"criterion declares is reported as such and exits non-zero. " +
			"criterion --changed compares the current corpus, uncommitted edits " +
			"included, against the corpus at the merge-base of HEAD and the " +
			"repository's default branch, or of HEAD and the revision --since names, " +
			"and lists every acceptance criterion whose content hash differs, " +
			"classifying each as changed, added or removed. Its answer is a YAML " +
			"mapping holding base (the revision and the merge-base commit) and then " +
			"criteria; it exits 0 whatever it lists, or 1 on a non-empty list under " +
			"--exit-code. It needs git on PATH.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if changed {
				if len(args) != 0 || of != "" || tag != "" {
					return invocationError("--changed cannot be combined with an identifier, --of or --tag")
				}
				return runCriterionChanged(streams, *configPath, since, exitCode, jsonOut)
			}
			if c.Flags().Changed("since") {
				return invocationError("--since applies only to --changed")
			}
			if exitCode {
				return invocationError("--exit-code applies only to --changed")
			}
			if len(args) == 1 && (of != "" || tag != "") {
				return invocationError("an identifier cannot be combined with --of or --tag")
			}
			if of != "" && tag != "" {
				return invocationError("--of and --tag cannot both be given")
			}
			if len(args) == 0 && of == "" && tag == "" {
				return invocationError("give an identifier, --of, --tag or --changed")
			}
			return runCriterionQuery(streams, *configPath, args, of, tag, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&changed, "changed", false,
		"list acceptance criteria whose content differs from the corpus at a git base")
	cmd.Flags().StringVar(&since, "since", "",
		"with --changed, compare against the merge-base of HEAD and this branch, tag or commit instead of the default branch")
	cmd.Flags().BoolVar(&exitCode, "exit-code", false,
		"with --changed, exit 1 when any criterion is listed")
	cmd.Flags().StringVar(&of, "of", "",
		"list every criterion beneath this feature")
	cmd.Flags().StringVar(&tag, "tag", "",
		"list every criterion carrying this qualifier")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the answer as JSON instead of YAML")
	return cmd
}

// runCriterionChanged compares the current corpus against the corpus at a
// git base and renders the criteria that differ (FEAT-GHK-008).
//
// The current side is corpus.Build over the working tree, so committed,
// staged, unstaged and untracked documents beneath the spec root are all
// part of it (AC-GHK-008.6). The base side is corpus.Build over the spec
// root as the base commit holds it, materialized by the Revisions
// component into a temporary directory and rebased there at the same
// repository-relative paths, so an unmoved document reports one path on
// both sides. Neither side reads, writes or requires specs.json: the
// persisted index is rebuilt by any query, so it cannot stand for an
// earlier state (AC-GHK-008.7).
func runCriterionChanged(streams *ioStreams, configPath, since string, exitCode, jsonOut bool) error {
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

	current, err := corpus.Build(opts)
	if err != nil {
		return fmt.Errorf("building the corpus: %w", err)
	}

	before, base, err := buildBaseIndex(opts, since)
	if err != nil {
		return err
	}

	result := criterionChangesResult{
		Base:     changeBase{Revision: base.Ref, Commit: base.Commit},
		Criteria: changedCriteria(before, current.Index()),
	}
	if exitCode && len(result.Criteria) > 0 {
		streams.exitCode = statusGateFailed
	}
	return renderCriterionChanges(streams, result, jsonOut)
}

// buildBaseIndex builds the corpus at the base the comparison runs
// against: the merge-base of HEAD and since, or of HEAD and the default
// branch when since is empty. The materialized copy is removed before it
// returns, on every path.
func buildBaseIndex(opts corpus.Options, since string) (specmodel.SpecIndex, revisions.Base, error) {
	repo, err := revisions.Open(opts.RepoRoot)
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, revisionError(err)
	}

	top, err := canonicalPath(repo.Root())
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, environmentError("%s cannot be resolved: %v", repo.Root(), err)
	}
	repoRel, err := relativeInside(top, opts.RepoRoot)
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, err
	}
	specRel, err := relativeInside(top, opts.Root)
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, err
	}

	base, err := repo.Base(since)
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, revisionError(err)
	}

	tree, err := repo.Materialize(base.Commit, filepath.ToSlash(specRel))
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, revisionError(err)
	}
	defer func() { _ = tree.Close() }()

	baseOpts := opts
	baseOpts.RepoRoot = filepath.Join(tree.Root(), repoRel)
	baseOpts.Root = filepath.Join(tree.Root(), specRel)
	// A spec root added since the base is absent from the copy; an empty
	// directory makes every criterion in it read as added (AC-GHK-008.3).
	if err := os.MkdirAll(baseOpts.Root, 0o755); err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, environmentError("%s cannot be created: %v", baseOpts.Root, err)
	}

	built, err := corpus.Build(baseOpts)
	if err != nil {
		return specmodel.SpecIndex{}, revisions.Base{}, fmt.Errorf("building the corpus at %s: %w", base.Commit, err)
	}
	return built.Index(), base, nil
}

// canonicalPath is path made absolute with every symbolic link resolved,
// the form git prints its top level in. Comparing two paths in any other
// form misreads a temporary directory under a symlink (macOS's /var) as
// lying outside the work tree.
func canonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// relativeInside is path relative to the work tree's top level, or an
// environment error naming both when path lies outside it
// (AC-GHK-008.13).
func relativeInside(top, path string) (string, error) {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	resolved, err := canonicalPath(path)
	if err != nil {
		return "", environmentError("%s cannot be resolved: %v", path, err)
	}
	rel, err := filepath.Rel(top, resolved)
	if err != nil || !filepath.IsLocal(rel) {
		return "", environmentError("%s lies outside the git work tree %s; a comparison reads the spec root from that repository's history, so move the spec root inside it",
			path, top)
	}
	return rel, nil
}

// revisionError classifies an error from the Revisions component under
// the exit status its class carries.
func revisionError(err error) error {
	var rerr *revisions.Error
	if !errors.As(err, &rerr) {
		return err
	}
	switch rerr.Class {
	case revisions.Invocation:
		return invocationError("%s", rerr.Message)
	case revisions.Environment:
		return environmentError("%s", rerr.Message)
	}
	return internalError("%s", rerr.Message)
}

// runCriterionQuery converts args and the --of, --tag and --json flags
// into one corpus lookup and renders the result — the whole of what this
// command's query surface does (AC-GHK-007.2, AC-GHK-007.3, AC-GHK-007.5).
//
// The corpus is opened warm (corpus.Open), the same as check, slug-exists,
// next-id and refs, so a query answers from the persisted index without
// re-walking the corpus when its inputs are unchanged (AC-GHK-007.6).
//
// An identifier argument selects one criterion by
// [corpus.Corpus.Criterion]; naming no record is reported distinctly and
// exits non-zero (AC-GHK-007.4). Otherwise --of selects every criterion
// [corpus.Corpus.CriteriaOf] one feature, or --tag every criterion
// [corpus.Corpus.CriteriaByTag] one qualifier — both are selections, and
// an empty selection exits zero rather than being confused with an
// unknown identifier.
func runCriterionQuery(streams *ioStreams, configPath string, args []string, of, tag string, jsonOut bool) error {
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
		id := args[0]
		criterion, found := c.Criterion(specmodel.CriterionId(id))
		if !found {
			fmt.Fprintf(streams.err, "%s: not found\n", id)
			streams.exitCode = statusGateFailed
			return nil
		}
		return renderCriteria(streams, []specmodel.Criterion{criterion}, jsonOut)
	}

	var criteria []specmodel.Criterion
	if of != "" {
		criteria = c.CriteriaOf(specmodel.FeatureId(of))
	} else {
		criteria = c.CriteriaByTag(tag)
	}
	return renderCriteria(streams, criteria, jsonOut)
}

// criteriaResult is the criterion query's answer: the matching records
// and nothing else, under the one key `criteria`, whatever the arguments
// selected — one identifier, --of or --tag — so a caller never branches
// on what it passed ([[cli-output-standards]]'s "One shape per command").
// Criteria is never nil, so an empty selection renders as `criteria: []`.
type criteriaResult struct {
	Criteria []specmodel.Criterion `json:"criteria"`
}

// renderCriteria writes the answer as YAML, or as JSON under --json, both
// carrying the same values in the same order (AC-GHK-007.2, AC-GHK-007.5).
// An empty selection is an answer, never the non-zero "not found" an
// unknown identifier produces (AC-GHK-007.4).
func renderCriteria(streams *ioStreams, criteria []specmodel.Criterion, asJSON bool) error {
	if criteria == nil {
		criteria = []specmodel.Criterion{}
	}
	result := criteriaResult{Criteria: criteria}
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("rendering the criteria as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}

// criterionChangesResult is the change listing's answer: the base it was
// compared against, then the criteria that differ (AC-GHK-008.11).
// Criteria is never nil, so an empty listing renders as `criteria: []`.
type criterionChangesResult struct {
	Base     changeBase        `json:"base"`
	Criteria []criterionChange `json:"criteria"`
}

// changeBase names the base: the revision as --since named it, or the
// default branch found, and the merge-base commit's full object name.
type changeBase struct {
	Revision string `json:"revision"`
	Commit   string `json:"commit"`
}

// criterionChange is one criterion's classification against the base:
// "changed", "added" or "removed", per AC-GHK-008.3, and the path of the
// document holding it — in the current corpus, or in the base for a
// removed one. A criterion present in both with an unchanged hash —
// whatever its position in either corpus — has no criterionChange at all
// (AC-GHK-008.2).
type criterionChange struct {
	Id             specmodel.CriterionId `json:"id"`
	Classification string                `json:"classification"`
	Path           string                `json:"path"`
}

const (
	classificationChanged = "changed"
	classificationAdded   = "added"
	classificationRemoved = "removed"
)

// changedCriteria compares base against current by criterion identifier
// and content hash, classifying each difference by whether the identifier
// is present in both, only in current, or only in base (AC-GHK-008.1,
// AC-GHK-008.3). Position is never read: two criteria with the same
// identifier and the same content hash are the same criterion, wherever
// each corpus places it (AC-GHK-008.2).
//
// The result is sorted by identifier, so the listing is byte-identical
// across runs and independent of both corpora's own entry order.
func changedCriteria(base, current specmodel.SpecIndex) []criterionChange {
	before := criterionHashes(base)
	after := criterionHashes(current)

	changes := []criterionChange{}
	for id, now := range after {
		then, existed := before[id]
		switch {
		case !existed:
			changes = append(changes, criterionChange{Id: id, Classification: classificationAdded, Path: now.path})
		case then.hash != now.hash:
			changes = append(changes, criterionChange{Id: id, Classification: classificationChanged, Path: now.path})
		}
	}
	for id, then := range before {
		if _, stillPresent := after[id]; !stillPresent {
			changes = append(changes, criterionChange{Id: id, Classification: classificationRemoved, Path: then.path})
		}
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Id < changes[j].Id })
	return changes
}

// criterionState is what the comparison reads of one criterion: its
// content hash, and the path of the document its feature is declared in.
type criterionState struct {
	hash specmodel.ContentHash
	path string
}

// criterionHashes flattens an index's nested features into a lookup from
// criterion identifier to its state — the whole of what
// [changedCriteria] reads from either side of the comparison.
func criterionHashes(index specmodel.SpecIndex) map[specmodel.CriterionId]criterionState {
	states := make(map[specmodel.CriterionId]criterionState)
	for _, feature := range index.Features {
		for _, criterion := range feature.Criteria {
			states[criterion.Id] = criterionState{hash: criterion.ContentHash, path: feature.Path}
		}
	}
	return states
}

// renderCriterionChanges writes the answer as YAML, or as JSON under
// --json, both carrying the same values in the same order.
func renderCriterionChanges(streams *ioStreams, result criterionChangesResult, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(result)
		if err != nil {
			return internalError("rendering the change listing as JSON: %v", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprint(streams.out, renderYAML(result))
	return nil
}
