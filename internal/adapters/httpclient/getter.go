package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Getter implements ports.HTTPGetter with size and timeout limits.
type Getter struct {
	Client      *http.Client
	UserAgent   string
	MaxBytes    int64
}

var _ ports.HTTPGetter = (*Getter)(nil)

// NewGetter builds a default getter.
func NewGetter(timeout time.Duration, userAgent string, maxBytes int64) *Getter {
	if maxBytes <= 0 {
		maxBytes = 50 << 20
	}
	return &Getter{
		Client: &http.Client{Timeout: timeout},
		UserAgent: userAgent,
		MaxBytes: maxBytes,
	}
}

// Get performs a GET request and returns the body capped at MaxBytes.
func (g *Getter) Get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if g.UserAgent != "" {
		req.Header.Set("User-Agent", g.UserAgent)
	}
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	lr := io.LimitReader(resp.Body, g.MaxBytes+1)
	body, err := io.ReadAll(lr)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > g.MaxBytes {
		return nil, fmt.Errorf("body exceeds max bytes (%d)", g.MaxBytes)
	}
	return body, nil
}
