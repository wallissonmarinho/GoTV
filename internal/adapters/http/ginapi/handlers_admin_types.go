package ginapi

// sourceCreateBody is the JSON body for POST /m3u-sources and POST /epg-sources.
type sourceCreateBody struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}
