package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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
// checked.
//
// It is deliberately not a Finding:
// [[contract-specs-configuration]] states this is the one configuration
// failure with no file to carry a span into ("WHEN specs.toml is absent,
// the CLI SHALL report the path it looked for and exit non-zero, and
// SHALL NOT infer a spec root" — AC-SPEC-008.3). A caller reports Path
// directly.
type ErrConfigNotFound struct {
	// Path is the location Load checked and found nothing at.
	Path string
}

func (e *ErrConfigNotFound) Error() string {
	return fmt.Sprintf("no specs.toml found at %s", e.Path)
}

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
func Load(path string, registry *Registry) (*Config, []specmodel.Finding, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil, &ErrConfigNotFound{Path: path}
		}
		return nil, nil, fmt.Errorf("cmd: reading %s: %w", path, err)
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
		return nil, nil, fmt.Errorf("cmd: re-encoding %s: %w", path, err)
	}
	if err := json.Unmarshal(encoded, &cfg); err != nil {
		return nil, nil, fmt.Errorf("cmd: decoding %s: %w", path, err)
	}

	active, found := registry.Lookup(cfg.Profile.Name)
	if !found {
		// checkConfigSchema and the profile-registry check above already
		// guarantee this is unreachable: a clean schema check means
		// cfg.Profile.Name is a non-empty string, and the registry lookup
		// just above would have raised RuleConfigUnknownProfile and made
		// report.HasErrors() true otherwise.
		return nil, nil, fmt.Errorf("cmd: %s: profile %q resolved after passing its own check", path, cfg.Profile.Name)
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
