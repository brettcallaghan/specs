package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/frontmatter"
	"github.com/brettcallaghan/agent-plugins/cli/core/mdlinks"
	"github.com/brettcallaghan/agent-plugins/cli/core/references"
)

// newLinksCmd builds the `links` subcommand: rewrite one document's
// `[[slug]]` wikilinks into markdown links, per
// container-spec-cli.md's Command Surface (modified — links) delta
// (AC-SPECMD-005.1 through AC-SPECMD-005.10).
//
// This is the sixth subcommand REQ-SPEC-011 anticipated growing into; it
// neither renames nor repurposes the five established here. Unlike refs,
// which answers what cites what and always exits zero, links transforms
// a document and can refuse.
func newLinksCmd(streams *ioStreams, configPath *string) *cobra.Command {
	var mapPath string

	cmd := &cobra.Command{
		Use:   "links <path>",
		Short: "Rewrite a document's wikilinks into markdown links",
		Long: "links rewrites the [[slug]] wikilinks in one markdown document into " +
			"ordinary markdown links: each citation targets the path a caller-supplied " +
			"--map names for its slug, or, given no map, the cited document's path " +
			"relative to the document being rewritten. Unlike refs, which reports what " +
			"cites what, links emits one document with its citations rewritten.",
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runLinks(streams, *configPath, args[0], mapPath)
		},
	}
	cmd.Flags().StringVar(&mapPath, "map", "",
		"path to a YAML link map resolving corpus slugs to link targets")
	return cmd
}

// runLinks converts the operand and --map into one mdlinks.Resolve /
// mdlinks.Substitute call and its result into two streams and an exit
// status. It holds no judgement of its own: what a citation resolves to
// is settled by cli/core/mdlinks against the references
// cli/core/references already extracted and resolved.
//
// The operand need not be a corpus member — a rewritten copy travels
// outside the spec root by construction (the FRD's Terminology) — so it
// is read directly rather than through the corpus's own membership-gated
// Document/CheckFor. The corpus is still opened warm, exactly as check,
// refs, slug-exists and next-id do, because resolving a citation still
// needs its index.
func runLinks(streams *ioStreams, configPath, operand, mapPath string) error {
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	if err != nil {
		return err
	}
	if findings != nil {
		return reportAndExit(streams, findings, false)
	}

	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		return err
	}

	c, err := openCorpusReadOnly(opts)
	if err != nil {
		return err
	}

	reported := normalizeOperand(operand)
	report := finding.NewReport()

	src, readErr := os.ReadFile(operand)
	if readErr != nil {
		report.Add(documentUnreadable(reported, readErr))
		return finishLinks(streams, report, nil, 0)
	}
	// Normalized the same way corpus.Build's own walk normalizes every
	// candidate file it reads (AC-SPEC-009.7): the operand is read
	// through a second, independent path from that walk — it need not be
	// a corpus member at all — and a CRLF checkout left unnormalized here
	// would move every reference's byte offset against the corpus's own
	// (already-normalized) spans, and would echo CRLF back in the
	// rewritten document a Windows checkout produced identically to a
	// Linux one on every other command.
	src = normalizeLineEndings(src)

	doc, perr := frontmatter.Parse(src)
	if perr != nil {
		// The block is present and could not be read. Reusing
		// finding.FrontmatterParse — rather than treating this as a
		// second "unreadable" case — keeps one rule for one fact ("this
		// document's frontmatter does not parse") regardless of which
		// subcommand discovered it. Stopping here, rather than falling
		// through with doc == nil, avoids scanning the raw frontmatter
		// block as if it were body text for `[[...]]` constructs.
		report.Add(finding.FrontmatterParse(reported, perr.Position, perr.Message))
		return finishLinks(streams, report, nil, 0)
	}

	refs := references.Extract(references.Options{
		Path:                reported,
		Source:              src,
		Frontmatter:         doc,
		ReferenceProperties: referenceProperties(cfg, doc),
	})
	references.Resolve(refs, c.Index())

	var linkMap mdlinks.LinkMap
	if mapPath != "" {
		loaded, mapFindings := LoadLinkMap(mapPath)
		if mapFindings != nil {
			report.Add(mapFindings...)
			return finishLinks(streams, report, nil, 0)
		}
		// loaded is non-nil even when it resolves zero entries
		// (LoadLinkMap's make(LinkMap, len(raw)) with len(raw) == 0 is
		// still a non-nil map). That non-nilness is carried through
		// unchanged: mdlinks.Options.Map's nil-vs-empty distinction is
		// what lets links/unmapped fire for a supplied map that simply
		// omits a slug, rather than being read as "no map was given"
		// (AC-SPECMD-006.3).
		linkMap = mdlinks.LinkMap(loaded)
	}

	targets, mdFindings := mdlinks.Resolve(refs, mdlinks.Options{Map: linkMap, MapPath: mapPath})
	report.Add(mdFindings...)

	out, count := mdlinks.Substitute(src, refs, targets)
	return finishLinks(streams, report, out, count)
}

// openCorpusReadOnly opens the corpus exactly as [openCorpus] does —
// served warm from the persisted index at the default location when it
// is current, built by walking the corpus otherwise — but never writes
// the index back when it had to build one.
//
// links is the one subcommand AC-SPECMD-005.3 binds ("creates, modifies
// and deletes no file"); openCorpus's best-effort persist of a freshly
// built index would violate that on a cold cache, so links opens through
// this narrower helper instead. Every other subcommand carries no such
// constraint and keeps using openCorpus, cache write included.
func openCorpusReadOnly(opts corpus.Options) (*corpus.Corpus, error) {
	dest := corpus.IndexPath(opts, "")
	c, _, err := corpus.Open(opts, dest)
	if err != nil {
		return nil, fmt.Errorf("cmd: opening the corpus: %w", err)
	}
	return c, nil
}

// referenceProperties looks up the reference-bearing frontmatter
// properties for doc's declared type, per the active profile — the same
// per-type registration corpus.Build reads, applied here to a document
// that need not be a corpus member. A document with no frontmatter, or
// one declaring a type the profile does not define, contributes no
// frontmatter reference: only its body wikilinks are extracted.
func referenceProperties(cfg *Config, doc *frontmatter.Document) []string {
	if doc == nil {
		return nil
	}
	property, ok := doc.Property("type")
	if !ok {
		return nil
	}
	declared, isString := property.Value.(string)
	if !isString {
		return nil
	}
	rules, known := cfg.Profile.DocumentTypes()[declared]
	if !known {
		return nil
	}
	return rules.ReferenceProperties
}

// normalizeLineEndings rewrites every CRLF in src to a bare LF.
//
// This mirrors corpus.Build's own unexported normalization step, applied
// here because the links operand is read through this package's own
// os.ReadFile rather than through corpus's walk — the operand need not
// be a corpus member, so it may never pass through corpus at all. See
// that function's doc comment (cli/core/corpus/corpus.go) for the full
// rationale; the rule itself is one line and is duplicated rather than
// exported for one caller.
func normalizeLineEndings(src []byte) []byte {
	if !bytes.Contains(src, []byte("\r\n")) {
		return src
	}
	return bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))
}

// normalizeOperand converts the operand to the forward-slashed form
// every path in the corpus's reference graph is already reported in, so
// a caller on Windows naming the operand with a backslash still resolves
// correctly against corpus-relative targets, and so a finding raised
// against the operand names it the same way a corpus finding would.
func normalizeOperand(p string) string {
	return path.Clean(filepath.ToSlash(p))
}

// finishLinks renders report's findings and, when the rewrite completed,
// the replacement count, to streams.err (AC-SPECMD-003.2) — never to the
// stream the rewritten document itself is written to. The document is
// written to streams.out only when the run raised no error finding
// (AC-SPECMD-002.2, AC-SPECMD-005.9); a run whose only findings are
// warnings still emits it (AC-SPECMD-002.6) and exits 0 (AC-SPECMD-002.8).
func finishLinks(streams *ioStreams, report *finding.Report, out []byte, count int) error {
	fmt.Fprint(streams.err, report.Text())

	if report.HasErrors() {
		streams.exitCode = 1
		return nil
	}

	fmt.Fprintf(streams.err, "%d link(s) rewritten\n", count)
	_, err := streams.out.Write(out)
	return err
}
