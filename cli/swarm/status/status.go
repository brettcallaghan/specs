// Package status is the Swarm CLI's Status Collector: it takes a run's
// measurements from the tracker and version control and returns them as
// a value — per the Status Collector component in
// specs/architecture/container-swarm-cli.md and FEAT-SWARM-CORE-003 of
// specs/features/swarm-core/swarm-core-frd.md.
//
// It formats nothing and decides nothing. Rendering the measurements and
// choosing the exit status they imply belong to the Status Reporter, and
// the split is what lets the report's wording be tested against a fixed
// measurement set with no subprocess anywhere in the test.
//
// # It reads no stored progress value
//
// A run was once read healthy through a five-hour outage in which nothing
// was dispatched, because health was a label written by the component
// that had stopped. Nothing here reads such a value: every number below
// comes from `git` or from the tracker, and a run epic carrying a
// contradicting health key in its metadata changes no answer this package
// gives.
//
// # A measurement it could not take is unestablished
//
// Every measurement carries whether it was established and, when it was
// not, why. That is the distinction the container exists to keep: "the
// query said zero" and "the query could not answer" are different facts,
// and a collector that returned a zero for the second would be the silent
// failure this whole CLI is built against.
package status

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/vcs"
)

// A worktree is matched to the dispatch holding it through the bead its
// HEAD commit closed and that bead's assignee — see
// [measureWorktreeOwners] — never through the tree's branch name: the
// mechanism that creates a dispatch's worktree does not control that name
// and does not carry the dispatch id in it. Measured against this
// harness's own worktrees, whose branches read `worktree-agent-<hash>`,
// not `bb/<dispatch-id>`.

// KindEpic is the tracker's own type for an epic. An epic is not work, so
// it is excluded from the counts and from the ready set.
const KindEpic = "epic"

// Unestablished is what a measurement carries when it could not be taken:
// whether it was established, and the reason it was not.
//
// Embedded in every measurement rather than expressed as a nil pointer,
// because a nil measurement carries no reason and a reader given one has
// to invent the explanation.
type Unestablished struct {
	// Established is true when the measurement was actually taken.
	Established bool
	// Reason is why it was not, empty when it was.
	Reason string
}

func taken() Unestablished { return Unestablished{Established: true} }

func missing(format string, a ...any) Unestablished {
	return Unestablished{Reason: fmt.Sprintf(format, a...)}
}

// Worktrees is the standing-worktree measurement.
type Worktrees struct {
	Unestablished
	// Standing is how many working trees are registered, after stale
	// registrations have been cleared.
	Standing int
	// Trees are the standing trees themselves, so a later measurement can
	// match one to the agent holding it.
	Trees []vcs.Worktree
}

// LastCommit is the sprint branch's most recent commit and its age.
//
// Age is carried rather than left to the caller to subtract, because it
// is the measurement the report gives and it has to be taken against the
// same instant every other measurement was.
type LastCommit struct {
	Unestablished
	SHA     string
	Subject string
	When    time.Time
	Age     time.Duration
}

// ReadySet is the size of the run's ready set — how many beads
// nothing is blocking right now.
type ReadySet struct {
	Unestablished
	// Depth is how many beads are ready.
	Count int
	// Beads are the ready beads, kept because per-agent drain is read
	// from the same query.
	Beads []tracker.Bead
}

// Dispatch is one agent's drain state, from the two queries it takes.
//
// **Both queries are required and neither is sufficient.** The ready
// query excludes work in progress, so on its own it reports an agent
// mid-bead as finished; the claimed set alone says nothing about the work
// waiting for it. An agent has drained when both are empty.
type Dispatch struct {
	Unestablished
	// Agent is the dispatch id, which is the assignee on its beads.
	Agent string
	// Ready are the ids of beads ready and assigned to it.
	Ready []string
	// InProgress are the ids of beads it has claimed.
	InProgress []string
	// Blocked are the ids of beads it holds that cannot proceed. A
	// blocked bead is not work in flight, so it does not keep an agent
	// from having drained — it is reported beside the verdict rather than
	// folded into it.
	Blocked []string
	// Drained is true when the agent holds neither ready nor in-progress
	// work.
	Drained bool
	// Worktree is the path of the standing worktree checked out on this
	// dispatch's branch, empty when none stands.
	Worktree string
}

// Completion is the run's bead count by state, over every bead in the
// run however nested.
type Completion struct {
	Unestablished
	// Total is every non-epic bead below the run epic, at any depth.
	Total int
	// Done, Blocked, Claimed, Open and Deferred partition Total by the
	// run's vocabulary of states; Unknown counts beads whose tracker
	// status this vocabulary does not cover.
	Done, Blocked, Claimed, Open, Deferred, Unknown int
	// Beads is every bead counted, so a caller can name one.
	Beads []tracker.Bead
}

// Scope is the live queue compared against the manifest's admission
// record.
type Scope struct {
	Unestablished
	// Additions are beads present in the run that `admitted` does not
	// name — work that entered after admission.
	Additions []string
	// Missing are beads `admitted` names that the run no longer holds —
	// closed out of it, or moved.
	Missing []string
	// Excluded are the beads admission deliberately left out, reported
	// beside the completion verdict so a run that finished its admitted
	// set while beads sat out reads as done over a narrowed scope rather
	// than as done.
	Excluded []string
}

// Conflict is a bead marked blocked whose dispatch's worktree still
// stands: the partial work is still on disk, so the block is unresolved
// rather than merely recorded.
type Conflict struct {
	// Bead is the blocked bead's id.
	Bead string
	// Agent is the dispatch holding it.
	Agent string
	// Worktree is where its partial work still stands.
	Worktree string
}

// Measurements is everything the collector took, as one value.
//
// There is no overall verdict field. Deciding what the numbers amount to
// is the reporter's, and a verdict recorded here would be a second
// opinion for it to disagree with.
type Measurements struct {
	// Run is the run epic the measurements are of.
	Run string
	// Branch is the sprint branch the history was read from, empty when
	// the manifest did not carry one.
	Branch string
	// At is the instant every age was taken against.
	At time.Time

	Worktrees  Worktrees
	LastCommit LastCommit
	Ready      ReadySet
	Completion Completion
	Scope      Scope

	// Dispatches are the agents holding beads in this run, sorted by id.
	Dispatches []Dispatch
	// Conflicts are blocked beads whose worktrees still stand.
	Conflicts []Conflict
}

// Deps are the seams the collector reads through. Each is an interface,
// so the unit tier substitutes a fake with no process anywhere.
type Deps struct {
	Tracker tracker.Tracker
	VCS     vcs.VersionControl
	Store   manifest.Store
	Runner  plan.Runner
	// Now supplies the instant ages are taken against, nil meaning
	// time.Now. A test fixes it so an age is a value rather than a
	// moving target.
	Now func() time.Time
}

func (d Deps) now() time.Time {
	if d.Now == nil {
		return time.Now()
	}
	return d.Now()
}

// Collect takes every measurement for run and returns them with the
// record of every invocation it made.
//
// **Each measurement is its own plan.** A plan stops at its first failed
// step by design — that ordering guard is what stops a command chaining a
// removal after a failed merge — and folding six independent reads into
// one plan would make the first failure erase five measurements that were
// perfectly takeable. So a tracker query that cannot answer costs exactly
// the line it was measuring, which is what AC-SWARM-CORE-003.5 asks for.
//
// **It executes.** Explain and dry-run are the command surface's modes
// and belong to the command that renders these measurements; a collector
// asked to explain would return a measurement set in which nothing was
// established, which is not what a reader of `--explain` wants to see.
// The same reasoning as [tracker.Descendants], for the same reason.
//
// The error return is for a condition that makes the whole measurement
// set meaningless — an empty run scope — not for a measurement that
// could not be taken. Those are reported in the value.
func Collect(ctx context.Context, deps Deps, run string) (Measurements, []plan.Record, error) {
	if run == "" {
		return Measurements{}, nil, fmt.Errorf("status: a run scope is required; an empty one would measure the whole tracker")
	}

	m := Measurements{Run: run, At: deps.now()}
	var records []plan.Record

	// The manifest first: it carries the sprint branch the history is
	// read from, so the commit measurement depends on it.
	doc, docTaken, recs := readManifest(ctx, deps, run)
	records = append(records, recs...)
	if docTaken.Established {
		if branch, ok := text(doc, "branch"); ok {
			m.Branch = branch
		}
	}

	m.Worktrees, recs = measureWorktrees(ctx, deps)
	records = append(records, recs...)

	m.LastCommit, recs = measureLastCommit(ctx, deps, m.Branch, m.At)
	records = append(records, recs...)

	m.Ready, recs = measureReadySet(ctx, deps, run)
	records = append(records, recs...)

	m.Completion, recs = measureCompletion(ctx, deps, run)
	records = append(records, recs...)

	owners, recs := measureWorktreeOwners(ctx, deps, m.Worktrees, m.Branch, m.Completion)
	records = append(records, recs...)

	m.Scope = compareScope(m.Completion, doc, docTaken)
	m.Dispatches = dispatches(m.Completion, m.Ready, owners)
	m.Conflicts = conflicts(m.Completion, owners)

	return m, records, nil
}

// readManifest fetches the run's record. A record that cannot be read is
// not fatal: the branch is then unknown and the scope comparison is
// unestablished, and every measurement that does not need it is still
// taken.
func readManifest(ctx context.Context, deps Deps, run string) (manifest.Document, Unestablished, []plan.Record) {
	if deps.Store == nil {
		return manifest.Document{}, missing("no manifest store is configured"), nil
	}
	steps := deps.Store.Read("manifest", run)
	report, records := run1(ctx, deps, "the run's manifest", steps)
	value, err := steps.ValueFrom(report)
	if err != nil {
		return manifest.Document{}, missing("the manifest could not be read: %s", err), records
	}
	doc, ok := value.(manifest.Document)
	if !ok {
		return manifest.Document{}, missing("the manifest read produced %T", value), records
	}
	return doc, taken(), records
}

func measureWorktrees(ctx context.Context, deps Deps) (Worktrees, []plan.Record) {
	if deps.VCS == nil {
		return Worktrees{Unestablished: missing("no version control is configured")}, nil
	}
	steps := deps.VCS.Worktrees("worktrees")
	report, records := run1(ctx, deps, "the standing worktrees", steps)
	value, err := steps.ValueFrom(report)
	if err != nil {
		return Worktrees{Unestablished: missing("the worktree list could not be read: %s", err)}, records
	}
	trees, ok := value.([]vcs.Worktree)
	if !ok {
		return Worktrees{Unestablished: missing("the worktree list produced %T", value)}, records
	}
	return Worktrees{Unestablished: taken(), Standing: len(trees), Trees: trees}, records
}

// measureLastCommit reads the newest commit on the sprint branch.
//
// A run with no branch recorded has no history to measure, and that is
// reported as unestablished rather than measured against HEAD: HEAD in
// the process's own directory is whatever the caller happens to have
// checked out, and an age taken from it would be a real number about the
// wrong tree.
func measureLastCommit(ctx context.Context, deps Deps, branch string, at time.Time) (LastCommit, []plan.Record) {
	if deps.VCS == nil {
		return LastCommit{Unestablished: missing("no version control is configured")}, nil
	}
	if branch == "" {
		return LastCommit{Unestablished: missing(
			"the manifest records no branch, so there is no sprint history to measure")}, nil
	}
	steps := deps.VCS.Commits("last-commit", branch, 1)
	report, records := run1(ctx, deps, "the last commit on "+branch, steps)
	value, err := steps.ValueFrom(report)
	if err != nil {
		return LastCommit{Unestablished: missing("the history of %s could not be read: %s", branch, err)}, records
	}
	commits, ok := value.([]vcs.Commit)
	if !ok {
		return LastCommit{Unestablished: missing("the history read produced %T", value)}, records
	}
	if len(commits) == 0 {
		return LastCommit{Unestablished: missing("%s carries no commits", branch)}, records
	}
	c := commits[0]
	return LastCommit{
		Unestablished: taken(),
		SHA:           c.SHA,
		Subject:       c.Subject,
		When:          c.When,
		Age:           at.Sub(c.When),
	}, records
}

func measureReadySet(ctx context.Context, deps Deps, run string) (ReadySet, []plan.Record) {
	if deps.Tracker == nil {
		return ReadySet{Unestablished: missing("no tracker is configured")}, nil
	}
	step := deps.Tracker.Ready("ready", run, tracker.ReadyFilter{ExcludeKinds: []string{KindEpic}})
	steps := plan.Steps{step}
	report, records := run1(ctx, deps, "the ready set under "+run, steps)
	value, err := steps.ValueFrom(report)
	if err != nil {
		return ReadySet{Unestablished: missing("the ready set could not be read: %s", err)}, records
	}
	beads, err := tracker.BeadsFrom(value)
	if err != nil {
		return ReadySet{Unestablished: missing("%s", err)}, records
	}
	return ReadySet{Unestablished: taken(), Count: len(beads), Beads: beads}, records
}

// measureCompletion walks the run's descendants and counts them by state.
//
// It walks rather than querying because the tracker's parent-scoped
// queries answer different questions and none of them is "every bead in
// this run": a run whose work sits under a nested epic has beads that one
// query sees and the other does not. [tracker.Descendants] carries the
// measurement behind that.
func measureCompletion(ctx context.Context, deps Deps, run string) (Completion, []plan.Record) {
	if deps.Tracker == nil {
		return Completion{Unestablished: missing("no tracker is configured")}, nil
	}
	found, records, err := tracker.Descendants(ctx, deps.Tracker, deps.Runner, run)
	if err != nil {
		return Completion{Unestablished: missing(
			"the run's beads could not be walked, so the count would be of an incomplete set: %s", err)}, records
	}

	c := Completion{Unestablished: taken()}
	for _, b := range found {
		if b.Kind == KindEpic {
			continue
		}
		c.Beads = append(c.Beads, b)
		c.Total++
		switch b.State {
		case tracker.StateDone:
			c.Done++
		case tracker.StateBlocked:
			c.Blocked++
		case tracker.StateClaimed:
			c.Claimed++
		case tracker.StateOpen:
			c.Open++
		case tracker.StateDeferred:
			c.Deferred++
		default:
			c.Unknown++
		}
	}
	return c, records
}

// compareScope names the beads in the run that admission did not, and the
// admitted beads the run no longer holds.
//
// An absent `admitted` is unestablished rather than "no additions". The
// two read identically in a report and mean opposite things: one is a run
// whose scope has not moved, the other a run nobody recorded a scope for.
func compareScope(c Completion, doc manifest.Document, docTaken Unestablished) Scope {
	if !docTaken.Established {
		return Scope{Unestablished: Unestablished{Reason: docTaken.Reason}}
	}
	if !c.Established {
		return Scope{Unestablished: missing(
			"the run's beads could not be walked, so the live queue is unknown")}
	}

	admitted, ok := strings_(doc, "admitted")
	if !ok {
		return Scope{Unestablished: missing(
			"the manifest records no admitted list, so there is nothing to compare the live queue against")}
	}

	live := map[string]bool{}
	for _, b := range c.Beads {
		live[b.ID] = true
	}
	inAdmitted := map[string]bool{}
	for _, id := range admitted {
		inAdmitted[id] = true
	}

	s := Scope{Unestablished: taken()}
	for id := range live {
		if !inAdmitted[id] {
			s.Additions = append(s.Additions, id)
		}
	}
	for _, id := range admitted {
		if !live[id] {
			s.Missing = append(s.Missing, id)
		}
	}
	s.Excluded = exclusions(doc)

	sort.Strings(s.Additions)
	sort.Strings(s.Missing)
	sort.Strings(s.Excluded)
	return s
}

// dispatches derives each agent's drain state from the two queries it
// takes: the ready set, and the beads the walk found claimed.
//
// Neither alone is right. `bd ready` excludes in-progress work, so an
// agent mid-bead has an empty ready queue and would read as finished; the
// claimed set says nothing about what is waiting. So an agent has drained
// when both are empty, and a dispatch is listed for any agent either
// query names.
func dispatches(c Completion, ready ReadySet, owners map[string]string) []Dispatch {
	byAgent := map[string]*Dispatch{}
	get := func(agent string) *Dispatch {
		d, ok := byAgent[agent]
		if !ok {
			d = &Dispatch{Agent: agent}
			byAgent[agent] = d
		}
		return d
	}

	if c.Established {
		for _, b := range c.Beads {
			if b.Assignee == "" {
				continue
			}
			// Every assignee gets an entry, whatever state its beads are
			// in. An agent whose beads have all closed is a dispatch that
			// has DRAINED, and a collector that only named agents still
			// holding work would leave the reporter unable to say so —
			// "drained" and "never existed" would print the same.
			d := get(b.Assignee)
			switch b.State {
			case tracker.StateClaimed:
				d.InProgress = append(d.InProgress, b.ID)
			case tracker.StateBlocked:
				d.Blocked = append(d.Blocked, b.ID)
			}
		}
	}
	if ready.Established {
		for _, b := range ready.Beads {
			if b.Assignee == "" {
				continue
			}
			d := get(b.Assignee)
			d.Ready = append(d.Ready, b.ID)
		}
	}

	out := make([]Dispatch, 0, len(byAgent))
	for _, d := range byAgent {
		// Both queries have to have answered for a drain verdict to mean
		// anything: with either missing, an agent holding work through
		// the query that failed would be reported drained.
		if c.Established && ready.Established {
			d.Unestablished = taken()
			d.Drained = len(d.Ready) == 0 && len(d.InProgress) == 0
		} else {
			d.Unestablished = missing(
				"drain takes both the ready set and the claimed set, and one of them could not be read")
		}
		d.Worktree = owners[d.Agent]
		sort.Strings(d.Ready)
		sort.Strings(d.InProgress)
		sort.Strings(d.Blocked)
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Agent < out[j].Agent })
	return out
}

// conflicts names every blocked bead whose dispatch's worktree still
// stands.
//
// The worktree is the durable record that a bead is unfinished: a blocked
// bead leaves its partial work in the tree rather than stashing it, so a
// standing tree beside a blocked bead is work somebody still has to
// resolve. A blocked bead whose tree is gone was already dealt with.
func conflicts(c Completion, owners map[string]string) []Conflict {
	if !c.Established {
		return nil
	}
	var out []Conflict
	for _, b := range c.Beads {
		if b.State != tracker.StateBlocked || b.Assignee == "" {
			continue
		}
		if path := owners[b.Assignee]; path != "" {
			out = append(out, Conflict{Bead: b.ID, Agent: b.Assignee, Worktree: path})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Bead < out[j].Bead })
	return out
}

// measureWorktreeOwners attributes each standing worktree, other than the
// one checked out on the sprint branch, to the dispatch that holds it:
// the assignee of the bead its HEAD commit closed.
//
// This reads one commit per worktree — [Deps.VCS.Commits] against its
// [vcs.Worktree.Head] — which is why it runs after both [measureWorktrees]
// and [measureCompletion]: it needs the standing trees to walk and the
// bead-to-assignee map to resolve each one against. A worktree whose HEAD
// is not a bead's own close commit, or whose bead this run's walk did not
// find, is not attributed — the safe default a caller (this run's
// coordinator, `sprint-cleanup.sh`) leaves standing rather than tearing
// down.
//
// Detachment is not consulted: a detached worktree's HEAD is a commit like
// any other, and the whole point of reading the commit instead of the
// branch is that the branch's name and shape carry nothing this container
// can rely on.
func measureWorktreeOwners(ctx context.Context, deps Deps, trees Worktrees, sprintBranch string, c Completion) (map[string]string, []plan.Record) {
	owners := map[string]string{}
	if !trees.Established || !c.Established {
		return owners, nil
	}

	byID := map[string]string{}
	for _, b := range c.Beads {
		if b.Assignee != "" {
			byID[b.ID] = b.Assignee
		}
	}

	var records []plan.Record
	for i, t := range trees.Trees {
		if t.Head == "" || (sprintBranch != "" && t.Branch == sprintBranch) {
			continue
		}
		steps := deps.VCS.Commits(fmt.Sprintf("worktree-owner-%d", i), t.Head, 1)
		report, recs := run1(ctx, deps, "the commit at "+t.Path, steps)
		records = append(records, recs...)
		value, err := steps.ValueFrom(report)
		if err != nil {
			continue
		}
		commits, ok := value.([]vcs.Commit)
		if !ok || len(commits) == 0 {
			continue
		}
		id := beadIDFromCloseSubject(commits[0].Subject)
		if id == "" {
			continue
		}
		if assignee, ok := byID[id]; ok && assignee != "" {
			owners[assignee] = t.Path
		}
	}
	return owners, records
}

// beadIDFromCloseSubject reads the bead id out of a commit subject shaped
// like a bead's own close — `<id> P<priority> <title>`, per
// `bb-close-bead.sh` — and returns empty for a subject that is not.
func beadIDFromCloseSubject(subject string) string {
	id, rest, found := strings.Cut(subject, " ")
	if !found || id == "" {
		return ""
	}
	if !priorityToken(rest) {
		return ""
	}
	return id
}

// priorityToken reports whether s begins with a close commit's priority
// token, "P" followed by a digit.
func priorityToken(s string) bool {
	if len(s) < 2 || s[0] != 'P' {
		return false
	}
	return s[1] >= '0' && s[1] <= '9'
}

// run1 interprets one fragment and returns its report with the records it
// produced.
func run1(ctx context.Context, deps Deps, name string, steps plan.Steps) (plan.Report, []plan.Record) {
	report, _ := deps.Runner.Run(ctx, plan.Plan{Name: name, Steps: steps}, plan.ModeExecute)
	return report, report.Records
}

// text reads a manifest key as a string, reporting whether it was
// written and decoded.
func text(doc manifest.Document, key string) (string, bool) {
	var s string
	ok, err := doc.Decode(key, &s)
	if err != nil || !ok || strings.TrimSpace(s) == "" {
		return "", false
	}
	return s, true
}

// strings_ reads a manifest key as a list of strings.
//
// The trailing underscore keeps the name off the `strings` package this
// file also uses; naming it `list` would say less about what it returns.
func strings_(doc manifest.Document, key string) ([]string, bool) {
	var out []string
	ok, err := doc.Decode(key, &out)
	if err != nil || !ok {
		return nil, false
	}
	return out, true
}

// exclusions reads the admission record's excluded entries as bead ids.
//
// The entries are objects carrying an id and the reason admission left
// the bead out; only the id is a measurement, and the reason is the
// manifest's to report. An entry that does not decode is skipped rather
// than failing the comparison: the exclusion list is reported beside the
// completion verdict, and losing the whole verdict to one malformed entry
// would be the larger loss.
func exclusions(doc manifest.Document) []string {
	var entries []struct {
		ID string `json:"id"`
	}
	ok, err := doc.Decode("excluded", &entries)
	if err != nil || !ok {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.ID != "" {
			out = append(out, e.ID)
		}
	}
	return out
}
