package models

import (
	"time"
)

type JobPosting struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PostedAt    time.Time `json:"posted_at"`
	RawText     string    `json:"raw_text"`
	ParentID    int       `json:"parent_id"`
}
