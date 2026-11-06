package main

import (
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
