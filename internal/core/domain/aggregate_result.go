package domain

import "time"

// MergeResult summarizes one merge run.
type MergeResult struct {
	OK              bool
	Message         string
	Errors          []string
	ChannelCount    int
	ProgrammeCount  int
	StartedAt       time.Time
	FinishedAt      time.Time
}
