package c4elements

// This file declares the architecture vocabulary a profile registers
// through Document Core's Profile extension point. The corpus package
// names these types for the extension point; they are declared here
// because this package is what reads them, and the corpus's checks call
// this package rather than the other way round.

// ElementLevel is the C4 level at which a registered document type
// declares an element.
//
// The four levels are C4 notation rather than any profile's vocabulary,
// so the host holds them; a profile chooses which of its types sits at
// each.
type ElementLevel string

// The four levels a registered element type may declare, outermost
// first.
const (
	LevelSystem    ElementLevel = "system"
	LevelContainer ElementLevel = "container"
	LevelModule    ElementLevel = "module"
	LevelComponent ElementLevel = "component"
)

// Valid reports whether level is one of the four C4 levels.
func (level ElementLevel) Valid() bool {
	switch level {
	case LevelSystem, LevelContainer, LevelModule, LevelComponent:
		return true
	}
	return false
}

// ArchitectureRules is what a profile tells the host about how its
// documents declare C4 elements.
//
// The host holds none of these values. It does not know which document
// type declares a container, what property names an element, or under
// which heading a parent lists its children — it knows only that some
// types declare elements at some level, that some properties carry an
// element's name and edges, and that some heading and record-block tag
// declare a child.
//
// Each property applies to a fixed set of the registered types, and the
// host's registration validation checks that the profile's union schema
// declares it on every one of them: EntityNameProperty and
// ConsumedContractsProperty on every element type, SystemProperty on the
// container-level types, ComposedModulesProperty on the container- and
// module-level types, and ProviderProperty on ContractType. A property
// left empty is not registered and is checked nowhere.
type ArchitectureRules struct {
	// Elements are the document types that declare an element, keyed by
	// the value documents declare in their type property, as the
	// profile's document-type table is. Every key must be a type the
	// profile defines.
	Elements map[string]ElementTypeRules

	// EntityNameProperty is the frontmatter property naming the element
	// a document declares — the name a diagram draws it under.
	EntityNameProperty string

	// SystemProperty is the frontmatter property in which a
	// container-level document names the system it belongs to.
	SystemProperty string

	// ComposedModulesProperty is the frontmatter property in which a
	// container- or module-level document lists the modules composed
	// into it.
	ComposedModulesProperty string

	// ConsumedContractsProperty is the frontmatter property listing the
	// contracts an element consumes.
	ConsumedContractsProperty string

	// ContractType is the document type that declares a contract. It
	// must be a type the profile defines; empty registers no contracts.
	ContractType string

	// ProviderProperty is the frontmatter property in which a contract
	// names the element that provides it.
	ProviderProperty string
}

// ElementTypeRules is what a profile tells the host about one document
// type that declares an element.
type ElementTypeRules struct {
	// Level is the C4 level the type declares an element at.
	Level ElementLevel

	// RosterHeading is the text of the heading under which a document of
	// this type declares its children. Empty, with RosterTag, for a type
	// that holds no children.
	RosterHeading string

	// RosterTag is the record-block tag that declares one child under
	// the roster heading. It must be one of the profile's record-block
	// tags.
	RosterTag string
}
