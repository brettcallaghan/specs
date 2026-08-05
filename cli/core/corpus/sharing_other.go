//go:build !windows

// The POSIX half of the atomic replacement in cache.go, which is the
// half with nothing to arrange. Both properties the replacement depends
// on come for free here: a rename replaces the destination without
// consulting anyone holding it open, and a handle opened beforehand keeps
// referring to what it opened. sharing_windows.go exists to obtain those
// same two properties on a platform that grants neither by default.
package corpus

import "os"

// openShared opens path for reading.
//
// AC: AC-SPEC-010.4
func openShared(path string) (*os.File, error) { return os.Open(path) }

// replaceFile renames source onto dest, replacing whatever is there.
//
// AC: AC-SPEC-010.4
func replaceFile(source, dest string) error { return os.Rename(source, dest) }
