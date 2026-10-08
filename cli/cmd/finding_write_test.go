package cmd

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteErrorTextNamesADestinationBeneathAFile pins the wording of a
// write beneath a regular file to "not a directory" whatever cause the
// platform reports: Windows reports that case as a missing path, which
// is simulated here with fs.ErrNotExist.
//
// AC: AC-SPEC-010.1
func TestWriteErrorTextNamesADestinationBeneathAFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "specs.toml")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		path string
		want string
	}{
		{"beneath a regular file", filepath.Join(file, "specs.json"), "not a directory"},
		{"deeper beneath a regular file", filepath.Join(file, "a", "specs.json"), "not a directory"},
		{"beneath a missing directory", filepath.Join(dir, "missing", "specs.json"), "no such file"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// The error's own path is the parent, as os.MkdirAll reports
			// it, to show the test is made on the destination.
			err := &fs.PathError{Op: "mkdir", Path: filepath.Dir(c.path), Err: fs.ErrNotExist}
			if got := writeErrorText(c.path, err); got != c.want {
				t.Errorf("writeErrorText = %q, want %q", got, c.want)
			}
		})
	}
}
