package inputserializer

import "context"

// InputSerializer common interface for all inputseralizer modules
type InputSerializer interface {
	GetSeralizerCh(ctx context.Context) (<-chan string, error)
}
