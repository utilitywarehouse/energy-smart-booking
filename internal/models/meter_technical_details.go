package models

import (
	"time"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
)

type GasMeterTechnicalDetails struct {
	MeterType platform.MeterTypeGas
	Capacity  float32
}

type ElectricityMeter struct {
	MeterType   platform.MeterTypeElec
	InstalledAt time.Time
}

type ElectricityMeterTechnicalDetails struct {
	ProfileClass                    platform.ProfileClass
	SettlementStandardConfiguration string
	Meters                          []ElectricityMeter
}

type ElectricityMeterRelatedMPAN struct {
	Relations []MPANRelation
}

type MPANRelation struct {
	Primary   string
	Secondary string
}
