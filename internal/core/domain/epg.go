package domain

// EPGChannel is one <channel> from XMLTV.
type EPGChannel struct {
	ID           string
	DisplayNames []string
}

// EPGProgramme is one <programme> from XMLTV.
type EPGProgramme struct {
	Channel string
	Start   string
	Stop    string
	Titles  []string
}

// EPGData holds parsed XMLTV content.
type EPGData struct {
	Channels   []EPGChannel
	Programmes []EPGProgramme
}
