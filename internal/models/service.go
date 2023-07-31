package models

import (
	"time"

	"github.com/utilitywarehouse/energy-pkg/domain"
)

type Service struct {
	ServiceID   string
	Mpxn        string
	OccupancyID string
	SupplyType  domain.SupplyType
	AccountID   string
	StartDate   *time.Time
	EndDate     *time.Time
	IsLive      bool
}
