package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// newCriterionCmd builds the `criterion` subcommand: list acceptance
// criteria that changed locally (CHG's --changed, AC-GHK-008.1), and query
// the corpus by criterion identifier, owning feature or qualifier tag
// (QRY's --of, --tag and --json, AC-GHK-007.2, AC-GHK-007.3).
//
// --changed and the query surface are mutually exclusive: --changed takes
// no identifier and answers a comparison against the persisted index,
// the query surface takes zero or one identifier plus the flags below and
// answers from the corpus directly, and the RunE closure branches between
// them before either flag-conversion function is called.
func newCriterionCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var changed bool
	var of string
	var tag string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "criterion [<id>]",
		Short: "Query an acceptance criterion, or list ones that changed locally",
		Long: "criterion <id> returns that criterion's record. criterion --of " +
			"<feature-id> returns every criterion beneath one feature, and criterion " +
			"--tag <qualifier> every criterion carrying one qualifier. --json writes " +
			"the matching records as JSON to standard output. An identifier no " +
			"criterion declares is reported as such and exits non-zero. " +
			"criterion --changed compares the current corpus against the persisted " +
			"index and lists every acceptance criterion whose content hash differs, " +
			"classifying each as changed, added or removed.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if changed {
				if len(args) != 0 || of != "" || tag != "" || jsonOut {
					return errors.New("specs criterion: --changed cannot be combined with an identifier, --of, --tag or --json")
				}
				return runCriterionChanged(streams, *configPath)
			}
			if len(args) == 1 && (of != "" || tag != "") {
				return errors.New("specs criterion: an identifier cannot be combined with --of or --tag")
			}
			if of != "" && tag != "" {
				return errors.New("specs criterion: --of and --tag cannot both be given")
			}
			if len(args) == 0 && of == "" && tag == "" {
				return errors.New("specs criterion: give an identifier, --of, --tag or --changed")
			}
			return runCriterionQuery(streams, *configPath, args, of, tag, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&changed, "changed", false,
		"list acceptance criteria whose content differs from the persisted index")
	cmd.Flags().StringVar(&of, "of", "",
		"list every criterion beneath this feature")
	cmd.Flags().StringVar(&tag, "tag", "",
		"list every criterion carrying this qualifier")
	cmd.Flags().BoolVar(&jsonOut, "json", false,
		"write the matching criterion records as JSON to standard output")
	return cmd
}

// runCriterionChanged compares the current corpus against the persisted
// index and renders the criteria that differ (AC-GHK-008.1).
//
// The persisted side is read with corpus.Load rather than corpus.Open.
// Open's warm-cache decision is one-sided: when the persisted index's
// input hash no longer matches, it silently rebuilds and hands back the
// *current* corpus, discarding the very earlier state this comparison
// needs (cache.go's openWarm). That is exactly the case --changed exists
// to report, so this command reads the artifact directly instead of
// through the cache. The current side always builds fresh
// (corpus.Build), independent of the cache being warm — a --changed run
// must see whatever the corpus is right now, not a servable answer.
//
// This command writes nothing back: unlike openCorpus's best-effort
// persist, doing so here would let a --changed run silently replace the
// baseline it just compared against with the state it is reporting as
// different from that baseline.
func runCriterionChanged(streams *ioStreams, configPath string) error {
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

	indexPath := corpus.IndexPath(opts, "")
	persisted, ok := loadPersistedIndex(indexPath)
	if !ok {
		fmt.Fprintf(streams.out, "no persisted index at %s; run `specs index` first\n",
			filepath.ToSlash(indexPath))
		return nil
	}

	built, err := corpus.Build(opts)
	if err != nil {
		return fmt.Errorf("cmd: building the corpus: %w", err)
	}

	changes := changedCriteria(persisted, built.Index())
	renderCriterionChanges(streams, changes)
	return nil
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
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	if err != nil {
		return err
	}
	if findings != nil {
		return reportAndExit(streams, findings, jsonOut)
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
			streams.exitCode = 1
			return nil
		}
		return renderCriterion(streams, criterion, jsonOut)
	}

	var criteria []specmodel.Criterion
	if of != "" {
		criteria = c.CriteriaOf(specmodel.FeatureId(of))
	} else {
		criteria = c.CriteriaByTag(tag)
	}
	return renderCriteria(streams, criteria, jsonOut)
}

// renderCriterion writes one criterion to streams.out: the generated
// [specmodel.Criterion] JSON in --json mode, and a single
// `<id>\t<name>` line otherwise — the same shape [renderCriteria] writes
// per entry, so a single lookup and a selection of one read alike.
func renderCriterion(streams *ioStreams, criterion specmodel.Criterion, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(criterion)
		if err != nil {
			return fmt.Errorf("cmd: rendering the criterion as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}
	fmt.Fprintf(streams.out, "%s\t%s\n", criterion.Id, criterionName(criterion))
	return nil
}

// renderCriteria writes a criterion selection to streams.out: the
// generated [specmodel.Criterion] JSON array in --json mode
// (AC-GHK-007.5), and one `<id>\t<name>` line per criterion otherwise. An
// empty selection writes a fixed sentence in text mode and `[]` in JSON
// mode — it is a valid answer, never the non-zero "not found" an unknown
// identifier produces (AC-GHK-007.4).
func renderCriteria(streams *ioStreams, criteria []specmodel.Criterion, asJSON bool) error {
	if criteria == nil {
		criteria = []specmodel.Criterion{}
	}
	if asJSON {
		data, err := json.Marshal(criteria)
		if err != nil {
			return fmt.Errorf("cmd: rendering the criterion selection as JSON: %w", err)
		}
		fmt.Fprintf(streams.out, "%s\n", data)
		return nil
	}

	if len(criteria) == 0 {
		fmt.Fprintln(streams.out, "no matching criteria")
		return nil
	}
	for _, criterion := range criteria {
		fmt.Fprintf(streams.out, "%s\t%s\n", criterion.Id, criterionName(criterion))
	}
	return nil
}

// loadPersistedIndex reads the index at path and reports whether it may
// be used as the earlier state AC-GHK-008.1 compares against.
//
// Every way it can fail to be usable — the file is absent, unreadable,
// not JSON, or written by an older index shape — is folded into "no
// persisted index" rather than surfaced as a distinct error, the same
// generosity corpus.Open's warm path applies when deciding whether to
// trust the cache. Unlike that path, a hash mismatch here is not a
// reason to treat the index as unusable — a differing input hash is
// what this command exists to explain, not a condition it rebuilds
// around.
func loadPersistedIndex(path string) (specmodel.SpecIndex, bool) {
	index, err := corpus.Load(path)
	if err != nil {
		return specmodel.SpecIndex{}, false
	}
	if index.SchemaVersion != specindex.SchemaVersion {
		return specmodel.SpecIndex{}, false
	}
	return index, true
}

// criterionChange is one criterion's classification against the earlier
// index: "changed", "added" or "removed", per AC-GHK-008.3. A criterion
// present in both with an unchanged hash — whatever its position in
// either corpus — has no criterionChange at all (AC-GHK-008.2).
type criterionChange struct {
	Id             specmodel.CriterionId
	Classification string
}

const (
	classificationChanged = "changed"
	classificationAdded   = "added"
	classificationRemoved = "removed"
)

// changedCriteria compares persisted against current by criterion
// identifier and content hash, classifying each difference by whether
// the identifier is present in both, only in current, or only in
// persisted (AC-GHK-008.1, AC-GHK-008.3). Position is never read: two
// criteria with the same identifier and the same content hash are the
// same criterion, wherever each corpus places it (AC-GHK-008.2).
//
// The result is sorted by identifier, so the listing is byte-identical
// across runs and independent of both corpora's own entry order.
func changedCriteria(persisted, current specmodel.SpecIndex) []criterionChange {
	before := criterionHashes(persisted)
	after := criterionHashes(current)

	var changes []criterionChange
	for id, afterHash := range after {
		beforeHash, existed := before[id]
		switch {
		case !existed:
			changes = append(changes, criterionChange{Id: id, Classification: classificationAdded})
		case beforeHash != afterHash:
			changes = append(changes, criterionChange{Id: id, Classification: classificationChanged})
		}
	}
	for id := range before {
		if _, stillPresent := after[id]; !stillPresent {
			changes = append(changes, criterionChange{Id: id, Classification: classificationRemoved})
		}
	}

	sort.Slice(changes, func(i, j int) bool { return changes[i].Id < changes[j].Id })
	return changes
}

// criterionHashes flattens an index's nested features into a lookup from
// criterion identifier to content hash — the whole of what
// [changedCriteria] reads from either side of the comparison.
func criterionHashes(index specmodel.SpecIndex) map[specmodel.CriterionId]specmodel.ContentHash {
	hashes := make(map[specmodel.CriterionId]specmodel.ContentHash)
	for _, feature := range index.Features {
		for _, criterion := range feature.Criteria {
			hashes[criterion.Id] = criterion.ContentHash
		}
	}
	return hashes
}

// renderCriterionChanges writes one line per change to streams.out —
// classification, then identifier, tab-separated — or a fixed sentence
// when the comparison found none. There is no --json rendering here:
// that belongs to QRY's --json, which spans this command's whole query
// surface rather than one flag's output.
func renderCriterionChanges(streams *ioStreams, changes []criterionChange) {
	if len(changes) == 0 {
		fmt.Fprintln(streams.out, "no local changes")
		return
	}
	for _, c := range changes {
		fmt.Fprintf(streams.out, "%s\t%s\n", c.Classification, c.Id)
	}
}
