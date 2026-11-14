package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type InstagramHosting struct {
	Enabled   bool `json:"enabled"`
	Available bool `json:"available"`
}

func (s *UploadService) InstagramHosting() InstagramHosting {
	if s.root == "" {
		return InstagramHosting{}
	}
	var value InstagramHosting
	raw, e := os.ReadFile(filepath.Join(s.root, "data", "publisher", "instagram-https.json"))
	if e == nil {
		if json.Unmarshal(raw, &value) != nil {
			value = InstagramHosting{}
		}
	}
	info, e := os.Stat(filepath.Join(s.root, ".tools", "cloudflared.exe"))
	value.Available = e == nil && info.Mode().IsRegular()
	return value
}
func (s *UploadService) SetInstagramHosting(enabled, consent bool) error {
	if s.root == "" {
		return errors.New("workspace unavailable")
	}
	if enabled && (!consent || !s.InstagramHosting().Available) {
		return errors.New("install cloudflared and accept temporary file sharing before enabling HTTPS")
	}
	path := filepath.Join(s.root, "data", "publisher", "instagram-https.json")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	raw, _ := json.Marshal(struct {
		Enabled bool `json:"enabled"`
	}{enabled})
	temp, e := os.CreateTemp(filepath.Dir(path), "hosting-*.tmp")
	if e != nil {
		return e
	}
	defer os.Remove(temp.Name())
	_, e = temp.Write(raw)
	if e == nil {
		e = temp.Sync()
	}
	ce := temp.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	return os.Rename(temp.Name(), path)
}
