// Property tests over the delimiter scanner.
//
// specs/modules/document-core/document-core-module.md names this the gate for any
// change to Spans: it is "the highest-risk component because every other
// position derives from it", so the obligation is properties that hold
// over generated inputs rather than a handful of examples.
//
// Generation is seeded from a constant, so a failure is reproducible and
// a passing run is not luck. Each property prints the generated input on
// failure — a property test that fails without showing its counterexample
// is a property test nobody can act on.
package span_test

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/core/span"
)

const (
	// iterations is per property. The generators are small and Find is
	// O(n), so this costs single-digit milliseconds.
	iterations = 400
	// seed fixes the generated corpus. Change it deliberately, never to
	// make a failure go away.
	seed = 0x5350414e5f5631 // "SPAN_V1"
)

// delimiterRunes and fillerRunes are disjoint. That is what lets the
// properties below assert exactly where a delimiter occurs: a delimiter
// built from the first set cannot appear by accident inside filler built
// from the second.
//
// The filler carries newlines, tabs, and multibyte codepoints, because
// line and column derivation is precisely what those exercise.
var (
	delimiterRunes = []rune("-<>!+%")
	fillerRunes    = []rune("ab c\n\txyz~é中🙂")
)

func newRand() *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
}

// genDelimiters returns an opening and a closing delimiter that a
// non-nesting scanner can distinguish. The closing delimiter is never a
// substring of the opening one: if it were, an opening delimiter
// appearing inside a region would carry its own terminator, and "the
// first closing delimiter ends the region" would locate a region no
// reader would agree with.
func genDelimiters(rnd *rand.Rand) (opening, closing string) {
	for {
		opening = genRunes(rnd, delimiterRunes, 1, 4)
		closing = genRunes(rnd, delimiterRunes, 1, 4)
		if !strings.Contains(opening, closing) {
			return opening, closing
		}
	}
}

func genFiller(rnd *rand.Rand) string {
	return genRunes(rnd, fillerRunes, 0, 24)
}

func genRunes(rnd *rand.Rand, alphabet []rune, minLen, maxLen int) string {
	n := minLen + rnd.IntN(maxLen-minLen+1)
	var b strings.Builder
	for range n {
		b.WriteRune(alphabet[rnd.IntN(len(alphabet))])
	}
	return b.String()
}

// A region built from an opening delimiter, a body and a closing
// delimiter is recovered exactly: Inner is the body and Outer is the
// whole construct, wherever in the document it sits.
func TestPropertyFindRecoversTheConstructedRegion(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body, suffix := genFiller(rnd), genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body + closing + suffix)

		region, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}
		if got := string(span.Slice(src, region.Inner)); got != body {
			t.Fatalf("case %d: Inner = %q, want %q\n%s", i, got, body, describe(src, opening, closing))
		}
		if got, want := string(span.Slice(src, region.Outer)), opening+body+closing; got != want {
			t.Fatalf("case %d: Outer = %q, want %q\n%s", i, got, want, describe(src, opening, closing))
		}
		if got, want := int(region.Outer.Offset), len(prefix); got != want {
			t.Fatalf("case %d: Outer.Offset = %d, want %d\n%s", i, got, want, describe(src, opening, closing))
		}
		if body == "" && region.Inner.Length != 0 {
			t.Fatalf("case %d: empty body gave Inner.Length = %d\n%s", i, region.Inner.Length, describe(src, opening, closing))
		}
	}
}

// Every span a located region carries addresses the document it came
// from, and the content is contained in the construct that delimits it.
func TestPropertySpansStayWithinTheDocument(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body, suffix := genFiller(rnd), genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body + closing + suffix)

		region, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}

		for name, s := range map[string]span.Span{"Outer": region.Outer, "Inner": region.Inner} {
			if s.Offset+s.Length > uint64(len(src)) {
				t.Fatalf("case %d: %s extends past the document: %d+%d > %d\n%s",
					i, name, s.Offset, s.Length, len(src), describe(src, opening, closing))
			}
		}
		if region.Inner.Offset < region.Outer.Offset ||
			region.Inner.Offset+region.Inner.Length > region.Outer.Offset+region.Outer.Length {
			t.Fatalf("case %d: Inner %+v is not contained in Outer %+v\n%s",
				i, region.Inner, region.Outer, describe(src, opening, closing))
		}
	}
}

// Line and column are derived from the byte offset, so they must agree
// with it — checked against an independent, deliberately naive walk over
// the source rather than against the production computation.
func TestPropertyLineAndColumnAgreeWithTheByteExtent(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body, suffix := genFiller(rnd), genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body + closing + suffix)

		region, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}

		for name, s := range map[string]span.Span{"Outer": region.Outer, "Inner": region.Inner} {
			line, column := referenceLineColumn(src, int(s.Offset))
			if s.StartLine != uint64(line) || s.StartColumn != uint64(column) {
				t.Fatalf("case %d: %s starts at %d:%d, want %d:%d\n%s",
					i, name, s.StartLine, s.StartColumn, line, column, describe(src, opening, closing))
			}
			line, column = referenceLineColumn(src, int(s.Offset+s.Length))
			if s.EndLine != uint64(line) || s.EndColumn != uint64(column) {
				t.Fatalf("case %d: %s ends at %d:%d, want %d:%d\n%s",
					i, name, s.EndLine, s.EndColumn, line, column, describe(src, opening, closing))
			}
		}
	}
}

// An opening delimiter with no closing delimiter after it is not a
// region. The scanner reports none rather than inventing one that runs
// to the end of the document.
func TestPropertyUnterminatedRegionIsNotARegion(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body := genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body)

		if region, ok := span.Find(src, opening, closing, 0); ok {
			t.Fatalf("case %d: Find located %+v in an unterminated document\n%s",
				i, region, describe(src, opening, closing))
		}
	}
}

// A document with no opening delimiter holds no region, whatever else is
// in it.
func TestPropertyAbsentOpeningDelimiterIsNotARegion(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		src := []byte(genFiller(rnd) + genFiller(rnd))

		if region, ok := span.Find(src, opening, closing, 0); ok {
			t.Fatalf("case %d: Find located %+v where no opening delimiter occurs\n%s",
				i, region, describe(src, opening, closing))
		}
	}
}

// The scanner does not nest: an opening delimiter inside a region is
// content, and the first closing delimiter after the outermost opening
// one ends the region.
func TestPropertyNestedOpeningDelimiterIsContent(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, head, tail, suffix := genFiller(rnd), genFiller(rnd), genFiller(rnd), genFiller(rnd)
		body := head + opening + tail
		src := []byte(prefix + opening + body + closing + suffix)

		region, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}
		if got := string(span.Slice(src, region.Inner)); got != body {
			t.Fatalf("case %d: Inner = %q, want %q — a nested opening delimiter is content\n%s",
				i, got, body, describe(src, opening, closing))
		}
	}
}

// A trailing closing delimiter is content of no region: the first one
// ends the region, and the second is left in the suffix.
func TestPropertyOnlyTheFirstClosingDelimiterEndsTheRegion(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body, tail := genFiller(rnd), genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body + closing + tail + closing)

		region, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}
		if got := string(span.Slice(src, region.Inner)); got != body {
			t.Fatalf("case %d: Inner = %q, want %q\n%s", i, got, body, describe(src, opening, closing))
		}
	}
}

// Scanning from the end of one region finds the next, in document order,
// and stops once the document holds no more.
func TestPropertySequentialScanFindsEveryRegionInOrder(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		count := 1 + rnd.IntN(4)

		var built strings.Builder
		bodies := make([]string, count)
		for r := range count {
			bodies[r] = genFiller(rnd)
			built.WriteString(genFiller(rnd))
			built.WriteString(opening)
			built.WriteString(bodies[r])
			built.WriteString(closing)
		}
		built.WriteString(genFiller(rnd))
		src := []byte(built.String())

		at := 0
		for r := range count {
			region, ok := span.Find(src, opening, closing, at)
			if !ok {
				t.Fatalf("case %d: no region %d of %d\n%s", i, r+1, count, describe(src, opening, closing))
			}
			if got := string(span.Slice(src, region.Inner)); got != bodies[r] {
				t.Fatalf("case %d: region %d Inner = %q, want %q\n%s",
					i, r+1, got, bodies[r], describe(src, opening, closing))
			}
			next := int(region.Outer.Offset + region.Outer.Length)
			if next <= at && r > 0 {
				t.Fatalf("case %d: scan did not advance past %d\n%s", i, at, describe(src, opening, closing))
			}
			at = next
		}
		if region, ok := span.Find(src, opening, closing, at); ok {
			t.Fatalf("case %d: found an extra region %+v after %d of %d\n%s",
				i, region, count, count, describe(src, opening, closing))
		}
	}
}

// Starting the scan anywhere at or before the region finds the same
// region, so a caller does not have to guess a start offset.
func TestPropertyAnyStartAtOrBeforeTheRegionFindsIt(t *testing.T) {
	rnd := newRand()

	for i := range iterations {
		opening, closing := genDelimiters(rnd)
		prefix, body, suffix := genFiller(rnd), genFiller(rnd), genFiller(rnd)
		src := []byte(prefix + opening + body + closing + suffix)

		want, ok := span.Find(src, opening, closing, 0)
		if !ok {
			t.Fatalf("case %d: Find reported no region\n%s", i, describe(src, opening, closing))
		}
		for _, from := range []int{-1, 0, len(prefix) / 2, len(prefix)} {
			got, ok := span.Find(src, opening, closing, from)
			if !ok || got != want {
				t.Fatalf("case %d: Find from %d gave (%+v, %t), want (%+v, true)\n%s",
					i, from, got, ok, want, describe(src, opening, closing))
			}
		}
	}
}

// Over arbitrary bytes and arbitrary delimiters drawn from the same
// alphabet — the adversarial case the generators above deliberately rule
// out — the scanner still either reports no region or reports one whose
// structure holds: it begins at the first opening delimiter at or after
// the start, opens with it, and ends with the closing one.
func TestPropertyStructureHoldsOverArbitraryInput(t *testing.T) {
	rnd := newRand()
	alphabet := []rune("ab-<>\n")

	for i := range iterations {
		src := []byte(genRunes(rnd, alphabet, 0, 40))
		opening := genRunes(rnd, alphabet, 1, 3)
		closing := genRunes(rnd, alphabet, 1, 3)
		from := rnd.IntN(len(src) + 2)

		region, ok := span.Find(src, opening, closing, from)
		if !ok {
			continue
		}

		outer := string(span.Slice(src, region.Outer))
		if !strings.HasPrefix(outer, opening) {
			t.Fatalf("case %d: Outer %q does not open with %q\n%s", i, outer, opening, describe(src, opening, closing))
		}
		if !strings.HasSuffix(outer, closing) {
			t.Fatalf("case %d: Outer %q does not close with %q\n%s", i, outer, closing, describe(src, opening, closing))
		}
		if got, want := string(span.Slice(src, region.Inner)),
			outer[len(opening):len(outer)-len(closing)]; got != want {
			t.Fatalf("case %d: Inner = %q, want %q\n%s", i, got, want, describe(src, opening, closing))
		}
		if got, want := int(region.Outer.Offset),
			from+strings.Index(string(src[from:]), opening); got != want {
			t.Fatalf("case %d: region opens at %d, but the first opening delimiter at or after %d is at %d\n%s",
				i, got, from, want, describe(src, opening, closing))
		}
		if region.Outer.Offset+region.Outer.Length > uint64(len(src)) {
			t.Fatalf("case %d: Outer %+v extends past the document\n%s", i, region.Outer, describe(src, opening, closing))
		}
	}
}

// referenceLineColumn is an independent, deliberately naive derivation of
// a one-based line and codepoint column, written differently from the
// implementation under test so that the two agreeing means something.
func referenceLineColumn(src []byte, offset int) (line, column int) {
	line, column = 1, 1
	for _, r := range string(src[:offset]) {
		if r == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return line, column
}

func describe(src []byte, opening, closing string) string {
	return fmt.Sprintf("  src      = %q\n  opening  = %q\n  closing  = %q", src, opening, closing)
}
