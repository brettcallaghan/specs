// Package gherkin parses a Gherkin source into the feature and criterion
// records the index carries.
//
// Everything here is true of Gherkin and none of it is true of any
// particular document vocabulary. Tags are strings, and which of them
// identifies a feature or a criterion is a question the caller answers
// through [Options] — no identifier pattern and no qualifier name appears
// in this package. That is what keeps parsing in the host: a vault of
// `.feature` files with no relation to this framework's document types can
// use it unchanged.
//
// A source need not declare a `Feature`. Where its content begins with any
// other keyword — a fenced block holding one `Scenario` — a synthetic
// `Feature` header is prepended before parsing, and no position the header
// occupies is ever reported. Every span returned is expressed in the
// coordinates of the document the source was read from, which is what
// makes a criterion written inside a fenced block addressable from the
// document containing it.
//
// A criterion's body is text. It is not decomposed into typed steps or
// keyed example rows: that would model one notation, and the criterion is
// the atomic unit every consumer here operates on.
package gherkin

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	cucumber "github.com/cucumber/gherkin/go/v42"
	messages "github.com/cucumber/messages/go/v34"

	"github.com/brettcallaghan/specs/cli/core/span"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

// FenceTag is the info string a fenced block holding Gherkin carries.
//
// It names a notation rather than a vocabulary, which is why it belongs
// here and not in a profile: a fence labelled this way holds Gherkin
// wherever it is written, and a caller with no document types at all
// still reads it the same way.
const FenceTag = "gherkin"

// FenceTags returns the info strings that mark a fenced block as a
// Gherkin source, in the form a fenced-block lookup takes them.
//
// A function rather than a package-level slice, so a caller cannot
// rewrite the vocabulary every later caller reads.
func FenceTags() []string { return []string{FenceTag} }

// syntheticHeader is prepended to a source whose content begins with a
// keyword other than a feature keyword, so that the parser is handed a
// document it accepts. It carries no name, because a name here would be
// invented text reported as though an author had written it.
const syntheticHeader = "Feature:\n"

// syntheticHeaderLines is the number of lines syntheticHeader adds, and
// therefore the offset between a parser line and a source line whenever
// the header was prepended.
const syntheticHeaderLines = 1

// featureKeywords are the keywords that open a feature in the dialect the
// parser reads by default. They are asked of the parser rather than
// written out, so that this package and the parser cannot disagree about
// what declares a feature.
var featureKeywords = cucumber.DialectsBuiltin().GetDialect(cucumber.DefaultDialect).FeatureKeywords()

// Options carries the caller's tag vocabulary.
//
// These two functions are the whole of what this package knows about
// meaning: it hands them the tags it read — as written, but for the
// leading `@` that makes a tag a tag — and takes back an identity or
// nothing. Either may be nil, which identifies nothing and leaves the
// records' identifiers empty; parsing still succeeds, and every tag is
// still carried, so a caller with no vocabulary loses nothing but the
// resolution it did not ask for.
type Options struct {
	// FeatureIdentity resolves a feature's tags to its identifier and the
	// code that identifier groups it under.
	FeatureIdentity func(tags []string) (specmodel.FeatureId, specmodel.FeatureCode, bool)

	// CriterionIdentity resolves a construct's tags to its criterion
	// identifier.
	//
	// It also decides how a scenario carrying `Examples` blocks divides:
	// where any block is identified, each block becomes its own criterion,
	// and where none is, the scenario is one criterion whose text covers
	// every block.
	CriterionIdentity func(tags []string) (specmodel.CriterionId, bool)
}

// Feature is one parsed feature.
type Feature struct {
	// Record is the index record, populated with everything Gherkin
	// establishes: identity where the vocabulary resolved it, name,
	// description and position.
	//
	// Its owning document, path and source form are established by the
	// caller that found the source, and its nested criteria are left empty
	// because a [specmodel.Criterion] carries no tags — the caller fills
	// them from Criteria once it has read the qualifiers it registers.
	Record specmodel.Feature

	// Tags are the feature's tags in the order written, each without its
	// leading `@`.
	Tags []string

	// Criteria are the feature's criteria in the order written, across
	// every `Rule` beneath it.
	Criteria []Criterion

	// PartialOutlines is every Scenario Outline beneath this feature whose
	// Examples blocks are inconsistently tagged, in the order written.
	PartialOutlines []PartialOutline
}

// PartialOutline is one Scenario Outline whose Examples blocks are
// inconsistently tagged: some carry an identifier and others do not.
//
// The defect is the inconsistency, not the several blocks — an outline
// whose blocks are all tagged is several criteria, and one whose blocks
// are all untagged is a single criterion written in named groups
// (AC-GHK-002.9). Both are admissible, which is why this is populated
// only when at least one block matched an identifier and at least one
// did not; a caller checking for it needs nothing but len > 0.
type PartialOutline struct {
	// Untagged is the span of each untagged block, in the order written.
	Untagged []span.Span
}

// Criterion is one parsed criterion: a `Scenario`, or one `Examples` block
// of a scenario whose blocks are separately identified.
type Criterion struct {
	// Record is the index record, populated with everything Gherkin
	// establishes: identity where the vocabulary resolved it, name, text,
	// description, rule, notation, content hash and position.
	//
	// Whether the criterion is a prohibition, whether it can only be
	// verified by hand, and at what level it is verified are read from
	// Tags by the caller whose vocabulary assigns those tags meaning.
	Record specmodel.Criterion

	// Tags are the criterion's tags in the order written, each without its
	// leading `@`.
	//
	// A criterion derived from an `Examples` block carries its scenario's
	// tags before its own: the block is half of one construct, and a
	// qualifier written on the scenario governs every block beneath it.
	Tags []string

	// Inherited is how many of Tags, from the front, are its scenario's
	// rather than its own: zero for a criterion that is a whole scenario,
	// and the scenario's tag count for one derived from an `Examples`
	// block. With FirstBlock it lets a caller judge a scenario's tag once
	// however many blocks carry it.
	Inherited int

	// FirstBlock is true for the criterion derived from a scenario's first
	// `Examples` block, and for a criterion that is a whole scenario.
	FirstBlock bool
}

// ParseError reports a source that would not parse.
//
// It is returned instead of records, never alongside them: a half-parsed
// source yields a feature missing criteria the author wrote, and an index
// that silently lost a criterion is worse than one that recorded why.
type ParseError struct {
	// Errors is every failure the parser reported, in its order, each
	// located in the containing document.
	Errors []specmodel.ParseError
}

func (e *ParseError) Error() string {
	if len(e.Errors) == 0 {
		return "gherkin: parse failed"
	}
	first := e.Errors[0]
	if len(e.Errors) == 1 {
		return fmt.Sprintf("gherkin: line %d: %s", first.Span.StartLine, first.Message)
	}
	return fmt.Sprintf("gherkin: line %d: %s (and %d more)",
		first.Span.StartLine, first.Message, len(e.Errors)-1)
}

// Parse parses all of src, for a source that is a document in its own
// right such as a `.feature` file.
func Parse(src []byte, opts Options) (*Feature, error) {
	return ParseRegion(src, span.Of(src, 0, len(src)), opts)
}

// ParseRegion parses the region of doc that at covers, for a source
// written inside a larger document such as a fenced block.
//
// Every position on the returned records is expressed in doc's own
// coordinates. A source declaring no feature returns no feature and no
// error; a source that would not parse returns a [*ParseError] and no
// records.
func ParseRegion(doc []byte, at span.Span, opts Options) (*Feature, error) {
	source := newSource(doc, at)
	if !source.hasContent() {
		return nil, nil
	}

	input := source.src
	if !source.declaresFeature() {
		source.header = syntheticHeaderLines
		input = append([]byte(syntheticHeader), source.src...)
	}

	document, err := cucumber.ParseGherkinDocument(bytes.NewReader(input), nodeIds())
	if err != nil {
		return nil, source.failure(err)
	}
	if document == nil || document.Feature == nil {
		return nil, nil
	}

	source.note(document.Comments)
	source.measure(document.Feature)
	return source.feature(document.Feature, opts), nil
}

// ContentHashOf returns the hash of a criterion whose name and text are as
// given, and of nothing else.
//
// Both are normalized first — line endings to LF, the common indentation
// removed, trailing whitespace stripped from each line — so that a
// criterion reindented by an edit to the document around it, or read from
// a CRLF checkout, hashes as it did before. A NUL byte separates the two,
// so that a name running into a text cannot collide with the same
// characters divided the other way.
func ContentHashOf(name, text string) specmodel.ContentHash {
	digest := sha256.New()
	digest.Write([]byte(normalized(name)))
	digest.Write([]byte{0})
	digest.Write([]byte(normalized(text)))
	return specmodel.ContentHash("sha256:" + hex.EncodeToString(digest.Sum(nil)))
}

// identify resolves tags to a criterion identifier, with no vocabulary
// resolving nothing.
func (o Options) identify(tags []string) (specmodel.CriterionId, bool) {
	if o.CriterionIdentity == nil {
		return "", false
	}
	return o.CriterionIdentity(tags)
}

// nodeIds returns the identifier generator the parser requires.
//
// Nothing this package produces carries one: a node identifier is a
// positional counter that shifts whenever anything above it is edited, so
// a criterion that merely moved would appear to have changed.
func nodeIds() func() string {
	minted := 0
	return func() string {
		minted++
		return strconv.Itoa(minted)
	}
}

// source is one Gherkin source, indexed by line, together with where it
// sits in the document it was read from.
type source struct {
	doc []byte
	at  span.Span
	src []byte

	// start and end hold, per source line, the byte offset of its first
	// byte and the offset just past its last, the line terminator excluded.
	start []int
	end   []int

	// header is the number of lines the synthetic header added, and so the
	// amount by which every parser line runs ahead of its source line.
	header int

	// comment records which source lines the parser read as comments, so
	// that a construct's extent can be trimmed back over the ones trailing
	// it. Comments inside a doc string are content and are absent here.
	comment map[int]bool

	// extent maps a construct's anchor line — the first of its tag lines,
	// or its keyword line where it carries none — to the last line of the
	// content beneath it.
	extent map[int]int
}

func newSource(doc []byte, at span.Span) *source {
	s := &source{
		doc:     doc,
		at:      at,
		src:     span.Slice(doc, at),
		comment: map[int]bool{},
		extent:  map[int]int{},
	}
	for offset := 0; offset < len(s.src); {
		s.start = append(s.start, offset)
		next := bytes.IndexByte(s.src[offset:], '\n')
		if next < 0 {
			s.end = append(s.end, len(s.src))
			break
		}
		s.end = append(s.end, offset+next)
		offset += next + 1
	}
	return s
}

func (s *source) lines() int { return len(s.start) }

func (s *source) line(n int) string {
	if n < 1 || n > s.lines() {
		return ""
	}
	return string(s.src[s.start[n-1]:s.end[n-1]])
}

// significant reports whether line n carries content: neither blank, nor a
// comment, nor a line of tags.
func (s *source) significant(n int) bool {
	text := strings.TrimSpace(s.line(n))
	return text != "" && !strings.HasPrefix(text, "#") && !strings.HasPrefix(text, "@")
}

// hasContent reports whether the source carries anything to parse. A
// source of blank lines, comments and tags alone declares no feature, and
// prepending a header to it would mint one nobody wrote.
func (s *source) hasContent() bool {
	for n := 1; n <= s.lines(); n++ {
		if s.significant(n) {
			return true
		}
	}
	return false
}

// declaresFeature reports whether the source's first significant line
// opens a feature.
func (s *source) declaresFeature() bool {
	for n := 1; n <= s.lines(); n++ {
		if !s.significant(n) {
			continue
		}
		text := strings.TrimSpace(s.line(n))
		for _, keyword := range featureKeywords {
			if strings.HasPrefix(text, keyword+":") {
				return true
			}
		}
		return false
	}
	return false
}

// at translates a parser location into its source line: the parser's line,
// less the lines the synthetic header added. A location inside the header
// clamps to the source's first line, so that no position the header
// occupies is ever reported.
func (s *source) sourceLine(location *messages.Location) int {
	if location == nil {
		return 1
	}
	return clamp(int(location.Line)-s.header, 1, s.lines())
}

// anchor returns the line a construct begins on: the first of its tag
// lines, or its keyword line where it carries none.
func (s *source) anchor(tags []*messages.Tag, keyword *messages.Location) int {
	anchor := s.sourceLine(keyword)
	for _, tag := range tags {
		if line := s.sourceLine(tag.Location); line < anchor {
			anchor = line
		}
	}
	return anchor
}

func (s *source) note(comments []*messages.Comment) {
	for _, comment := range comments {
		s.comment[s.sourceLine(comment.Location)] = true
	}
}

// measure records where every construct's content ends.
//
// A construct runs to the line before the next construct begins, trimmed
// back over trailing blank and comment lines. Deriving the end this way
// rather than from the last step, table row or doc string delimiter means
// no construct needs its own arithmetic, and a doc string's contents —
// which may look like anything, comments included — are inside the range
// rather than at its edge.
func (s *source) measure(feature *messages.Feature) {
	var anchors []int
	scenario := func(node *messages.Scenario) {
		anchors = append(anchors, s.anchor(node.Tags, node.Location))
		for _, examples := range node.Examples {
			anchors = append(anchors, s.anchor(examples.Tags, examples.Location))
		}
	}
	for _, child := range feature.Children {
		switch {
		case child.Background != nil:
			anchors = append(anchors, s.sourceLine(child.Background.Location))
		case child.Scenario != nil:
			scenario(child.Scenario)
		case child.Rule != nil:
			anchors = append(anchors, s.anchor(child.Rule.Tags, child.Rule.Location))
			for _, sub := range child.Rule.Children {
				switch {
				case sub.Background != nil:
					anchors = append(anchors, s.sourceLine(sub.Background.Location))
				case sub.Scenario != nil:
					scenario(sub.Scenario)
				}
			}
		}
	}

	sort.Ints(anchors)
	for i, anchor := range anchors {
		last := s.lines()
		if i+1 < len(anchors) {
			last = anchors[i+1] - 1
		}
		s.extent[anchor] = s.trimmed(anchor, last)
	}
}

// trimmed walks back from last over blank and comment lines, which belong
// to whatever follows them rather than to the construct they trail.
func (s *source) trimmed(first, last int) int {
	last = clamp(last, 1, s.lines())
	for last > first && (strings.TrimSpace(s.line(last)) == "" || s.comment[last]) {
		last--
	}
	return clamp(last, first, s.lines())
}

// spanOf returns the span covering source lines first through last, in the
// coordinates of the document the source was read from. It stops at the
// end of last's text, so a span never carries the line terminator that
// separates it from what follows.
func (s *source) spanOf(first, last int) span.Span {
	if s.lines() == 0 {
		return s.at
	}
	first = clamp(first, 1, s.lines())
	last = clamp(last, first, s.lines())
	offset := int(s.at.Offset) + s.start[first-1]
	return span.Of(s.doc, offset, s.end[last-1]-s.start[first-1])
}

// feature builds the feature record and every criterion beneath it.
func (s *source) feature(parsed *messages.Feature, opts Options) *Feature {
	record := specmodel.Feature{
		Name: parsed.Name,
		// The feature occupies the whole source: a Gherkin document
		// declares one feature, and where the source is a fenced block this
		// is the Gherkin inside the fence, the fence itself being the
		// caller's to record.
		Span: s.at,
	}
	if description := dedented(parsed.Description); description != "" {
		record.Description = &description
	}
	tags := tagNames(parsed.Tags)
	if opts.FeatureIdentity != nil {
		if id, code, ok := opts.FeatureIdentity(tags); ok {
			record.Id, record.Code = id, code
		}
	}

	built := &Feature{Record: record, Tags: tags}
	add := func(scenario *messages.Scenario, rule string) {
		criteria, partial := s.criteria(scenario, rule, opts)
		built.Criteria = append(built.Criteria, criteria...)
		if partial != nil {
			built.PartialOutlines = append(built.PartialOutlines, *partial)
		}
	}
	for _, child := range parsed.Children {
		switch {
		case child.Scenario != nil:
			add(child.Scenario, "")
		case child.Rule != nil:
			for _, sub := range child.Rule.Children {
				if sub.Scenario != nil {
					add(sub.Scenario, child.Rule.Name)
				}
			}
		}
	}
	return built
}

// lineRange is a run of source lines, empty when from is zero.
type lineRange struct{ from, to int }

// criteria builds the criteria one scenario yields: one per identified
// `Examples` block where any block is identified, and otherwise one for
// the scenario, its text covering every block it carries. The second
// return is non-nil exactly when the outline's blocks are inconsistently
// tagged (AC-GHK-002.7) — see [PartialOutline].
func (s *source) criteria(parsed *messages.Scenario, rule string, opts Options) ([]Criterion, *PartialOutline) {
	anchor := s.anchor(parsed.Tags, parsed.Location)
	tags := tagNames(parsed.Tags)

	var steps lineRange
	if len(parsed.Steps) > 0 {
		steps = lineRange{from: s.sourceLine(parsed.Steps[0].Location), to: s.extent[anchor]}
	}

	divided := false
	for _, examples := range parsed.Examples {
		if _, ok := opts.identify(tagNames(examples.Tags)); ok {
			divided = true
		}
	}

	if divided {
		// Every block becomes a criterion, identified or not. A block left
		// untagged beside a tagged one is a defect for a check to report,
		// and a check can only report what the records carry.
		built := make([]Criterion, 0, len(parsed.Examples))
		var untagged []span.Span
		anyTagged := false
		for i, examples := range parsed.Examples {
			own := tagNames(examples.Tags)
			id, ok := opts.identify(own)
			block := s.block(examples)
			position := s.spanOf(s.anchor(examples.Tags, examples.Location), block.to)
			if ok {
				anyTagged = true
			} else {
				untagged = append(untagged, position)
			}
			criterion := s.criterion(draft{
				id:          id,
				name:        examples.Name,
				description: examples.Description,
				rule:        rule,
				tags:        append(append([]string{}, tags...), own...),
				text:        s.textOf(steps, block),
				position:    position,
			})
			criterion.Inherited = len(tags)
			criterion.FirstBlock = i == 0
			built = append(built, criterion)
		}
		var partial *PartialOutline
		if anyTagged && len(untagged) > 0 {
			partial = &PartialOutline{Untagged: untagged}
		}
		return built, partial
	}

	last := s.extent[anchor]
	ranges := []lineRange{steps}
	for _, examples := range parsed.Examples {
		block := s.block(examples)
		ranges = append(ranges, block)
		last = block.to
	}
	id, _ := opts.identify(tags)
	return []Criterion{s.criterion(draft{
		id:          id,
		name:        parsed.Name,
		description: parsed.Description,
		rule:        rule,
		tags:        tags,
		text:        s.textOf(ranges...),
		position:    s.spanOf(anchor, last),
	})}, nil
}

// block returns the lines of an Examples block: its keyword line through
// its last row. Its tags are excluded, being carried as tags rather than
// as text.
func (s *source) block(examples *messages.Examples) lineRange {
	return lineRange{
		from: s.sourceLine(examples.Location),
		to:   s.extent[s.anchor(examples.Tags, examples.Location)],
	}
}

// draft is what a criterion is built from, before the parts Gherkin does
// not establish are filled in by a caller that has a vocabulary.
type draft struct {
	id          specmodel.CriterionId
	name        string
	description string
	rule        string
	tags        []string
	text        string
	position    span.Span
}

func (s *source) criterion(d draft) Criterion {
	record := specmodel.Criterion{
		Id:          d.id,
		Text:        d.text,
		Notation:    specmodel.CriterionNotationGherkin,
		ContentHash: ContentHashOf(d.name, d.text),
		Span:        d.position,
	}
	if d.name != "" {
		record.Name = &d.name
	}
	if description := dedented(d.description); description != "" {
		record.Description = &description
	}
	if d.rule != "" {
		record.Rule = &d.rule
	}
	return Criterion{Record: record, Tags: d.tags, FirstBlock: true}
}

// textOf returns the source lines of every range given, verbatim but for
// the common indentation of the whole, which is removed so that a
// criterion reads the same wherever it was indented to.
func (s *source) textOf(ranges ...lineRange) string {
	var lines []string
	for _, r := range ranges {
		if r.from == 0 {
			continue
		}
		for n := r.from; n <= clamp(r.to, r.from, s.lines()); n++ {
			lines = append(lines, s.line(n))
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(dedent(lines), "\n")
}

// failure turns the parser's own errors into located ones.
func (s *source) failure(err error) error {
	reported := flatten(err)
	located := make([]specmodel.ParseError, 0, len(reported))
	for _, one := range reported {
		located = append(located, s.locate(one))
	}
	return &ParseError{Errors: located}
}

// locationPattern matches the position the parser writes at the head of
// every message it reports.
var locationPattern = regexp.MustCompile(`(?s)^\((\d+):(\d+)\):\s*(.*)$`)

// locate splits a parser error into its position and the expectation it
// states, and expresses the position in the containing document.
func (s *source) locate(err error) specmodel.ParseError {
	match := locationPattern.FindStringSubmatch(err.Error())
	if match == nil {
		return specmodel.ParseError{Message: err.Error(), Span: s.spanOf(1, 1)}
	}
	reported, _ := strconv.Atoi(match[1])
	column, _ := strconv.Atoi(match[2])
	line := clamp(reported-s.header, 1, s.lines())

	within := clamp(column-1, 0, s.end[line-1]-s.start[line-1])
	offset := int(s.at.Offset) + s.start[line-1] + within
	return specmodel.ParseError{
		Message: match[3],
		Span:    span.Of(s.doc, offset, s.end[line-1]-s.start[line-1]-within),
	}
}

// flatten returns the errors err carries. The parser reports several at
// once through a slice type it does not export, so the shape is read
// rather than asserted.
func flatten(err error) []error {
	if err == nil {
		return nil
	}
	value := reflect.ValueOf(err)
	if value.Kind() != reflect.Slice {
		return []error{err}
	}
	var flattened []error
	for i := range value.Len() {
		one, ok := value.Index(i).Interface().(error)
		if !ok {
			continue
		}
		flattened = append(flattened, flatten(one)...)
	}
	if len(flattened) == 0 {
		return []error{err}
	}
	return flattened
}

// tagNames returns tags as written, each without the leading `@` that
// makes it a tag: the identifier is the value, and the `@` is only how it
// is written down.
func tagNames(tags []*messages.Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, strings.TrimPrefix(tag.Name, "@"))
	}
	return names
}

// dedented removes a description's common indentation and the blank lines
// around it.
func dedented(text string) string {
	lines := dedent(strings.Split(text, "\n"))
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// normalized is the form text is hashed in: LF line endings, no trailing
// whitespace on any line, and no common indentation.
func normalized(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t\r")
	}
	return strings.Join(dedent(lines), "\n")
}

// dedent removes the longest run of leading whitespace every non-blank
// line shares.
func dedent(lines []string) []string {
	common := ""
	found := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !found {
			common, found = indent, true
			continue
		}
		common = commonPrefix(common, indent)
	}
	if common == "" {
		return lines
	}
	dedented := make([]string, len(lines))
	for i, line := range lines {
		dedented[i] = strings.TrimPrefix(line, common)
	}
	return dedented
}

func commonPrefix(a, b string) string {
	limit := min(len(a), len(b))
	for i := range limit {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:limit]
}

func clamp(value, low, high int) int {
	if high < low {
		return low
	}
	return min(max(value, low), high)
}
