package ports

import "context"

// HTTPGetter downloads remote playlist or EPG bodies.
type HTTPGetter interface {
	Get(ctx context.Context, url string) ([]byte, error)
}
