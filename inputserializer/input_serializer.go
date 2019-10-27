package inputserializer

import "context"

// InputSerializer common interface for all inputseralizer modules
type InputSerializer interface {
	GetSerializerCh(ctx context.Context) (<-chan string, error)
}
