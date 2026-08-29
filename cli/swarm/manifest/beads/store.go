// Package beads implements [manifest.Store] over a beads run epic's
// metadata object — per the Beads Manifest Store component in
// specs/architecture/container-swarm-cli.md.
//
// It reaches beads through the Beads Tracker rather than invoking the
// tool itself, so the rule that only two components in this container
// name a tool still holds.
//
// Named for beads rather than for trackers in general, deliberately. An
// arbitrary JSON object attached to an issue is a beads feature; a store
// for a tracker without one would keep the record somewhere that tracker
// does have — a fenced block in the epic's body, a pinned comment, a
// per-issue property API — and satisfy the same interface. Which of those
// any given tracker supports is that store's author's problem, which is
// the point of the seam.
package beads

import (
	"encoding/json"
	"fmt"
	"strings"

	swarmexec "github.com/brettcallaghan/agent-plugins/cli/swarm/exec"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/manifest"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/plan"
	"github.com/brettcallaghan/agent-plugins/cli/swarm/tracker"
	beadstracker "github.com/brettcallaghan/agent-plugins/cli/swarm/tracker/beads"
)

// Store is [manifest.Store] over the run epic's beads metadata object.
//
// Tracker is the Beads Tracker it reads and writes through. The concrete
// type rather than the interface, because the metadata operation this
// store needs is not one the Tracker interface carries.
type Store struct {
	Tracker beadstracker.Tracker
}

var _ manifest.Store = Store{}

// Read declares the step that fetches the record whole.
func (s Store) Read(prefix, run string) plan.Steps {
	step := s.Tracker.Bead(prefix, run)
	step.Description = "read " + run + "'s run manifest"
	step.Decode = documentFrom(step.Decode)
	return plan.Steps{step}
}

// Write declares the guard and the write.
//
// Two steps, and the first is a read: the condition — that no key this
// write covers has moved since the caller fetched the record — can only
// be checked against the record as it is now. The guard's decoder is
// where the comparison happens, so a moved key makes the guard fail and
// the plan's own ordering policy is what stops the write. There is no
// conjunction to write and none to forget.
//
// The payload is known when the plan is built, because the record's own
// store merges at the top level: writing only the changed keys leaves
// every other key exactly as another writer left it. That is what keeps
// the whole fragment renderable before anything runs.
func (s Store) Write(prefix, run string, changes map[string]json.RawMessage, expect manifest.Document) plan.Steps {
	guard := s.Tracker.Bead(prefix+"-guard", run)
	guard.Description = "check no key this write covers has moved"
	guard.OnFailure = plan.Stop
	guard.Decode = guarding(documentFrom(guard.Decode), expect, changes)

	payload, err := json.Marshal(changes)
	if err != nil {
		// A payload that cannot be encoded is a defect in the caller, and
		// the plan still has to render. The step refuses when it would
		// have taken effect, which is where the reason belongs.
		payload = []byte("{}")
		refusal := fmt.Errorf("manifest/beads: encoding the write for %s: %w", run, err)
		write := s.Tracker.WriteMetadata(prefix+"-write", run, payload)
		write.Decode = func(swarmexec.Result) (any, error) { return nil, refusal }
		return plan.Steps{guard, write}
	}

	write := s.Tracker.WriteMetadata(prefix+"-write", run, payload)
	write.Description = fmt.Sprintf("merge %s into %s's run manifest", keyList(changes), run)
	write.Decode = documentFrom(write.Decode)
	return plan.Steps{guard, write}
}

// documentFrom wraps a bead decoder so the step's value is the manifest
// document rather than the bead carrying it.
func documentFrom(decode plan.Decoder) plan.Decoder {
	return func(res swarmexec.Result) (any, error) {
		value, err := decode(res)
		if err != nil {
			return nil, err
		}
		bead, err := tracker.BeadFrom(value)
		if err != nil {
			return nil, err
		}
		return manifest.ParseDocument(bead.Metadata)
	}
}

// guarding wraps a document decoder with the concurrent-write check.
func guarding(decode plan.Decoder, expect manifest.Document, changes map[string]json.RawMessage) plan.Decoder {
	return func(res swarmexec.Result) (any, error) {
		value, err := decode(res)
		if err != nil {
			return nil, err
		}
		current, ok := value.(manifest.Document)
		if !ok {
			return nil, fmt.Errorf("manifest/beads: the guard read %T, not a manifest document", value)
		}
		if moved := manifest.Moved(expect, current, changes); len(moved) > 0 {
			return nil, fmt.Errorf(
				"manifest/beads: refusing the write — %s changed since it was read; re-read the manifest and try again",
				strings.Join(moved, ", "))
		}
		return current, nil
	}
}

// keyList renders the keys a write covers, for the step's description.
func keyList(changes map[string]json.RawMessage) string {
	names := make([]string, 0, len(changes))
	for name := range changes {
		names = append(names, name)
	}
	if len(names) == 0 {
		return "nothing"
	}
	sortStrings(names)
	return strings.Join(names, ", ")
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
