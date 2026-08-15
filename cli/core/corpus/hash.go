package corpus

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"sort"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/specindex"
)

// hashAlgorithm labels the digest every input hash carries, so that a
// later change of algorithm is legible in the artifact rather than
// silently comparing unlike values: a hash written under one label never
// equals one written under another, and the mismatch rebuilds.
const hashAlgorithm = "sha256"

// inputHasher accumulates the content hash of one build's inputs.
//
// It is deliberately a hash of *content*, never of modification time.
// That is the pinned decision behind the whole warm-cache path: a fresh
// CI checkout gives every file a new mtime and identical bytes, and it
// must still be served warm, because the warm-index latency budget of
// FEAT-SPEC-009.1 is claimed in exactly that environment. An mtime-based
// cache is cheaper and would be cold on every CI run.
//
// Two properties make the result comparable across machines:
//
//   - Only repository-relative paths are hashed, never absolute ones, so
//     the same content checked out at a different absolute path hashes
//     the same.
//   - Files are folded in path order rather than walk order, so the
//     filesystem's iteration order cannot reach the digest.
type inputHasher struct {
	files []fileDigest
}

// fileDigest is one source file's contribution: its reported path, what
// kind of source it is, and the digest of its bytes.
type fileDigest struct {
	path   string
	kind   sourceKind
	digest [sha256.Size]byte
	// unread marks a file that could not be read. Its digest is
	// therefore not a digest of anything, and it is folded in under a
	// different label so that "unreadable" can never collide with the
	// content of a file that *was* read.
	unread bool
}

// add folds one source file into the hash. A file that could not be read
// still contributes, because an unreadable candidate is an index entry
// (with its failure recorded) and not an absence: if it later becomes
// readable, the index changes and the hash must change with it.
//
// A feature source contributes on the same terms as a candidate. It is no
// corpus member, but it decides the index's `features` collection, so a
// hash that covered only candidates would serve a persisted index built
// before a `.feature` file was edited — a stale answer to every query
// over features.
func (h *inputHasher) add(kind sourceKind, reported string, src []byte, readErr error) {
	entry := fileDigest{path: reported, kind: kind, unread: readErr != nil}
	if readErr == nil {
		entry.digest = sha256.Sum256(src)
	}
	h.files = append(h.files, entry)
}

// label names the field one source file is folded in under, so that
// neither the two kinds of source nor read and unread content can ever
// collide.
func (f fileDigest) label() string {
	switch {
	case f.kind == featureSource && f.unread:
		return "unreadable-feature:"
	case f.kind == featureSource:
		return "feature:"
	case f.unread:
		return "unreadable:"
	default:
		return "file:"
	}
}

// sum returns the input hash for a build with the given options, in the
// form `sha256:` followed by lowercase hexadecimal.
//
// The corpus's files are not the whole of the index's inputs. The
// workspace identifier is written onto the index; the active profile's
// schema and per-type table decide which candidates are members and which
// are excluded; its registered feature tags decide what a feature's
// records carry; and the schema version fixes the shape emitted. A change
// to any of them changes the index while every file on disk stands still,
// so all of them are folded in. What is deliberately *not* folded in is
// the spec root's absolute path — see [inputHasher].
func (h *inputHasher) sum(opts Options) string {
	sum := sha256.New()

	// A version tag on the hash construction itself, so that a future
	// change to what is hashed invalidates every index written under the
	// old rules rather than comparing two incomparable digests.
	field(sum, "input-hash-version", []byte("2"))
	field(sum, "schema-version", []byte(specindex.SchemaVersion))
	field(sum, "workspace", []byte(opts.Workspace))
	field(sum, "profile-schema", opts.Profile.Schema())

	types := opts.Profile.DocumentTypes()
	names := make([]string, 0, len(types))
	for name := range types {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rules := types[name]
		field(sum, "type:"+name, fmt.Appendf(nil, "%s\x00%t", rules.SlugPattern, rules.Numbered))
		// The reference-bearing properties are folded in because they
		// decide the index's *content*: a property registered today holds
		// a reference the emitted graph carries and did not carry
		// yesterday, while every file on disk stands still. Without this a
		// profile that registers a new one would be served the index built
		// before it — a stale answer, which the whole cache decision is
		// one-sided to avoid.
		//
		// The registration's other members are deliberately not folded in.
		// Typed references, pairings and the lifecycle are read by checks
		// that run against the live profile on every invocation and reach
		// nothing the index records, so a change to one of them cannot
		// make a persisted index stale — and folding it in would rebuild
		// the corpus to produce identical bytes.
		field(sum, "references:"+name, []byte(strings.Join(rules.ReferenceProperties, "\x00")))
	}

	foldFeatureTags(sum, opts.Profile.FeatureTags())

	sort.Slice(h.files, func(i, j int) bool { return h.files[i].path < h.files[j].path })
	for _, file := range h.files {
		field(sum, file.label()+file.path, file.digest[:])
	}

	return hashAlgorithm + ":" + hex.EncodeToString(sum.Sum(nil))
}

// foldFeatureTags folds the registered feature-tag vocabulary into a
// running hash, alongside the per-type table.
//
// A changed pattern changes the index while every file on disk stands
// still: an identifier tag that matched yesterday's pattern and not
// today's is recorded on no feature, a changed owner prefix attributes a
// feature source to a different document, and a qualifier registered
// today populates a criterion property that was empty before. All three
// are the index's own content, so all three are folded in.
//
// The shared captures are deliberately not folded in, by the same rule
// the per-type table's other members are left out: they are read only by
// the check that compares a criterion identifier against its feature's,
// which runs against the live profile on every invocation and reaches
// nothing the index records.
//
// Qualifiers are folded in tag order rather than registration order, so
// that two profiles registering the same vocabulary in a different order
// are not rebuilt against each other.
func foldFeatureTags(sum hash.Hash, tags FeatureTagRules) {
	field(sum, "feature-pattern", []byte(tags.FeaturePattern))
	field(sum, "criterion-pattern", []byte(tags.CriterionPattern))
	field(sum, "owner-prefix", []byte(tags.OwnerPrefix))

	qualifiers := append([]QualifierTag(nil), tags.Qualifiers...)
	sort.Slice(qualifiers, func(i, j int) bool { return qualifiers[i].Tag < qualifiers[j].Tag })
	for _, qualifier := range qualifiers {
		field(sum, "qualifier:"+qualifier.Tag,
			fmt.Appendf(nil, "%s\x00%s", qualifier.Meaning, qualifier.Level))
	}
}

// field folds one named value into a running hash, length-prefixed so
// that no two different sequences of inputs can produce the same byte
// stream — without it, a file named `a` holding `bc` and a file named
// `ab` holding `c` would hash alike.
func field(h hash.Hash, name string, value []byte) {
	fmt.Fprintf(h, "%s\x00%d\x00", name, len(value))
	h.Write(value)
}

// Fingerprint returns the content hash of the inputs a [Build] with the
// same options would read, without parsing any of them.
//
// This is the cheap half of the warm-cache path: it reads the same bytes
// the build reads — every candidate and every feature source, on the one
// walk — and does none of the frontmatter parsing, Gherkin parsing,
// envelope decoding, classification or checking that dominate a build. A caller
// compares the result against the hash recorded on a persisted index to
// decide whether that index may be served (AC-SPEC-010.5) or must be
// rebuilt (AC-SPEC-010.6).
//
// The result is identical to the hash [Build] records on the index it
// returns, for the same options over the same corpus — both are fed by
// the same walk.
func Fingerprint(opts Options) (string, error) {
	specRoot, repoRoot, err := opts.roots()
	if err != nil {
		return "", err
	}

	var hasher inputHasher
	if err := scanSources(specRoot, repoRoot, hasher.add); err != nil {
		return "", err
	}
	return hasher.sum(opts), nil
}
