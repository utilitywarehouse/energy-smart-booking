package models

import (
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
)

func BookingTariffTypeToLowribeckTariffType(bookingTariffType bookingv1.TariffType) lowribeckv1.TariffType {
	switch bookingTariffType {
	case bookingv1.TariffType_TARIFF_TYPE_CREDIT:
		return lowribeckv1.TariffType_TARIFF_TYPE_CREDIT
	case bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT:
		return lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT
	}

	return lowribeckv1.TariffType_TARIFF_TYPE_UNKNOWN
}

func BookingLowribeckTariffTypeToBookingAPITariffType(lowribeckTariffType lowribeckv1.TariffType) bookingv1.TariffType {
	switch lowribeckTariffType {
	case lowribeckv1.TariffType_TARIFF_TYPE_CREDIT:
		return bookingv1.TariffType_TARIFF_TYPE_CREDIT
	case lowribeckv1.TariffType_TARIFF_TYPE_PREPAYMENT:
		return bookingv1.TariffType_TARIFF_TYPE_PREPAYMENT
	}

	return bookingv1.TariffType_TARIFF_TYPE_UNKNOWN
}
