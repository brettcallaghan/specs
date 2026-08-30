package cli

import (
	"io"

	"github.com/brettcallaghan/agent-plugins/cli/swarm/swarmmodel"
)

// The answer: what a command prints when the full result was not asked
// for.
//
// The reader is an agent that knows which command it ran and nothing
// else. So the answer carries the properties the command established and
// nothing about how it established them — no header naming the command
// the caller typed, no transcript nobody asked for, no summary of a
// mapping the reader can see, and no outcome line restating the exit
// status the caller already holds. The provenance is one flag away.

// answer builds the document a command prints.
//
// The exit status selects the shape, because a reader needs to know which
// shape they are holding before they parse it and the status is a value
// they already route on: clear renders the properties and their values,
// and every other status renders a verdict per property.
func answer(r Result) yamlNode {
	if r.Status() == StatusClear {
		established := yMap{}
		for _, f := range r.Facts {
			if !f.Established {
				continue
			}
			established = append(established, yPair{f.Name, factNode(f)})
		}
		return established
	}

	verdicts := yMap{}
	failing := 0
	for _, f := range r.Facts {
		issues := ySeq{}
		if !f.Established {
			failing++
			if f.Reason != "" {
				issues = append(issues, yStr(f.Reason))
			}
			// The resolution is the actionable half and it has to reach
			// the caller. It goes inside the property rather than beside
			// the mapping, because the answer is a mapping of properties
			// and a sibling key would collide with the first command
			// that reports a property by that name.
			if r.Resolution != "" {
				issues = append(issues, yStr(r.Resolution))
			}
		}
		verdicts = append(verdicts, yPair{f.Name, yMap{
			{"ok", yBool(f.Established)},
			{"value", factNode(f)},
			{"issues", issues},
		}})
	}

	if failing > 0 {
		return verdicts
	}

	// Nothing a property carries explains this status — a refusal that
	// established nothing, or a command whose properties are all in hand
	// and whose finding is about the run rather than about any one of
	// them. It is reported where a refusal reports one, and the document
	// stays decodable because `issues` maps to a sequence where a
	// property maps to a mapping.
	issues := ySeq{}
	if r.Summary != "" {
		issues = append(issues, yStr(r.Summary))
	}
	if r.Resolution != "" {
		issues = append(issues, yStr(r.Resolution))
	}
	return append(yMap{{"issues", issues}}, verdicts...)
}

// factNode renders a property's value as its own type: a number
// unquoted, a boolean as true or false, an absent value as null, a list
// as a sequence.
func factNode(f Fact) yamlNode {
	if !f.Established || f.Value == nil {
		return yNull()
	}
	switch f.Value.Kind {
	case swarmmodel.FactKindText:
		if f.Value.Text != nil {
			return yStr(*f.Value.Text)
		}
	case swarmmodel.FactKindCount:
		if f.Value.Count != nil {
			return yInt(*f.Value.Count)
		}
	case swarmmodel.FactKindFlag:
		if f.Value.Flag != nil {
			return yBool(*f.Value.Flag)
		}
	case swarmmodel.FactKindList:
		if f.Value.Items != nil {
			items := ySeq{}
			for _, item := range *f.Value.Items {
				items = append(items, yStr(item))
			}
			return items
		}
	}
	return yNull()
}

// renderAnswer writes the answer to w, as JSON when asJSON is set and as
// YAML otherwise.
func renderAnswer(w io.Writer, r Result, asJSON bool) error {
	node := answer(r)
	if asJSON {
		return writeJSON(w, node)
	}
	return writeYAML(w, node)
}
