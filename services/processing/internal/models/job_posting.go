package models

import (
	"time"
)

type JobPosting struct {
	ID                   string
	Title                string
	Company              string
	Location             string
	Description          string
	Technologies         []string
	ExperienceLevel      string
	CompensationMin      float64
	CompensationMax      float64
	CompensationCurrency string
	CompensationPeriod   string
	RemotePolicy         string
	Source               string
	SourceURL            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	RawData              string
}
