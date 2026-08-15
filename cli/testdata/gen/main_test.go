// The gate for the generator's own correctness: regenerate must overwrite
// every artifact a case declares, and must not write one it does not.
//
// This program's silent-skip class of bug has recurred twice with the same
// signature: it reports "regenerated <case>" for every case while some
// artifacts were never written, either because the case directory was
// resolved relatively after chdir'ing into a scratch dir (so every write
// landed in a directory this program then deleted), or because a declared
// artifact kind had no writer branch at all. Neither leaves any trace in
// stdout or the exit status — the only observable is the file on disk not
// changing. So the gate corrupts a copy of every artifact a case declares,
// runs regenerate over the copy, and requires each one to have been
// overwritten; it also requires no artifact the case does not declare to
// exist afterward, catching the fix's own bug of writing a --json payload
// under a checks.log name.
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

// knownArtifactKinds is every artifact name regenerate can write. A case
// declaring one of these must see it change; a case not declaring one must
// never see it appear.
func knownArtifactKinds() []string {
	kinds := append([]string{}, stdoutArtifacts...)
	kinds = append(kinds, stderrArtifact, manifestArtifact)
	return kinds
}

// TestRegenerateOverwritesEveryDeclaredArtifact replays every case under
// cli/testdata/cases from a corrupted copy of its expected/ directory and
// requires regenerate to have overwritten each artifact the case declares,
// and to have written no artifact it does not.
func TestRegenerateOverwritesEveryDeclaredArtifact(t *testing.T) {
	const casesDir = "../cases"

	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatalf("reading %s: %v", casesDir, err)
	}

	corruption := []byte("CORRUPTED-BY-GATE-TEST\n")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			// The scratch copy lives under "." (this package's own
			// directory), NOT under t.TempDir() (which os.MkdirTemp puts
			// under the OS temp dir, several directory levels shallower
			// than this package). That depth difference matters: a
			// relative caseDir resolved from a shallower cwd after
			// regenerate's internal chdir happens to walk back "up to
			// root and down again" and land on the right place anyway,
			// which would mask exactly the bug this test exists to catch.
			// A same-depth, pure-descent path — the shape main() actually
			// builds from a relative casesDir argument — does not get
			// that accidental rescue.
			scratch, err := os.MkdirTemp(".", "gate-scratch-*")
			if err != nil {
				t.Fatalf("creating scratch dir: %v", err)
			}
			t.Cleanup(func() { os.RemoveAll(scratch) })

			caseDir := filepath.Join(scratch, name)
			if err := copyTree(filepath.Join(casesDir, name), caseDir); err != nil {
				t.Fatalf("copying fixture case into scratch: %v", err)
			}

			var manifest caseManifest
			if _, err := toml.DecodeFile(filepath.Join(caseDir, "case.toml"), &manifest); err != nil {
				t.Fatalf("decoding case.toml: %v", err)
			}
			if len(manifest.Artifacts) == 0 {
				t.Fatalf("case %s declares no artifacts in case.toml; this gate has nothing to verify for it", name)
			}

			// Corrupt every artifact the case declares. Anything still
			// holding this exact content after regenerate ran was never
			// written.
			for _, artifact := range manifest.Artifacts {
				p := filepath.Join(caseDir, "expected", artifact)
				if err := os.WriteFile(p, corruption, 0o644); err != nil {
					t.Fatalf("corrupting %s: %v", p, err)
				}
			}

			// caseDir is already relative (scratch is "gate-scratch-*",
			// not "./gate-scratch-*" — MkdirTemp cleans the leading
			// "./"), the same shape main() gives regenerate: os.Args[1]
			// joined with the entry name, descending from the working
			// directory the program was launched in. regenerate chdirs
			// internally; an already-absolute caseDir would resolve
			// correctly regardless of the chdir and could never reproduce
			// the historical bug, where a relative dir was read again
			// after the chdir and silently missed every file.
			if filepath.IsAbs(caseDir) {
				t.Fatalf("caseDir %q is absolute; this test requires a relative path to exercise the chdir hazard", caseDir)
			}

			if err := regenerate(caseDir, name); err != nil {
				t.Fatalf("regenerate: %v", err)
			}

			for _, artifact := range manifest.Artifacts {
				p := filepath.Join(caseDir, "expected", artifact)
				got, err := os.ReadFile(p)
				if err != nil {
					t.Fatalf("reading regenerated %s: %v", p, err)
				}
				if bytes.Equal(got, corruption) {
					t.Errorf("regenerate left %s corrupted — it declares this artifact but did not write it", artifact)
				}
			}

			for _, kind := range knownArtifactKinds() {
				if manifest.declares(kind) {
					continue
				}
				p := filepath.Join(caseDir, "expected", kind)
				if _, err := os.Stat(p); err == nil {
					t.Errorf("regenerate wrote %s, which case.toml does not declare", kind)
				}
			}
		})
	}
}
