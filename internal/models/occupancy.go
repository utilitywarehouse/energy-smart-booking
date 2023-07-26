package models

import "time"

type Occupancy struct {
	OccupancyID string
	SiteID      string
	AccountID   string
	CreatedAt   time.Time
}

func (o Occupancy) IsEmpty() bool {
	return o.AccountID == "" &&
		o.CreatedAt == time.Time{} &&
		o.OccupancyID == "" &&
		o.SiteID == ""
}
