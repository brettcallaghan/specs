package corpus

import (
	"fmt"
	"sort"

	"github.com/goccy/go-yaml"

	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// recordTagsOf returns the info strings profile registers as record
// blocks, in registration order. A profile registering none yields nil.
func recordTagsOf(profile Profile) []string {
	var tags []string
	for _, registered := range profile.RecordBlocks() {
		tags = append(tags, registered.Tag)
	}
	return tags
}

// sortedTags returns a sorted copy of tags, so two profiles registering
// the same tags in a different order hash alike.
func sortedTags(tags []string) []string {
	sorted := append([]string(nil), tags...)
	sort.Strings(sorted)
	return sorted
}

// recordBlocksOf returns every fenced block in src whose info string is
// one of tags, as the index records them: tag and span, never content.
//
// The blocks are recorded whether or not their content is valid. Judging
// a record is a check's job; a block that fails it is still a place a
// reader can be sent. The search starts at byte offset from, the start of
// the body, so a fence written inside frontmatter YAML is not a record;
// the spans are in whole-file coordinates, the same as the heading spans.
func recordBlocksOf(src []byte, from int, tags []string) []specmodel.RecordBlock {
	var records []specmodel.RecordBlock
	for _, block := range markdown.FencedFrom(src, from, tags) {
		records = append(records, specmodel.RecordBlock{Tag: block.Tag, Span: block.Outer})
	}
	return records
}

// attachRecordBlocks attaches each record to the innermost heading whose
// section span contains it and returns the records no section contains.
// Each record is attached exactly once.
//
// The innermost containing heading is the one starting last: sections
// nest, so every heading whose section contains a block starts at or
// before it, and the nearest such start is the deepest.
func attachRecordBlocks(headings []specmodel.Heading, records []specmodel.RecordBlock) []specmodel.RecordBlock {
	var unattached []specmodel.RecordBlock
	for _, record := range records {
		innermost := -1
		for i, heading := range headings {
			section := heading.SectionSpan
			if record.Span.Offset < section.Offset || record.Span.Offset+record.Span.Length > section.Offset+section.Length {
				continue
			}
			if innermost < 0 || section.Offset >= headings[innermost].SectionSpan.Offset {
				innermost = i
			}
		}
		if innermost < 0 {
			unattached = append(unattached, record)
			continue
		}
		headings[innermost].RecordBlocks = append(headings[innermost].RecordBlocks, record)
	}
	return unattached
}

// DecodeRecordBlock reads the content of a recorded block out of src
// through its span and decodes it as YAML into target, which is a
// pointer to the caller's own Go value.
//
// src is the bytes of the document the block was recorded in. The block
// is re-located from the body through Markdown Structure, as the index
// searched, and its content read as the
// parser reads it, container prefixes (a blockquote's "> ") removed, so
// every command and check decodes a record the same way. A span outside
// src, a span that does not cover a fenced block of the block's tag, or
// content that is not YAML fitting target is an error.
func DecodeRecordBlock(src []byte, block specmodel.RecordBlock, target any) error {
	end := block.Span.Offset + block.Span.Length
	if end > uint64(len(src)) {
		return fmt.Errorf("record block span [%d,%d) lies outside a %d-byte document",
			block.Span.Offset, end, len(src))
	}
	for _, found := range markdown.FencedFrom(src, bodyOffset(src), []string{block.Tag}) {
		if found.Outer.Offset == block.Span.Offset && found.Outer.Length == block.Span.Length {
			return yaml.Unmarshal(found.Text, target)
		}
	}
	return fmt.Errorf("record block span [%d,%d) is not a fenced %q block",
		block.Span.Offset, end, block.Tag)
}
