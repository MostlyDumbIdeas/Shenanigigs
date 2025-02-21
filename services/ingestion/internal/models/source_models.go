package models

import (
	"time"
)

type SourcePost struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Time   int64  `json:"time"`
	Parent int    `json:"parent"`
	Kids   []int  `json:"kids"`
	By     string `json:"by"`
}

func (p *SourcePost) ToJobPosting() *JobPosting {
	return &JobPosting{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Text,
		PostedAt:    time.Unix(p.Time, 0),
		RawText:     p.Text,
		ParentID:    p.Parent,
	}
}
