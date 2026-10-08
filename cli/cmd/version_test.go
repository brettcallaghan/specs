package cmd

import (
	"runtime/debug"
	"strings"
	"testing"
)

// AC: AC-SPEC-011.7
func TestVersionReport(t *testing.T) {
	t.Run("ldflags version overrides the module version", func(t *testing.T) {
		info := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}
		got := renderYAML(buildVersionReport("v1.2.3", info))
		want := "version: v1.2.3\nrevision: null\ntime: null\nmodified: null\n"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
	t.Run("module version when nothing is stamped", func(t *testing.T) {
		info := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}
		if got := buildVersionReport("", info).Version; got != "(devel)" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("absent VCS settings render null", func(t *testing.T) {
		got := renderYAML(buildVersionReport("", &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}))
		for _, k := range []string{"revision: null", "time: null", "modified: null"} {
			if !strings.Contains(got, k) {
				t.Fatalf("missing %q in %q", k, got)
			}
		}
	})
	t.Run("vcs.modified true renders true", func(t *testing.T) {
		info := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}, Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abc123"},
			{Key: "vcs.time", Value: "2026-10-08T00:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
		}}
		got := renderYAML(buildVersionReport("", info))
		if !strings.Contains(got, "modified: true") || !strings.Contains(got, "revision: abc123") {
			t.Fatalf("got %q", got)
		}
	})
}
