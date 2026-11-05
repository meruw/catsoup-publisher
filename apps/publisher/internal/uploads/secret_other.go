//go:build !windows

package uploads

import "errors"

func protect([]byte, string, bool) ([]byte, error) {
	return nil, errors.New("account storage requires Windows")
}
