package models

import bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"

func PlatformSourceToBookingSource(platformSource bookingv1.Platform) bookingv1.BookingSource {
	switch platformSource {
	case bookingv1.Platform_PLATFORM_APP:
		return bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_APP
	case bookingv1.Platform_PLATFORM_MY_ACCOUNT:
		return bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_MY_ACCOUNT
	case bookingv1.Platform_PLATFORM_WEB:
		return bookingv1.BookingSource_BOOKING_SOURCE_PLATFORM_WEB
	}

	return bookingv1.BookingSource_BOOKING_SOURCE_UNKNOWN
}
