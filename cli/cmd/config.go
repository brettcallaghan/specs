package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/brettcallaghan/agent-plugins/cli/core/corpus"
	"github.com/brettcallaghan/agent-plugins/cli/core/finding"
	"github.com/brettcallaghan/agent-plugins/cli/core/specmodel"
)

// Config is a fully resolved and checked specs.toml: exactly what
// [[contract-specs-configuration]] says the resolved configuration
// determines — the spec root, the workspace identifier, the activated
// profile, and the resolved house conventions (AC-SPEC-008.2).
//
// [Load] returns a nil *Config whenever it also returns at least one
// finding: [[contract-specs-configuration]]'s "SHALL NOT fall back to a
// default for that key" is honored here by never partially resolving a
// Config from a file with any known defect, rather than by resolving
// everything except the offending key.
type Config struct {
	// SpecRoot is [workspace].root resolved to an absolute path relative
	// to the directory containing specs.toml (AC-SPEC-008.1), so the spec
	// root is correct regardless of the working directory Load is invoked
	// from.
	SpecRoot string
	// Workspace is [workspace].name, declared once on every index this
	// configuration's corpus builds.
	Workspace string
	// Profile is the profile [profile].name selected from the registry
	// passed to Load.
	Profile corpus.Profile
	// Conventions is the resolved house conventions: specs.toml's own
	// values where it set them, and Profile's registered defaults
	// elsewhere ([[contract-specs-configuration]]'s resolution rule).
	Conventions specmodel.ConventionsConfig
}

// ErrConfigNotFound is returned by [Load] when no file exists at the path
// checked, and by [discoverConfig] when no directory it searched holds
// one.
//
// It is deliberately not a Finding:
// [[contract-specs-configuration]] states this is the one configuration
// failure with no file to carry a span into ("WHEN specs.toml is absent,
// the CLI SHALL report every directory it searched and exit with the
// environment-error status, and SHALL NOT infer a spec root" —
// AC-SPEC-008.3). A caller reports it directly.
type ErrConfigNotFound struct {
	// Path is the location Load checked and found nothing at. It is set
	// for a file --config named, and empty for a failed search.
	Path string
	// Searched is every directory a search checked, nearest first. It is
	// set for a failed search, and empty for a file --config named.
	Searched []string
	// RepoRoot is the repository root the search stopped at, or empty
	// where it ran to the filesystem root (AC-SPEC-008.9).
	RepoRoot string
}

func (e *ErrConfigNotFound) Error() string {
	if len(e.Searched) == 0 {
		return fmt.Sprintf("no specs.toml found at %s", e.Path)
	}
	limit := "the filesystem root, no repository root being above it"
	if e.RepoRoot != "" {
		limit = "the repository root " + e.RepoRoot
	}
	return fmt.Sprintf("no %s found in %s (searched up to %s); pass --config to name one",
		configFileName, strings.Join(e.Searched, ", "), limit)
}

// exitStatus classifies a missing configuration as an environment error
// (status 3): the invocation was correct, and the repository it ran in
// lacks the file it needs.
func (e *ErrConfigNotFound) exitStatus() int { return statusEnvironment }

// Load reads, checks, and resolves the specs.toml file at path, selecting
// a profile from registry.
//
// Every defect specs.toml carries is reported as a finding in one run
// (AC-SPEC-008.4) rather than Load stopping at the first — a TOML syntax
// error is the one exception, because BurntSushi/toml has no partial tree
// to keep checking once parsing itself has failed. An unregistered
// [profile].name is checked and reported independently of whether the
// rest of the file is schema-valid, so it is never silently absorbed by
// an unrelated defect elsewhere in the same run.
//
// Load returns a non-nil error only for [ErrConfigNotFound] or a failure
// this package cannot itself recover from (the generated configuration
// schema failing to compile, most plausibly signalling a build defect
// rather than anything about the repository being checked). Every defect
// in the checked repository's own specs.toml is a finding, never an
// error.
//
// Once [profile].name resolves to a registered profile, Load also runs
// [corpus.ValidateRegistration] against it and reports any defect as a
// finding positioned at [profile].name (AC-SPECREF-004.3) — the one
// registration-level check that cannot be expressed as a specs.toml
// schema constraint, because what it validates is the compiled-in
// profile the name selected, not the text of the file itself.
func Load(path string, registry *Registry) (*Config, []specmodel.Finding, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, &ErrConfigNotFound{Path: path}
		}
		return nil, nil, fmt.Errorf("reading %s: %w", path, err)
	}

	raw, syntaxFindings := decodeConfig(path, src)
	if syntaxFindings != nil {
		return nil, syntaxFindings, nil
	}

	loc := locateConfig(src)
	report := finding.NewReport()

	schemaFindings, err := checkConfigSchema(path, src, raw, loc)
	if err != nil {
		return nil, nil, err
	}
	report.Add(schemaFindings...)

	if name, ok := profileNameFrom(raw); ok {
		if _, found := registry.Lookup(name); !found {
			at := loc.valueSpan("/profile/name", src)
			report.Add(configUnknownProfile(path, at, name, registry.Names()))
		}
	}

	if report.HasErrors() {
		return nil, report.Model().Findings, nil
	}

	var cfg specmodel.SpecsConfig
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("re-encoding %s: %w", path, err)
	}
	if err := json.Unmarshal(encoded, &cfg); err != nil {
		return nil, nil, fmt.Errorf("decoding %s: %w", path, err)
	}

	active, found := registry.Lookup(cfg.Profile.Name)
	if !found {
		// checkConfigSchema and the profile-registry check above already
		// guarantee this is unreachable: a clean schema check means
		// cfg.Profile.Name is a non-empty string, and the registry lookup
		// just above would have raised RuleConfigUnknownProfile and made
		// report.HasErrors() true otherwise.
		return nil, nil, fmt.Errorf("%s: profile %q resolved after passing its own check", path, cfg.Profile.Name)
	}

	// A registration is compiled in, so a defect in it is not the
	// repository's specs.toml being wrong in the way the checks above
	// catch — it is the *activated profile* contradicting itself (a
	// pairing over a property nothing extracts, a typed reference naming
	// an undefined target type, a lifecycle status outside the profile's
	// own vocabulary). [profile].name is the only place in specs.toml
	// that selected this profile, so every defect is positioned there
	// (AC-SPECREF-004.3), and — the same "never fall back to a default"
	// discipline this function's own doc comment states — a defect here
	// means Load reports it and returns no Config, rather than resolving
	// one that activates a registration known to be broken.
	if defects := corpus.ValidateRegistration(active); len(defects) > 0 {
		at := loc.valueSpan("/profile/name", src)
		for _, defect := range defects {
			report.Add(configRegistrationDefect(path, at, defect))
		}
		return nil, report.Model().Findings, nil
	}

	resolved := &Config{
		SpecRoot:    filepath.Join(filepath.Dir(path), cfg.Workspace.Root),
		Workspace:   cfg.Workspace.Name,
		Profile:     active,
		Conventions: resolveConventions(cfg.Conventions, active.Conventions()),
	}
	return resolved, nil, nil
}

// resolveConventions applies [[contract-specs-configuration]]'s
// resolution rule: a key specs.toml set wins, and the active profile's
// registered default fills every key it left absent.
func resolveConventions(fromConfig *specmodel.ConventionsConfig, profileDefaults specmodel.ConventionsConfig) specmodel.ConventionsConfig {
	resolved := profileDefaults
	if fromConfig != nil && fromConfig.LineWidth != nil {
		resolved.LineWidth = fromConfig.LineWidth
	}
	return resolved
}

// configFileName is the file a search looks for in each directory.
const configFileName = "specs.toml"

// loadConfig resolves the configuration a command runs against: the file
// --config named, read and searched for nowhere else (AC-SPEC-008.8), or,
// where --config was not given, the nearest specs.toml found by
// [discoverConfig] (AC-SPEC-008.7). It returns the path it loaded, which
// the caller hands on to [corpusOptions] so the paths that file declares
// resolve beside it (AC-SPEC-008.1).
//
// A discovered path is written relative to the working directory: from
// the directory holding specs.toml it is exactly "specs.toml", as an
// explicit default would have been, so a finding positioned in the file
// names it the same way wherever it was found.
func loadConfig(configPath string) (*Config, string, []specmodel.Finding, error) {
	if configPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, "", nil, environmentError("reading the working directory: %w", err)
		}
		found, err := discoverConfig(wd)
		if err != nil {
			return nil, "", nil, err
		}
		configPath = found
		if rel, err := filepath.Rel(wd, found); err == nil {
			configPath = rel
		}
	}
	cfg, findings, err := Load(configPath, BuiltinRegistry())
	return cfg, configPath, findings, err
}

// discoverConfig searches workingDir and then each directory above it for
// specs.toml, and returns the absolute path of the first one found, so
// the nearest wins (AC-SPEC-008.7).
//
// The search stops after the repository root: the nearest directory
// holding a .git entry, a directory in a clone and a file in a worktree
// or submodule. Outside any repository it runs to the filesystem root
// (AC-SPEC-008.9). The root is found by looking for .git rather than by
// running git, because discovery runs on every invocation and must work
// where git is not installed; [[adr-031]]'s git dependency is change
// detection's alone.
//
// A search that finds nothing returns [*ErrConfigNotFound] naming every
// directory it checked (AC-SPEC-008.3).
func discoverConfig(workingDir string) (string, error) {
	dir, err := filepath.Abs(workingDir)
	if err != nil {
		return "", environmentError("resolving the working directory %s: %w", workingDir, err)
	}
	notFound := &ErrConfigNotFound{}
	for {
		candidate := filepath.Join(dir, configFileName)
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
			return candidate, nil
		}
		notFound.Searched = append(notFound.Searched, dir)
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			notFound.RepoRoot = dir
			return "", notFound
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", notFound
		}
		dir = parent
	}
}
