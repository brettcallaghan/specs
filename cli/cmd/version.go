package cmd

import (
	"runtime/debug"
	"strconv"

	"github.com/spf13/cobra"
)

// version is the release version stamped into a release build with
//
//	-ldflags "-X github.com/brettcallaghan/specs/cli/cmd.version=<v>"
//
// It is empty in every other build, which then reports the module
// version the toolchain recorded ("(devel)" for a local build).
var version string

// versionReport is the answer to `specs --version` (AC-SPEC-011.7): a
// mapping rendered by [renderYAML], keys in declaration order. Revision,
// Time and Modified are the binary's version-control build settings and
// are null where the build recorded none (for example a `go install`
// of a tagged module carries no VCS settings).
type versionReport struct {
	Version  string  `json:"version"`
	Revision *string `json:"revision"`
	Time     *string `json:"time"`
	Modified *bool   `json:"modified"`
}

// buildVersionReport assembles the report from stamped (the -X value,
// may be empty) and info (may be nil when build information is
// unavailable).
func buildVersionReport(stamped string, info *debug.BuildInfo) versionReport {
	report := versionReport{Version: stamped}
	if info == nil {
		if report.Version == "" {
			report.Version = "(devel)"
		}
		return report
	}
	if report.Version == "" {
		report.Version = info.Main.Version
	}
	if report.Version == "" {
		report.Version = "(devel)"
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			v := s.Value
			report.Revision = &v
		case "vcs.time":
			v := s.Value
			report.Time = &v
		case "vcs.modified":
			if b, err := strconv.ParseBool(s.Value); err == nil {
				report.Modified = &b
			}
		}
	}
	return report
}

// installVersionFlag gives the root a --version flag that prints the
// build report as YAML and exits 0 before any subcommand dispatch.
func installVersionFlag(root *cobra.Command) {
	info, _ := debug.ReadBuildInfo()
	root.Version = "build"
	root.SetVersionTemplate(renderYAML(buildVersionReport(version, info)))
	root.Flags().Bool("version", false,
		"print the build's version, revision, time and modified state as YAML")
}
