package models

import (
	"encoding/json"
	"time"
)

type JobPosting struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PostedAt    time.Time `json:"posted_at"`
	RawText     string    `json:"raw_text"`
	ParentID    int       `json:"parent_id"`
}

func (p JobPosting) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p JobPosting) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &p)
}
