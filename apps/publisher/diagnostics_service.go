package main

import (
	"bytes"
	"catsoup/publisher/internal/diagnostics"
	"encoding/json"
	"errors"
)

// DiagnosticsService exposes bounded, structured logs and fixed native actions.
type DiagnosticsService struct {
	store    *diagnostics.Store
	copyText func(string) bool
	reveal   func(string) error
	export   func([]byte) (bool, error)
}

func (s *DiagnosticsService) Query(q diagnostics.Query) (diagnostics.Result, error) {
	return s.store.Query(q)
}
func (s *DiagnosticsService) Listen(enabled bool) { s.store.Listen(enabled) }
func (s *DiagnosticsService) OpenFolder() error   { return s.reveal(s.store.Directory()) }
func (s *DiagnosticsService) CopyEvent(id int64) error {
	if id < 1 || id > 9007199254740990 {
		return errors.New("invalid event")
	}
	result, err := s.store.Query(diagnostics.Query{Before: id + 1})
	if err != nil {
		return err
	}
	for _, event := range result.Events {
		if event.ID == id {
			raw, err := json.MarshalIndent(event, "", "  ")
			if err != nil {
				return err
			}
			if !s.copyText(string(raw)) {
				return errors.New("could not copy event")
			}
			return nil
		}
	}
	return errors.New("event expired; refresh the console")
}

// CopyResults copies chronological JSONL, including matches beyond the UI page.
// A zero count leaves the clipboard intact.
func (s *DiagnosticsService) CopyResults(q diagnostics.Query, minutes int) (int, error) {
	result, err := s.store.CopyResults(q, minutes)
	if err != nil {
		return 0, err
	}
	if len(result.Events) == 0 {
		return 0, nil
	}
	var raw bytes.Buffer
	encoder := json.NewEncoder(&raw)
	for i := len(result.Events) - 1; i >= 0; i-- {
		if err := encoder.Encode(result.Events[i]); err != nil {
			return 0, err
		}
		if raw.Len() > 8*1024*1024 {
			return 0, errors.New("copy exceeds 8 MB; narrow the filters or time range")
		}
	}
	if !s.copyText(raw.String()) {
		return 0, errors.New("could not copy diagnostics")
	}
	return len(result.Events), nil
}
func (s *DiagnosticsService) Export(q diagnostics.Query) (bool, error) {
	result, err := s.store.Query(q)
	if err != nil {
		return false, err
	}
	raw := []byte{}
	for i := len(result.Events) - 1; i >= 0; i-- {
		line, err := json.Marshal(result.Events[i])
		if err != nil {
			return false, err
		}
		raw = append(raw, line...)
		raw = append(raw, '\n')
	}
	return s.export(raw)
}
