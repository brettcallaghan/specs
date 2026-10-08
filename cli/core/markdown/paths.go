package markdown

import (
	"slices"
	"strings"
)

// Outline is what a heading path is derived from: one heading's text as
// authored and its level, in document order.
//
// It is the two properties every record of a heading carries — a parsed
// [Heading] and an indexed one alike — so a caller holding either derives
// paths without this package depending on the other's shape.
type Outline struct {
	Text  string
	Level int
}

// PathSeparator joins the segments of a heading path, and separates the
// segments of a selection or anchor written as one.
const PathSeparator = "#"

// HeadingPaths returns, for each heading of outline, its heading path:
// the identifiers of its enclosing headings from the outermost down,
// followed by its own.
//
// A heading encloses every heading after it up to the next one at its own
// level or higher — the same nesting a heading's section follows — so a
// level skipped between a heading and one nested under it changes
// nothing, and a document with no top-level heading has paths that begin
// at its shallowest.
func HeadingPaths(outline []Outline) [][]string {
	paths := make([][]string, len(outline))
	var enclosing []int
	for i, heading := range outline {
		for len(enclosing) > 0 && outline[enclosing[len(enclosing)-1]].Level >= heading.Level {
			enclosing = enclosing[:len(enclosing)-1]
		}
		path := make([]string, 0, len(enclosing)+1)
		for _, at := range enclosing {
			path = append(path, Identifier(outline[at].Text))
		}
		paths[i] = append(path, Identifier(heading.Text))
		enclosing = append(enclosing, i)
	}
	return paths
}

// JoinPath writes a heading path as one string, its segments joined by
// [PathSeparator].
func JoinPath(path []string) string {
	return strings.Join(path, PathSeparator)
}

// MatchesPath reports whether query satisfies the heading whose heading
// path is path, taken alone.
//
// query is one or more segments separated by [PathSeparator], each
// compared on its identifier. The last names the heading itself; every
// earlier one names a heading enclosing it, outermost first, at any depth
// — so an intermediate heading may be left out. A single segment is
// exactly an identifier comparison. A query holding an empty segment
// matches nothing: "A##B" and "A#" name no heading.
//
// Satisfying a query is not the same as being selected by it: a shallow
// heading's full path can be satisfied by a deeper same-named heading
// too. [SelectPaths] decides selection over a whole outline, and is what
// a caller answering a query uses.
func MatchesPath(query string, path []string) bool {
	if len(path) == 0 {
		return false
	}
	segments := strings.Split(query, PathSeparator)
	if len(segments) == 1 {
		return Identifier(query) == path[len(path)-1]
	}
	for _, segment := range segments {
		if strings.TrimSpace(segment) == "" {
			return false
		}
	}
	if Identifier(segments[len(segments)-1]) != path[len(path)-1] {
		return false
	}
	ancestors := path[:len(path)-1]
	for _, segment := range segments[:len(segments)-1] {
		wanted := Identifier(segment)
		for len(ancestors) > 0 && ancestors[0] != wanted {
			ancestors = ancestors[1:]
		}
		if len(ancestors) == 0 {
			return false
		}
		ancestors = ancestors[1:]
	}
	return true
}

// SelectPaths reports, for each heading path of one document's outline,
// whether query selects that heading.
//
// A query of more than one segment that equals some heading's full
// heading path, segment for segment on identifiers, selects exactly the
// headings with that full path — so a path the CLI prints pastes back as
// a query naming that heading alone, unless two sibling headings share
// identical text all the way up. Any other query selects every heading
// [MatchesPath] says it satisfies.
//
// A single segment is never narrowed this way: it stays an identifier
// comparison over every heading, so a top-level heading's one-segment
// path selects each same-named heading of the document, as every
// selection and anchor written before paths existed always did.
func SelectPaths(query string, paths [][]string) []bool {
	selected := make([]bool, len(paths))
	if segments := strings.Split(query, PathSeparator); len(segments) > 1 {
		wanted := make([]string, len(segments))
		for i, segment := range segments {
			wanted[i] = Identifier(segment)
		}
		exact := false
		for i, path := range paths {
			if slices.Equal(path, wanted) && MatchesPath(query, path) {
				selected[i] = true
				exact = true
			}
		}
		if exact {
			return selected
		}
	}
	for i, path := range paths {
		selected[i] = MatchesPath(query, path)
	}
	return selected
}
