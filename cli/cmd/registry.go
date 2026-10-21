// Package cmd is the Spec CLI's Command Surface
// ([[container-spec-cli]]): the process-level concerns Document Core and
// the profile modules deliberately do not own — configuration discovery
// and checking here, and (from a later change) argument parsing, output
// rendering and exit codes.
//
// This package holds no domain logic of its own. Loading and checking
// specs.toml is orchestration: it decodes the file, validates it against
// the generated configuration schema, and selects a profile by name from
// a registry — but the schema itself, the profile vocabulary, and the
// corpus and finding machinery all come from cli/core and cli/profile.
// Per specs/containers/spec-cli/spec-cli-container.md's Change guidance, a
// capability that belongs in a module is invoked from here, never
// reimplemented here.
package cmd

import (
	"sort"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/profile"
)

// Registry maps a profile name — the value an adopting repository writes
// as `[profile].name` in specs.toml — to the [corpus.Profile] it
// activates.
//
// This is the customization seam [[contract-specs-configuration]]
// describes: "the profile registry... is where a second taxonomy is
// wired in — a registry of profiles, each carrying its own single
// schema." A Registry holds this framework's built-in alongside any
// other profile a caller registers; nothing about the shape assumes
// specs is the only entry, or the last one.
//
// A Registry is not safe for concurrent registration and lookup; build
// it once at startup and treat it as read-only afterward, the same
// discipline [[schemaregistry.Registry]] documents for its own cache.
type Registry struct {
	profiles map[string]corpus.Profile
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{profiles: map[string]corpus.Profile{}}
}

// BuiltinRegistry returns a Registry with this framework's own profile —
// cli/profile's specs — already registered under its name.
//
// This is the one place in the Command Surface that imports cli/profile:
// Document Core's extension point means cli/core never does, and a
// caller wanting only a different profile, or no built-in at all, uses
// [NewRegistry] and [Registry.Register] directly instead.
func BuiltinRegistry() *Registry {
	r := NewRegistry()
	r.Register(profile.Name, profile.New())
	return r
}

// Register adds p to the registry under name, replacing any profile
// previously registered under the same name.
func (r *Registry) Register(name string, p corpus.Profile) {
	r.profiles[name] = p
}

// Lookup returns the profile registered under name, and whether one was.
func (r *Registry) Lookup(name string) (corpus.Profile, bool) {
	p, ok := r.profiles[name]
	return p, ok
}

// Names returns every registered profile name, sorted, for a finding that
// must list them (specs-configuration-contract.md: "IF [profile].name
// names no profile in the registry... a finding listing the registered
// names").
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.profiles))
	for name := range r.profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
