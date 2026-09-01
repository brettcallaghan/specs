package cli

import (
	"encoding/json"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"

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
			if !satisfied(f) {
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
		if !satisfied(f) {
			failing++
			// One element per issue, each a WHOLE issue: what is wrong
			// and, where there is one, what would put it right — in one
			// string. Split across two elements, a reader has to join
			// them back up, and a second issue is indistinguishable from
			// the second half of the first. The fact carries the joined
			// sentence because only the command knows the resolution
			// for THIS property.
			resolution := f.Resolution
			if resolution == "" {
				resolution = r.Resolution
			}
			issues = append(issues, yStr(issueOf(f.Reason, resolution)))
		}
		verdicts = append(verdicts, yPair{f.Name, yMap{
			{"ok", yBool(satisfied(f))},
			// The value is whatever the property holds, which is null
			// when nothing could be read and is the offending value when
			// something could be read and is wrong. A null in the second
			// case would hide which of the two the reader is looking at.
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
	if condition := issueOf(r.Summary, r.Resolution); condition != "" {
		issues = append(issues, yStr(condition))
	}
	return append(yMap{{"issues", issues}}, verdicts...)
}

// satisfied reports whether a property is both readable and right.
// Established is not the same as satisfactory: a value that was read and
// is wrong is established and carries a reason.
func satisfied(f Fact) bool { return f.Established && f.Reason == "" }

// issueOf joins what is wrong with what would put it right, as the one
// sentence a reader acts on.
func issueOf(reason, resolution string) string {
	switch {
	case reason == "":
		return resolution
	case resolution == "":
		return reason
	}
	if !strings.HasSuffix(reason, ".") {
		reason += "."
	}
	return reason + " " + resolution
}

// factNode renders a property's value as its own type: a number
// unquoted, a boolean as true or false, an absent value as null, a list
// as a sequence.
func factNode(f Fact) yamlNode {
	if f.Value == nil {
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
	case swarmmodel.FactKindStructured:
		if f.Value.Structured != nil {
			return structuredNode(*f.Value.Structured)
		}
	}
	return yNull()
}

// structuredNode parses a structured value's JSON and renders it as real
// structure, so an object arrives as a mapping rather than as an escaped
// string in the middle of otherwise readable YAML.
//
// An object's keys are sorted, because encoding/json hands back a map
// and the source order is already gone by the time this sees it. Every
// other ordering in this package is the command's own.
func structuredNode(document string) yamlNode {
	var decoded any
	if err := json.Unmarshal([]byte(document), &decoded); err != nil {
		// Unreachable: [Result.Validate] refuses a structured value that
		// is not valid JSON. Rendering the text is the honest fallback —
		// it loses the structure and loses nothing else.
		return yStr(document)
	}
	return jsonNode(decoded)
}

func jsonNode(value any) yamlNode {
	switch v := value.(type) {
	case nil:
		return yNull()
	case bool:
		return yBool(v)
	case float64:
		if v == math.Trunc(v) && math.Abs(v) < 1<<53 {
			return yInt(int64(v))
		}
		return yScalar{lit: strconv.FormatFloat(v, 'g', -1, 64)}
	case string:
		return yStr(v)
	case []any:
		items := ySeq{}
		for _, item := range v {
			items = append(items, jsonNode(item))
		}
		return items
	case map[string]any:
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		fields := yMap{}
		for _, key := range keys {
			fields = append(fields, yPair{key, jsonNode(v[key])})
		}
		return fields
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
