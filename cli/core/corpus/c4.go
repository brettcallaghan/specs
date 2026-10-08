package corpus

import (
	"fmt"
	"slices"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/c4diagram"
	"github.com/brettcallaghan/agent-plugins/cli/core/c4elements"
	"github.com/brettcallaghan/agent-plugins/cli/core/c4model"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// C4Elements returns the corpus's architectural elements, each drawing
// document's scope and the resolution of every node drawn in one, read
// through the active profile's architecture vocabulary.
//
// Diagrams are read from each drawing document's source as it stands at
// the time of the call, since the index does not record them. A profile
// registering no architecture vocabulary yields an empty set and no
// error.
func (c *Corpus) C4Elements() (*c4elements.Set, error) {
	return c4elements.Build(c.c4Input())
}

// C4Model returns the corpus's C4 model, carrying workspace as the
// configured workspace name: the element tree nested by the containment
// edges the corpus declares, the module collection, the relationships and
// the excluded statements. It is assembled from the same element set
// [Corpus.C4Elements] returns.
//
// A non-empty root narrows the model to the element it names and its
// descendants and sets root_declared; a root naming no element answers a
// model holding nothing, with root_declared false.
func (c *Corpus) C4Model(workspace, root string) (*specmodel.C4Model, error) {
	return c4model.Build(c4model.Input{Elements: c.c4Input(), Workspace: workspace, Root: root})
}

// c4Input is what the C4 Elements package reads from this corpus,
// through the active profile's architecture vocabulary.
func (c *Corpus) c4Input() c4elements.Input {
	var vocabulary ArchitectureRules
	if c.profile != nil {
		vocabulary = c.profile.Architecture()
	}
	return c4elements.Input{
		Vocabulary: vocabulary,
		Entries:    c.index.Entries,
		References: c.index.References,
		Source:     c.Source,
		Decode:     DecodeRecordBlock,
	}
}

// c4NodeFindings raises the first half of the c4 rule family over every
// drawing document: c4/parse for a statement the parser could not read
// and, unverified, for a `mermaid` block that is not a C4 diagram it
// reads; c4/undeclared-node for a node naming no element in scope that
// is neither external nor a passive store; and c4/ambiguous-node for one
// naming several.
//
// Only a member whose type the profile registers as declaring an
// element draws, so a diagram anywhere else — a feature blueprint's —
// raises nothing, and a profile registering no architecture vocabulary
// raises nothing at all.
func c4NodeFindings(set *c4elements.Set) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, document := range set.Documents() {
		for _, diagram := range document.Diagrams {
			if !diagram.Parsed.Read() {
				findings = append(findings,
					finding.C4UnreadDiagram(document.Path, diagram.Block, diagram.Parsed.Type))
				continue
			}
			for _, failed := range diagram.Parsed.Errors {
				findings = append(findings,
					finding.C4Parse(document.Path, failed.Span, failed.Statement, failed.Reason))
			}
			for _, node := range diagram.Nodes {
				if node.External || node.PassiveStore {
					continue
				}
				switch len(node.Matches) {
				case 0:
					findings = append(findings,
						finding.C4UndeclaredNode(document.Path, node.Span, node.Alias, node.Label))
				case 1:
				default:
					matches := make([]string, 0, len(node.Matches))
					for _, element := range node.Matches {
						matches = append(matches, element.Slug)
					}
					findings = append(findings,
						finding.C4AmbiguousNode(document.Path, node.Span, node.Alias, node.Label, matches))
				}
			}
		}
	}
	return findings
}

// c4RelationshipFindings raises the second half of the c4 rule family:
// every rule about an arrow and the tags on it.
//
// Per statement, in every C4 diagram a drawing document holds:
// c4/bidirectional at a `BiRel`; c4/duplicate-edge at the later of two
// statements with one source alias and one target alias after `Rel_Back`
// normalisation; c4/interaction-conflict at a statement tagged both
// `sync` and `async`; c4/contract-unresolved for a `contract:` tag naming
// no member of the registered contract type; and c4/tag-separator for a
// tags argument holding a comma. Per drawing — a statement whose two
// endpoints each name exactly one element — c4/contract-mismatch for a
// resolved contract tag the source does not consume or the target does
// not provide. Per relationship, across the whole corpus,
// c4/drawing-conflict for each drawing that differs from the first in
// technology, description or tags.
func c4RelationshipFindings(set *c4elements.Set) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, document := range set.Documents() {
		for _, diagram := range document.Diagrams {
			findings = append(findings, statementFindings(set, document.Path, diagram)...)
			for _, drawing := range diagram.Drawings {
				findings = append(findings, contractMismatchFindings(set, drawing)...)
			}
		}
	}
	for _, relationship := range set.Relationships() {
		findings = append(findings, drawingConflictFindings(relationship)...)
	}
	return findings
}

// statementFindings raises the rules judged from one relationship
// statement, and the duplicate-edge rule judged within its diagram.
func statementFindings(set *c4elements.Set, path string, diagram c4elements.Diagram) []specmodel.Finding {
	var findings []specmodel.Finding
	edges := map[[2]string]bool{}
	for _, statement := range diagram.Parsed.Statements {
		rel := statement.Relationship
		if statement.Kind != c4diagram.KindRelationship || rel == nil {
			continue
		}
		source, target := rel.Source.Value, rel.Target.Value
		if rel.Bidirectional {
			findings = append(findings, finding.C4Bidirectional(path, statement.Span, source, target))
		} else {
			edge := [2]string{source, target}
			if edges[edge] {
				findings = append(findings, finding.C4DuplicateEdge(path, statement.Span, source, target))
			}
			edges[edge] = true
		}

		interactions := map[string]bool{}
		for _, tag := range rel.Tags.Reserved {
			switch tag.Name {
			case c4diagram.TagSync, c4diagram.TagAsync:
				interactions[tag.Name] = true
			case c4diagram.TagContract:
				if _, declared := set.Contract(tag.Slug); !declared {
					findings = append(findings,
						finding.C4ContractUnresolved(path, statement.Span, tag.Slug, source, target))
				}
			}
		}
		if interactions[c4diagram.TagSync] && interactions[c4diagram.TagAsync] {
			findings = append(findings, finding.C4InteractionConflict(path, statement.Span, source, target))
		}
		if rel.Tags.Comma {
			findings = append(findings, finding.C4TagSeparator(path, statement.Span, rel.Tags.Raw))
		}
	}
	return findings
}

// contractMismatchFindings reports each resolved contract tag on a
// drawing that its source element does not consume or its target
// element does not provide. An unresolved contract tag is
// c4/contract-unresolved's, and a statement whose endpoints do not
// resolve is not a drawing, so neither is reported here.
func contractMismatchFindings(set *c4elements.Set, drawing c4elements.Drawing) []specmodel.Finding {
	var findings []specmodel.Finding
	for _, tag := range drawing.Tags.Reserved {
		if tag.Name != c4diagram.TagContract {
			continue
		}
		contract, declared := set.Contract(tag.Slug)
		if !declared {
			continue
		}
		var why []string
		if !drawing.Source.Consumes(tag.Slug) {
			why = append(why, fmt.Sprintf("%s does not list it among the contracts it consumes", drawing.Source.Slug))
		}
		if !contract.ProvidedBy(drawing.Target) {
			why = append(why, fmt.Sprintf("its provider is not %s", drawing.Target.Slug))
		}
		if len(why) == 0 {
			continue
		}
		findings = append(findings, finding.C4ContractMismatch(drawing.Path, drawing.Span, tag.Slug,
			drawing.SourceAlias, drawing.TargetAlias, strings.Join(why, ", and ")))
	}
	return findings
}

// drawingConflictFindings reports every drawing of a relationship that
// differs from its first drawing — by path, then by span — in
// technology, description or tags, naming the first.
func drawingConflictFindings(relationship c4elements.Relationship) []specmodel.Finding {
	if len(relationship.Drawings) < 2 {
		return nil
	}
	first := relationship.Drawings[0]
	var findings []specmodel.Finding
	for _, drawing := range relationship.Drawings[1:] {
		var differs []string
		if drawing.Technology != first.Technology {
			differs = append(differs, "technology")
		}
		if drawing.Description != first.Description {
			differs = append(differs, "description")
		}
		if !sameTags(drawing.Tags, first.Tags) {
			differs = append(differs, "tags")
		}
		if len(differs) == 0 {
			continue
		}
		findings = append(findings, finding.C4DrawingConflict(drawing.Path, drawing.Span,
			drawing.SourceAlias, drawing.TargetAlias, first.Path, first.Span, strings.Join(differs, ", ")))
	}
	return findings
}

// sameTags reports whether two tags arguments read as the same tags, in
// the same order. A contract tag's slug is compared the way contract
// resolution compares slugs (c4elements' slugKey): without regard to
// letter case or surrounding space. Unreserved tags are free text and
// compare exactly.
func sameTags(a, b c4diagram.Tags) bool {
	return slices.EqualFunc(a.Reserved, b.Reserved, func(x, y c4diagram.ReservedTag) bool {
		return x.Name == y.Name && strings.ToLower(strings.TrimSpace(x.Slug)) == strings.ToLower(strings.TrimSpace(y.Slug))
	}) && slices.Equal(a.Unreserved, b.Unreserved)
}
