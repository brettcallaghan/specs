package corpus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/span"
	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// candidateExtension is the only file extension discovery considers.
//
// Matching is case-insensitive so that a corpus checked out on a
// case-preserving filesystem yields the same membership as the same
// corpus on a case-sensitive one. Every other file — an emitted
// specs.json beside the documents it describes, a README.rst, an image
// — is passed over silently rather than recorded as an exclusion
// (AC-SPEC-001.6): an exclusion is a statement that a file looked like a
// specification and was not one, and only a markdown file can make that
// claim.
const candidateExtension = ".md"

// typeProperty is the frontmatter property whose value selects the
// document type. The host knows this property's *name* and nothing about
// its values: which values are legal is entirely [Profile.DocumentTypes].
const typeProperty = "type"

// slugProperty is the frontmatter property carrying the identifier every
// other document references a member by, and the one duplicate detection
// compares.
const slugProperty = "slug"

// Options configures one corpus build.
type Options struct {
	// RepoRoot is the directory every path this build emits is reported
	// relative to — in the index and in findings alike (AC-SPEC-009.8).
	//
	// It is the directory holding the repository's configuration file,
	// not the spec root: an index that reported paths relative to the
	// spec root would name files that no other tool in the repository
	// could resolve. Defaults to Root when empty.
	RepoRoot string

	// Root is the spec root: the one directory discovery walks, and the
	// boundary no read crosses (AC-SPEC-001.5).
	Root string

	// Workspace is the workspace identifier from configuration, declared
	// once on the index (AC-SPEC-002.3) so that indexes emitted by
	// several repositories merge into one graph without two documents
	// that both declare `adr-001` colliding.
	Workspace string

	// Profile is the active profile: the whole of this build's knowledge
	// of which document types exist.
	Profile Profile
}

// Corpus is one built corpus: the index of every document beneath the
// spec root, and the findings discovery raised over it.
//
// It is built once and read many times. Nothing about it is lazy — a
// query answers from memory — because the queries it serves are asked
// one per process invocation and the walk that fills it is the expensive
// part (AC-SPEC-009.3).
//
// Persisting this index and serving it warm from a content hash are a
// separate concern and live elsewhere; nothing here writes to the
// filesystem.
type Corpus struct {
	repoRoot string
	index    specmodel.SpecIndex
	findings []specmodel.Finding
	byPath   map[string]int
	bySlug   map[string]int

	// profile is the active profile the corpus was built under, retained
	// so that Check can reach its schema and per-type rules without the
	// caller supplying them again — and without this package importing
	// anything but the Profile extension point it already depends on.
	profile Profile
	// types is opts.Profile.DocumentTypes(), retained for the same reason:
	// Check needs each member's slug pattern, and Build has already read
	// it once.
	types map[string]DocumentTypeRules
	// docs is the parsed frontmatter document for every member whose
	// frontmatter decoded into the envelope, keyed by reported path. Check
	// reads from here rather than re-parsing (AC-SPEC-003 at corpus
	// level); a member with a parse error carries no entry.
	docs map[string]*frontmatter.Document

	// registryState is Check's lazily-compiled Schema Registry. It is
	// built at most once regardless of how many times Check or CheckFor is
	// called.
	registryState registryState
}

// member is one corpus member during a build, carrying the facts
// duplicate detection and later checking need and the index does not
// model: the slug as the document declared it, where that slug sits in
// the file, and the parsed document itself.
type member struct {
	path   string
	slug   string
	slugAt specmodel.Span
	doc    *frontmatter.Document
}

// Build walks the configured spec root, decides membership from
// frontmatter and declared type, and assembles the index every query
// reads.
//
// It orchestrates the parsers and holds no parsing logic of its own:
// frontmatter extraction is [frontmatter.Parse]'s, positions are
// [span]'s, and the vocabulary of document types is the profile's. What
// belongs here is the walk, the membership decision, and the ordering
// that makes the result identical on every platform.
//
// A file that cannot be read or whose frontmatter cannot be parsed is
// recorded against itself and the rest of the corpus is still indexed
// (AC-SPEC-002.4); Build returns an error only when the build as a whole
// cannot start — no profile, no spec root, or a spec root that cannot be
// opened.
func Build(opts Options) (*Corpus, error) {
	if opts.Profile == nil {
		return nil, errors.New("corpus: no profile supplied")
	}
	if strings.TrimSpace(opts.Root) == "" {
		return nil, errors.New("corpus: no spec root configured")
	}

	specRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, fmt.Errorf("corpus: resolving the spec root: %w", err)
	}
	repoRoot := opts.RepoRoot
	if strings.TrimSpace(repoRoot) == "" {
		repoRoot = specRoot
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("corpus: resolving the repository root: %w", err)
	}

	// Every read of a candidate goes through this handle rather than
	// through a bare os.ReadFile, so that a path which resolves outside
	// the spec root is refused by the operating system rather than by a
	// check this package would have to get right (AC-SPEC-001.5).
	root, err := os.OpenRoot(specRoot)
	if err != nil {
		return nil, fmt.Errorf("corpus: opening the spec root: %w", err)
	}
	defer root.Close()

	b := &builder{
		root:      root,
		specRoot:  specRoot,
		repoRoot:  repoRoot,
		profile:   opts.Profile,
		types:     opts.Profile.DocumentTypes(),
		index:     specindex.New(opts.Workspace),
		byPath:    map[string]int{},
		alsoKnown: map[string]int{},
		docs:      map[string]*frontmatter.Document{},
	}

	if err := filepath.WalkDir(specRoot, b.visit); err != nil {
		return nil, fmt.Errorf("corpus: walking the spec root: %w", err)
	}

	return b.finish(), nil
}

// builder is the mutable state of one build.
type builder struct {
	root     *os.Root
	specRoot string
	repoRoot string
	profile  Profile
	types    map[string]DocumentTypeRules

	index   specmodel.SpecIndex
	members []member

	byPath map[string]int
	// alsoKnown maps a lowercased slug to the index of the first entry
	// declaring it, which is what makes the slug query case-insensitive
	// in the same way duplicate detection is (AC-SPEC-003.4).
	alsoKnown map[string]int
	// docs is every member's parsed frontmatter document, keyed by
	// reported path, carried onto the built Corpus for Check to read.
	docs map[string]*frontmatter.Document
}

// visit classifies one entry of the walk.
//
// Only regular files are candidates. A symlink is deliberately not one:
// filepath.WalkDir does not descend into a symlinked directory, so
// refusing symlinked files as well means the walk has no way at all to
// leave the spec root (AC-SPEC-001.5).
func (b *builder) visit(osPath string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if d.IsDir() || !d.Type().IsRegular() {
		return nil
	}
	if !strings.EqualFold(filepath.Ext(d.Name()), candidateExtension) {
		return nil
	}

	rel, err := filepath.Rel(b.specRoot, osPath)
	if err != nil {
		return err
	}
	reported, err := b.reportedPath(osPath)
	if err != nil {
		return err
	}

	src, err := b.read(rel)
	if err != nil {
		// An unreadable candidate is recorded against itself for the same
		// reason an unparsable one is: losing a file silently is worse
		// than an index that says which file it could not read.
		b.addEntry(specmodel.IndexEntry{
			Path:       reported,
			ParseError: &specmodel.ParseError{Message: err.Error(), Position: span.Of(nil, 0, 0)},
		}, member{})
		return nil
	}

	b.classify(reported, src)
	return nil
}

// classify decides what one candidate file is, and records it.
func (b *builder) classify(reported string, src []byte) {
	doc, perr := frontmatter.Parse(src)

	switch {
	case perr != nil:
		// The block is there and could not be read. That is a member with
		// a recorded failure, not an exclusion: the file claimed to be a
		// specification and the claim could not be checked
		// (AC-SPEC-002.4).
		b.addEntry(specmodel.IndexEntry{Path: reported, ParseError: perr}, member{})

	case doc == nil:
		// No frontmatter block at all (AC-SPEC-001.3).
		b.index.Exclusions = append(b.index.Exclusions, specmodel.IndexExclusion{
			Path:   reported,
			Reason: specmodel.ExclusionReasonNoFrontmatter,
		})

	default:
		b.classifyDocument(reported, doc)
	}
}

// classifyDocument decides membership for a file whose frontmatter
// parsed, from the type it declares.
func (b *builder) classifyDocument(reported string, doc *frontmatter.Document) {
	declared, declaredOK := declaredType(doc)
	if _, known := b.types[declared]; !known || !declaredOK {
		// A type the active profile does not define, or no type at all.
		// Both are `unknown-type`, and the declared value is reported
		// where there is one, so a reviewing agent can judge whether the
		// profile or the document is wrong (AC-SPEC-001.4).
		exclusion := specmodel.IndexExclusion{
			Path:   reported,
			Reason: specmodel.ExclusionReasonUnknownType,
		}
		if declaredOK {
			value := declared
			exclusion.Type = &value
		}
		b.index.Exclusions = append(b.index.Exclusions, exclusion)
		return
	}

	envelope, perr := decodeFrontmatter(doc)
	if perr != nil {
		// The block parsed as YAML and does not fit the frontmatter
		// envelope — a status outside the vocabulary, a timestamp that is
		// not one. Recorded against this member so the rest of the corpus
		// still indexes (AC-SPEC-002.4).
		b.addEntry(specmodel.IndexEntry{Path: reported, ParseError: perr}, member{})
		return
	}

	b.docs[reported] = doc
	b.addEntry(
		specmodel.IndexEntry{Path: reported, Frontmatter: envelope},
		member{path: reported, slug: string(envelope.Slug), slugAt: slugSpan(doc), doc: doc},
	)
}

// addEntry records one entry, and the member facts duplicate detection
// will need. A zero member — an entry with a parse failure — declares no
// slug and takes part in no slug comparison.
func (b *builder) addEntry(entry specmodel.IndexEntry, m member) {
	b.index.Entries = append(b.index.Entries, entry)
	if m.slug != "" {
		b.members = append(b.members, m)
	}
}

// read returns the bytes of one candidate, named relative to the spec
// root so the read cannot address a file outside it.
func (b *builder) read(rel string) ([]byte, error) {
	f, err := b.root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// reportedPath converts an absolute filesystem path to the form every
// index entry, exclusion and finding uses: relative to the repository
// root and forward-slashed on every platform (AC-SPEC-009.8).
func (b *builder) reportedPath(osPath string) (string, error) {
	rel, err := filepath.Rel(b.repoRoot, osPath)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// finish orders everything the walk collected and assembles the Corpus.
//
// Ordering is this component's responsibility rather than the caller's.
// Directory iteration order is a filesystem property, not a corpus
// property, so entries and exclusions are sorted by path and findings
// through the one authority for finding order — [finding.Report] — so
// that the same corpus content yields identical output wherever it is
// built.
func (b *builder) finish() *Corpus {
	sort.Slice(b.index.Entries, func(i, j int) bool {
		return b.index.Entries[i].Path < b.index.Entries[j].Path
	})
	sort.Slice(b.index.Exclusions, func(i, j int) bool {
		return b.index.Exclusions[i].Path < b.index.Exclusions[j].Path
	})
	sort.Slice(b.members, func(i, j int) bool {
		return b.members[i].path < b.members[j].path
	})

	for i, entry := range b.index.Entries {
		b.byPath[entry.Path] = i
		if entry.Frontmatter == nil {
			continue
		}
		key := strings.ToLower(string(entry.Frontmatter.Slug))
		if key == "" {
			continue
		}
		if _, seen := b.alsoKnown[key]; !seen {
			b.alsoKnown[key] = i
		}
	}

	report := finding.NewReport()
	report.Add(duplicateSlugs(b.members)...)

	return &Corpus{
		repoRoot: b.repoRoot,
		index:    b.index,
		findings: report.Model().Findings,
		byPath:   b.byPath,
		bySlug:   b.alsoKnown,
		profile:  b.profile,
		types:    b.types,
		docs:     b.docs,
	}
}

// duplicateSlugs reports every collision between two members' slugs,
// compared without regard to letter case (AC-SPEC-003.4).
//
// One finding is raised against each member of a colliding group naming
// each other member of it (AC-SPEC-003.3), so that checking any one of
// them surfaces the collision. For the two-document case the acceptance
// criterion describes, that is exactly two findings; for a larger group
// it is every ordered pair, because a member that named only one of its
// several collisions would leave the others invisible to anyone checking
// that file.
//
// members must already be ordered by path, which is what makes the
// findings' order a function of corpus content alone.
func duplicateSlugs(members []member) []specmodel.Finding {
	groups := map[string][]int{}
	keys := make([]string, 0, len(members))
	for i, m := range members {
		key := strings.ToLower(m.slug)
		if _, seen := groups[key]; !seen {
			keys = append(keys, key)
		}
		groups[key] = append(groups[key], i)
	}
	sort.Strings(keys)

	var findings []specmodel.Finding
	for _, key := range keys {
		group := groups[key]
		if len(group) < 2 {
			continue
		}
		for _, i := range group {
			for _, j := range group {
				if i == j {
					continue
				}
				findings = append(findings, finding.DuplicateSlug(
					members[i].path, members[i].slugAt, members[i].slug, members[j].path))
			}
		}
	}
	return findings
}

// declaredType returns the value a document declares in its type
// property, and whether it declared one at all.
//
// A non-string value is reported as its source text rather than as Go's
// rendering of whatever the YAML parser inferred, so `type: 3` is
// excluded reporting `3` and the author can see what they wrote.
func declaredType(doc *frontmatter.Document) (string, bool) {
	property, ok := doc.Property(typeProperty)
	if !ok {
		return "", false
	}
	if value, isString := property.Value.(string); isString {
		return value, true
	}
	return property.Raw, true
}

// slugSpan locates a document's slug property, falling back to the
// frontmatter block itself for a document that declares no slug — every
// position this framework reports is a span, so there is no case in
// which a finding has none.
func slugSpan(doc *frontmatter.Document) specmodel.Span {
	if property, ok := doc.Property(slugProperty); ok {
		return property.Position
	}
	return doc.Region.Inner
}

// decodeFrontmatter converts a parsed block into the typed envelope an
// index entry carries (AC-SPEC-002.1).
//
// The conversion runs through JSON because the generated records carry
// json tags and no YAML ones. A value of the wrong shape for the
// envelope is a located failure rather than an error the caller raises:
// it is recorded against its own document and the walk continues.
func decodeFrontmatter(doc *frontmatter.Document) (*specmodel.DocumentFrontmatter, *specmodel.ParseError) {
	raw, err := doc.JSON()
	if err != nil {
		return nil, &specmodel.ParseError{
			Message:  fmt.Sprintf("the frontmatter block could not be read: %s", err),
			Position: doc.Region.Inner,
		}
	}

	var envelope specmodel.DocumentFrontmatter
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, &specmodel.ParseError{
			Message:  fmt.Sprintf("the frontmatter block does not fit the document envelope: %s", err),
			Position: envelopeErrorSpan(doc, err),
		}
	}
	return &envelope, nil
}

// envelopeErrorSpan points a decode failure at the offending property
// where the decoder names one, and at the whole block where it does not.
func envelopeErrorSpan(doc *frontmatter.Document, err error) specmodel.Span {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		name, _, _ := strings.Cut(typeErr.Field, ".")
		if property, ok := doc.Property(name); ok {
			return property.Position
		}
	}
	return doc.Region.Inner
}

// Index returns the built index.
//
// Its collections are the corpus's own and are already ordered; a caller
// that modifies them corrupts every later read, so callers treat the
// result as read-only.
func (c *Corpus) Index() specmodel.SpecIndex {
	return c.index
}

// Findings returns every finding discovery raised over the whole corpus,
// in the deterministic order [finding.Report] fixes.
func (c *Corpus) Findings() []specmodel.Finding {
	return c.findings
}

// FindingsFor returns the findings raised against one file.
//
// This is how a single-document check is served: the corpus is built
// from the whole spec root and then narrowed, rather than the named file
// being read alone, because a defect like a duplicate slug is a property
// of the corpus and is invisible to a document read in isolation
// (AC-SPEC-001.2).
//
// path may be absolute or repository-relative, and may use either
// platform's separator.
func (c *Corpus) FindingsFor(path string) []specmodel.Finding {
	wanted := c.normalize(path)
	var found []specmodel.Finding
	for _, f := range c.findings {
		if f.Path == wanted {
			found = append(found, f)
		}
	}
	return found
}

// Entry returns the index entry for one file.
func (c *Corpus) Entry(path string) (specmodel.IndexEntry, bool) {
	i, ok := c.byPath[c.normalize(path)]
	if !ok {
		return specmodel.IndexEntry{}, false
	}
	return c.index.Entries[i], true
}

// Document returns the parsed frontmatter document for one corpus member,
// so that a caller building further checks over it — the Schema Registry,
// a slug-pattern check — reads the same parse Build already performed
// rather than re-parsing the file. Present exactly when [Corpus.Entry]
// reports a non-nil Frontmatter for the same path; a member with a parse
// error carries neither.
func (c *Corpus) Document(path string) (*frontmatter.Document, bool) {
	doc, ok := c.docs[c.normalize(path)]
	return doc, ok
}

// Slug reports whether a corpus member declares the given slug and, if
// one does, returns its entry — the query AC-SPEC-007.2 defines.
//
// The comparison ignores letter case, the same way duplicate detection
// does: a corpus cannot hold `frd-checkout` and `FRD-Checkout` as
// distinct identifiers, so a query for either must find the other.
// Where two members do collide, the entry returned is the first in path
// order, and both are reported as findings.
func (c *Corpus) Slug(slug string) (specmodel.IndexEntry, bool) {
	i, ok := c.bySlug[strings.ToLower(strings.TrimSpace(slug))]
	if !ok {
		return specmodel.IndexEntry{}, false
	}
	return c.index.Entries[i], true
}

// normalize converts a caller's path to the reported form: relative to
// the repository root, forward-slashed, and cleaned.
func (c *Corpus) normalize(p string) string {
	if filepath.IsAbs(p) {
		if rel, err := filepath.Rel(c.repoRoot, p); err == nil {
			p = rel
		}
	}
	return path.Clean(filepath.ToSlash(p))
}
