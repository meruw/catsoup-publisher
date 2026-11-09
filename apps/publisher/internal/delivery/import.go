package delivery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ImportResult struct {
	ID       string `json:"id"`
	Existing bool   `json:"existing"`
}

// Import copies only declared package files. Validation finishes before the
// episode becomes visible; the source and all approval records stay unchanged.
func (r *Reader) Import(source string) (ImportResult, error) {
	var result ImportResult
	src, err := os.OpenRoot(source)
	if err != nil {
		return result, errors.New("Choose the episode folder containing manifest.json")
	}
	defer src.Close()
	f, err := src.Open("manifest.json")
	if err != nil {
		return result, errors.New("This folder has no manifest.json. Choose the episode delivery folder")
	}
	raw, err := io.ReadAll(io.LimitReader(f, 2*1024*1024+1))
	f.Close()
	if err != nil || len(raw) > 2*1024*1024 {
		return result, errors.New("Could not read delivery manifest (maximum 2 MiB)")
	}
	var man Manifest
	if json.Unmarshal(raw, &man) != nil || !identifier.MatchString(man.ID) || man.Contract != "catsoup-delivery" || man.ContractVersion != 2 {
		return result, errors.New("A valid delivery contract v2 is required")
	}
	parent := filepath.Join(r.root, "deliveries")
	if err = os.MkdirAll(parent, 0700); err != nil {
		return result, err
	}
	canonical, err := filepath.EvalSymlinks(parent)
	if err != nil || !within(r.root, canonical) {
		return result, errors.New("Delivery directory escapes the workspace")
	}
	parent = canonical
	target := filepath.Join(parent, man.ID)
	if _, err = os.Lstat(target); err == nil {
		path, pathErr := r.manifestPath(man.ID)
		if pathErr == nil {
			saved, readErr := os.ReadFile(path)
			if readErr == nil && bytes.Equal(raw, saved) {
				if _, inspectErr := r.Inspect(man.ID); inspectErr == nil {
					return ImportResult{ID: man.ID, Existing: true}, nil
				}
			}
		}
		return result, errors.New("This episode already exists with different or invalid files. Nothing was replaced")
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	files := append([]Text(nil), man.Texts...)
	for _, v := range man.Videos {
		files = append(files, Text{Path: v.Path, SHA256: v.SHA256})
	}
	if len(files) == 0 || len(files) > 256 {
		return result, errors.New("Delivery must declare between 1 and 256 files")
	}
	stage, err := os.MkdirTemp(parent, ".import-")
	if err != nil {
		return result, err
	}
	// Only this freshly allocated directory is removed, never a selected source.
	defer os.RemoveAll(stage)
	base := filepath.Join(stage, "deliveries", man.ID)
	if err = os.MkdirAll(base, 0700); err != nil {
		return result, err
	}
	seen := map[string]bool{"manifest.json": true}
	var total int64
	for _, item := range files {
		rel := filepath.FromSlash(item.Path)
		if !filepath.IsLocal(rel) || strings.Contains(rel, ":") {
			return result, errors.New("Package files must use relative paths inside the delivery folder")
		}
		key := strings.ToLower(filepath.Clean(rel))
		if seen[key] {
			return result, errors.New("Duplicate package file path")
		}
		seen[key] = true
		hash, hashErr := hex.DecodeString(item.SHA256)
		if hashErr != nil || len(hash) != sha256.Size {
			return result, errors.New("Package file has no valid SHA-256 hash")
		}
		input, openErr := src.Open(rel)
		if openErr != nil {
			return result, fmt.Errorf("Cannot read package file %q", item.Path)
		}
		info, statErr := input.Stat()
		if statErr != nil || !info.Mode().IsRegular() {
			input.Close()
			return result, errors.New("Package files must be regular files")
		}
		total += info.Size()
		if total > 64<<30 {
			input.Close()
			return result, errors.New("Delivery exceeds the 64 GiB import limit")
		}
		dest := filepath.Join(base, rel)
		if err = os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
			input.Close()
			return result, err
		}
		output, openErr := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if openErr != nil {
			input.Close()
			return result, openErr
		}
		digest := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(output, digest), io.LimitReader(input, info.Size()+1))
		closeErr := output.Close()
		input.Close()
		if copyErr != nil || closeErr != nil || n != info.Size() || !bytes.Equal(digest.Sum(nil), hash) {
			return result, fmt.Errorf("File %q changed or does not match its manifest; export the delivery again", item.Path)
		}
	}
	if err = os.WriteFile(filepath.Join(base, "manifest.json"), raw, 0600); err != nil {
		return result, err
	}
	staged, err := New(stage, r.validator)
	if err != nil {
		return result, err
	}
	if _, err = staged.Inspect(man.ID); err != nil {
		return result, err
	}
	// Rename on the same volume exposes a complete package without touching drafts.
	if _, err = os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		return result, errors.New("Episode appeared during import; nothing was replaced")
	}
	if err = os.Rename(base, target); err != nil {
		return result, fmt.Errorf("Could not finish importing the delivery: %w", err)
	}
	return ImportResult{ID: man.ID}, nil
}
