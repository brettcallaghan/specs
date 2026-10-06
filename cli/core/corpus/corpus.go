package corpus

import (
	"bytes"
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
	"sync"

	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/markdown"
	"github.com/brettcallaghan/agent-plugins/cli/core/references"
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
// It is built once and read many times. Nothing about a built corpus is
// lazy — a query answers from memory — because the queries it serves are
// asked one per process invocation and the walk that fills it is the
// expensive part (AC-SPEC-009.3).
//
// A corpus served warm from a persisted index (see [Open]) is the one
// exception: it holds the same index and answers the same queries, but
// it has parsed no document, so it reads and parses one on demand when a
// caller asks for something only the source can answer — a slug's
// position, or a schema check. That is what makes a single-document
// check against a warm index cost one file rather than the corpus
// (AC-SPEC-009.2).
type Corpus struct {
	repoRoot string
	// specRoot is retained so that a corpus served warm can read one
	// member's source on demand, through the same containment guarantee
	// [Build] reads under.
	specRoot string
	index    specmodel.SpecIndex
	findings []specmodel.Finding
	// failures is every Gherkin source the walk could not parse. It is not
	// on the index — a persisted index records what was extracted, not
	// what would not parse — so a corpus served warm carries none.
	failures []SourceFailure
	byPath   map[string]int
	bySlug   map[string]int

	// outbound and inbound are the reference graph's two directions, each
	// holding positions into index.References keyed by a document's path:
	// outbound by the path a reference occurs in, inbound by the path it
	// resolved to. Both are built once, by [Corpus.indexReferences], from
	// the one collection the index carries.
	outbound map[string][]int
	inbound  map[string][]int

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
	//
	// A warm corpus fills it on demand rather than at build time, which is
	// the one thing about a Corpus that is written after construction —
	// hence the mutex, so that two callers reading the same warm corpus
	// cannot race on it.
	docs   map[string]*frontmatter.Document
	docsMu sync.Mutex

	// registryState is Check's lazily-compiled Schema Registry. It is
	// built at most once regardless of how many times Check or CheckFor is
	// called.
	registryState registryState

	// warm reports whether this corpus was served from a persisted index
	// rather than built by walking the corpus, which is the one condition
	// under which docs is filled on demand rather than by the walk.
	warm bool
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
	specRoot, repoRoot, err := opts.roots()
	if err != nil {
		return nil, err
	}

	b := &builder{
		specRoot:   specRoot,
		repoRoot:   repoRoot,
		profile:    opts.Profile,
		types:      opts.Profile.DocumentTypes(),
		vocabulary: newVocabulary(opts.Profile.FeatureTags()),
		index:      specindex.New(opts.Workspace),
		byPath:     map[string]int{},
		alsoKnown:  map[string]int{},
		docs:       map[string]*frontmatter.Document{},
	}

	if err := scanSources(specRoot, repoRoot, b.visit); err != nil {
		return nil, err
	}

	return b.finish(opts), nil
}

// roots resolves one operation's spec root and repository root to
// absolute paths, and refuses an operation that cannot be run at all.
//
// A build with no repository root reports paths relative to the spec
// root, which is the only defensible fallback: an index has to name its
// members relative to *something*, and the spec root is the one directory
// every operation has.
//
// It is a method on Options rather than a step inside Build because
// [Fingerprint] must resolve the same inputs the same way — a fingerprint
// taken over a differently-resolved root would compare two different
// corpora.
func (opts Options) roots() (specRoot, repoRoot string, err error) {
	if opts.Profile == nil {
		return "", "", errors.New("corpus: no profile supplied")
	}
	if strings.TrimSpace(opts.Root) == "" {
		return "", "", errors.New("corpus: no spec root configured")
	}

	specRoot, err = filepath.Abs(opts.Root)
	if err != nil {
		return "", "", fmt.Errorf("corpus: resolving the spec root: %w", err)
	}
	repoRoot = opts.RepoRoot
	if strings.TrimSpace(repoRoot) == "" {
		repoRoot = specRoot
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		return "", "", fmt.Errorf("corpus: resolving the repository root: %w", err)
	}
	return specRoot, repoRoot, nil
}

// sourceKind is what one file the walk offers is: a candidate the
// membership decision is made over, or a feature source.
//
// The two are offered by one walk rather than by two passes over the same
// tree, which is what makes a feature source's bytes reach the input hash
// on the same terms a candidate's do. A hash covering only candidates
// would serve a stale index to every query over features while a
// `.feature` file was edited under it — a silent wrong answer rather than
// a visible failure.
type sourceKind int

const (
	// markdownCandidate is a `.md` file: a candidate for corpus
	// membership, admitted or excluded by what its frontmatter declares.
	markdownCandidate sourceKind = iota

	// featureSource is a `.feature` file: a Gherkin source that is no
	// document and no candidate, and produces neither an entry nor an
	// exclusion.
	featureSource
)

// scanSources walks specRoot and offers every source file to visit,
// naming it in the form the index reports — relative to repoRoot and
// forward-slashed on every platform (AC-SPEC-009.8).
//
// This is the only walk in the package. [Build] consumes it to assemble
// the index and [Fingerprint] consumes it to hash the same bytes without
// parsing any of them, so the two can never disagree about which files
// the corpus is made of — the property the whole warm-cache path rests
// on.
//
// A file that cannot be read is offered with readErr set and nil bytes
// rather than aborting the walk: losing a file silently is worse than
// recording that it could not be read.
func scanSources(specRoot, repoRoot string, visit func(kind sourceKind, reported string, src []byte, readErr error)) error {
	// Every read of a candidate goes through this handle rather than
	// through a bare os.ReadFile, so that a path which resolves outside
	// the spec root is refused by the operating system rather than by a
	// check this package would have to get right (AC-SPEC-001.5).
	root, err := os.OpenRoot(specRoot)
	if err != nil {
		return fmt.Errorf("corpus: opening the spec root: %w", err)
	}
	defer root.Close()

	walk := func(osPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Only regular files are candidates. A symlink is deliberately
		// not one: filepath.WalkDir does not descend into a symlinked
		// directory, so refusing symlinked files as well means the walk
		// has no way at all to leave the spec root (AC-SPEC-001.5).
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		var kind sourceKind
		switch extension := filepath.Ext(d.Name()); {
		case strings.EqualFold(extension, candidateExtension):
			kind = markdownCandidate
		case strings.EqualFold(extension, featureExtension):
			kind = featureSource
		default:
			return nil
		}

		rel, err := filepath.Rel(specRoot, osPath)
		if err != nil {
			return err
		}
		reported, err := filepath.Rel(repoRoot, osPath)
		if err != nil {
			return err
		}

		// Both kinds are read through the one handle and the one line-ending
		// normalization every other read passes through, so a feature source
		// read from a CRLF checkout hashes and spans as it does anywhere else.
		src, readErr := readUnder(root, rel)
		visit(kind, filepath.ToSlash(reported), src, readErr)
		return nil
	}

	if err := filepath.WalkDir(specRoot, walk); err != nil {
		return fmt.Errorf("corpus: walking the spec root: %w", err)
	}
	return nil
}

// readUnder returns the bytes of one file named relative to an open root,
// so the read cannot address a file outside it.
//
// The bytes are returned with line endings normalized by
// [normalizeLineEndings], which is what makes this the one place a
// CRLF checkout stops being visible to the rest of the package.
func readUnder(root *os.Root, rel string) ([]byte, error) {
	f, err := root.Open(rel)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return normalizeLineEndings(src), nil
}

// normalizeLineEndings rewrites every CRLF in src to a bare LF.
//
// This is what AC-SPEC-009.7 rests on. Git converts LF to CRLF in the
// working tree on Windows (`core.autocrlf`), so the *same commit* is a
// different sequence of bytes there. Without this, every byte offset in
// a span, and the content hash the warm-cache path compares, would move
// with the checkout's line endings — and the same corpus would produce
// different output on Windows than on Linux and macOS, which is exactly
// what AC-SPEC-009.7 forbids.
//
// It is applied here, in [readUnder], rather than in each parser,
// because [scanCandidates] is the only walk in the package: [Build] and
// [Fingerprint] both read through it, so normalizing once means the
// index's spans and the input hash can never disagree about which bytes
// a file is made of.
//
// The consequence, recorded deliberately: a reported byte offset is an
// offset into the normalized text, not into the on-disk file, on a
// platform where the two differ. Line and column — which is what a
// human or an editor acts on — are unaffected, and cross-platform
// reproducibility is worth more than a byte offset that is only correct
// on one of the three platforms.
//
// A lone CR is left alone. It is not a line ending in any checkout this
// tool sees, and rewriting it would change content rather than restore
// it.
func normalizeLineEndings(src []byte) []byte {
	if !bytes.Contains(src, []byte("\r\n")) {
		return src
	}
	return bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
}

// builder is the mutable state of one build.
type builder struct {
	specRoot string
	repoRoot string
	profile  Profile
	types    map[string]DocumentTypeRules
	// vocabulary is opts.Profile.FeatureTags() compiled once, the seam
	// through which the registered patterns reach the Gherkin component.
	vocabulary vocabulary

	// hasher accumulates the content hash of everything the walk reads,
	// so that the fingerprint the emitted index records is a by-product of
	// the build rather than a second pass over the same files.
	hasher inputHasher

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
	// references is every reference the walk extracted, in walk order.
	// They are resolved and ordered in [builder.finish], once the whole
	// corpus is known: what a reference reached is a question about the
	// corpus rather than about the document it was written in.
	references []specmodel.Reference
	// features is every feature the walk extracted, in walk order, ordered
	// in [builder.finish] by the path it was written in and then by its
	// position within that file.
	features []specmodel.Feature
	// failures is every Gherkin source that would not parse, recorded
	// against its own file so the walk continues (AC-SPEC-002.4).
	failures []SourceFailure
}

// visit classifies one source the walk offered, and folds its bytes into
// the build's input hash.
func (b *builder) visit(kind sourceKind, reported string, src []byte, readErr error) {
	b.hasher.add(kind, reported, src, readErr)

	if kind == featureSource {
		// A feature source is no candidate: it produces no entry and no
		// exclusion whatever it holds, so membership is never consulted for
		// it and AC-SPEC-001.6's silent pass-over stands.
		b.visitFeatureSource(reported, src, readErr)
		return
	}

	if readErr != nil {
		// An unreadable candidate is recorded against itself for the same
		// reason an unparsable one is: losing a file silently is worse
		// than an index that says which file it could not read.
		// No bytes were read, so there is no body to find a heading in;
		// addEntry records the empty collection.
		b.addEntry(specmodel.IndexEntry{
			Path:       reported,
			ParseError: &specmodel.ParseError{Message: readErr.Error(), Span: span.Of(nil, 0, 0)},
		}, member{})
		return
	}

	b.classify(reported, src)
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
		b.addEntry(specmodel.IndexEntry{
			Path:       reported,
			ParseError: perr,
			Headings:   headingsOf(src, nil),
		}, member{})

	case doc == nil:
		// No frontmatter block at all (AC-SPEC-001.3).
		b.index.Exclusions = append(b.index.Exclusions, specmodel.IndexExclusion{
			Path:   reported,
			Reason: specmodel.ExclusionReasonNoFrontmatter,
		})

	default:
		b.classifyDocument(reported, src, doc)
	}
}

// classifyDocument decides membership for a file whose frontmatter
// parsed, from the type it declares.
//
// src is the file's bytes, carried this far so that a member's references
// are extracted from the read the walk already performed rather than from
// a second one.
func (b *builder) classifyDocument(reported string, src []byte, doc *frontmatter.Document) {
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
		b.addEntry(specmodel.IndexEntry{
			Path:       reported,
			ParseError: perr,
			Headings:   headingsOf(src, doc),
		}, member{})
		return
	}

	b.docs[reported] = doc
	b.addEntry(
		specmodel.IndexEntry{
			Path:        reported,
			Frontmatter: envelope,
			Headings:    headingsOf(src, doc),
		},
		member{path: reported, slug: string(envelope.Slug), slugAt: slugSpan(doc), doc: doc},
	)
	b.extractReferences(reported, src, doc, b.types[declared])
	b.extractBlockFeatures(reported, src, doc)
}

// headingsOf returns the headings of one candidate's body, in the order
// they are written and in the document's own coordinates.
//
// The search starts at the body rather than at the start of the file. A
// frontmatter block's closing delimiter underlines the YAML above it,
// which CommonMark reads as a setext heading, so a whole-file search
// reports the entire block as a level-2 heading — the one respect in
// which a heading search differs from the fenced-block search beside it,
// which is offered the whole file precisely so its spans need no second
// arithmetic.
//
// doc is the parsed frontmatter where there is one, so that the offset
// costs nothing beyond the parse the walk already performed. Where the
// block could not be parsed there is no Document to ask, and the block is
// located again from the source: a member whose frontmatter failed to
// parse still carries its headings, a heading not depending on an
// envelope.
func headingsOf(src []byte, doc *frontmatter.Document) []specmodel.Heading {
	from := 0
	switch {
	case doc != nil:
		from = int(doc.Region.Outer.Offset + doc.Region.Outer.Length)
	default:
		// The Frontmatter component owns the format, so it is asked
		// where the block is rather than the rule being re-derived here:
		// its closing delimiter carries a leading newline, and a search
		// for a bare "---" would close the block at the first such run
		// inside a value and leave the body starting mid-block. Where no
		// block is located — none present, or one that never closes —
		// the whole source is the body.
		if region, ok := frontmatter.Locate(src); ok {
			from = int(region.Outer.Offset + region.Outer.Length)
		}
	}

	found := markdown.HeadingsFrom(src, from)
	headings := make([]specmodel.Heading, 0, len(found))
	for _, heading := range found {
		headings = append(headings, specmodel.Heading{
			Text:        heading.Text,
			Level:       uint64(heading.Level),
			Span:        heading.Position,
			SectionSpan: heading.Section,
			// Record blocks attach to a heading once a profile registers
			// a tag; until then every heading carries none rather than
			// null.
			RecordBlocks: []specmodel.RecordBlock{},
		})
	}
	return headings
}

// extractReferences records the references one member makes, from the
// bytes the walk has already read and the frontmatter it has already
// parsed — which is what keeps the reference graph costing the walk it
// rides on rather than a second pass over the corpus.
//
// Only a member contributes references. A file that is not a
// specification, or is one whose frontmatter could not be read, has no
// declared type, and the properties that hold a reference in frontmatter
// are registered per type — so there is nothing to ask the profile for
// and no document whose graph the answer would belong to.
func (b *builder) extractReferences(reported string, src []byte, doc *frontmatter.Document, rules DocumentTypeRules) {
	b.references = append(b.references, references.Extract(references.Options{
		Path:                reported,
		Source:              src,
		Frontmatter:         doc,
		ReferenceProperties: rules.ReferenceProperties,
	})...)
}

// addEntry records one entry, and the member facts duplicate detection
// will need. A zero member — an entry with a parse failure — declares no
// slug and takes part in no slug comparison.
func (b *builder) addEntry(entry specmodel.IndexEntry, m member) {
	// Every collection on the entry is non-nil for the reason
	// [specindex.New] gives for the index's own: an entry with no
	// headings marshals them as [] rather than as null, so the emitted
	// index compares byte for byte across platforms and a consumer
	// iterating a collection never distinguishes "empty" from "absent".
	if entry.Headings == nil {
		entry.Headings = []specmodel.Heading{}
	}
	if entry.RecordBlocks == nil {
		entry.RecordBlocks = []specmodel.RecordBlock{}
	}
	b.index.Entries = append(b.index.Entries, entry)
	if m.slug != "" {
		b.members = append(b.members, m)
	}
}

// finish orders everything the walk collected and assembles the Corpus.
//
// Ordering is this component's responsibility rather than the caller's.
// Directory iteration order is a filesystem property, not a corpus
// property, so entries and exclusions are sorted by path and findings
// through the one authority for finding order — [finding.Report] — so
// that the same corpus content yields identical output wherever it is
// built.
func (b *builder) finish(opts Options) *Corpus {
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

	// Resolution runs here rather than during the walk because it is a
	// question about the whole corpus: a reference read from the first
	// file may name the last, so nothing can be resolved until every
	// member's slug is known. It reads no document — the walk is still
	// the only pass over the corpus.
	b.index.References = append(b.index.References, orderedReferences(b.references)...)
	references.Resolve(b.index.References, b.index)
	b.index.Features = append(b.index.Features, orderedFeatures(b.features)...)

	report := finding.NewReport()
	report.Add(duplicateSlugs(b.members)...)

	// The index records what it was built from, which is what lets a
	// later invocation serve this same index warm instead of rebuilding
	// (AC-SPEC-010.5).
	hash := b.hasher.sum(opts)
	b.index.InputHash = &hash

	built := &Corpus{
		repoRoot: b.repoRoot,
		specRoot: b.specRoot,
		index:    b.index,
		findings: report.Model().Findings,
		failures: b.failures,
		byPath:   b.byPath,
		bySlug:   b.alsoKnown,
		profile:  b.profile,
		types:    b.types,
		docs:     b.docs,
	}
	built.indexReferences()
	return built
}

// orderedReferences returns refs in the order the emitted index carries
// them: by the path of the document the reference occurs in, then by its
// position within that document.
//
// One rule serves two requirements. A document's outbound references come
// out ordered by position, which is what AC-SPECREF-001.5 asks for, and
// the collection as a whole is a function of corpus content rather than
// of the filesystem's iteration order, which is what makes the emitted
// index byte-identical across platforms and across runs (AC-SPEC-009.7).
//
// Sorting by byte offset rather than by line and column is deliberate:
// the offset is the authoritative extent of a span, and two references
// cannot begin at the same offset in the same document, so the order is
// total.
func orderedReferences(refs []specmodel.Reference) []specmodel.Reference {
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Path != refs[j].Path {
			return refs[i].Path < refs[j].Path
		}
		return refs[i].Span.Offset < refs[j].Span.Offset
	})
	return refs
}

// orderedFeatures returns features in the order the emitted index carries
// them: by the path the feature is written in, then by its position
// within that file.
//
// The same rule the reference collection is ordered by, for the same
// reason: the order is a function of corpus content rather than of the
// filesystem's iteration order, which is what makes the emitted index
// byte-identical across platforms and across runs (AC-SPEC-009.7). Two
// features cannot begin at the same offset in the same file, so the order
// is total.
func orderedFeatures(features []specmodel.Feature) []specmodel.Feature {
	sort.Slice(features, func(i, j int) bool {
		if features[i].Path != features[j].Path {
			return features[i].Path < features[j].Path
		}
		return features[i].Span.Offset < features[j].Span.Offset
	})
	return features
}

// indexReferences builds the two direction lookups over the index's one
// reference collection, once, when the corpus is built or opened.
//
// Both directions index the same rows rather than storing an edge and its
// reverse ([[adr-018]]): outbound is keyed by the path a reference occurs
// in, inbound by the path it resolved to, and each map holds positions
// into [Corpus.index].References. A reference that resolved to nothing has
// no inbound key, which is the honest shape — a broken edge has one end.
//
// The anchor-only form is the one reference whose target is not what its
// slug half reached: its slug is empty and resolves to nothing, and the
// document it points into is the one holding it (AC-SPECANCH-002.6). It
// is recorded inbound against that document, so that a query for what
// reaches a heading answers with every reference that reaches it rather
// than with every reference written somewhere else (AC-SPECREF-001.8).
// Only where its anchor resolved: one that names no heading of its own
// document reaches nothing and has one end like any other broken edge.
//
// The values come out in the collection's own order, so a lookup answers
// in path-then-position order without sorting anything a second time.
func (c *Corpus) indexReferences() {
	c.outbound = map[string][]int{}
	c.inbound = map[string][]int{}

	for i, reference := range c.index.References {
		c.outbound[reference.Path] = append(c.outbound[reference.Path], i)
		if resolved, ok := reference.Resolution.AsResolved(); ok {
			c.inbound[resolved.TargetPath] = append(c.inbound[resolved.TargetPath], i)
			continue
		}
		if references.ResolvedSelfAnchor(reference) {
			c.inbound[reference.Path] = append(c.inbound[reference.Path], i)
		}
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

// declaredSlug returns the slug a document declares, as it declared it.
//
// It reads the parsed block rather than the typed envelope on purpose. A
// feature written in a document is attributed to the slug that document
// actually wrote, and the envelope's value is a validated one that a
// document with a malformed or absent slug does not reach — which would
// leave its features attributed to nothing while the document itself
// indexes. A non-string value is reported as its source text, for the
// same reason [declaredType] reports one.
func declaredSlug(doc *frontmatter.Document) specmodel.Slug {
	property, ok := doc.Property(slugProperty)
	if !ok {
		return ""
	}
	if value, isString := property.Value.(string); isString {
		return specmodel.Slug(value)
	}
	return specmodel.Slug(property.Raw)
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

// envelopeShape is [specmodel.DocumentFrontmatter] with its generated
// UnmarshalJSON deliberately dropped.
//
// A defined type does not inherit the methods of the type it is derived
// from, so decoding through this one runs encoding/json's default struct
// path: the same json tags and the same type checking, without the
// generated decoder's required-field enforcement.
//
// That enforcement is right everywhere else the CLI decodes a generated
// record — a specs.toml missing `workspace.name` or a persisted index
// missing an entry path should fail loudly. It is wrong *here*, because a
// missing required property in a document's frontmatter is the subject of
// a check rather than a precondition of reading one. Enforcing it at
// decode makes every such document a `frontmatter/parse` failure and
// leaves `frontmatter/missing-property` unable to fire at all — a check
// that reports success because it never runs. The schema owns
// required-ness and says which property and where; this decode owns shape.
type envelopeShape specmodel.DocumentFrontmatter

// decodeFrontmatter converts a parsed block into the typed envelope an
// index entry carries (AC-SPEC-002.1).
//
// The conversion runs through JSON because the generated records carry
// json tags and no YAML ones. A value of the wrong shape for the
// envelope is a located failure rather than an error the caller raises:
// it is recorded against its own document and the walk continues. A
// value that is merely absent is not a failure here at all — see
// [envelopeShape].
func decodeFrontmatter(doc *frontmatter.Document) (*specmodel.DocumentFrontmatter, *specmodel.ParseError) {
	raw, err := doc.JSON()
	if err != nil {
		return nil, &specmodel.ParseError{
			Message: fmt.Sprintf("the frontmatter block could not be read: %s", err),
			Span:    doc.Region.Inner,
		}
	}

	var shape envelopeShape
	if err := json.Unmarshal(raw, &shape); err != nil {
		return nil, &specmodel.ParseError{
			Message: fmt.Sprintf("the frontmatter block does not fit the document envelope: %s", err),
			Span:    envelopeErrorSpan(doc, err),
		}
	}
	envelope := specmodel.DocumentFrontmatter(shape)
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
	return c.document(c.normalize(path))
}

// document returns the parsed document behind one member, named in
// reported form.
//
// A built corpus answers from the walk. A corpus served warm from a
// persisted index has parsed nothing, so it reads and parses the one file
// asked for and remembers it — which is what keeps a single-document
// check against a warm index proportional to that document rather than to
// the corpus (AC-SPEC-009.2). A file that cannot be read or parsed is
// reported as absent, exactly as a member with a recorded parse error is.
func (c *Corpus) document(reported string) (*frontmatter.Document, bool) {
	c.docsMu.Lock()
	defer c.docsMu.Unlock()

	if doc, ok := c.docs[reported]; ok {
		return doc, true
	}
	if !c.warm {
		return nil, false
	}

	doc, err := c.readDocument(reported)
	if err != nil || doc == nil {
		return nil, false
	}
	c.docs[reported] = doc
	return doc, true
}

// readDocument reads and parses one member's source, through the same
// os.Root containment the walk reads under (AC-SPEC-001.5).
func (c *Corpus) readDocument(reported string) (*frontmatter.Document, error) {
	rel, err := filepath.Rel(c.specRoot, filepath.Join(c.repoRoot, filepath.FromSlash(reported)))
	if err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(c.specRoot)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	src, err := readUnder(root, rel)
	if err != nil {
		return nil, err
	}

	doc, perr := frontmatter.Parse(src)
	if perr != nil {
		return nil, fmt.Errorf("corpus: reading %s: %s", reported, perr.Message)
	}
	return doc, nil
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

// OutboundReferences returns every reference one document makes: the
// `[[slug]]` links in its body and the values of the reference-bearing
// frontmatter properties the active profile registers for its type, in
// the order they occur in the document.
//
// The result is empty for a document that references nothing, and for one
// the corpus does not hold — a document that is not a member makes no
// reference the graph records, and neither case is an error.
//
// path may be absolute or repository-relative, and may use either
// platform's separator.
func (c *Corpus) OutboundReferences(path string) []specmodel.Reference {
	return c.referencesAt(c.outbound, path)
}

// InboundReferences returns every reference that resolved to one
// document: the reverse direction, derived from the same collection the
// outbound references are read from rather than stored a second time.
//
// Each returned reference is cited at its position in the *referring*
// document, because that is the file a reader opens to change it. The
// result is ordered by referring path and then by position within it, and
// is empty for a document nothing references — which is an answer, not a
// failure.
//
// path may be absolute or repository-relative, and may use either
// platform's separator.
func (c *Corpus) InboundReferences(path string) []specmodel.Reference {
	return c.referencesAt(c.inbound, path)
}

// referencesAt returns the references one direction's lookup holds for a
// path, copied out of the index so that a caller cannot reorder the
// collection every later query reads.
func (c *Corpus) referencesAt(direction map[string][]int, path string) []specmodel.Reference {
	positions := direction[c.normalize(path)]
	if len(positions) == 0 {
		return nil
	}
	found := make([]specmodel.Reference, 0, len(positions))
	for _, i := range positions {
		found = append(found, c.index.References[i])
	}
	return found
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
