package mapper

import (
	"errors"
	"fmt"
)

var (
	ErrAppointmentNotFound      = errors.New("no appointments found")
	ErrAppointmentOutOfRange    = errors.New("appointment out of range")
	ErrAppointmentAlreadyExists = errors.New("appointment already exists")
	ErrInternalError            = errors.New("internal server error")
)

type InvalidType string

const (
	InvalidPostcode        InvalidType = "postcode"
	InvalidReference       InvalidType = "reference"
	InvalidSite            InvalidType = "site"
	InvalidAppointmentDate InvalidType = "appointment date"
	InvalidAppointmentTime InvalidType = "appointment time"
)

type InvalidRequestError struct {
	Parameter InvalidType
}

func (m *InvalidRequestError) Error() string {
	return fmt.Errorf("invalid request [%s]", m.Parameter).Error()
}

func (m *InvalidRequestError) GetParameter() InvalidType {
	return m.Parameter
}

func NewInvalidRequestError(parameter InvalidType) *InvalidRequestError {
	return &InvalidRequestError{
		Parameter: parameter,
	}
}
