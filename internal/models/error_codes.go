package models

import (
	"errors"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
)

var (
	ErrUnknownErrorCode = errors.New("unknown error code")
)

func BookingLowriBeckErrorCodeToBookingErrorCode(errorCode lowribeckv1.BookingErrorCodes) bookingv1.BookingErrorCodes {
	switch errorCode {
	case lowribeckv1.BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE:
		return bookingv1.BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE
	case lowribeckv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS:
		return bookingv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS
	case lowribeckv1.BookingErrorCodes_BOOKING_ERROR_UNSET:
		return bookingv1.BookingErrorCodes_BOOKING_ERROR_UNSET
	}

	return bookingv1.BookingErrorCodes_BOOKING_UNKNOWN
}
