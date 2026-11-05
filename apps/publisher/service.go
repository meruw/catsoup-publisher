package main

import "catsoup/publisher/internal/delivery"

// PublisherService is the narrow desktop boundary. Domain code never imports Wails.
// No generic filesystem, shell, approval or publishing operations are exposed.
type PublisherService struct{ reader *delivery.Reader }

func (s *PublisherService) ListDeliveries() ([]delivery.Entry, error) { return s.reader.List() }
func (s *PublisherService) GetEpisode(id string) (delivery.Episode, error) {
	return s.reader.Inspect(id)
}
