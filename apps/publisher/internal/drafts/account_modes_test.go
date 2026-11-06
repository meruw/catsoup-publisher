package drafts

import (
	"path/filepath"
	"testing"
)

func TestXModeSurvivesRestartAndDowngrade(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	browser := Profile{Platform: "x", Language: "es", ID: "cats_es", Name: "@cats_es"}
	api := Profile{Platform: "x", Language: "en", ID: "123", Name: "@cats_en"}
	if err = s.PutProfile(browser, []byte{}, ""); err != nil {
		t.Fatal(err)
	}
	if err = s.PutProfile(api, []byte("sealed-fixture"), ""); err != nil {
		t.Fatal(err)
	}
	s.Close()
	s, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	rows, err := s.Profiles()
	if err != nil || len(rows) != 2 {
		t.Fatal(rows, err)
	}
	for _, p := range rows {
		if p.API != (p.Language == "en") {
			t.Fatal("incorrect mode", p)
		}
	}
	api.ID = "cats_en"
	if err = s.PutProfile(api, []byte{}, "123"); err != nil {
		t.Fatal(err)
	}
	raw, err := s.Credential("x", "en", "cats_en")
	if err != nil || len(raw) != 0 {
		t.Fatal("API token retained on downgrade", err)
	}
	if _, err = s.Credential("x", "en", "123"); err == nil {
		t.Fatal("old identity still reachable")
	}
	rows, err = s.Profiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range rows {
		if p.API {
			t.Fatal("API mode survived downgrade")
		}
	}
}
