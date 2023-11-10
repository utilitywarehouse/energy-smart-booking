package models

import (
	"fmt"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	energy "github.com/utilitywarehouse/energy-pkg/domain"
)

type OrderSupply struct {
	MPXN       string
	TariffType bookingv1.TariffType
}

func (os OrderSupply) IsEmpty() bool {
	return os.MPXN == "" && os.TariffType == bookingv1.TariffType_TARIFF_TYPE_UNKNOWN
}

func DeduceOrderSupplies(orderSupplies []OrderSupply) (OrderSupply, OrderSupply, error) {
	var mpan, mprn OrderSupply
	for i, orderSupply := range orderSupplies {
		mpxn, err := energy.NewMeterPointNumber(orderSupply.MPXN)
		if err != nil {
			return OrderSupply{}, OrderSupply{}, fmt.Errorf("invalid meterpoint number (%s): %v", orderSupply.MPXN, err)
		}
		// We want the first electricity MPAN
		if mpxn.SupplyType() == energy.SupplyTypeElectricity && mpan.IsEmpty() {
			mpan = orderSupplies[i]
		} else if mpxn.SupplyType() == energy.SupplyTypeGas {
			mprn = orderSupplies[i]
		}
	}
	return mpan, mprn, nil
}
