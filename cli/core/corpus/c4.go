package corpus

import (
	"github.com/brettcallaghan/agent-plugins/cli/core/c4elements"
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
	var vocabulary ArchitectureRules
	if c.profile != nil {
		vocabulary = c.profile.Architecture()
	}
	return c4elements.Build(c4elements.Input{
		Vocabulary: vocabulary,
		Entries:    c.index.Entries,
		References: c.index.References,
		Source:     c.Source,
		Decode:     DecodeRecordBlock,
	})
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
