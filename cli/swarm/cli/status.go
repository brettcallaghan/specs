// The `swarm status` command: turns the Status Collector's measurement
// set into the report a caller reads and the exit status it implies. Per
// the Status Reporter component in specs/architecture/container-swarm-cli.md
// and FEAT-SWARM-CORE-003 of specs/features/swarm-core/swarm-core-frd.md.
//
// It does not go through [Spec] and [newSubcommand] the way the manifest
// commands do. [status.Collect] is not one static plan a caller can
// explain before running it — its completion walk queries one level of
// the run's tree at a time, and the next query depends on what the last
// one answered, so there is no fixed sequence to print ahead of running
// it (see [tracker.Descendants]'s own doc comment). This file declares
// the same five contract flags by hand instead, and reuses [finish],
// [refusal] and the [Refuse*] constructors so the exit-status mapping and
// the two output forms are still the surface's, not a second copy of
// them. The five contract flags it declares by hand are the five
// [newSubcommand] declares, --full-result included.
//
// Explain and dry-run still mean what the contract says: explain prints
// what this command would measure and measures nothing; dry-run measures
// live state, which for a command that performs no write at all is
// exactly what execute does. There is no write to refuse, so the two
// differ only in the mode name the report carries.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/status"
)

// defaultStallAfter is the threshold in force when --stall-after is not
// given. It is the value the derived health check this command replaces
// used for the same comparison — nothing standing, work ready, and this
// long since the last commit.
const defaultStallAfter = 30 * time.Minute

// statusCmd builds `swarm status`.
func statusCmd(streams *ioStreams, deps status.Deps) *cobra.Command {
	var (
		explain    bool
		dryRun     bool
		asJSON     bool
		fullResult bool
		runID      string
		stallAfter time.Duration
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Report a run's status from state",
		Long: "Reports the standing worktree count, the sprint branch's last commit " +
			"and its age, the ready set's size, each dispatch's drain state and the " +
			"run's completion — measured from git and the tracker on every call, never " +
			"read from a value the run stored about itself. It never asserts that a run " +
			"has stalled: it names what it measured and the threshold it compared " +
			"against, and leaves the reading to whoever can see the harness.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			req := Request{
				Command:    cmd.CommandPath(),
				Run:        runID,
				JSON:       asJSON,
				FullResult: fullResult || explain || dryRun,
			}

			if strings.TrimSpace(runID) == "" {
				return finish(streams, req, refusal(req, RefuseArguments(
					"%s is scoped to one run; pass --run <run-epic-id>", req.Command)))
			}

			switch {
			case explain && dryRun:
				return finish(streams, req, refusal(req, RefuseArguments(
					"--explain and --dry-run select different interpreters; pass one")))
			case explain:
				req.Mode = plan.ModeExplain
			case dryRun:
				req.Mode = plan.ModeDryRun
			default:
				req.Mode = plan.ModeExecute
			}

			if req.Mode == plan.ModeExplain {
				return finish(streams, req, explainStatus(req, stallAfter))
			}

			m, records, err := status.Collect(context.Background(), deps, runID)
			if err != nil {
				return finish(streams, req, refusal(req, RefuseUnrepairable("%s", err)))
			}
			return finish(streams, req, reportStatus(req, m, records, stallAfter))
		},
	}

	cmd.Flags().BoolVar(&explain, "explain", false,
		"describe what this command would measure, and measure nothing")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false,
		"measure live state; status performs no write, so this matches execute")
	cmd.Flags().BoolVar(&asJSON, "json", false,
		"machine-readable output carrying the same facts as the default form")
	cmd.Flags().StringVar(&runID, "run", "",
		"the run epic this command is scoped to")
	cmd.Flags().BoolVar(&fullResult, "full-result", false,
		"print the whole command result — the mode, the outcome, the facts and the invocations behind them — instead of the answer alone")
	cmd.Flags().DurationVar(&stallAfter, "stall-after", defaultStallAfter,
		"the last-commit age, with the ready set non-empty and no dispatch worktree "+
			"standing, that raises the attention status")

	return cmd
}

// explainStatus describes the measurements this command would take
// without taking any. The sequence is not a [plan.Plan]: the completion
// walk's depth is not known until each level of the run's tree answers,
// so what is printed is the description [tracker.Descendants] itself
// gives for a command that walks, rather than a plan explain cannot
// render truthfully.
func explainStatus(req Request, stallAfter time.Duration) Result {
	lines := []string{
		"swarm status measures a run; it does not store or replay one, so its steps " +
			"are described here rather than printed as a fixed plan:",
		"  bd show <run> --json                                (the run manifest, for its branch and admission record)",
		"  git worktree prune                                  (clear stale registrations)",
		"  git worktree list --porcelain                       (the standing worktrees)",
		"  git log --format=... -n 1 <branch>                  (<branch> is the manifest read's answer)",
		"  bd ready --parent <run> --exclude-type epic --json  (the ready set)",
		"  bd list --parent <run> --all --limit 0 --json       (one query per level of the run's tree; " +
			"the depth is not known until each level answers)",
		fmt.Sprintf("the attention threshold in force is %s (--stall-after)", stallAfter),
	}
	return req.stamp(Result{
		Outcome:    OutcomeClear,
		Summary:    "explained; nothing was measured",
		Transcript: lines,
	})
}

// reportStatus turns one measurement set into the report a caller reads
// and the outcome it implies. It takes no dependency but the measurements
// themselves, so its wording and its verdict are testable against a
// fixed value with no subprocess anywhere in the test.
func reportStatus(req Request, m status.Measurements, records []plan.Record, stallAfter time.Duration) Result {
	manifestInv := argvsFor(records, "manifest")
	worktreeInv := argvsFor(records, "worktrees-prune", "worktrees-list")
	commitInv := argvsFor(records, "last-commit")
	readyInv := argvsFor(records, "ready")
	completionInv := childrenArgvs(records)

	// A fact this reporter states about the run as a whole — the
	// threshold, the observability disclaimer, an unestablished
	// measurement's own line — names the invocation that grounds it when
	// there is one, and falls back to the manifest read otherwise. The
	// manifest read is the one invocation [status.Collect] always
	// attempts, so it is always available to cite.
	fallback := func(primary [][]string) [][]string {
		if len(primary) > 0 {
			return primary
		}
		return manifestInv
	}

	result := req.stamp(Result{Outcome: OutcomeClear})
	var unestablished []string
	note := func(name, reason string) { unestablished = append(unestablished, name+": "+reason) }

	// Stated on every report, whatever else it finds: the threshold this
	// run's progress is compared against, and what the harness cannot
	// tell this command in place of that comparison.
	observed := append(append(append([][]string{}, worktreeInv...), commitInv...), readyInv...)
	result.Facts = append(result.Facts,
		Text("stall_after", stallAfter.String(), fallback(commitInv)...),
		Text("observability",
			"whether a subagent is still working cannot be observed from outside the harness; "+
				"this reports the standing worktree count, the last commit's age and the ready "+
				"set's size in its place",
			fallback(observed)...),
	)

	if m.Worktrees.Established {
		result.Facts = append(result.Facts,
			Count("worktrees_standing", m.Worktrees.Standing, worktreeInv...),
			Count("dispatch_worktrees_standing", dispatchWorktreeCount(m.Dispatches), worktreeInv...),
		)
	} else {
		result.Facts = append(result.Facts, Unestablished("worktrees_standing", m.Worktrees.Reason, fallback(nil)...))
		note("worktrees_standing", m.Worktrees.Reason)
	}

	if m.LastCommit.Established {
		result.Facts = append(result.Facts,
			Text("last_commit_sha", m.LastCommit.SHA, commitInv...),
			Text("last_commit_subject", m.LastCommit.Subject, commitInv...),
			Text("last_commit_at", m.LastCommit.When.UTC().Format(time.RFC3339), commitInv...),
			Count("last_commit_age_seconds", int(m.LastCommit.Age.Round(time.Second).Seconds()), commitInv...),
		)
	} else {
		result.Facts = append(result.Facts, Unestablished("last_commit", m.LastCommit.Reason, fallback(nil)...))
		note("last_commit", m.LastCommit.Reason)
	}

	if m.Ready.Established {
		result.Facts = append(result.Facts, Count("ready_count", m.Ready.Count, readyInv...))
	} else {
		result.Facts = append(result.Facts, Unestablished("ready_count", m.Ready.Reason, fallback(nil)...))
		note("ready_count", m.Ready.Reason)
	}

	if m.Completion.Established {
		result.Facts = append(result.Facts,
			Count("completion_total", m.Completion.Total, completionInv...),
			Count("completion_done", m.Completion.Done, completionInv...),
			Count("completion_blocked", m.Completion.Blocked, completionInv...),
			Count("completion_claimed", m.Completion.Claimed, completionInv...),
			Count("completion_open", m.Completion.Open, completionInv...),
			Count("completion_deferred", m.Completion.Deferred, completionInv...),
			Count("completion_unknown", m.Completion.Unknown, completionInv...),
		)
	} else {
		result.Facts = append(result.Facts, Unestablished("completion_total", m.Completion.Reason, fallback(nil)...))
		note("completion_total", m.Completion.Reason)
	}

	if m.Scope.Established {
		// Reported unconditionally, so a run whose admitted work is all
		// closed while admission excluded others reads the excluded list
		// beside the completion verdict above rather than behind a
		// separate condition this reporter would have to detect —
		// AC-SWARM-CORE-003.8.
		result.Facts = append(result.Facts,
			List("scope_additions", m.Scope.Additions, fallback(nil)...),
			List("scope_missing", m.Scope.Missing, fallback(nil)...),
			List("scope_excluded", m.Scope.Excluded, fallback(nil)...),
		)
	} else {
		result.Facts = append(result.Facts, Unestablished("scope_additions", m.Scope.Reason, fallback(nil)...))
		note("scope", m.Scope.Reason)
	}

	dispatchInv := fallback(append(append([][]string{}, readyInv...), completionInv...))
	result.Facts = append(result.Facts,
		Structured("dispatches", dispatchDocument(m.Dispatches), dispatchInv...))

	conflictInv := fallback(append(append([][]string{}, worktreeInv...), completionInv...))
	result.Facts = append(result.Facts,
		Structured("conflicts", conflictDocument(m.Conflicts), conflictInv...))

	switch {
	case len(unestablished) > 0:
		// AC-SWARM-CORE-003.5, .003.10: a measurement that could not be
		// taken is reported as unknown above and exits non-zero here,
		// whatever else this run's live state would otherwise report.
		result.Outcome = OutcomeAttention
		result.Summary = fmt.Sprintf("%d measurement(s) could not be taken: %s",
			len(unestablished), strings.Join(unestablished, "; "))
		result.Resolution = "investigate why the named measurement failed, then re-run `swarm status`"

	case m.Ready.Count > 0 && dispatchWorktreeCount(m.Dispatches) == 0 && m.LastCommit.Age > stallAfter:
		// AC-SWARM-CORE-003.1: measurements only, never a diagnosis —
		// AC-SWARM-CORE-003.11 is why this says "no commit for Xs" and
		// not "stalled".
		result.Outcome = OutcomeAttention
		result.Summary = fmt.Sprintf(
			"no dispatch worktree stands; %d bead(s) ready; no commit for %s, over the %s threshold",
			m.Ready.Count, m.LastCommit.Age.Round(time.Second), stallAfter)
		result.Resolution = "dispatch a new batch, or check why none is running"

	default:
		result.Summary = fmt.Sprintf("%d/%d bead(s) done; %d ready; last commit %s ago",
			m.Completion.Done, m.Completion.Total, m.Ready.Count, m.LastCommit.Age.Round(time.Second))
	}

	return result
}

// dispatchWorktreeCount is how many standing worktrees the collector
// attributed to a dispatch, as opposed to the main checkout or one it
// could not attribute at all. A real repository always has at least one
// worktree — the main checkout — so the raw [status.Worktrees.Standing]
// count the collector reports is never the right measure of whether an
// agent is running; the collector's own attribution, carried on each
// [status.Dispatch.Worktree], is.
func dispatchWorktreeCount(dispatches []status.Dispatch) int {
	n := 0
	for _, d := range dispatches {
		if d.Worktree != "" {
			n++
		}
	}
	return n
}

// dispatchDocument renders every dispatch as a JSON document the surface
// emits as real structure.
//
// One field per thing a reader asks about, rather than one formatted line
// carrying several. A line reading "dev-1: in progress: run.2 (worktree:
// /tmp/x)" states the agent, its drain state, the beads it holds and
// where its worktree stands — four facts a reader has to parse back out,
// and a machine has to split on a punctuation convention nothing
// enforces.
func dispatchDocument(dispatches []status.Dispatch) string {
	type entry struct {
		Agent       string   `json:"agent"`
		Established bool     `json:"established"`
		Reason      string   `json:"reason,omitempty"`
		Drained     *bool    `json:"drained,omitempty"`
		Ready       []string `json:"ready,omitempty"`
		InProgress  []string `json:"in_progress,omitempty"`
		Blocked     []string `json:"blocked,omitempty"`
		Worktree    string   `json:"worktree,omitempty"`
	}

	entries := []entry{}
	for _, d := range dispatches {
		e := entry{Agent: d.Agent, Established: d.Established, Reason: d.Reason, Worktree: d.Worktree}
		if d.Established {
			drained := d.Drained
			e.Drained = &drained
			e.Ready, e.InProgress, e.Blocked = d.Ready, d.InProgress, d.Blocked
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Agent < entries[j].Agent })
	return marshal(entries)
}

// conflictDocument renders every bead two agents hold at once.
func conflictDocument(conflicts []status.Conflict) string {
	type entry struct {
		Bead     string `json:"bead"`
		Agent    string `json:"agent"`
		Worktree string `json:"worktree"`
	}

	entries := []entry{}
	for _, c := range conflicts {
		entries = append(entries, entry{Bead: c.Bead, Agent: c.Agent, Worktree: c.Worktree})
	}
	return marshal(entries)
}

// marshal encodes a document for a structured fact. A value this package
// built cannot fail to encode, and a fact carrying invalid JSON is
// refused by [Result.Validate] rather than rendered, so the error has
// nowhere useful to go and an empty list is the safe residue.
func marshal(v any) string {
	encoded, err := json.Marshal(v)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

// argvsFor returns the resolved argv of every record named in names, in
// the order names lists them. A step that never ran contributes nothing,
// which is what happens when the measurement that step belongs to was
// never taken.
func argvsFor(records []plan.Record, names ...string) [][]string {
	var out [][]string
	for _, want := range names {
		for _, r := range records {
			if r.Step == want && len(r.Invocation.Name) > 0 {
				out = append(out, append([]string{r.Invocation.Name}, r.Invocation.Args...))
			}
		}
	}
	return out
}

// childrenArgvs returns the argv of every completion-walk query, however
// many levels the run's tree turned out to have. The walk's step names
// are "children-<depth>-<index>", generated only once each level is
// reached, which is why they cannot be named in advance the way
// [argvsFor]'s other callers are.
func childrenArgvs(records []plan.Record) [][]string {
	var out [][]string
	for _, r := range records {
		if strings.HasPrefix(r.Step, "children-") && len(r.Invocation.Name) > 0 {
			out = append(out, append([]string{r.Invocation.Name}, r.Invocation.Args...))
		}
	}
	return out
}
