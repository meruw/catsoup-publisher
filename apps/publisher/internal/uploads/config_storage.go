package uploads

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const developerHeader = "CATSOUP-DEVELOPER-DPAPI-v1\n"
const maxDeveloperConfig = 32768

var developerMu sync.Mutex
var developerFiles = []string{"google-desktop.json", "tiktok-desktop.json", "meta-desktop.json", "x-desktop.json"}

func developerFile(file string) bool {
	for _, known := range developerFiles {
		if file == known {
			return true
		}
	}
	return false
}
func developerPath(file string) (string, error) {
	if !developerFile(file) {
		return "", errors.New("unsupported developer configuration")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "CATsoup Publisher", file), nil
}
func saveDeveloperConfiguration(file, raw string) error {
	developerMu.Lock()
	defer developerMu.Unlock()
	path, err := developerPath(file)
	if err != nil {
		return err
	}
	if len(raw) > maxDeveloperConfig || !json.Valid([]byte(raw)) {
		return errors.New("invalid developer configuration")
	}
	return sealDeveloperConfiguration(path, file, []byte(raw))
}

// Encrypt and verify before replacing the original. No plaintext temporary file
// or backup is written. The exact JSON bytes survive migration.
func sealDeveloperConfiguration(path, file string, raw []byte) error {
	scope := "developer-config/" + file
	encrypted, err := protect(raw, scope, true)
	if err != nil {
		return errors.New("Windows could not protect the developer credentials")
	}
	restored, err := protect(encrypted, scope, false)
	if err != nil || !bytes.Equal(restored, raw) {
		return errors.New("could not verify protected developer credentials")
	}
	return writeConfigurationBytes(path, append([]byte(developerHeader), encrypted...))
}
func writeConfigurationBytes(path string, raw []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "client-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(raw); err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmp.Name(), path)
}
func readDeveloperConfiguration(file string) ([]byte, error) {
	developerMu.Lock()
	defer developerMu.Unlock()
	path, err := developerPath(file)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(f, maxDeveloperConfig+16385))
	f.Close()
	if err != nil || len(raw) > maxDeveloperConfig+16384 {
		return nil, errors.New("invalid developer credential file")
	}
	if bytes.HasPrefix(raw, []byte(developerHeader)) {
		decoded, err := protect(raw[len(developerHeader):], "developer-config/"+file, false)
		if err != nil || len(decoded) > maxDeveloperConfig || !json.Valid(decoded) {
			return nil, errors.New("could not unlock developer credentials; use the original Windows user or replace them in Settings")
		}
		return decoded, nil
	}
	// Reject damaged/unknown envelopes; only a valid legacy JSON document migrates.
	if len(raw) == 0 || len(raw) > maxDeveloperConfig || !json.Valid(raw) {
		return nil, errors.New("invalid developer credentials; replace them in Settings")
	}
	if err := sealDeveloperConfiguration(path, file, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// MigrateDeveloperCredentials also verifies already protected files. It never
// modifies account tokens, client receipts, callback consent or publishing data.
func MigrateDeveloperCredentials() ([]string, error) {
	var ready []string
	var failures []error
	for _, file := range developerFiles {
		if _, err := readDeveloperConfiguration(file); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				failures = append(failures, fmt.Errorf("%s: developer credential protection failed", file))
			}
		} else {
			ready = append(ready, file)
		}
	}
	return ready, errors.Join(failures...)
}
