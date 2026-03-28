package merge

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
)

// NormalizeStreamURL trims whitespace; used as dedupe key.
func NormalizeStreamURL(u string) string {
	return strings.TrimSpace(u)
}

// NormalizeName lowercases and trims for loose EPG matching.
func NormalizeName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// StableTvgIDFromURL generates a deterministic tvg-id when the playlist omits it.
func StableTvgIDFromURL(rawURL string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(rawURL)))
	return fmt.Sprintf("gotv-%x", h[:8])
}

// MergeChannelsByURLOrder merges batches in order; first URL wins.
func MergeChannelsByURLOrder(batches [][]domain.Channel) []domain.Channel {
	seen := make(map[string]struct{})
	var out []domain.Channel
	for _, batch := range batches {
		for _, c := range batch {
			key := NormalizeStreamURL(c.URL)
			if key == "" {
				continue
			}
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			nc := c
			if strings.TrimSpace(nc.TVGID) == "" {
				nc.TVGID = StableTvgIDFromURL(nc.URL)
			}
			out = append(out, nc)
		}
	}
	return out
}

// RemapEPG aligns programme channel refs to merged playlist tvg-ids.
func RemapEPG(
	merged []domain.Channel,
	perSourceChannels [][]domain.Channel,
	perSourceEPG []*domain.EPGData,
) *domain.EPGData {
	urlToTvg := make(map[string]string)
	for _, ch := range merged {
		urlToTvg[NormalizeStreamURL(ch.URL)] = strings.TrimSpace(ch.TVGID)
	}

	out := &domain.EPGData{}
	chSeen := make(map[string]struct{})
	progSeen := make(map[string]struct{})

	for i, epg := range perSourceEPG {
		if epg == nil || i >= len(perSourceChannels) {
			continue
		}
		orig := perSourceChannels[i]
		xmlIDToURL := mapXMLChannelIDToStreamURL(epg.Channels, orig)

		for _, p := range epg.Programmes {
			u, ok := xmlIDToURL[p.Channel]
			if !ok {
				continue
			}
			finalID, ok := urlToTvg[NormalizeStreamURL(u)]
			if !ok || finalID == "" {
				continue
			}
			pk := finalID + "|" + p.Start + "|" + p.Stop + "|" + strings.Join(p.Titles, ";")
			if _, dup := progSeen[pk]; dup {
				continue
			}
			progSeen[pk] = struct{}{}
			out.Programmes = append(out.Programmes, domain.EPGProgramme{
				Channel: finalID,
				Start:   p.Start,
				Stop:    p.Stop,
				Titles:  append([]string(nil), p.Titles...),
			})
		}
	}

	for _, ch := range merged {
		id := strings.TrimSpace(ch.TVGID)
		if id == "" {
			continue
		}
		if _, ok := chSeen[id]; ok {
			continue
		}
		chSeen[id] = struct{}{}
		dn := ch.Name
		if strings.TrimSpace(ch.TVGName) != "" {
			dn = ch.TVGName
		}
		out.Channels = append(out.Channels, domain.EPGChannel{
			ID:           id,
			DisplayNames: []string{dn},
		})
	}

	return out
}

func mapXMLChannelIDToStreamURL(epgChs []domain.EPGChannel, m3u []domain.Channel) map[string]string {
	out := make(map[string]string)
	for _, xc := range epgChs {
		id := strings.TrimSpace(xc.ID)
		if id == "" {
			continue
		}
		if u := matchXMLChannelToURL(xc, m3u); u != "" {
			out[id] = u
		}
	}
	return out
}

// MergeIndependentEPGs unions several XMLTV documents (channels and programmes deduped).
func MergeIndependentEPGs(epgs []*domain.EPGData) *domain.EPGData {
	out := &domain.EPGData{}
	chSeen := make(map[string]struct{})
	progSeen := make(map[string]struct{})
	for _, epg := range epgs {
		if epg == nil {
			continue
		}
		for _, c := range epg.Channels {
			id := strings.TrimSpace(c.ID)
			if id == "" {
				continue
			}
			if _, dup := chSeen[id]; dup {
				continue
			}
			chSeen[id] = struct{}{}
			out.Channels = append(out.Channels, domain.EPGChannel{
				ID:           id,
				DisplayNames: append([]string(nil), c.DisplayNames...),
			})
		}
		for _, p := range epg.Programmes {
			ch := strings.TrimSpace(p.Channel)
			if ch == "" {
				continue
			}
			pk := ch + "|" + p.Start + "|" + p.Stop + "|" + strings.Join(p.Titles, ";")
			if _, dup := progSeen[pk]; dup {
				continue
			}
			progSeen[pk] = struct{}{}
			out.Programmes = append(out.Programmes, domain.EPGProgramme{
				Channel: ch,
				Start:   p.Start,
				Stop:    p.Stop,
				Titles:  append([]string(nil), p.Titles...),
			})
		}
	}
	return out
}

// BuildOutputEPGForPlaylist keeps programmes that reference a tvg-id present in the merged playlist.
func BuildOutputEPGForPlaylist(merged []domain.Channel, mergedEPG *domain.EPGData) *domain.EPGData {
	if mergedEPG == nil {
		mergedEPG = &domain.EPGData{}
	}
	tvg := make(map[string]struct{})
	for _, ch := range merged {
		id := strings.TrimSpace(ch.TVGID)
		if id == "" {
			id = StableTvgIDFromURL(ch.URL)
		}
		tvg[id] = struct{}{}
	}
	out := &domain.EPGData{}
	for _, ch := range merged {
		id := strings.TrimSpace(ch.TVGID)
		if id == "" {
			id = StableTvgIDFromURL(ch.URL)
		}
		dn := ch.Name
		if strings.TrimSpace(ch.TVGName) != "" {
			dn = ch.TVGName
		}
		out.Channels = append(out.Channels, domain.EPGChannel{
			ID:           id,
			DisplayNames: []string{dn},
		})
	}
	progSeen := make(map[string]struct{})
	for _, p := range mergedEPG.Programmes {
		cid := strings.TrimSpace(p.Channel)
		if _, ok := tvg[cid]; !ok {
			continue
		}
		pk := cid + "|" + p.Start + "|" + p.Stop + "|" + strings.Join(p.Titles, ";")
		if _, dup := progSeen[pk]; dup {
			continue
		}
		progSeen[pk] = struct{}{}
		out.Programmes = append(out.Programmes, domain.EPGProgramme{
			Channel: cid,
			Start:   p.Start,
			Stop:    p.Stop,
			Titles:  append([]string(nil), p.Titles...),
		})
	}
	return out
}

func matchXMLChannelToURL(xc domain.EPGChannel, m3u []domain.Channel) string {
	id := strings.TrimSpace(xc.ID)
	for _, c := range m3u {
		if strings.TrimSpace(c.TVGID) == id {
			return c.URL
		}
	}
	xmlCanon := canonicalEPGIDKey(id)
	if xmlCanon != "" {
		for _, c := range m3u {
			if canonicalEPGIDKey(c.TVGID) == xmlCanon {
				return c.URL
			}
		}
	}
	for _, c := range m3u {
		for _, label := range m3uDisplayLabels(c) {
			cn := NormalizeName(label)
			if cn == "" {
				continue
			}
			for _, dn := range xc.DisplayNames {
				if NormalizeName(dn) == cn {
					return c.URL
				}
			}
		}
	}
	if len(m3u) == 1 && len(xc.DisplayNames) == 0 {
		return m3u[0].URL
	}
	return ""
}

func m3uDisplayLabels(c domain.Channel) []string {
	var out []string
	if n := strings.TrimSpace(c.Name); n != "" {
		out = append(out, n)
	}
	if tn := strings.TrimSpace(c.TVGName); tn != "" {
		out = append(out, tn)
	}
	return out
}
