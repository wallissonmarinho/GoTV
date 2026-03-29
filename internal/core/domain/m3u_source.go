package domain

import "time"

// M3USource is a registered remote M3U playlist URL.
type M3USource struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Label     string    `json:"label,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
