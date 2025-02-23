package models

import (
	"encoding/json"
	"strconv"
	"time"
)

type SourcePost struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Text        string   `json:"text"`
	By          string   `json:"by"`
	Time        int64    `json:"time"`
	Kids        IntSlice `json:"kids"`
	Type        string   `json:"type"`
	URL         string   `json:"url"`
	Score       int      `json:"score"`
	Dead        bool     `json:"dead"`
	Deleted     bool     `json:"deleted"`
	Descendants int      `json:"descendants"`
}

type IntSlice []int

func (s IntSlice) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *IntSlice) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func (p SourcePost) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p SourcePost) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &p)
}

func (p *SourcePost) ToJobPosting() *JobPosting {
	return &JobPosting{
		ID:          strconv.Itoa(p.ID),
		Title:       p.Title,
		Description: p.Text,
		PostedAt:    time.Unix(p.Time, 0),
		RawText:     p.Text,
	}
}
