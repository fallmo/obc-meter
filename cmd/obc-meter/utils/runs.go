package utils

import (
	"time"
)

type Run struct {
	ID             int       `json:"id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	SucceededCount int       `json:"succeeded_count"`
	UpdatedCount   int       `json:"updated_count"`
	FailedCount    int       `json:"failed_count"`
	SucceededUids  []string  `json:"succeeded_uids"`
	UpdatedUids    []string  `json:"updated_uids"`
	FailedUids     []string  `json:"failed_uids"`
	// trigger (manual vs. automatic?)
}
