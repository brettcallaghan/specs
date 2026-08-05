//go:build windows

package corpus

import (
	"os"
	"syscall"
)

// openShared opens path for reading in a way that does not block a
// concurrent atomic replacement of it.
//
// This is the root fix for the Windows half of AC-SPEC-010.4, and the
// reason it cannot be os.Open. Go opens files with FILE_SHARE_READ and
// FILE_SHARE_WRITE but *not* FILE_SHARE_DELETE (golang/go#32088), and
// MoveFileEx cannot replace a file whose open handles withhold delete
// sharing — so on Windows an ordinary reader of the index vetoes the very
// rewrite this package performs. Waiting the veto out only works while
// readers leave gaps; a reader that reopens the file continuously blocks
// the writer indefinitely, which is livelock rather than contention.
//
// Granting FILE_SHARE_DELETE removes the veto and, in doing so, buys the
// POSIX semantics the rest of this package assumes: the replacement
// proceeds, and this handle goes on referring to the file it opened —
// the old, complete index — rather than seeing bytes change underneath
// it. A reader still observes one whole version or the other, which is
// exactly what the criterion requires.
//
// AC: AC-SPEC-010.4
func openShared(path string) (*os.File, error) {
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}

	handle, err := syscall.CreateFile(
		name,
		syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		// A *os.PathError wrapping the raw Errno, which is the shape the
		// rest of the package matches on: os.IsNotExist keeps working for
		// callers, and sharingConflict can still see ERROR_SHARING_VIOLATION
		// underneath.
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}
