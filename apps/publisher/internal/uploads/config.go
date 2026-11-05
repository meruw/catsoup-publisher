package uploads

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func Configure(platform, raw string) error {
	if len(raw) > 32768 {
		return errors.New("client configuration is too large")
	}
	var data struct {
		Installed struct {
			ID     string `json:"client_id"`
			Secret string `json:"client_secret"`
		}
		Key       string `json:"client_key"`
		Secret    string `json:"client_secret"`
		App       string `json:"app_id"`
		AppSecret string `json:"app_secret"`
		Config    string `json:"config_id"`
	}
	if json.Unmarshal([]byte(raw), &data) != nil {
		return errors.New("select a valid client JSON file")
	}
	var file string
	switch platform {
	case "youtube":
		if data.Installed.ID == "" || data.Installed.Secret == "" {
			return errors.New("select Google Desktop app JSON")
		}
		file = "google-desktop.json"
	case "tiktok":
		if data.Key == "" || data.Secret == "" {
			return errors.New("client_key and client_secret are required")
		}
		file = "tiktok-desktop.json"
	case "instagram":
		if data.App == "" || data.AppSecret == "" || data.Config == "" {
			return errors.New("app_id, app_secret and config_id are required")
		}
		file = "meta-desktop.json"
	default:
		return errors.New("unsupported client configuration")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir = filepath.Join(dir, "CATsoup Publisher")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "client-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.WriteString(raw)
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, file))
}
