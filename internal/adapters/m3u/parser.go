package m3u

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Parser implements ports.M3UParser.
type Parser struct{}

var _ ports.M3UParser = (*Parser)(nil)

var reAttr = regexp.MustCompile(`([a-zA-Z0-9-]+)="([^"]*)"`)

// Parse decodes extended M3U entries.
func (Parser) Parse(content string) ([]domain.Channel, error) {
	if !strings.HasPrefix(strings.TrimSpace(content), "#EXTM3U") && strings.Contains(content, "#EXTINF") {
		// tolerate missing header
		content = "#EXTM3U\n" + content
	}
	lines := strings.Split(content, "\n")
	var out []domain.Channel
	var pending *domain.Channel

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		lineTrim := strings.TrimSpace(line)
		if strings.HasPrefix(lineTrim, "#EXTINF:") {
			ch, err := parseExtInf(lineTrim)
			if err != nil {
				return nil, err
			}
			pending = &ch
			continue
		}
		if pending != nil && lineTrim != "" && !strings.HasPrefix(lineTrim, "#") {
			pending.URL = lineTrim
			out = append(out, *pending)
			pending = nil
		}
	}
	return out, nil
}

func parseExtInf(line string) (domain.Channel, error) {
	rest := strings.TrimPrefix(line, "#EXTINF:")
	comma := strings.LastIndex(rest, ",")
	if comma < 0 {
		return domain.Channel{}, fmt.Errorf("extinf missing comma")
	}
	attrPart := rest[:comma]
	name := strings.TrimSpace(rest[comma+1:])

	ch := domain.Channel{Name: name}
	for _, m := range reAttr.FindAllStringSubmatch(attrPart, -1) {
		if len(m) < 3 {
			continue
		}
		switch strings.ToLower(m[1]) {
		case "tvg-id":
			ch.TVGID = m[2]
		case "tvg-name":
			ch.TVGName = m[2]
		case "group-title":
			ch.Group = m[2]
		case "tvg-logo":
			ch.Logo = m[2]
		}
	}
	return ch, nil
}
