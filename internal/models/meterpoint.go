package models

import bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"

type Meterpoint struct {
	MPXN       string
	TariffType bookingv1.TariffType
}

func (m Meterpoint) IsEmpty() bool {
	return m.MPXN == "" && m.TariffType == bookingv1.TariffType_TARIFF_TYPE_UNKNOWN
}
