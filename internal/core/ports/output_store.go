package ports

// OutputStore holds the last merged M3U/EPG bytes and run metadata.
type OutputStore interface {
	Set(m3u, epg []byte, meta OutputMeta)
	Get() (m3u, epg []byte, ok bool)
	Meta() OutputMeta
}

// OutputMeta describes the last aggregation.
type OutputMeta struct {
	OK             bool
	Message        string
	ChannelCount   int
	ProgrammeCount int
}
