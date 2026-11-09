package main

import (
	"catsoup/publisher/internal/delivery"
	"errors"
	"sync"
)

// PublisherService is the narrow desktop boundary. Domain code never imports Wails.
// No generic filesystem, shell, approval or publishing operations are exposed.
type PublisherService struct {
	reader         *delivery.Reader
	chooseDelivery func() (string, error)
	importing      sync.Mutex
}

func (s *PublisherService) ListDeliveries() ([]delivery.Entry, error) { return s.reader.List() }
func (s *PublisherService) GetEpisode(id string) (delivery.Episode, error) {
	return s.reader.Inspect(id)
}

// ImportDelivery accepts no frontend path: access starts with an explicit native selection.
func (s *PublisherService) ImportDelivery() (*delivery.ImportResult, error) {
	if !s.importing.TryLock() {
		return nil, errors.New("An import is already running")
	}
	defer s.importing.Unlock()
	if s.chooseDelivery == nil {
		return nil, errors.New("Folder selection is unavailable")
	}
	path, err := s.chooseDelivery()
	if err != nil || path == "" {
		return nil, err
	}
	result, err := s.reader.Import(path)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
