package m3u

import (
	"fmt"
	"strings"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Writer implements ports.M3UPlaylistEncoder.
type Writer struct{}

var _ ports.M3UPlaylistEncoder = (*Writer)(nil)

// Build returns UTF-8 M3U bytes.
func (Writer) Build(channels []domain.Channel) []byte {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	for _, c := range channels {
		b.WriteString("#EXTINF:-1")
		if id := strings.TrimSpace(c.TVGID); id != "" {
			fmt.Fprintf(&b, ` tvg-id=%q`, id)
		}
		if tn := strings.TrimSpace(c.TVGName); tn != "" {
			fmt.Fprintf(&b, ` tvg-name=%q`, tn)
		} else if c.Name != "" {
			fmt.Fprintf(&b, ` tvg-name=%q`, c.Name)
		}
		if g := strings.TrimSpace(c.Group); g != "" {
			fmt.Fprintf(&b, ` group-title=%q`, g)
		}
		if logo := strings.TrimSpace(c.Logo); logo != "" {
			fmt.Fprintf(&b, ` tvg-logo=%q`, logo)
		}
		b.WriteString(",")
		b.WriteString(strings.TrimSpace(c.Name))
		b.WriteString("\n")
		b.WriteString(strings.TrimSpace(c.URL))
		b.WriteString("\n")
	}
	return []byte(b.String())
}
