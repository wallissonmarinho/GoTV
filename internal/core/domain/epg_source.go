package domain

import "time"

// EPGSource is a registered remote XMLTV URL.
type EPGSource struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	Label     string    `json:"label,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
