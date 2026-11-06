//go:build !windows

package diagnostics

import "errors"

func lockDirectory(string) (func(), error) {
	return nil, errors.New("disk diagnostics require the Windows desktop runtime")
}
