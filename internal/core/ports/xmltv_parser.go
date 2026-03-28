package ports

import "github.com/wallissonmarinho/GoTV/internal/core/domain"

// XMLTVParser parses XMLTV bytes.
type XMLTVParser interface {
	Parse(data []byte) (*domain.EPGData, error)
}
