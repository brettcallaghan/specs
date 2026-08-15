package cmd

import (
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
// criteria that changed locally, per the CHG delta's --changed flag
// (AC-GHK-008.1).
//
// Querying a criterion by identifier — the plain `spec criterion <id>`
// the eventual Use line will advertise, plus --code, --of, --tag and
// --json — is agent-plugins-wvd.10's addition and is not built here.
// This file carries only what CHG needs, so QRY extends the same command
// rather than replacing it.
func newCriterionCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var changed bool

	cmd := &cobra.Command{
		Use:   "criterion",
		Short: "List acceptance criteria that changed locally",
		Long: "criterion --changed compares the current corpus against the persisted " +
			"index and lists every acceptance criterion whose content hash differs, " +
			"classifying each as changed, added or removed.",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if !changed {
				return errors.New("spec criterion: querying without --changed is not yet supported")
			}
			return runCriterionChanged(streams, *configPath)
		},
	}
	cmd.Flags().BoolVar(&changed, "changed", false,
		"list acceptance criteria whose content differs from the persisted index")
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
		fmt.Fprintf(streams.out, "no persisted index at %s; run `spec index` first\n",
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
