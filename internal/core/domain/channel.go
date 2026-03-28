package domain

// Channel represents one stream entry from an M3U (after parse).
type Channel struct {
	Name    string
	URL     string
	TVGID   string
	TVGName string
	Group   string
	Logo    string
}
