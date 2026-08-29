// Package contract carries what the swarm CLI's JSON surfaces promise,
// beyond the shapes the generated model already fixes — per
// specs/contracts/contract-swarm-json.md.
//
// The shapes themselves are generated from schemas/swarm/*.strata and
// live in cli/swarm/swarmmodel; nothing here restates them. What lives
// here is the small amount of structure a schema cannot express on its
// own: which struct declares the keys that become required at a given
// stage of a run, derived from the stage's own name rather than from a
// table that could disagree with it.
package contract

import (
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// StageStructName is the name of the schema definition declaring the
// manifest keys that become required at stage.
//
// The correspondence is a naming rule rather than a table: the stage
// "pre-flight" is declared by "PreFlightKeys". A table would be a second
// place a stage could be recorded, and the point of declaring each key in
// exactly one struct is that there is no second place.
func StageStructName(stage swarmmodel.RunStage) string {
	var name strings.Builder
	for _, part := range strings.Split(string(stage), "-") {
		if part == "" {
			continue
		}
		name.WriteString(strings.ToUpper(part[:1]))
		name.WriteString(part[1:])
	}
	name.WriteString("Keys")
	return name.String()
}

// Stages returns the run's stages in the order a run reaches them, which
// is the order the manifest's key structs inherit in.
func Stages() []swarmmodel.RunStage {
	return []swarmmodel.RunStage{
		swarmmodel.RunStagePreFlight,
		swarmmodel.RunStageScope,
		swarmmodel.RunStagePlan,
		swarmmodel.RunStageDispatch,
		swarmmodel.RunStageRetro,
	}
}
