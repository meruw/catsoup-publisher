// Package delivery reads production packages without modifying or approving them.
// It deliberately has no dependency on the desktop framework.
package delivery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var identifier = regexp.MustCompile(`^catsoup(?:-[a-z0-9]+)+$`)

type Video struct {
	ID        string  `json:"id"`
	Lang      *string `json:"lang"`
	Kind      string  `json:"kind"`
	Narration bool    `json:"narration"`
	Path      string  `json:"path"`
	SHA256    string  `json:"sha256"`
	Measured  struct {
		Duration   float64 `json:"duration_s"`
		Resolution string  `json:"resolution"`
		FPS        float64 `json:"fps"`
	} `json:"measured"`
}
type Text struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Manifest struct {
	Contract        string   `json:"contract"`
	ContractVersion int      `json:"contract_version"`
	ID              string   `json:"episode_id"`
	Title           string   `json:"internal_title"`
	Version         string   `json:"version"`
	Status          string   `json:"status"`
	Publishable     bool     `json:"publishable"`
	Reason          string   `json:"not_publishable_reason"`
	Languages       []string `json:"languages"`
	Videos          []Video  `json:"videos"`
	Texts           []Text   `json:"texts"`
}
type Episode struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason"`
	Ready       bool     `json:"ready"`
	Languages   []string `json:"languages"`
	Videos      []Video  `json:"videos"`
	Fingerprint string   `json:"fingerprint"`
	Warnings    []string `json:"warnings"`
}
type Entry struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Valid   bool   `json:"valid"`
	Problem string `json:"problem"`
}
type Verdict struct {
	Errors      []string `json:"errors"`
	Warnings    []string `json:"warnings"`
	Publishable bool     `json:"publishable"`
}
type Validator interface {
	Check(context.Context, string) (Verdict, error)
}
type Reader struct {
	root      string
	validator Validator
}

func New(root string, validator Validator) (*Reader, error) {
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, err
	}
	if validator == nil {
		return nil, errors.New("delivery validator is required")
	}
	return &Reader{root: canonical, validator: validator}, nil
}
func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
func (r *Reader) manifestPath(id string) (string, error) {
	if !identifier.MatchString(id) {
		return "", errors.New("invalid episode identifier")
	}
	parent := filepath.Join(r.root, "deliveries")
	path, err := filepath.EvalSymlinks(filepath.Join(parent, id, "manifest.json"))
	if err != nil {
		return "", errors.New("delivery manifest not found")
	}
	if !within(parent, path) {
		return "", errors.New("manifest escapes deliveries directory")
	}
	if filepath.Base(filepath.Dir(path)) != id {
		return "", errors.New("manifest folder does not match episode")
	}
	return path, nil
}
func (r *Reader) resolve(base, id, relative string) (string, error) {
	if relative == "" || filepath.IsAbs(relative) || strings.Contains(relative, ":") || strings.HasPrefix(relative, `\`) {
		return "", errors.New("relative delivery path required")
	}
	path, err := filepath.EvalSymlinks(filepath.Join(base, filepath.FromSlash(relative)))
	if err != nil {
		return "", errors.New("delivery file missing")
	}
	allowed := within(base, path)
	if id == "catsoup-poc-streetlight" {
		for _, dir := range []string{"render", "audio", "textos"} {
			allowed = allowed || within(filepath.Join(r.root, dir), path)
		}
	}
	if !allowed {
		return "", errors.New("delivery path escapes its allowed folder")
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("delivery file must be a regular file")
	}
	return path, nil
}
func readManifest(path string) (Manifest, []byte, error) {
	var man Manifest
	file, err := os.Open(path)
	if err != nil {
		return man, nil, err
	}
	defer file.Close()
	raw, err := io.ReadAll(io.LimitReader(file, 2*1024*1024+1))
	if err != nil {
		return man, nil, err
	}
	if len(raw) > 2*1024*1024 {
		return man, nil, errors.New("manifest exceeds 2 MiB")
	}
	if err = json.Unmarshal(raw, &man); err != nil {
		return man, nil, errors.New("invalid manifest JSON")
	}
	return man, raw, nil
}
func (r *Reader) Inspect(id string) (Episode, error) {
	var episode Episode
	path, err := r.manifestPath(id)
	if err != nil {
		return episode, err
	}
	man, raw, err := readManifest(path)
	if err != nil {
		return episode, err
	}
	if man.Contract != "catsoup-delivery" || man.ContractVersion != 2 || man.ID != id || man.Version == "" {
		return episode, errors.New("delivery contract v2 and matching episode ID required")
	}
	if len(man.Languages) == 0 || len(man.Videos) == 0 {
		return episode, errors.New("delivery must contain languages and videos")
	}
	langs := map[string]bool{}
	for _, lang := range man.Languages {
		if (lang != "es" && lang != "en") || langs[lang] {
			return episode, errors.New("unique ES/EN languages required")
		}
		langs[lang] = true
	}
	ids := map[string]bool{}
	files := []Text{}
	for _, video := range man.Videos {
		if video.Kind != "final" && video.Kind != "clean" {
			return episode, errors.New("invalid video kind")
		}
		if strings.ToLower(filepath.Ext(video.Path)) != ".mp4" {
			return episode, errors.New("video must be MP4")
		}
		files = append(files, Text{ID: video.ID, Path: video.Path, SHA256: video.SHA256})
	}
	files = append(files, man.Texts...)
	// Resolve before calling the shared validator, which reads every declared file.
	for _, item := range files {
		if item.ID == "" || ids[item.ID] {
			return episode, errors.New("missing or duplicate item ID")
		}
		ids[item.ID] = true
		if _, err = r.resolve(filepath.Dir(path), id, item.Path); err != nil {
			return episode, err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	verdict, err := r.validator.Check(ctx, path)
	if err != nil {
		return episode, err
	}
	if len(verdict.Errors) > 0 {
		return episode, fmt.Errorf("invalid delivery: %s", strings.Join(verdict.Errors, "; "))
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return episode, err
	}
	if string(current) != string(raw) {
		return episode, errors.New("delivery changed during validation; refresh")
	}
	digest := sha256.Sum256(raw)
	episode = Episode{ID: id, Title: man.Title, Version: man.Version, Status: man.Status, Reason: man.Reason,
		Ready: verdict.Publishable && id != "catsoup-poc-streetlight", Languages: man.Languages, Videos: man.Videos,
		Fingerprint: hex.EncodeToString(digest[:]), Warnings: verdict.Warnings}
	if episode.Title == "" {
		episode.Title = id
	}
	if episode.Warnings == nil {
		episode.Warnings = []string{}
	}
	return episode, nil
}
func (r *Reader) List() ([]Entry, error) {
	paths, err := filepath.Glob(filepath.Join(r.root, "deliveries", "*", "manifest.json"))
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(paths))
	for _, path := range paths {
		id := filepath.Base(filepath.Dir(path))
		entry := Entry{ID: id, Title: id}
		episode, err := r.Inspect(id)
		if err != nil {
			entry.Problem = err.Error()
		} else {
			entry.Title = episode.Title
			entry.Valid = true
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
func (r *Reader) Media(id, videoID string) (string, error) {
	episode, err := r.Inspect(id)
	if err != nil {
		return "", err
	}
	path, err := r.manifestPath(id)
	if err != nil {
		return "", err
	}
	for _, v := range episode.Videos {
		if v.ID == videoID {
			return r.resolve(filepath.Dir(path), id, v.Path)
		}
	}
	return "", errors.New("unknown video identifier")
}
