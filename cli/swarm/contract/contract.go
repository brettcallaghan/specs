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

// Version is the version of contract-swarm-json this binary produces
// and reads. It is carried in every result, so a consumer branches on the
// version it was handed rather than on the binary's own.
//
// It is stated here rather than beside the renderer because more than one
// thing has to agree with it: the result carries it, and every end-to-end
// fixture records the version it was recorded against and is checked
// against this. The contract document's own frontmatter is a third copy,
// and nothing checks that it agrees — keep the two in step by hand.
const Version = "2.0"

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

// ManifestType is the generated struct carrying the whole declared key
// set: the last stage struct's keys, plus the keys that belong to no
// stage.
func ManifestType() reflect.Type { return reflect.TypeOf(swarmmodel.RunManifest{}) }

// AlwaysOptional returns the manifest keys whose absence is legitimate at
// every stage.
//
// Derived rather than listed: a key declared directly on RunManifest and
// in no stage struct is a key with no stage, and having no stage is what
// "never required" means. So the placement of a declaration says it, and
// there is no second list to keep in step.
func AlwaysOptional() []string {
	staged := map[string]bool{}
	for _, decl := range StageDeclarations() {
		for i := 0; i < decl.Type.NumField(); i++ {
			staged[wireName(decl.Type.Field(i))] = true
		}
	}

	manifest := ManifestType()
	var out []string
	for i := 0; i < manifest.NumField(); i++ {
		if name := wireName(manifest.Field(i)); name != "" && !staged[name] {
			out = append(out, name)
		}
	}
	return out
}

// wireName is the JSON key a generated field carries, empty when the
// field is not serialized.
func wireName(field reflect.StructField) string {
	tag, ok := field.Tag.Lookup("json")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "-" {
		return ""
	}
	return name
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
