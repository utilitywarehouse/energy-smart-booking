package models

import bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"

type OrderSupply struct {
	MPXN       string
	TariffType bookingv1.TariffType
}

func (os OrderSupply) IsEmpty() bool {
	return os.MPXN == "" && os.TariffType == bookingv1.TariffType_TARIFF_TYPE_UNKNOWN
}
