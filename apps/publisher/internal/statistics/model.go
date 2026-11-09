// Package statistics defines provider-neutral measurements, without credentials.
package statistics

// Decimal strings preserve exact counters across the desktop bridge.
type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Sample struct {
	Items     []Metric `json:"items"`
	UpdatedAt string   `json:"updatedAt"`
}
