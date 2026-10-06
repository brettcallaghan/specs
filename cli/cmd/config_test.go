// Tests for Load — resolving and checking specs.toml and selecting a
// profile from a registry.
//
// The fixture profile below is a stub rather than cli/profile's real
// spec-driven-dev, so these tests exercise this package's own resolution
// and checking logic independent of that profile's document vocabulary;
// TestBuiltinRegistryRegistersSpecDrivenDev is the one test that reaches
// the real profile, proving the registration itself.
package cmd_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brettcallaghan/agent-plugins/cli/cmd"
	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// stubProfile stands in for a real profile, the same fixture shape
// cli/core/corpus's own tests use.
type stubProfile struct {
	schema      []byte
	types       map[string]corpus.DocumentTypeRules
	conventions specmodel.ConventionsConfig
	lifecycle   corpus.LifecycleRules
	featureTags corpus.FeatureTagRules
}

func (p stubProfile) Schema() []byte                                     { return p.schema }
func (p stubProfile) DocumentTypes() map[string]corpus.DocumentTypeRules { return p.types }
func (p stubProfile) Conventions() specmodel.ConventionsConfig           { return p.conventions }
func (p stubProfile) Lifecycle() corpus.LifecycleRules                   { return p.lifecycle }
func (p stubProfile) FeatureTags() corpus.FeatureTagRules                { return p.featureTags }

// testRegistry registers exactly one profile, "test-profile", with a
// line_width default of 100 — distinct from spec-driven-dev's 80, so a
// test asserting the default came from the active profile cannot pass by
// coincidence.
func testRegistry() *cmd.Registry {
	registry := cmd.NewRegistry()
	width := uint64(100)
	registry.Register("test-profile", stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"note": {SlugPattern: `^note-[0-9]+$`},
		},
		conventions: specmodel.ConventionsConfig{LineWidth: &width},
	})
	return registry
}

// writeConfig writes content as specs.toml under dir and returns its
// path.
func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "specs.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing specs.toml: %v", err)
	}
	return path
}

// AC: AC-SPEC-008.1
func TestLoadResolvesSpecRootRelativeToTheConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs/specs"

[profile]
name = "test-profile"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}

	want := filepath.Join(dir, "docs", "specs")
	if cfg.SpecRoot != want {
		t.Errorf("SpecRoot = %q, want %q (relative to the config file's directory, not the cwd)", cfg.SpecRoot, want)
	}
}

// AC: AC-SPEC-008.2
func TestLoadResolvesWorkspaceProfileAndConventions(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs"

[profile]
name = "test-profile"

[conventions]
line_width = 72
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}

	if cfg.Workspace != "acme" {
		t.Errorf("Workspace = %q, want %q", cfg.Workspace, "acme")
	}
	if cfg.Profile == nil {
		t.Fatal("Profile is nil")
	}
	// The activated profile is reachable, not merely named: its own
	// document types (and, through them, the union schema and slug
	// patterns) resolve from here.
	if _, ok := cfg.Profile.DocumentTypes()["note"]; !ok {
		t.Error("the activated profile's document types are not reachable from Config")
	}
	if cfg.Conventions.LineWidth == nil || *cfg.Conventions.LineWidth != 72 {
		t.Errorf("LineWidth = %v, want 72 (specs.toml's own value)", cfg.Conventions.LineWidth)
	}
}

// Absent [conventions] keys resolve from the active profile
// (specs-configuration-contract.md's resolution rule).
func TestLoadResolvesAbsentConventionsFromTheActiveProfile(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs"

[profile]
name = "test-profile"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %+v, want none", findings)
	}
	if cfg.Conventions.LineWidth == nil || *cfg.Conventions.LineWidth != 100 {
		t.Errorf("LineWidth = %v, want the active profile's default of 100", cfg.Conventions.LineWidth)
	}
}

// AC: AC-SPEC-008.3
func TestLoadReportsTheAbsentConfigPathAndInfersNoSpecRoot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "specs.toml")

	cfg, findings, err := cmd.Load(path, testRegistry())
	if cfg != nil {
		t.Error("Load resolved a Config for a repository with no specs.toml")
	}
	if findings != nil {
		t.Errorf("findings = %+v, want none: an absent file is not a finding, there is no span to carry", findings)
	}
	var notFound *cmd.ErrConfigNotFound
	if !errors.As(err, &notFound) {
		t.Fatalf("err = %v (%T), want *cmd.ErrConfigNotFound", err, err)
	}
	if notFound.Path != path {
		t.Errorf("Path = %q, want %q (the path looked for)", notFound.Path, path)
	}
}

// AC: AC-SPEC-008.4, AC-SPEC-008.5
func TestLoadReportsAnInvalidKeyWithItsSpanAndDoesNotDefaultIt(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = 42

[profile]
name = "test-profile"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config despite an invalid key — the offending key was defaulted rather than reported")
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}

	f := findings[0]
	if f.Rule != cmd.RuleConfigSchema {
		t.Errorf("rule = %q, want %q", f.Rule, cmd.RuleConfigSchema)
	}
	if f.Span.StartLine == 0 {
		t.Error("the finding carries no span")
	}
	detail, ok := f.Detail.Value.(specmodel.FindingSchemaViolation)
	if !ok {
		t.Fatalf("detail = %T, want FindingSchemaViolation", f.Detail.Value)
	}
	if detail.Pointer != "/workspace/root" {
		t.Errorf("pointer = %q, want %q", detail.Pointer, "/workspace/root")
	}
}

// Every defect specs.toml carries is reported in one run, not just the
// first.
func TestLoadReportsEveryDefectInOneRun(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = 42
bogus = "nope"

[profile]
name = "test-profile"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config despite two invalid keys")
	}

	var sawInvalidType, sawUndeclaredKey bool
	for _, f := range findings {
		switch f.Rule {
		case cmd.RuleConfigSchema:
			sawInvalidType = true
		case cmd.RuleConfigUndeclaredKey:
			sawUndeclaredKey = true
		}
	}
	if !sawInvalidType {
		t.Error("the invalid root value was not reported")
	}
	if !sawUndeclaredKey {
		t.Error("the undeclared key was not reported")
	}
	if len(findings) < 2 {
		t.Fatalf("got %d findings, want at least 2 (one run, not stop-at-first): %+v", len(findings), findings)
	}
}

// IF [profile].name names no profile in the registry, a finding lists the
// registered names.
func TestLoadReportsAnUnknownProfileListingRegisteredNames(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs"

[profile]
name = "does-not-exist"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config naming an unregistered profile")
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}

	f := findings[0]
	if f.Rule != cmd.RuleConfigUnknownProfile {
		t.Errorf("rule = %q, want %q", f.Rule, cmd.RuleConfigUnknownProfile)
	}
	if !strings.Contains(f.Message, "test-profile") {
		t.Errorf("message %q does not list the registered profile name", f.Message)
	}
}

// A table the schema requires and specs.toml omits entirely is reported
// as a missing property, located against the whole file since there is no
// [profile] table to point at.
func TestLoadReportsAMissingRequiredTable(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs"
`)

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config with no [profile] table")
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}
	if findings[0].Rule != cmd.RuleConfigMissingProperty {
		t.Errorf("rule = %q, want %q", findings[0].Rule, cmd.RuleConfigMissingProperty)
	}
	detail, ok := findings[0].Detail.Value.(specmodel.FindingMissingProperty)
	if !ok {
		t.Fatalf("detail = %T, want FindingMissingProperty", findings[0].Detail.Value)
	}
	if detail.Property != "profile" {
		t.Errorf("property = %q, want %q", detail.Property, "profile")
	}
}

// A TOML syntax error is reported as a finding rather than surfacing as a
// bare decode failure — checking the configuration covers malformed TOML,
// not only well-formed-but-invalid TOML.
func TestLoadReportsATOMLSyntaxError(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, "workspace = [unterminated\n")

	cfg, findings, err := cmd.Load(path, testRegistry())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config from an unparsable file")
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}
	if findings[0].Span.StartLine == 0 {
		t.Error("the syntax finding carries no span")
	}
}

// BuiltinRegistry holds this framework's own profile, so an adopting
// repository can select "spec-driven-dev" without registering it itself.
func TestBuiltinRegistryRegistersSpecDrivenDev(t *testing.T) {
	registry := cmd.BuiltinRegistry()
	if _, ok := registry.Lookup("spec-driven-dev"); !ok {
		t.Error("BuiltinRegistry does not register spec-driven-dev")
	}
}

// registryWithBrokenRegistration registers "broken-profile": a pairing
// over "successor" on document type "note", where "successor" is never
// declared in ReferenceProperties — one of the three defects
// [corpus.ValidateRegistration] finds, per its own doc comment.
func registryWithBrokenRegistration() *cmd.Registry {
	registry := cmd.NewRegistry()
	registry.Register("broken-profile", stubProfile{
		schema: []byte(`{"definitions":{}}`),
		types: map[string]corpus.DocumentTypeRules{
			"note": {
				SlugPattern: `^note-[0-9]+$`,
				Pairings: []specmodel.PropertyPairing{
					{Property: "successor", Counterpart: "predecessor"},
				},
			},
		},
	})
	return registry
}

// A profile registration [corpus.ValidateRegistration] rejects is
// reported as a finding naming the offending key, positioned at
// [profile].name — the only file position a compiled-in registration
// has — and Load falls back to no Config rather than activating it.
//
// AC: AC-SPECREF-004.3
func TestLoadReportsARegistrationDefectAtProfileNamePosition(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
[workspace]
name = "acme"
root = "docs"

[profile]
name = "broken-profile"
`)

	cfg, findings, err := cmd.Load(path, registryWithBrokenRegistration())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg != nil {
		t.Error("Load resolved a Config for a profile whose registration is defective — the offending key was activated rather than reported")
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %+v", len(findings), findings)
	}

	f := findings[0]
	if f.Rule != cmd.RuleConfigInvalidRegistration {
		t.Errorf("rule = %q, want %q", f.Rule, cmd.RuleConfigInvalidRegistration)
	}
	if !strings.Contains(f.Message, "note.successor") {
		t.Errorf("message %q does not name the offending key", f.Message)
	}
	if f.Span.StartLine == 0 {
		t.Error("the finding carries no span")
	}
	detail, ok := f.Detail.Value.(specmodel.FindingSchemaViolation)
	if !ok {
		t.Fatalf("detail = %T, want FindingSchemaViolation", f.Detail.Value)
	}
	if detail.Pointer != "/profile/name" {
		t.Errorf("pointer = %q, want %q (the only file position a compiled-in registration has)", detail.Pointer, "/profile/name")
	}
}

// Registry.Names is sorted, so a finding listing them is deterministic
// across runs regardless of Go's randomized map iteration.
func TestRegistryNamesIsSorted(t *testing.T) {
	registry := cmd.NewRegistry()
	registry.Register("zebra", stubProfile{})
	registry.Register("alpha", stubProfile{})

	names := registry.Names()
	if len(names) != 2 || names[0] != "alpha" || names[1] != "zebra" {
		t.Errorf("Names = %v, want [alpha zebra]", names)
	}
}
