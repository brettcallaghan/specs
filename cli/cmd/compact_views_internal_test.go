// Whitebox test of the claim the compact and full views of `feature` and
// `headings` share: they list the same items in the same order, the
// compact one only keeping fewer fields of each. Run over this
// repository's own corpus, with and without a selection.
package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brettcallaghan/specs/cli/core/corpus"
	"github.com/brettcallaghan/specs/cli/core/specmodel"
)

func repoCorpus(t *testing.T) *corpus.Corpus {
	t.Helper()
	configPath, err := filepath.Abs("../../specs.toml")
	if err != nil {
		t.Fatal(err)
	}
	cfg, configPath, findings, err := loadConfig(configPath)
	if err != nil || findings != nil {
		t.Fatalf("loadConfig = %v, %v", findings, err)
	}
	opts, err := corpusOptions(cfg, configPath)
	if err != nil {
		t.Fatal(err)
	}
	built, err := corpus.Build(opts)
	if err != nil {
		t.Fatal(err)
	}
	return built
}

// listed decodes a rendered JSON answer's list under key into one value per
// item, reading field from each.
func listed(t *testing.T, result any, key, field string) []any {
	t.Helper()
	var out bytes.Buffer
	streams := &ioStreams{out: &out, err: &out}
	var err error
	switch r := result.(type) {
	case headingsResult:
		err = renderHeadings(streams, r, true)
	case headingsCompactResult:
		err = renderHeadings(streams, r, true)
	}
	if err != nil {
		t.Fatal(err)
	}
	var answer map[string][]map[string]any
	if err := json.Unmarshal(out.Bytes(), &answer); err != nil {
		t.Fatalf("decoding %s: %v", out.String(), err)
	}
	var values []any
	for _, item := range answer[key] {
		values = append(values, item[field])
	}
	return values
}

// AC: AC-SPECSTR-003.16, AC-SPECSTR-003.1
func TestHeadingsCompactAndFullListTheSameMatchesInTheSameOrder(t *testing.T) {
	index := repoCorpus(t).Index()
	for _, operand := range []string{"", "module-document-core", "#open-questions", "#nonexistent"} {
		matches, _ := selectHeadings(index, parseSelection(operand), nil)
		full := headingsResult{Matches: matches}
		compact := compactHeadings(full)
		if got, want := listed(t, compact, "matches", "ref"), listed(t, full, "matches", "ref"); !reflect.DeepEqual(got, want) {
			t.Errorf("%q: compact refs %v, full refs %v", operand, got, want)
		}
		for _, field := range []string{"path", "section"} {
			if got := listed(t, compact, "matches", field); len(got) > 0 && got[0] != nil {
				t.Errorf("%q: compact view carries %q", operand, field)
			}
		}
		for i, match := range compact.Matches {
			if match.StartLine == 0 || match.StartLine >= full.Matches[i].Section.StartLine {
				t.Errorf("%q: start_line %d is not the heading line before its section at %d", operand, match.StartLine, full.Matches[i].Section.StartLine)
			}
		}
	}
}

// AC: AC-GHK-007.10, AC-GHK-007.11
func TestFeatureCompactAndFullListTheSameFeaturesInTheSameOrder(t *testing.T) {
	c := repoCorpus(t)
	all := c.Features()
	if len(all) == 0 {
		t.Fatal("this repository's corpus declares no feature")
	}
	for name, features := range map[string][]specmodel.Feature{
		"every feature": all,
		"one code":      c.FeaturesByCode(all[0].Code),
		"one feature":   all[:1],
	} {
		full := featuresResult{Features: features}
		compact := compactFeatures(full)
		if len(compact.Features) != len(features) {
			t.Fatalf("%s: compact lists %d features, full %d", name, len(compact.Features), len(features))
		}
		for i, summary := range compact.Features {
			if summary.Id != features[i].Id || summary.CriteriaCount != len(features[i].Criteria) {
				t.Errorf("%s: item %d = %+v, want %s with %d criteria", name, i, summary, features[i].Id, len(features[i].Criteria))
			}
		}
	}
}
