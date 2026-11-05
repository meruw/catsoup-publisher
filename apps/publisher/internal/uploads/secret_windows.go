package uploads

import (
	"errors"
	"golang.org/x/sys/windows"
	"unsafe"
)

// DPAPI binds secrets to the current Windows user. The entropy binds each blob
// to its destination so a swapped credential cannot be used for another slot.
func protect(data []byte, scope string, encrypt bool) ([]byte, error) {
	if len(data) == 0 || len(scope) == 0 {
		return nil, errors.New("empty account credential")
	}
	entropy := []byte(scope)
	in := windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
	extra := windows.DataBlob{Size: uint32(len(entropy)), Data: &entropy[0]}
	var out windows.DataBlob
	var err error
	if encrypt {
		err = windows.CryptProtectData(&in, nil, &extra, 0, nil, 1, &out)
	} else {
		err = windows.CryptUnprotectData(&in, nil, &extra, 0, nil, 1, &out)
	}
	if err != nil {
		return nil, errors.New("Windows could not unlock this account; reconnect it from Accounts")
	}
	defer windows.LocalFree(windows.Handle(uintptr(unsafe.Pointer(out.Data))))
	return append([]byte(nil), unsafe.Slice(out.Data, int(out.Size))...), nil
}
