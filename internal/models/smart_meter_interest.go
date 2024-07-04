package models

import "time"

type SmartMeterInterest struct {
	RegistrationID string
	AccountID      string
	Interested     bool
	Reason         string
	CreatedAt      time.Time
}
