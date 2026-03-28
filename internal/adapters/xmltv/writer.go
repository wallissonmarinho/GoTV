package xmltv

import (
	"encoding/xml"

	"github.com/wallissonmarinho/GoTV/internal/core/domain"
	"github.com/wallissonmarinho/GoTV/internal/core/ports"
)

// Writer implements ports.XMLTVEncoder.
type Writer struct{}

var _ ports.XMLTVEncoder = (*Writer)(nil)

type xmlTVDoc struct {
	XMLName    xml.Name    `xml:"tv"`
	Channels   []xmlTVCh   `xml:"channel,omitempty"`
	Programmes []xmlTVProg `xml:"programme,omitempty"`
}

type xmlTVCh struct {
	ID           string   `xml:"id,attr"`
	DisplayNames []string `xml:"display-name"`
}

type xmlTVProg struct {
	Channel string   `xml:"channel,attr"`
	Start   string   `xml:"start,attr"`
	Stop    string   `xml:"stop,attr"`
	Titles  []string `xml:"title"`
}

// Build serializes EPGData to XMLTV.
func (Writer) Build(data *domain.EPGData) []byte {
	if data == nil {
		data = &domain.EPGData{}
	}
	doc := xmlTVDoc{
		Channels:   make([]xmlTVCh, 0, len(data.Channels)),
		Programmes: make([]xmlTVProg, 0, len(data.Programmes)),
	}
	for _, ch := range data.Channels {
		doc.Channels = append(doc.Channels, xmlTVCh{
			ID:           ch.ID,
			DisplayNames: append([]string(nil), ch.DisplayNames...),
		})
	}
	for _, p := range data.Programmes {
		doc.Programmes = append(doc.Programmes, xmlTVProg{
			Channel: p.Channel,
			Start:   p.Start,
			Stop:    p.Stop,
			Titles:  append([]string(nil), p.Titles...),
		})
	}
	out, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return []byte(`<?xml version="1.0" encoding="UTF-8"?><tv></tv>`)
	}
	h := []byte(xml.Header)
	return append(h, out...)
}
