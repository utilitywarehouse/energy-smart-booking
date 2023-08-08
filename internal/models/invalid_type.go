package models

type InvalidType string

const (
	InvalidPostcode        InvalidType = "postcode"
	InvalidReference       InvalidType = "reference"
	InvalidSite            InvalidType = "site"
	InvalidAppointmentDate InvalidType = "appointment date"
	InvalidAppointmentTime InvalidType = "appointment time"
)
