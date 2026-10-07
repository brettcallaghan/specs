// Package c4elements holds the corpus's architectural elements, each
// drawing document's scope, and the resolution of diagram nodes to
// elements.
//
// It owns that judgement once, so the checks that report on a diagram
// and the model that assembles the corpus's architecture cannot disagree
// about what a node is. It reads the index and the architecture
// vocabulary a profile registers ([ArchitectureRules]), and names no
// document type, property or heading of its own. Diagrams are read from
// each drawing document's source at the time of the run; the index does
// not record them.
package c4elements

import (
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/c4diagram"
	"github.com/brettcallaghan/agent-plugins/cli/core/envelope"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// recordSlugKey is the key a record block names its element's slug
// under. A record block declares an element exactly as a blueprint's
// frontmatter does, and the slug is the one identity property every
// document carries under this name, so it is the host's own rather than
// a registered one.
const recordSlugKey = "slug"

// Input is what [Build] reads.
type Input struct {
	// Vocabulary is the active profile's architecture registration. Its
	// zero value yields a set with no elements and no drawing documents.
	Vocabulary ArchitectureRules

	// Entries are the index's entries.
	Entries []specmodel.IndexEntry

	// References are the index's references, resolved. A roster's body
	// references and a composed-modules property's values are read from
	// here.
	References []specmodel.Reference

	// Source returns the bytes of the member at an entry's path, as the
	// corpus read it. It is called once per member whose type declares an
	// element. A member it fails for is kept, nameless and drawing nothing.
	Source func(path string) ([]byte, error)

	// Decode decodes a recorded record block's content out of its
	// document's source into target. It is the corpus's own record-block
	// decoding function, so a roster record is read exactly as every
	// other reader of a record block reads it.
	Decode func(src []byte, block specmodel.RecordBlock, target any) error
}

// Element is one architectural element of the corpus.
type Element struct {
	// Slug is the element's identity, as its first declaration writes it.
	Slug string

	// Name is the element's entity name: its blueprint's entity-name
	// property where a blueprint declares it, otherwise the first record
	// block's. A diagram node names an element by this.
	Name string

	// Level is the blueprint's registered level where a blueprint
	// declares the element. An element declared only by roster record
	// blocks is a component: a roster lists the components of the
	// container or module holding it.
	Level ElementLevel

	// Path is the path of the blueprint declaring the element, and empty
	// for an element only record blocks declare.
	Path string

	// Declarations are every declaration of the element — its
	// blueprint's frontmatter and each record block naming its slug —
	// ordered by path and then by span.
	Declarations []specmodel.C4Declaration

	// ConsumedContracts are the contract slugs the element's blueprint
	// lists in its consumed-contracts property, as written and without
	// any anchor. A record block declares none.
	ConsumedContracts []string

	// System is the slug the element's blueprint names in its system
	// property, as written and without any anchor, and empty where it
	// names none.
	System string

	// ComposedModules are the slugs the element's blueprint lists in its
	// composed-modules property, as written and without any anchor.
	ComposedModules []string
}

// Contract is one corpus member of the registered contract type.
type Contract struct {
	// Slug is the contract's slug, and Path the member declaring it.
	Slug string
	Path string

	// Providers are the slugs the contract names in its provider
	// property, as written and without any anchor.
	Providers []string
}

// Consumes reports whether the element lists slug among its consumed
// contracts, compared without regard to letter case.
func (e *Element) Consumes(slug string) bool {
	for _, consumed := range e.ConsumedContracts {
		if slugKey(consumed) == slugKey(slug) {
			return true
		}
	}
	return false
}

// ProvidedBy reports whether the contract names element as its provider.
func (c *Contract) ProvidedBy(element *Element) bool {
	for _, provider := range c.Providers {
		if slugKey(provider) == slugKey(element.Slug) {
			return true
		}
	}
	return false
}

// Document is one drawing document: a corpus member whose type the
// profile registers as declaring an element.
type Document struct {
	// Path is the member's path.
	Path string

	// Element is the element the document declares.
	Element *Element

	// Scope is every element a node drawn in this document may name,
	// ordered by slug: the document's own element; the elements its
	// roster declares, by record block or by a reference to a
	// component-level blueprint placed under the roster heading; the
	// elements its composed-modules property names; and every system-
	// and container-level element in the corpus.
	Scope []*Element

	// Diagrams are the document's Mermaid blocks, in the order written.
	Diagrams []Diagram

	// Roster is every child the document's roster declares, in the order
	// written: each record block of its roster tag, and each reference
	// under its roster heading resolving to a component-level blueprint.
	Roster []Child
}

// Child is one declaration of an element in a parent's roster.
type Child struct {
	// Element is the element declared.
	Element *Element

	// Span covers the declaration: the record block, fences included, or
	// the reference.
	Span specmodel.Span
}

// Diagram is one Mermaid block of a drawing document.
type Diagram struct {
	// Block covers the fenced block, fences included.
	Block specmodel.Span

	// Parsed is the block as the C4 diagram parser read it. A diagram it
	// does not read carries its type and nothing else.
	Parsed c4diagram.Diagram

	// Nodes are the diagram's element and external node statements, in
	// the order written.
	Nodes []Node

	// Drawings are the relationship statements that resolved, in the
	// order written.
	Drawings []Drawing

	// Unresolved are the relationship statements that did not, in the
	// order written. A `BiRel` statement is in neither: it states no
	// single direction, so it draws no relationship and is not an
	// endpoint failure either.
	Unresolved []Unresolved
}

// Drawing is one relationship statement whose two endpoints each name
// exactly one element.
type Drawing struct {
	// Source and Target are the elements at the two ends, in dependency
	// direction: Source depends on Target, after `Rel_Back` normalisation.
	Source *Element
	Target *Element

	// Label, Technology and Description are the statement's arguments as
	// written, empty where absent.
	Label       string
	Technology  string
	Description string

	// Tags are the statement's tags, as the parser read them.
	Tags c4diagram.Tags

	// View is the view the diagram draws.
	View specmodel.C4View

	// SourceAlias and TargetAlias are the endpoint aliases as written, in
	// dependency direction.
	SourceAlias string
	TargetAlias string

	// Path is the drawing document's path, and Span covers the statement.
	Path string
	Span specmodel.Span

	// Statement is the statement's index into [Diagram.Parsed]'s
	// statements.
	Statement int
}

// Unresolved is one relationship statement an endpoint of which does not
// name exactly one element.
type Unresolved struct {
	// Reason is why, as the first failing check finds it: every endpoint
	// is checked for being external, then for being a passive store, then
	// for naming no element — an alias with no node in the diagram
	// included — then for naming several, the source endpoint before the
	// target at each step.
	Reason specmodel.C4ExclusionReason

	// SourceEndpoint reports whether the endpoint Reason is about is the
	// source, in dependency direction, rather than the target.
	SourceEndpoint bool

	// SourceAlias and TargetAlias are the endpoint aliases as written, in
	// dependency direction.
	SourceAlias string
	TargetAlias string

	// Path is the drawing document's path, and Span covers the statement.
	Path string
	Span specmodel.Span

	// Statement is the statement's index into [Diagram.Parsed]'s
	// statements.
	Statement int
}

// Relationship is every drawing of one relationship: drawings whose
// source element, target element and label are equal.
type Relationship struct {
	Source *Element
	Target *Element
	Label  string

	// Drawings are the relationship's drawings, ordered by path and then
	// by span.
	Drawings []Drawing
}

// Node is one node statement of a diagram, resolved.
type Node struct {
	// Statement is the node's index into [Diagram.Parsed]'s statements.
	Statement int

	// Alias and Label are the node's alias and label arguments as
	// written.
	Alias string
	Label string

	// Span covers the statement.
	Span specmodel.Span

	// External marks an external node — a person, or any `_Ext` form —
	// which names no element of the corpus and is not resolved.
	External bool

	// PassiveStore marks a Db or Queue node whose label matches no
	// element in scope: a store the diagram draws rather than an element
	// it fails to name.
	PassiveStore bool

	// Matches are the in-scope elements whose name equals the label
	// exactly, ordered by slug. A node naming one element has one.
	Matches []*Element
}

// Set is the corpus's elements and its drawing documents.
type Set struct {
	elements      []*Element
	bySlug        map[string]*Element
	documents     []*Document
	relationships []Relationship
	contracts     map[string]*Contract
}

// Contract returns the member of the registered contract type declaring
// slug, compared without regard to letter case. A profile registering no
// contract type has none.
func (s *Set) Contract(slug string) (*Contract, bool) {
	contract, ok := s.contracts[slugKey(slug)]
	return contract, ok
}

// Contracts returns every member of the registered contract type,
// ordered by slug.
func (s *Set) Contracts() []*Contract {
	contracts := make([]*Contract, 0, len(s.contracts))
	for _, contract := range s.contracts {
		contracts = append(contracts, contract)
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].Slug < contracts[j].Slug })
	return contracts
}

// Elements returns every element, ordered by slug.
func (s *Set) Elements() []*Element { return s.elements }

// Element returns the element declaring slug, compared without regard to
// letter case as the corpus compares slugs.
func (s *Set) Element(slug string) (*Element, bool) {
	element, ok := s.bySlug[slugKey(slug)]
	return element, ok
}

// Documents returns every drawing document, ordered by path.
func (s *Set) Documents() []*Document { return s.documents }

// Relationships returns every relationship drawn anywhere in the corpus,
// ordered by source slug, then target slug, then label.
func (s *Set) Relationships() []Relationship { return s.relationships }

// member is one corpus member whose type declares an element, as the
// build reads it.
type member struct {
	entry   specmodel.IndexEntry
	rules   ElementTypeRules
	src     []byte
	values  map[string]any
	element *Element
	// roster covers each of the member's roster headings: the heading
	// line and the section beneath it, so a heading nested in the section
	// and the roster heading itself both lie inside. Only Offset and
	// Length are set; the span is compared, never reported.
	roster []specmodel.Span
}

// Build reads the element set, every drawing document's scope, and the
// resolution of every node drawn in one.
//
// A member whose source cannot be read declares its element without a
// name and draws nothing; a member whose source does not hold what the
// vocabulary expects — no entity name, an undecodable record block —
// declares what it can and is reported by the checks that own that
// judgement. The error return is reserved; no read failure sets it.
func Build(in Input) (*Set, error) {
	set := &Set{bySlug: map[string]*Element{}, contracts: map[string]*Contract{}}
	vocabulary := in.Vocabulary
	if len(vocabulary.Elements) == 0 {
		return set, nil
	}
	set.contracts = contractsOf(in)

	members, err := readMembers(in)
	if err != nil {
		return nil, err
	}

	byPath := map[string]*member{}
	for _, m := range members {
		byPath[m.entry.Path] = m
		m.element = set.declare(envelope.Of(m.entry.Frontmatter).Slug, specmodel.C4Declaration{
			Path: m.entry.Path,
			Span: frontmatterSpan(m.src),
		})
		if m.element.Path == "" {
			m.element.Path = m.entry.Path
			m.element.Level = m.rules.Level
			m.element.Name = stringValue(m.values[vocabulary.EntityNameProperty])
		}
	}

	for _, ref := range in.References {
		m, ok := byPath[ref.Path]
		if !ok || m.element.Path != m.entry.Path {
			continue
		}
		switch {
		case fromProperty(ref, vocabulary.ConsumedContractsProperty):
			m.element.ConsumedContracts = append(m.element.ConsumedContracts, ref.Slug)
		case fromProperty(ref, vocabulary.SystemProperty) && m.element.System == "":
			m.element.System = ref.Slug
		case fromProperty(ref, vocabulary.ComposedModulesProperty):
			m.element.ComposedModules = append(m.element.ComposedModules, ref.Slug)
		}
	}

	refsByPath := map[string][]specmodel.Reference{}
	for _, ref := range in.References {
		refsByPath[ref.Path] = append(refsByPath[ref.Path], ref)
	}

	rosters := map[*member][]Child{}
	for _, m := range members {
		rosters[m] = set.declareRoster(in, vocabulary, m)
	}
	set.finishElements()

	for _, m := range members {
		document := &Document{Path: m.entry.Path, Element: m.element}
		document.Roster = rosterOf(refsByPath[m.entry.Path], m, rosters[m], byPath)
		document.Scope = scopeOf(refsByPath[m.entry.Path], vocabulary, set, m, document.Roster, byPath)
		document.Diagrams = diagramsOf(m.entry.Path, m.src, document.Scope)
		set.documents = append(set.documents, document)
	}
	sort.Slice(set.documents, func(i, j int) bool { return set.documents[i].Path < set.documents[j].Path })
	set.relationships = group(set.documents)
	return set, nil
}

// group gathers every drawing into the relationship it draws: one per
// distinct source element, target element and label.
func group(documents []*Document) []Relationship {
	type key struct {
		source, target *Element
		label          string
	}
	index := map[key]int{}
	var relationships []Relationship
	for _, document := range documents {
		for _, diagram := range document.Diagrams {
			for _, drawing := range diagram.Drawings {
				k := key{drawing.Source, drawing.Target, drawing.Label}
				i, ok := index[k]
				if !ok {
					i = len(relationships)
					index[k] = i
					relationships = append(relationships, Relationship{
						Source: drawing.Source, Target: drawing.Target, Label: drawing.Label,
					})
				}
				relationships[i].Drawings = append(relationships[i].Drawings, drawing)
			}
		}
	}
	for _, relationship := range relationships {
		sort.SliceStable(relationship.Drawings, func(i, j int) bool {
			a, b := relationship.Drawings[i], relationship.Drawings[j]
			if a.Path != b.Path {
				return a.Path < b.Path
			}
			return a.Span.Offset < b.Span.Offset
		})
	}
	sort.SliceStable(relationships, func(i, j int) bool {
		a, b := relationships[i], relationships[j]
		if a.Source.Slug != b.Source.Slug {
			return a.Source.Slug < b.Source.Slug
		}
		if a.Target.Slug != b.Target.Slug {
			return a.Target.Slug < b.Target.Slug
		}
		return a.Label < b.Label
	})
	return relationships
}

// contractsOf returns every member of the registered contract type, keyed
// by slug, with the providers its provider property names.
func contractsOf(in Input) map[string]*Contract {
	contracts := map[string]*Contract{}
	if in.Vocabulary.ContractType == "" {
		return contracts
	}
	byPath := map[string]*Contract{}
	for _, entry := range in.Entries {
		shared := envelope.Of(entry.Frontmatter)
		if entry.Frontmatter == nil || string(shared.Type) != in.Vocabulary.ContractType {
			continue
		}
		key := slugKey(string(shared.Slug))
		if _, seen := contracts[key]; seen {
			continue
		}
		contract := &Contract{Slug: string(shared.Slug), Path: entry.Path}
		contracts[key] = contract
		byPath[entry.Path] = contract
	}
	for _, ref := range in.References {
		contract, ok := byPath[ref.Path]
		if ok && fromProperty(ref, in.Vocabulary.ProviderProperty) {
			contract.Providers = append(contract.Providers, ref.Slug)
		}
	}
	return contracts
}

// fromProperty reports whether ref was written in the frontmatter
// property named property. An empty property is registered nowhere.
func fromProperty(ref specmodel.Reference, property string) bool {
	origin, ok := ref.Origin.AsFrontmatterProperty()
	return ok && property != "" && origin.Property == property
}

// readMembers returns every member whose type declares an element, in
// path order, with its source read and its frontmatter values parsed.
func readMembers(in Input) ([]*member, error) {
	var members []*member
	for _, entry := range in.Entries {
		if entry.Frontmatter == nil {
			continue
		}
		rules, registered := in.Vocabulary.Elements[string(envelope.Of(entry.Frontmatter).Type)]
		if !registered {
			continue
		}
		// A member whose source cannot be read stays, nameless and
		// drawing nothing, so one unreadable file does not drop the rest.
		src, err := in.Source(entry.Path)
		m := &member{entry: entry, rules: rules, values: map[string]any{}}
		if err == nil {
			m.src = src
			if doc, perr := frontmatter.Parse(src); perr == nil && doc != nil {
				m.values = doc.Values()
			}
		}
		if err == nil && rules.RosterHeading != "" {
			wanted := markdown.Identifier(rules.RosterHeading)
			for _, heading := range entry.Headings {
				if markdown.Identifier(heading.Text) == wanted {
					m.roster = append(m.roster, specmodel.Span{
						Offset: heading.Span.Offset,
						Length: heading.SectionSpan.Offset + heading.SectionSpan.Length - heading.Span.Offset,
					})
				}
			}
		}
		members = append(members, m)
	}
	sort.Slice(members, func(i, j int) bool { return members[i].entry.Path < members[j].entry.Path })
	return members, nil
}

// declareRoster declares an element for every record block of the
// member's roster tag attached to its roster heading or to a heading
// nested in its section, and returns those elements in the order
// written. A member whose type registers no roster declares none.
func (s *Set) declareRoster(in Input, vocabulary ArchitectureRules, m *member) []Child {
	if m.rules.RosterTag == "" || len(m.roster) == 0 || in.Decode == nil {
		return nil
	}
	var declared []Child
	for _, heading := range m.entry.Headings {
		if !insideAny(heading.Span, m.roster) {
			continue
		}
		for _, block := range heading.RecordBlocks {
			if block.Tag != m.rules.RosterTag {
				continue
			}
			var record map[string]any
			if err := in.Decode(m.src, block, &record); err != nil {
				continue
			}
			slug := stringValue(record[recordSlugKey])
			name := stringValue(record[vocabulary.EntityNameProperty])
			if slug == "" || name == "" {
				continue
			}
			element := s.declare(specmodel.Slug(slug), specmodel.C4Declaration{Path: m.entry.Path, Span: block.Span})
			if element.Name == "" {
				element.Name = name
			}
			declared = append(declared, Child{Element: element, Span: block.Span})
		}
	}
	return declared
}

// declare records one declaration of slug and returns the element
// holding it, creating the element at its first declaration.
func (s *Set) declare(slug specmodel.Slug, at specmodel.C4Declaration) *Element {
	key := slugKey(string(slug))
	element, ok := s.bySlug[key]
	if key == "" {
		// A member with no slug declares an element of its own, which no
		// declaration merges with and no slug lookup finds.
		element, ok = nil, false
	}
	if !ok {
		element = &Element{Slug: string(slug)}
		if key != "" {
			s.bySlug[key] = element
		}
		s.elements = append(s.elements, element)
	}
	element.Declarations = append(element.Declarations, at)
	return element
}

// finishElements orders the elements by slug and each element's
// declarations by path and then by span, and gives an element no
// blueprint declares its level.
func (s *Set) finishElements() {
	for _, element := range s.elements {
		if element.Path == "" {
			element.Level = LevelComponent
		}
		sort.SliceStable(element.Declarations, func(i, j int) bool {
			a, b := element.Declarations[i], element.Declarations[j]
			if a.Path != b.Path {
				return a.Path < b.Path
			}
			return a.Span.Offset < b.Span.Offset
		})
	}
	sort.SliceStable(s.elements, func(i, j int) bool {
		a, b := s.elements[i], s.elements[j]
		if a.Slug != b.Slug {
			return a.Slug < b.Slug
		}
		return a.Path < b.Path
	})
}

// scopeOf returns the elements a node drawn in m may name, ordered by
// slug and each held once.
func scopeOf(refs []specmodel.Reference, vocabulary ArchitectureRules, set *Set, m *member, roster []Child, byPath map[string]*member) []*Element {
	scope := map[*Element]bool{m.element: true}
	for _, child := range roster {
		scope[child.Element] = true
	}
	for _, ref := range refs {
		resolved, ok := ref.Resolution.AsResolved()
		if !ok {
			continue
		}
		target, ok := byPath[resolved.TargetPath]
		if !ok {
			continue
		}
		if fromProperty(ref, vocabulary.ComposedModulesProperty) {
			scope[target.element] = true
		}
	}
	for _, element := range set.elements {
		if element.Path != "" && (element.Level == LevelSystem || element.Level == LevelContainer) {
			scope[element] = true
		}
	}

	ordered := make([]*Element, 0, len(scope))
	for element := range scope {
		ordered = append(ordered, element)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Slug != ordered[j].Slug {
			return ordered[i].Slug < ordered[j].Slug
		}
		return ordered[i].Path < ordered[j].Path
	})
	return ordered
}

// rosterOf returns every child m's roster declares, in the order written:
// the record blocks already declared, and each body reference placed
// under a roster heading that resolves to a component-level blueprint.
func rosterOf(refs []specmodel.Reference, m *member, records []Child, byPath map[string]*member) []Child {
	roster := append([]Child(nil), records...)
	for _, ref := range refs {
		if _, body := ref.Origin.AsBody(); !body || !insideAny(ref.Span, m.roster) {
			continue
		}
		resolved, ok := ref.Resolution.AsResolved()
		if !ok {
			continue
		}
		if target, ok := byPath[resolved.TargetPath]; ok && target.rules.Level == LevelComponent {
			roster = append(roster, Child{Element: target.element, Span: ref.Span})
		}
	}
	sort.SliceStable(roster, func(i, j int) bool { return roster[i].Span.Offset < roster[j].Span.Offset })
	return roster
}

// diagramsOf parses every Mermaid block in src, resolves its nodes
// against scope, and resolves its relationship statements against its
// nodes.
func diagramsOf(path string, src []byte, scope []*Element) []Diagram {
	var diagrams []Diagram
	for _, block := range markdown.FencedFrom(src, bodyOffset(src), []string{c4diagram.FenceTag}) {
		diagram := Diagram{Block: block.Outer, Parsed: c4diagram.Parse(src, block.Inner)}
		byAlias := map[string]int{}
		duplicated := map[string]bool{}
		for i, statement := range diagram.Parsed.Statements {
			if statement.Kind != c4diagram.KindElement && statement.Kind != c4diagram.KindExternal {
				continue
			}
			node := resolve(i, statement, scope)
			if _, seen := byAlias[node.Alias]; !seen {
				byAlias[node.Alias] = len(diagram.Nodes)
			} else {
				duplicated[node.Alias] = true
			}
			diagram.Nodes = append(diagram.Nodes, node)
		}
		view := views[diagram.Parsed.Type]
		for i, statement := range diagram.Parsed.Statements {
			rel := statement.Relationship
			if statement.Kind != c4diagram.KindRelationship || rel == nil || rel.Bidirectional {
				continue
			}
			ends := [2]*Node{}
			var many [2]bool
			for e, alias := range []string{rel.Source.Value, rel.Target.Value} {
				if n, ok := byAlias[alias]; ok {
					ends[e] = &diagram.Nodes[n]
					many[e] = duplicated[alias]
				}
			}
			if reason, source, failed := endpointFailure(ends, many); failed {
				diagram.Unresolved = append(diagram.Unresolved, Unresolved{
					Reason:         reason,
					SourceEndpoint: source,
					SourceAlias:    rel.Source.Value,
					TargetAlias:    rel.Target.Value,
					Path:           path,
					Span:           statement.Span,
					Statement:      i,
				})
				continue
			}
			diagram.Drawings = append(diagram.Drawings, Drawing{
				Source:      ends[0].Matches[0],
				Target:      ends[1].Matches[0],
				Label:       param(statement, c4diagram.ParamLabel),
				Technology:  param(statement, c4diagram.ParamTechnology),
				Description: param(statement, c4diagram.ParamDescription),
				Tags:        rel.Tags,
				View:        view,
				SourceAlias: rel.Source.Value,
				TargetAlias: rel.Target.Value,
				Path:        path,
				Span:        statement.Span,
				Statement:   i,
			})
		}
		diagrams = append(diagrams, diagram)
	}
	return diagrams
}

// views maps each diagram type the parser reads to the view it draws.
var views = map[string]specmodel.C4View{
	c4diagram.TypeContext:   specmodel.C4ViewContext,
	c4diagram.TypeContainer: specmodel.C4ViewContainer,
	c4diagram.TypeComponent: specmodel.C4ViewComponent,
}

// endpointFailure returns why a relationship's endpoints, source then
// target, do not each name exactly one element, and which of them it is
// about. A nil endpoint is an alias that names no node in the diagram;
// many[e] marks an endpoint whose alias more than one node carries,
// which is ambiguous however each node resolves.
func endpointFailure(ends [2]*Node, many [2]bool) (specmodel.C4ExclusionReason, bool, bool) {
	checks := []struct {
		reason specmodel.C4ExclusionReason
		fails  func(n *Node, many bool) bool
	}{
		// A duplicated alias names no single node, so the single-node
		// reasons do not apply to it and it reaches ambiguous_node.
		{specmodel.C4ExclusionReasonExternalEndpoint, func(n *Node, many bool) bool { return !many && n != nil && n.External }},
		{specmodel.C4ExclusionReasonPassiveStore, func(n *Node, many bool) bool { return !many && n != nil && n.PassiveStore }},
		{specmodel.C4ExclusionReasonUndeclaredNode, func(n *Node, many bool) bool { return !many && (n == nil || len(n.Matches) == 0) }},
		{specmodel.C4ExclusionReasonAmbiguousNode, func(n *Node, many bool) bool { return many || len(n.Matches) > 1 }},
	}
	for _, check := range checks {
		for e, end := range ends {
			if check.fails(end, many[e]) {
				return check.reason, e == 0, true
			}
		}
	}
	return "", false, false
}

// param returns the argument a statement gave the named parameter, and
// empty where it gave none.
func param(statement c4diagram.Statement, name string) string {
	arg, _ := statement.Param(name)
	return arg.Value
}

// resolve resolves one node statement against scope: an external node
// is not resolved, any other node matches every in-scope element whose
// name equals its label as written, and a Db or Queue node matching none
// is a passive store.
func resolve(index int, statement c4diagram.Statement, scope []*Element) Node {
	alias, _ := statement.Param(c4diagram.ParamAlias)
	label, _ := statement.Param(c4diagram.ParamLabel)
	node := Node{
		Statement: index,
		Alias:     alias.Value,
		Label:     label.Value,
		Span:      statement.Span,
		External:  statement.Kind == c4diagram.KindExternal,
	}
	if node.External {
		return node
	}
	for _, element := range scope {
		// An element with no entity name is named by no node.
		if element.Name == "" {
			continue
		}
		if element.Name == node.Label {
			node.Matches = append(node.Matches, element)
		}
	}
	node.PassiveStore = len(node.Matches) == 0 && statement.Store
	return node
}

// frontmatterSpan covers src's frontmatter block, delimiters included —
// the declaration a blueprint makes.
func frontmatterSpan(src []byte) specmodel.Span {
	region, ok := frontmatter.Locate(src)
	if !ok {
		return specmodel.Span{}
	}
	return region.Outer
}

// bodyOffset returns where src's body starts: after its frontmatter
// block, or at the start where there is none.
func bodyOffset(src []byte) int {
	region, ok := frontmatter.Locate(src)
	if !ok {
		return 0
	}
	return int(region.Outer.Offset + region.Outer.Length)
}

// insideAny reports whether at lies within one of sections.
func insideAny(at specmodel.Span, sections []specmodel.Span) bool {
	for _, section := range sections {
		if at.Offset >= section.Offset && at.Offset+at.Length <= section.Offset+section.Length {
			return true
		}
	}
	return false
}

// stringValue returns value when it is a string, and empty otherwise.
func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

// slugKey is the form two slugs are compared in.
func slugKey(slug string) string {
	return strings.ToLower(strings.TrimSpace(slug))
}
