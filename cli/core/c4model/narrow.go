package c4model

import (
	"github.com/brettcallaghan/agent-plugins/cli/core/c4elements"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// narrow returns full narrowed to root and its descendants.
//
// A system's descendants are its containers and everything they hold; a
// container's and a module's are its components, the modules composed
// into it, and theirs; a component has none. The root's ancestors appear
// only as the chain from the top of the tree down to it, each holding no
// other child — no other component, container or composed module. A
// module root, and every module among the descendants, appears in
// `modules`; an orphaned root appears in its orphaned collection.
//
// The narrowed model holds only the relationships whose source and target
// are both the root or a descendant, and only the excluded relationships
// whose statement sits in the blueprint of one. Derived properties keep
// the values the whole corpus gives them.
func (a *assembly) narrow(full *specmodel.C4Model, root *c4elements.Element) *specmodel.C4Model {
	out := &specmodel.C4Model{
		SchemaVersion:         full.SchemaVersion,
		Workspace:             full.Workspace,
		SoftwareSystems:       []specmodel.C4System{},
		OrphanedContainers:    []specmodel.C4Container{},
		OrphanedComponents:    []specmodel.C4Component{},
		Modules:               []specmodel.C4Module{},
		Relationships:         []specmodel.C4Relationship{},
		ExcludedRelationships: []specmodel.C4ExcludedRelationship{},
	}

	modules := map[string]specmodel.C4Module{}
	for _, module := range full.Modules {
		modules[slugKey(string(module.Slug))] = module
	}
	held := map[string]bool{}
	included := map[string]bool{}
	var includeModule func(slug string)
	includeComponents := func(components []specmodel.C4Component) {
		for _, component := range components {
			held[slugKey(string(component.Slug))] = true
		}
	}
	includeModule = func(slug string) {
		key := slugKey(slug)
		module, ok := modules[key]
		if !ok || included[key] {
			return
		}
		included[key] = true
		held[key] = true
		includeComponents(module.Components)
		for _, composed := range module.Modules {
			includeModule(string(composed))
		}
	}
	includeContainer := func(container specmodel.C4Container) {
		held[slugKey(string(container.Slug))] = true
		includeComponents(container.Components)
		for _, composed := range container.Modules {
			includeModule(string(composed))
		}
	}

	key := slugKey(root.Slug)
	switch root.Level {
	case c4elements.LevelSystem:
		for _, system := range full.SoftwareSystems {
			if slugKey(string(system.Slug)) == key {
				held[key] = true
				for _, container := range system.Containers {
					includeContainer(container)
				}
				out.SoftwareSystems = append(out.SoftwareSystems, system)
			}
		}
	case c4elements.LevelContainer:
		for _, system := range full.SoftwareSystems {
			for _, container := range system.Containers {
				if slugKey(string(container.Slug)) == key {
					includeContainer(container)
					system.Containers = []specmodel.C4Container{container}
					out.SoftwareSystems = append(out.SoftwareSystems, system)
				}
			}
		}
		for _, container := range full.OrphanedContainers {
			if slugKey(string(container.Slug)) == key {
				includeContainer(container)
				out.OrphanedContainers = append(out.OrphanedContainers, container)
			}
		}
	case c4elements.LevelModule:
		includeModule(root.Slug)
	default:
		held[key] = true
		a.placeComponent(full, key, out)
	}

	if root.Level != c4elements.LevelComponent {
		for _, module := range full.Modules {
			if included[slugKey(string(module.Slug))] {
				out.Modules = append(out.Modules, module)
			}
		}
	}

	blueprints := map[string]bool{}
	for _, element := range a.set.Elements() {
		if element.Path != "" && held[slugKey(element.Slug)] {
			blueprints[element.Path] = true
		}
	}
	for _, relationship := range full.Relationships {
		if held[slugKey(string(relationship.Source))] && held[slugKey(string(relationship.Target))] {
			out.Relationships = append(out.Relationships, relationship)
		}
	}
	for _, excluded := range full.ExcludedRelationships {
		if blueprints[excluded.Path] {
			out.ExcludedRelationships = append(out.ExcludedRelationships, excluded)
		}
	}
	return out
}

// placeComponent places the component whose slug key is key into out
// beneath the chain of its ancestors in full, each holding only the next
// link of the chain, or in the orphaned components where it has no
// parent.
func (a *assembly) placeComponent(full *specmodel.C4Model, key string, out *specmodel.C4Model) {
	only := func(components []specmodel.C4Component) ([]specmodel.C4Component, bool) {
		for _, component := range components {
			if slugKey(string(component.Slug)) == key {
				return []specmodel.C4Component{component}, true
			}
		}
		return nil, false
	}
	for _, system := range full.SoftwareSystems {
		for _, container := range system.Containers {
			if components, ok := only(container.Components); ok {
				container.Components, container.Modules = components, []specmodel.Slug{}
				system.Containers = []specmodel.C4Container{container}
				out.SoftwareSystems = append(out.SoftwareSystems, system)
				return
			}
		}
	}
	for _, container := range full.OrphanedContainers {
		if components, ok := only(container.Components); ok {
			container.Components, container.Modules = components, []specmodel.Slug{}
			out.OrphanedContainers = append(out.OrphanedContainers, container)
			return
		}
	}
	for _, module := range full.Modules {
		if components, ok := only(module.Components); ok {
			module.Components, module.Modules = components, []specmodel.Slug{}
			out.Modules = append(out.Modules, module)
			return
		}
	}
	if components, ok := only(full.OrphanedComponents); ok {
		out.OrphanedComponents = components
	}
}
