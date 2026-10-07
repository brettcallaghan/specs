package corpus

import "github.com/brettcallaghan/agent-plugins/cli/core/c4elements"

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
