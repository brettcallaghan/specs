package corpus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// IndexFileName is the fixed name of the index artifact: one file, in the
// configured spec root, that every skill body, CI configuration and hook
// can name without configuration (AC-SPEC-010.1).
//
// It sits beside the documents it describes rather than in a build
// directory because it is a description of them, and discovery reads only
// `.md` files (AC-SPEC-001.6), so it is never mistaken for a corpus
// member.
const IndexFileName = "specs.json"

// StdoutTarget is the output-path value that selects standard output
// instead of a file (AC-SPEC-010.3).
const StdoutTarget = "-"

// tempPattern names the temporary file an atomic replacement writes
// through. It is dot-prefixed and does not end in `.md`, so a concurrent
// build cannot see it as a candidate however briefly it exists.
const tempPattern = ".specs-*.json.tmp"

// IndexPath resolves where one build writes its index.
//
// An empty out is the default: [IndexFileName] inside the configured spec
// root (AC-SPEC-010.1). [StdoutTarget] stays itself, and is the one value
// that names no file at all (AC-SPEC-010.3). Anything else is taken as
// given and used instead of the default (AC-SPEC-010.2) — resolved by the
// caller's working directory, not by the spec root, because an option
// naming `build/spec-index.json` means that path in the repository the
// caller is standing in.
func IndexPath(opts Options, out string) string {
	switch trimmed := strings.TrimSpace(out); trimmed {
	case "":
		return filepath.Join(opts.Root, IndexFileName)
	case StdoutTarget:
		return StdoutTarget
	default:
		return trimmed
	}
}

// Marshal renders an index as the exact bytes the artifact holds:
// two-space-indented JSON with a trailing newline.
//
// Rendering lives here rather than at each call site so that the file, the
// stream written to standard output, and the bytes any test compares are
// the same bytes — which is what makes "identical output whether or not a
// warm index is present" a property of one function rather than a promise
// spread across callers.
func Marshal(index specmodel.SpecIndex) ([]byte, error) {
	encoded, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("corpus: rendering the index: %w", err)
	}
	return append(encoded, '\n'), nil
}

// Write writes an index to dest.
//
// dest of [StdoutTarget] writes to stdout and touches no file — in
// particular it neither creates nor modifies the default artifact
// (AC-SPEC-010.3). Any other dest is a file path: its parent directories
// are created if absent, and an existing file there is replaced
// atomically (AC-SPEC-010.4).
func Write(index specmodel.SpecIndex, dest string, stdout io.Writer) error {
	data, err := Marshal(index)
	if err != nil {
		return err
	}

	if dest == StdoutTarget {
		if stdout == nil {
			return errors.New("corpus: writing the index to standard output: no writer supplied")
		}
		if _, err := stdout.Write(data); err != nil {
			return fmt.Errorf("corpus: writing the index to standard output: %w", err)
		}
		return nil
	}
	return writeAtomic(dest, data)
}

// writeAtomic replaces the file at dest with data, so that a concurrent
// reader observes either the old complete index or the new one and never
// a partially written file (AC-SPEC-010.4).
//
// The mechanism is write-to-temporary-then-rename, in dest's own
// directory so the rename stays within one filesystem and is therefore a
// single atomic operation rather than a copy. Writing in place would be
// visibly wrong here: the index is rebuilt by a write-time hook while
// queries read it, so a truncate-then-write would hand some reader a
// truncated document on any run long enough to interleave.
func writeAtomic(dest string, data []byte) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("corpus: creating %s: %w", dir, err)
	}

	temp, err := os.CreateTemp(dir, tempPattern)
	if err != nil {
		return fmt.Errorf("corpus: creating a temporary file beside %s: %w", dest, err)
	}
	name := temp.Name()
	// A no-op once the rename below has succeeded, and the cleanup that
	// keeps a failed write from leaving debris beside the artifact.
	defer os.Remove(name)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("corpus: writing %s: %w", dest, err)
	}
	// The rename is atomic with respect to other readers of this
	// directory; the sync is what makes the *content* durable before the
	// name points at it, so a crash cannot leave the artifact's name
	// resolving to an empty file.
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("corpus: writing %s: %w", dest, err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("corpus: writing %s: %w", dest, err)
	}
	// os.CreateTemp creates with 0o600; the artifact is published output
	// and is readable by whatever later reads it.
	if err := os.Chmod(name, 0o644); err != nil {
		return fmt.Errorf("corpus: writing %s: %w", dest, err)
	}
	if err := os.Rename(name, dest); err != nil {
		return fmt.Errorf("corpus: replacing %s: %w", dest, err)
	}
	return nil
}

// Load reads a persisted index from path.
//
// It is the inverse of [Write] and performs no validation beyond decoding:
// whether the index may be *used* is [Open]'s question, and it is answered
// by the recorded input hash rather than by anything the file says about
// itself.
func Load(path string) (specmodel.SpecIndex, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return specmodel.SpecIndex{}, fmt.Errorf("corpus: reading the index at %s: %w", path, err)
	}

	var index specmodel.SpecIndex
	if err := json.Unmarshal(raw, &index); err != nil {
		return specmodel.SpecIndex{}, fmt.Errorf("corpus: reading the index at %s: %w", path, err)
	}
	return index, nil
}

// Open returns the corpus for opts, served from the persisted index at
// path when that index records the corpus's current input hash, and built
// by walking the corpus otherwise. warm reports which happened.
//
// This is the whole of the cache decision, and it is deliberately
// one-sided: every condition that is not "this index was built from
// exactly these inputs" rebuilds. An absent file, an unreadable one, one
// written by an older index shape, one carrying no hash, one whose hash
// differs by a byte (AC-SPEC-010.6) — all of them rebuild. A stale answer
// is a silent wrong answer, and rebuilding costs a second at worst.
//
// The two paths are interchangeable by construction: the warm corpus holds
// the very index the cold one wrote, so a caller cannot tell them apart
// from the output (and deleting the artifact is therefore always safe).
func Open(opts Options, path string) (*Corpus, bool, error) {
	if warm := openWarm(opts, path); warm != nil {
		return warm, true, nil
	}
	built, err := Build(opts)
	if err != nil {
		return nil, false, err
	}
	return built, false, nil
}

// openWarm returns the corpus served from the persisted index at path, or
// nil when there is no index there that may be served.
//
// It reports no errors. Every failure it can meet — the file is absent,
// unreadable, not JSON, or names a member that has since gone — is a
// reason to rebuild rather than a reason to fail, and [Build] is where any
// genuine failure to read the corpus surfaces.
func openWarm(opts Options, path string) *Corpus {
	if path == "" || path == StdoutTarget {
		return nil
	}

	index, err := Load(path)
	if err != nil {
		return nil
	}
	if index.InputHash == nil || *index.InputHash == "" {
		return nil
	}
	// A persisted index of a different shape may not be reasoned about by
	// this version's queries, whatever its hash says.
	if index.SchemaVersion != specindex.SchemaVersion {
		return nil
	}

	current, err := Fingerprint(opts)
	if err != nil || current != *index.InputHash {
		return nil
	}

	warm, err := fromIndex(index, opts)
	if err != nil {
		return nil
	}
	return warm
}

// fromIndex hydrates a Corpus from a persisted index, rebuilding the
// lookups and findings a built corpus carries alongside it.
//
// The index itself is taken as read — that is the point of a warm cache —
// but findings are not on the index, so they are recomputed here. That
// costs nothing in the ordinary case: a duplicate-slug finding needs the
// position of the offending slug, so only documents that actually collide
// are read, and a corpus with no collisions reads none.
func fromIndex(index specmodel.SpecIndex, opts Options) (*Corpus, error) {
	specRoot, repoRoot, err := opts.roots()
	if err != nil {
		return nil, err
	}

	// Non-nil collections, matching what specindex.New guarantees, so that
	// a hydrated index renders byte-identically to a built one even when
	// the file it came from was written by something laxer.
	if index.Entries == nil {
		index.Entries = []specmodel.IndexEntry{}
	}
	if index.Exclusions == nil {
		index.Exclusions = []specmodel.IndexExclusion{}
	}

	c := &Corpus{
		repoRoot: repoRoot,
		specRoot: specRoot,
		index:    index,
		byPath:   make(map[string]int, len(index.Entries)),
		bySlug:   make(map[string]int, len(index.Entries)),
		profile:  opts.Profile,
		types:    opts.Profile.DocumentTypes(),
		docs:     map[string]*frontmatter.Document{},
		warm:     true,
	}

	for i, entry := range index.Entries {
		c.byPath[entry.Path] = i
		if entry.Frontmatter == nil {
			continue
		}
		key := strings.ToLower(string(entry.Frontmatter.Slug))
		if key == "" {
			continue
		}
		if _, seen := c.bySlug[key]; !seen {
			c.bySlug[key] = i
		}
	}

	members, err := c.collidingMembers()
	if err != nil {
		return nil, err
	}
	report := finding.NewReport()
	report.Add(duplicateSlugs(members)...)
	c.findings = report.Model().Findings

	return c, nil
}

// collidingMembers returns, in path order, every member of the persisted
// index that shares its slug with another — the only members a
// duplicate-slug finding is raised against, and so the only ones whose
// source has to be read to place that finding.
func (c *Corpus) collidingMembers() ([]member, error) {
	groups := map[string][]int{}
	for i, entry := range c.index.Entries {
		if entry.Frontmatter == nil {
			continue
		}
		slug := string(entry.Frontmatter.Slug)
		if slug == "" {
			continue
		}
		key := strings.ToLower(slug)
		groups[key] = append(groups[key], i)
	}

	var members []member
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		for _, i := range group {
			entry := c.index.Entries[i]
			doc, ok := c.document(entry.Path)
			if !ok {
				return nil, fmt.Errorf(
					"corpus: the persisted index names %s, whose source cannot be read", entry.Path)
			}
			members = append(members, member{
				path:   entry.Path,
				slug:   string(entry.Frontmatter.Slug),
				slugAt: slugSpan(doc),
				doc:    doc,
			})
		}
	}
	sort.Slice(members, func(i, j int) bool { return members[i].path < members[j].path })
	return members, nil
}
