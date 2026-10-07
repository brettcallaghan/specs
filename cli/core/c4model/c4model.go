// Package c4model assembles the corpus's C4 model: the element tree,
// the module collection, the relationships and the excluded statements,
// as the generated [specmodel.C4Model].
//
// It is a projection of what the C4 Elements package resolves and
// decides nothing that package already owns: which members declare an
// element, which declarations are one element, and what a roster
// declares. It nests those elements by the containment edges the corpus
// declares — a container's system property, a host's composed-modules
// property, a child's place in its parent's roster — and never by
// directory layout or by a diagram. It renders nothing.
//
// Like the C4 Elements package, it names no document type, property or
// heading of its own: the properties it copies are named by the
// vocabulary a profile registers, or by the model's own shape.
package c4model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/c4diagram"
	"github.com/brettcallaghan/agent-plugins/cli/core/c4elements"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// SchemaVersion is the version of the model's shape this package emits.
const SchemaVersion = "0.1"

// Input is what [Build] reads.
type Input struct {
	// Elements is what the C4 Elements package reads; Build resolves the
	// element set from it, and reads each element's properties out of the
	// same sources through its Source and Decode.
	Elements c4elements.Input

	// Workspace is the configured workspace name the model carries.
	Workspace string

	// Root, when not empty, narrows the model to the element it names and
	// its descendants. It is compared without regard to letter case.
	Root string
}

// UnknownRootError reports a root slug that names no element.
type UnknownRootError struct {
	Slug string
}

func (e *UnknownRootError) Error() string {
	return fmt.Sprintf("%s names no element", e.Slug)
}

// Build assembles the corpus's C4 model.
//
// Every system-level element is placed in the model's systems, every
// module-level element in its modules, and every other element under its
// parent or in an orphaned collection. Every collection of elements is
// ordered by slug, and every collection of slugs likewise. No element is
// created from a diagram node. An element with no slug is identified by
// nothing and is left out.
func Build(in Input) (*specmodel.C4Model, error) {
	set, err := c4elements.Build(in.Elements)
	if err != nil {
		return nil, err
	}
	a := &assembly{in: in, set: set, entries: map[string]specmodel.IndexEntry{}}
	for _, entry := range in.Elements.Entries {
		a.entries[entry.Path] = entry
	}
	if in.Root == "" {
		return a.model(), nil
	}
	root, ok := set.Element(in.Root)
	if !ok || root.Slug == "" {
		return nil, &UnknownRootError{Slug: in.Root}
	}
	return a.narrow(a.model(), root), nil
}

// assembly is one run of [Build].
type assembly struct {
	in      Input
	set     *c4elements.Set
	entries map[string]specmodel.IndexEntry
	// sources caches each document's bytes, read once.
	sources map[string][]byte
}

func (a *assembly) model() *specmodel.C4Model {
	model := &specmodel.C4Model{
		SchemaVersion:         SchemaVersion,
		Workspace:             a.in.Workspace,
		SoftwareSystems:       []specmodel.C4System{},
		OrphanedContainers:    []specmodel.C4Container{},
		OrphanedComponents:    []specmodel.C4Component{},
		Modules:               []specmodel.C4Module{},
		Relationships:         a.relationships(),
		ExcludedRelationships: a.excluded(),
	}

	parents := a.parents()
	components := map[*c4elements.Element][]specmodel.C4Component{}
	containers := map[*c4elements.Element][]specmodel.C4Container{}

	// Components first, so each container and module finds its own
	// already placed; the element set is ordered by slug, so each
	// collection is too.
	for _, element := range a.elements(c4elements.LevelComponent) {
		component := a.component(element)
		if parent, ok := parents[element]; ok {
			components[parent] = append(components[parent], component)
			continue
		}
		model.OrphanedComponents = append(model.OrphanedComponents, component)
	}

	for _, element := range a.elements(c4elements.LevelContainer) {
		container := specmodel.C4Container{
			Components: nonNil(components[element]),
			Modules:    slugsOf(element.ComposedModules),
		}
		a.fill(element, &container)
		if system, ok := a.systemOf(element); ok {
			containers[system] = append(containers[system], container)
			continue
		}
		model.OrphanedContainers = append(model.OrphanedContainers, container)
	}

	for _, element := range a.elements(c4elements.LevelSystem) {
		system := specmodel.C4System{Containers: nonNil(containers[element])}
		a.fill(element, &system)
		model.SoftwareSystems = append(model.SoftwareSystems, system)
	}

	for _, element := range a.elements(c4elements.LevelModule) {
		values := a.values(element)
		module := specmodel.C4Module{
			ContainedBy: a.containedBy(element),
			Components:  nonNil(components[element]),
			Modules:     slugsOf(element.ComposedModules),
		}
		a.fill(element, &module)
		copyProperties(values, &moduleProperties{
			Kind: &module.Kind, Visibility: &module.Visibility, Version: &module.Version,
		})
		model.Modules = append(model.Modules, module)
	}
	return model
}

// relationships returns one relationship for each group of drawings C4
// Elements reports as one relationship, ordered by source slug, then
// target slug, then description. Its arguments are those of the drawing
// that sorts first by path and then by span.
func (a *assembly) relationships() []specmodel.C4Relationship {
	relationships := []specmodel.C4Relationship{}
	for _, group := range a.set.Relationships() {
		if len(group.Drawings) == 0 || group.Source.Slug == "" || group.Target.Slug == "" {
			continue
		}
		first := group.Drawings[0]
		relationship := specmodel.C4Relationship{
			Source:      specmodel.Citation(group.Source.Slug),
			Target:      specmodel.Citation(group.Target.Slug),
			Level:       levelOf(group.Source, group.Target),
			Description: first.Label,
			Technology:  present(first.Technology),
			Detail:      present(first.Description),
			DrawnIn:     []specmodel.C4Drawing{},
			Interaction: interactionOf(first.Tags),
			Contract:    a.contractOf(group.Source, group.Target, first.Tags),
		}
		if len(first.Tags.Unreserved) > 0 {
			tags := append([]string{}, first.Tags.Unreserved...)
			relationship.Tags = &tags
		}
		for _, drawing := range group.Drawings {
			relationship.DrawnIn = append(relationship.DrawnIn, specmodel.C4Drawing{
				Path:        drawing.Path,
				Span:        drawing.Span,
				View:        drawing.View,
				SourceAlias: drawing.SourceAlias,
				TargetAlias: drawing.TargetAlias,
			})
		}
		relationships = append(relationships, relationship)
	}
	return relationships
}

// interactionOf returns the interaction a drawing's reserved tags state:
// sync or async when it carries exactly one of them, and none when it
// carries both or neither.
func interactionOf(tags c4diagram.Tags) *specmodel.C4Interaction {
	var sync, async bool
	for _, tag := range tags.Reserved {
		switch tag.Name {
		case c4diagram.TagSync:
			sync = true
		case c4diagram.TagAsync:
			async = true
		}
	}
	var interaction specmodel.C4Interaction
	switch {
	case sync && !async:
		interaction = specmodel.C4InteractionSync
	case async && !sync:
		interaction = specmodel.C4InteractionAsync
	default:
		return nil
	}
	return &interaction
}

// contractOf returns the contract a relationship runs over: the slug the
// first contract tag of its first drawing names, whether or not it
// resolves; or, with no such tag, the one contract the source consumes
// whose provider is the target. Zero or several such contracts yield
// none.
func (a *assembly) contractOf(source, target *c4elements.Element, tags c4diagram.Tags) *specmodel.C4ContractLink {
	for _, tag := range tags.Reserved {
		if tag.Name == c4diagram.TagContract && tag.Slug != "" {
			return &specmodel.C4ContractLink{Slug: specmodel.Citation(tag.Slug), Source: specmodel.C4ContractSourceTag}
		}
	}
	var consumed []string
	decode(a.values(source), a.in.Elements.Vocabulary.ConsumedContractsProperty, &consumed)
	var fits []*c4elements.Contract
	seen := map[*c4elements.Contract]bool{}
	for _, slug := range consumed {
		contract, ok := a.set.Contract(slug)
		if !ok || seen[contract] || !contract.ProvidedBy(target) {
			continue
		}
		seen[contract] = true
		fits = append(fits, contract)
	}
	if len(fits) != 1 {
		return nil
	}
	return &specmodel.C4ContractLink{Slug: specmodel.Citation(fits[0].Slug), Source: specmodel.C4ContractSourceDerived}
}

// excluded returns every relationship statement the model does not
// admit, ordered by path and then by span: each statement C4 Elements
// leaves unresolved, with the reason it gives, and each `BiRel`
// statement, which states no single dependency. A diagram the parser does
// not read — a dynamic or deployment view among them — holds no
// statement, and a statement it could not read is not one, so neither
// contributes.
func (a *assembly) excluded() []specmodel.C4ExcludedRelationship {
	excluded := []specmodel.C4ExcludedRelationship{}
	for _, document := range a.set.Documents() {
		for _, diagram := range document.Diagrams {
			for _, unresolved := range diagram.Unresolved {
				excluded = append(excluded, a.exclusion(document.Path, unresolved.Span, unresolved.Reason))
			}
			for _, statement := range diagram.Parsed.Statements {
				if statement.Kind == c4diagram.KindRelationship && statement.Relationship != nil && statement.Relationship.Bidirectional {
					excluded = append(excluded, a.exclusion(document.Path, statement.Span, specmodel.C4ExclusionReasonBidirectional))
				}
			}
		}
	}
	sort.SliceStable(excluded, func(i, j int) bool {
		if excluded[i].Path != excluded[j].Path {
			return excluded[i].Path < excluded[j].Path
		}
		return excluded[i].Span.Offset < excluded[j].Span.Offset
	})
	return excluded
}

// exclusion records the statement at at in the document at path as
// excluded for reason, with its text as written.
func (a *assembly) exclusion(path string, at specmodel.Span, reason specmodel.C4ExclusionReason) specmodel.C4ExcludedRelationship {
	var text string
	if src := a.source(path); at.Offset+at.Length <= uint64(len(src)) {
		text = string(src[at.Offset : at.Offset+at.Length])
	}
	return specmodel.C4ExcludedRelationship{Path: path, Span: at, Reason: reason, Text: text}
}

// levelOf returns the level a relationship holds at: the finer of its
// endpoints' levels, where a system is coarser than a container and a
// container than a module or a component, both recorded as component.
func levelOf(source, target *c4elements.Element) specmodel.C4Level {
	rank := func(level c4elements.ElementLevel) int {
		switch level {
		case c4elements.LevelSystem:
			return 0
		case c4elements.LevelContainer:
			return 1
		}
		return 2
	}
	switch max(rank(source.Level), rank(target.Level)) {
	case 0:
		return specmodel.C4LevelSystem
	case 1:
		return specmodel.C4LevelContainer
	}
	return specmodel.C4LevelComponent
}

// present returns a pointer to value, or nil where it is empty: an absent
// argument and one written as the empty string both leave a property
// absent.
func present(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// elements returns every element at level that has a slug, ordered by
// slug.
func (a *assembly) elements(level c4elements.ElementLevel) []*c4elements.Element {
	var out []*c4elements.Element
	for _, element := range a.set.Elements() {
		if element.Level == level && element.Slug != "" {
			out = append(out, element)
		}
	}
	return out
}

// component returns element as a component, with its blueprint's kind
// where it has a blueprint.
func (a *assembly) component(element *c4elements.Element) specmodel.C4Component {
	var component specmodel.C4Component
	a.fill(element, &component)
	if element.Path != "" {
		copyProperties(a.values(element), &componentProperties{Kind: &component.Kind})
	}
	return component
}

// parents returns the container or module each component-level element
// is placed under: the parent whose roster declaration of it sorts first
// by path and then by span. A component no container or module roster
// declares has none.
func (a *assembly) parents() map[*c4elements.Element]*c4elements.Element {
	parents := map[*c4elements.Element]*c4elements.Element{}
	// Documents are ordered by path and each roster by span, so the first
	// declaration met is the one that sorts first.
	for _, document := range a.set.Documents() {
		parent := document.Element
		if parent.Slug == "" || (parent.Level != c4elements.LevelContainer && parent.Level != c4elements.LevelModule) {
			continue
		}
		for _, child := range document.Roster {
			element := child.Element
			if element == parent || element.Level != c4elements.LevelComponent {
				continue
			}
			if _, placed := parents[element]; !placed {
				parents[element] = parent
			}
		}
	}
	return parents
}

// systemOf returns the system-level element a container's system
// property names, if it names one.
func (a *assembly) systemOf(container *c4elements.Element) (*c4elements.Element, bool) {
	if container.System == "" {
		return nil, false
	}
	system, ok := a.set.Element(container.System)
	if !ok || system.Level != c4elements.LevelSystem || system.Slug == "" {
		return nil, false
	}
	return system, true
}

// containedBy returns every container and module whose composed-modules
// property names module, ordered by slug. It is computed over the whole
// element set.
func (a *assembly) containedBy(module *c4elements.Element) []specmodel.Citation {
	hosts := []specmodel.Citation{}
	for _, host := range a.set.Elements() {
		if host.Slug == "" || (host.Level != c4elements.LevelContainer && host.Level != c4elements.LevelModule) {
			continue
		}
		for _, composed := range host.ComposedModules {
			if slugKey(composed) == slugKey(module.Slug) {
				hosts = append(hosts, specmodel.Citation(host.Slug))
				break
			}
		}
	}
	return hosts
}

// contractsProvided returns every contract whose provider property names
// element, ordered by slug.
func (a *assembly) contractsProvided(element *c4elements.Element) []specmodel.Citation {
	provided := []specmodel.Citation{}
	for _, contract := range a.set.Contracts() {
		if contract.ProvidedBy(element) {
			provided = append(provided, specmodel.Citation(contract.Slug))
		}
	}
	return provided
}

// fill sets the properties every element carries on target, a pointer to
// one of the model's element types: its slug, declarations and derived
// contracts, and the properties copied from its source declaration.
func (a *assembly) fill(element *c4elements.Element, target any) {
	values := a.values(element)
	common := specmodel.C4Element{
		Slug:              specmodel.Slug(element.Slug),
		Declarations:      append([]specmodel.C4Declaration{}, element.Declarations...),
		ContractsProvided: a.contractsProvided(element),
	}
	vocabulary := a.in.Elements.Vocabulary
	decode(values, vocabulary.EntityNameProperty, &common.EntityName)
	decode(values, vocabulary.ConsumedContractsProperty, &common.ContractsConsumed)
	copyProperties(values, &elementProperties{
		Description: &common.Description, Technology: &common.Technology, Tags: &common.Tags,
	})

	// The model's element types repeat C4Element's fields rather than
	// embedding it, so each is set by name.
	out := reflect.ValueOf(target).Elem()
	in := reflect.ValueOf(common)
	for i := 0; i < in.NumField(); i++ {
		if field := out.FieldByName(in.Type().Field(i).Name); field.IsValid() {
			field.Set(in.Field(i))
		}
	}
}

// values returns the properties of element's source declaration: its
// blueprint's frontmatter where it has a blueprint, and otherwise the
// declaration that sorts first by path and then by span. A declaration
// that cannot be read yields none.
func (a *assembly) values(element *c4elements.Element) map[string]any {
	if element.Path != "" {
		src := a.source(element.Path)
		if doc, err := frontmatter.Parse(src); err == nil && doc != nil {
			return doc.Values()
		}
		return nil
	}
	if len(element.Declarations) == 0 || a.in.Elements.Decode == nil {
		return nil
	}
	first := element.Declarations[0]
	entry, ok := a.entries[first.Path]
	if !ok {
		return nil
	}
	for _, heading := range entry.Headings {
		for _, block := range heading.RecordBlocks {
			if block.Span.Offset != first.Span.Offset || block.Span.Length != first.Span.Length {
				continue
			}
			var record map[string]any
			if err := a.in.Elements.Decode(a.source(first.Path), block, &record); err != nil {
				return nil
			}
			return record
		}
	}
	return nil
}

// source returns the bytes of the member at path, read once, and nil
// where it cannot be read.
func (a *assembly) source(path string) []byte {
	if a.sources == nil {
		a.sources = map[string][]byte{}
	}
	if src, ok := a.sources[path]; ok {
		return src
	}
	var src []byte
	if a.in.Elements.Source != nil {
		src, _ = a.in.Elements.Source(path)
	}
	a.sources[path] = src
	return src
}

// elementProperties, moduleProperties and componentProperties name the
// properties copied from a source declaration by the model's own JSON
// names. Each field points at the model field it fills.
type elementProperties struct {
	Description *string    `json:"description"`
	Technology  **string   `json:"technology"`
	Tags        **[]string `json:"tags"`
}

type moduleProperties struct {
	Kind       **specmodel.BlueprintKind `json:"kind"`
	Visibility **specmodel.Visibility    `json:"visibility"`
	Version    **string                  `json:"version"`
}

type componentProperties struct {
	Kind **specmodel.BlueprintKind `json:"kind"`
}

// copyProperties decodes, for each field of the struct props points at,
// the value values holds under the field's JSON name into the location
// the field points at. A property values does not carry, or carries in a
// form the model cannot hold, is left as it was.
func copyProperties(values map[string]any, props any) {
	fields := reflect.ValueOf(props).Elem()
	for i := 0; i < fields.NumField(); i++ {
		name, _, _ := strings.Cut(fields.Type().Field(i).Tag.Get("json"), ",")
		decode(values, name, fields.Field(i).Interface())
	}
}

// decode decodes the value values holds under name into target, a
// pointer. A property values does not carry, or carries in a form target
// cannot hold, leaves target as it was.
func decode(values map[string]any, name string, target any) {
	value, ok := values[name]
	if name == "" || !ok || value == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	fresh := reflect.New(reflect.TypeOf(target).Elem())
	if err := json.Unmarshal(raw, fresh.Interface()); err != nil {
		return
	}
	reflect.ValueOf(target).Elem().Set(fresh.Elem())
}

// slugsOf returns slugs as the model's slugs, ordered and each held once.
func slugsOf(slugs []string) []specmodel.Slug {
	out := []specmodel.Slug{}
	seen := map[string]bool{}
	for _, slug := range slugs {
		if seen[slug] {
			continue
		}
		seen[slug] = true
		out = append(out, specmodel.Slug(slug))
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// nonNil returns items, or an empty collection where there are none, so
// the model emits an empty list rather than null.
func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

// slugKey is the form two slugs are compared in.
func slugKey(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
