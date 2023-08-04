package models

import (
	"errors"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
)

var (
	ErrUnknownErrorCode = errors.New("unknown error code")
)

func AvailabilityLowriBeckErrorCodeToBookingErrorCode(errorCode lowribeckv1.AvailabilityErrorCodes) (bookingv1.AvailabilityErrorCodes, error) {
	switch errorCode {
	case lowribeckv1.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR:
		return bookingv1.AvailabilityErrorCodes_AVAILABILITY_INTERNAL_ERROR, nil
	case lowribeckv1.AvailabilityErrorCodes_AVAILABILITY_NO_AVAILABLE_SLOTS:
		return bookingv1.AvailabilityErrorCodes_AVAILABILITY_NO_AVAILABLE_SLOTS, nil
	case lowribeckv1.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST:
		return bookingv1.AvailabilityErrorCodes_AVAILABILITY_INVALID_REQUEST, nil
	}

	return -1, ErrUnknownErrorCode
}

func BookingLowriBeckErrorCodeToBookingErrorCode(errorCode lowribeckv1.BookingErrorCodes) (bookingv1.BookingErrorCodes, error) {
	switch errorCode {
	case lowribeckv1.BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE:
		return bookingv1.BookingErrorCodes_BOOKING_APPOINTMENT_UNAVAILABLE, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS:
		return bookingv1.BookingErrorCodes_BOOKING_DUPLICATE_JOB_EXISTS, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_INTERNAL_ERROR:
		return bookingv1.BookingErrorCodes_BOOKING_INTERNAL_ERROR, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_INVALID_SITE:
		return bookingv1.BookingErrorCodes_BOOKING_INVALID_SITE, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_POSTCODE_REFERENCE_MISMATCH:
		return bookingv1.BookingErrorCodes_BOOKING_POSTCODE_REFERENCE_MISMATCH, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_INVALID_REQUEST:
		return bookingv1.BookingErrorCodes_BOOKING_INVALID_REQUEST, nil
	case lowribeckv1.BookingErrorCodes_BOOKING_NO_AVAILABLE_SLOTS:
		return bookingv1.BookingErrorCodes_BOOKING_NO_AVAILABLE_SLOTS, nil
	}

	return -1, ErrUnknownErrorCode
}
