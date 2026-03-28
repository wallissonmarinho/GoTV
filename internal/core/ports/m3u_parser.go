package ports

import "github.com/wallissonmarinho/GoTV/internal/core/domain"

// M3UParser parses M3U text into channels.
type M3UParser interface {
	Parse(content string) ([]domain.Channel, error)
}
