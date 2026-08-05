//go:build !windows

package corpus

import "os"

// openShared opens path for reading.
//
// Everywhere but Windows this is os.Open and there is nothing to arrange:
// a rename replaces the destination's directory entry without consulting
// anyone holding the file, and a handle opened before the replacement
// keeps referring to what it opened. The Windows implementation exists to
// obtain those same two properties; see open_windows.go for why they have
// to be asked for there.
//
// AC: AC-SPEC-010.4
func openShared(path string) (*os.File, error) { return os.Open(path) }
