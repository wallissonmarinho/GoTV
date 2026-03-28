package domain

import "time"

// MergeSnapshot is persisted merged output metadata and optional payloads.
type MergeSnapshot struct {
	M3U             []byte
	EPG             []byte
	OK              bool
	Message         string
	LastError       string
	ChannelCount    int
	ProgrammeCount  int
	StartedAt       time.Time
	FinishedAt      time.Time
}
