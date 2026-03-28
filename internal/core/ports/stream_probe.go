package ports

import "context"

// StreamProbe checks whether a stream URL responds (best-effort).
type StreamProbe interface {
	Probe(ctx context.Context, streamURL string) (ok bool, inconclusive bool)
}
