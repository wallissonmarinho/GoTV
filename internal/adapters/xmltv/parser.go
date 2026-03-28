package xmltv

import (
	"encoding/xml"
	"strings"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Parser implements ports.XMLTVParser.
type Parser struct{}

var _ ports.XMLTVParser = (*Parser)(nil)

type doc struct {
	XMLName   xml.Name    `xml:"tv"`
	Channels  []rawChannel `xml:"channel"`
	Programmes []rawProg   `xml:"programme"`
}

type rawChannel struct {
	ID           string        `xml:"id,attr"`
	DisplayNames []displayName `xml:"display-name"`
}

type displayName struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

type rawProg struct {
	Channel string        `xml:"channel,attr"`
	Start   string        `xml:"start,attr"`
	Stop    string        `xml:"stop,attr"`
	Titles  []titleEl     `xml:"title"`
}

type titleEl struct {
	Lang string `xml:"lang,attr"`
	Text string `xml:",chardata"`
}

// Parse loads XMLTV into domain.EPGData.
func (Parser) Parse(data []byte) (*domain.EPGData, error) {
	var d doc
	if err := xml.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	out := &domain.EPGData{}
	for _, c := range d.Channels {
		var names []string
		for _, dn := range c.DisplayNames {
			t := strings.TrimSpace(dn.Text)
			if t != "" {
				names = append(names, t)
			}
		}
		out.Channels = append(out.Channels, domain.EPGChannel{
			ID:           strings.TrimSpace(c.ID),
			DisplayNames: names,
		})
	}
	for _, p := range d.Programmes {
		var titles []string
		for _, t := range p.Titles {
			x := strings.TrimSpace(t.Text)
			if x != "" {
				titles = append(titles, x)
			}
		}
		out.Programmes = append(out.Programmes, domain.EPGProgramme{
			Channel: strings.TrimSpace(p.Channel),
			Start:   strings.TrimSpace(p.Start),
			Stop:    strings.TrimSpace(p.Stop),
			Titles:  titles,
		})
	}
	return out, nil
}
