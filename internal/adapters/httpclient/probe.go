package httpclient

import (
	"context"
	"net/http"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// StreamProbe implements ports.StreamProbe using a ranged GET.
type StreamProbe struct {
	Client *http.Client
	UA     string
}

var _ ports.StreamProbe = (*StreamProbe)(nil)

// NewStreamProbe returns a probe with short timeout.
func NewStreamProbe(timeout time.Duration, ua string) *StreamProbe {
	return &StreamProbe{
		Client: &http.Client{Timeout: timeout},
		UA:     ua,
	}
}

// Probe tries HEAD, then ranged GET if HEAD does not succeed.
func (p *StreamProbe) Probe(ctx context.Context, streamURL string) (ok bool, inconclusive bool) {
	if headOK, _ := p.head(ctx, streamURL); headOK {
		return true, false
	}
	return p.rangedGet(ctx, streamURL)
}

func (p *StreamProbe) head(ctx context.Context, streamURL string) (ok bool, inconclusive bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, streamURL, nil)
	if err != nil {
		return false, false
	}
	if p.UA != "" {
		req.Header.Set("User-Agent", p.UA)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return false, true
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	if code == http.StatusMethodNotAllowed || code == http.StatusNotImplemented {
		return false, true
	}
	if code == http.StatusUnauthorized || code == http.StatusForbidden {
		return false, true
	}
	if code >= 200 && code < 300 {
		return true, false
	}
	if code >= 300 && code < 400 {
		return true, false
	}
	return false, code >= 500
}

func (p *StreamProbe) rangedGet(ctx context.Context, streamURL string) (ok bool, inconclusive bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return false, false
	}
	req.Header.Set("Range", "bytes=0-2048")
	if p.UA != "" {
		req.Header.Set("User-Agent", p.UA)
	}
	resp, err := p.Client.Do(req)
	if err != nil {
		return false, true
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	if code == http.StatusUnauthorized || code == http.StatusForbidden {
		return false, true
	}
	if code >= 200 && code < 300 {
		return true, false
	}
	if code == http.StatusRequestedRangeNotSatisfiable {
		return true, false
	}
	return false, code >= 500
}
