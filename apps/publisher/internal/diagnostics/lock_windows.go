package diagnostics

import (
	"golang.org/x/sys/windows"
	"path/filepath"
)

func lockDirectory(dir string) (func(), error) {
	path, err := windows.UTF16PtrFromString(filepath.Join(dir, "writer.lock"))
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, err
	}
	return func() { _ = windows.CloseHandle(handle) }, nil
}
