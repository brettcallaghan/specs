package markdown

import "strings"

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

// MatchesPath reports whether query selects the heading whose heading
// path is path.
//
// query is one or more segments separated by [PathSeparator], each
// compared on its identifier. The last names the heading itself; every
// earlier one names a heading enclosing it, outermost first, at any depth
// — so an intermediate heading may be left out, and the full path pastes
// back as a query selecting the same heading. A single segment is exactly
// an identifier comparison. A query holding an empty segment matches
// nothing: "A##B" and "A#" name no heading.
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
