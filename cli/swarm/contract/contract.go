// Package contract carries what the swarm CLI's JSON surfaces promise,
// beyond the shapes the generated model already fixes — per
// specs/contracts/contract-swarm-json.md.
//
// The shapes themselves are generated from schemas/swarm/*.strata and
// live in cli/swarm/swarmmodel; nothing here restates them. What lives
// here is the small amount of structure a schema cannot express on its
// own: the order of the run's stages, and which generated struct declares
// the manifest keys that become required at each.
package contract

import (
	"reflect"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// StageDeclaration pairs a stage of a run with the generated struct
// declaring the manifest keys that become required at it.
type StageDeclaration struct {
	Stage swarmmodel.RunStage
	// Type is the generated struct. Because the structs inherit in run
	// order, each carries every key an earlier stage declares as well as
	// its own; the first stage carrying a key is the stage that key
	// becomes required at.
	Type reflect.Type
}

// StageDeclarations returns the chain, in the order a run reaches it.
//
// This slice is the only place the chain is written in Go. The schema
// states it a second time, structurally, by which struct inherits which —
// and cli/swarm/contract's own tests assert the two agree, so a stage
// added to one and not the other fails rather than drifting.
func StageDeclarations() []StageDeclaration {
	return []StageDeclaration{
		{swarmmodel.RunStagePreFlight, reflect.TypeOf(swarmmodel.PreFlightKeys{})},
		{swarmmodel.RunStageScope, reflect.TypeOf(swarmmodel.ScopeKeys{})},
		{swarmmodel.RunStagePlan, reflect.TypeOf(swarmmodel.PlanKeys{})},
		{swarmmodel.RunStageDispatch, reflect.TypeOf(swarmmodel.DispatchKeys{})},
		{swarmmodel.RunStageRetro, reflect.TypeOf(swarmmodel.RetroKeys{})},
	}
}

// Stages returns the run's stages in the order a run reaches them.
func Stages() []swarmmodel.RunStage {
	declarations := StageDeclarations()
	out := make([]swarmmodel.RunStage, 0, len(declarations))
	for _, d := range declarations {
		out = append(out, d.Stage)
	}
	return out
}

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
