//go:build windows

// The Windows half of the atomic replacement in cache.go. Replacing a
// file that readers hold open takes two accommodations here, and neither
// is optional: the reader has to stop vetoing the replacement
// ([openShared]), and the replacement has to stop being refused for the
// name still being in use ([replaceFile]). Clearing only the first is not
// progress — measured on the Windows runner, it left the writer failing
// exactly as before.
package corpus

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Access rights and information classes the portable syscall package does
// not name. Values from the Windows SDK headers (winnt.h, minwinbase.h,
// WinBase.h).
const (
	deleteAccess = 0x00010000 // DELETE
	synchronize  = 0x00100000 // SYNCHRONIZE

	// fileRenameInfoEx is FILE_INFO_BY_HANDLE_CLASS 22, the rename class
	// whose payload carries a flags word where the original carries a bare
	// ReplaceIfExists boolean. The flags are what this file is here for.
	fileRenameInfoEx = 22

	fileRenameReplaceIfExists = 0x00000001
	fileRenamePosixSemantics  = 0x00000002
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procSetFileInformationByHandle = kernel32.NewProc("SetFileInformationByHandle")
)

// fileRenameInfo mirrors FILE_RENAME_INFO. FileName is declared as a
// one-element array because the real structure is variable-length: the
// name follows the header inline, and callers allocate a buffer sized for
// it rather than using this type directly.
type fileRenameInfo struct {
	Flags          uint32
	RootDirectory  syscall.Handle
	FileNameLength uint32
	FileName       [1]uint16
}

// renameInfoLayout exists only so [unsafe.Offsetof] has something to
// measure. The name's offset is not the header's size — the structure
// carries trailing padding on 64-bit — so computing it by subtraction
// would place the name two bytes late and rename to a truncated path.
var renameInfoLayout fileRenameInfo

// openShared opens path for reading without blocking a concurrent
// replacement of it.
//
// Go opens files with FILE_SHARE_READ and FILE_SHARE_WRITE but not
// FILE_SHARE_DELETE (golang/go#32088, declined in 2019 for lack of
// consensus, with no opt-in added since). A replacement needs DELETE
// access on the destination, so an ordinary Go reader refuses it outright
// — the reader vetoes the write.
//
// Granting FILE_SHARE_DELETE withdraws the veto. It is necessary and, on
// its own, not sufficient: see [replaceFile] for the second gate. It also
// buys the property the rest of the package assumes, that this handle
// goes on referring to the file it opened rather than seeing bytes change
// underneath it.
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
		// callers, and sharingConflict can still see the code underneath.
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

// replaceFile renames source onto dest, replacing whatever is there.
//
// os.Rename is MoveFileEx(MOVEFILE_REPLACE_EXISTING) on Windows, in every
// Go release through 1.26, and MoveFileEx will not replace a file that
// has open handles at all. The rule is stated without a share-mode
// exception: "A file cannot be renamed if a file with the same name
// exists and has open handles" (FILE_RENAME_INFORMATION, "Special rules
// for renaming open files"). Sharing flags on the reader do not reach
// this — which is why [openShared] alone did not fix it.
//
// FILE_RENAME_POSIX_SEMANTICS is the documented exception, and it is
// documented in precisely the terms this package needs: with
// FILE_RENAME_REPLACE_IF_EXISTS it will "allow replacing a file even if
// there are existing handles to it. Existing handles to the replaced file
// continue to be valid for operations such as read and write. Any
// subsequent opens of the target name will open the renamed file, not the
// replaced file." That is the POSIX rename this code was written against,
// available through SetFileInformationByHandle's FileRenameInfoEx class
// and not through os.Rename.
//
// MoveFileEx is still tried first, and the POSIX path is the fallback
// rather than the default. It is the better-trodden call, it is what
// succeeds whenever nothing holds the destination, and keeping it in
// front means the new code runs only on the contended path it exists for.
// Rust's standard library arranges it the same way. When the fallback
// cannot help either — a filesystem without the class, which reports
// ERROR_INVALID_PARAMETER — the original MoveFileEx error is what the
// caller sees, because that is the failure that describes their machine.
//
// AC: AC-SPEC-010.4
func replaceFile(source, dest string) error {
	err := os.Rename(source, dest)
	if err == nil || !os.IsPermission(err) {
		return err
	}
	if posix := renamePosix(source, dest); posix == nil {
		return nil
	}
	return err
}

// renamePosix performs the replacement with POSIX semantics, which
// requires a handle to the *source* opened for DELETE — the rename is an
// operation on the file being moved, not on the directory.
func renamePosix(source, dest string) error {
	absolute, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	target, err := syscall.UTF16FromString(absolute)
	if err != nil {
		return err
	}
	// UTF16FromString terminates the string; FileNameLength counts the
	// name's own bytes, so the NUL is dropped rather than described.
	target = target[:len(target)-1]

	from, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	handle, err := syscall.CreateFile(
		from,
		deleteAccess|synchronize,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return &os.PathError{Op: "open", Path: source, Err: err}
	}
	defer syscall.CloseHandle(handle)

	const nameOffset = unsafe.Offsetof(renameInfoLayout.FileName)
	size := int(nameOffset) + len(target)*2
	buffer := make([]byte, size)

	info := (*fileRenameInfo)(unsafe.Pointer(&buffer[0]))
	info.Flags = fileRenameReplaceIfExists | fileRenamePosixSemantics
	info.RootDirectory = 0
	info.FileNameLength = uint32(len(target) * 2)
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(&buffer[nameOffset])), len(target)), target)

	ok, _, errno := procSetFileInformationByHandle.Call(
		uintptr(handle),
		fileRenameInfoEx,
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(size),
	)
	if ok == 0 {
		return &os.LinkError{Op: "rename", Old: source, New: dest, Err: errno}
	}
	return nil
}
