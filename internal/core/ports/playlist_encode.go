package ports

import "github.com/wallissonmarinho/GoTV/internal/core/domain"

// M3UPlaylistEncoder serializes merged channels to M3U text.
type M3UPlaylistEncoder interface {
	Build(channels []domain.Channel) []byte
}

// XMLTVEncoder serializes EPG data to XMLTV XML.
type XMLTVEncoder interface {
	Build(data *domain.EPGData) []byte
}
