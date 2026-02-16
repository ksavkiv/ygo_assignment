package destination

import (
	"context"
	"encoding/json"
	"strings"
)

// StubFetcher is a placeholder that returns static data.
// Replace with real external API calls (weather, attractions, etc.).
type StubFetcher struct{}

func NewStubFetcher() *StubFetcher {
	return &StubFetcher{}
}

func (f *StubFetcher) Fetch(_ context.Context, city string) (*Destination, error) {
	return &Destination{
		City:     strings.ToLower(city),
		Country:  "unknown",
		Metadata: json.RawMessage(`{"source":"stub"}`),
	}, nil
}
