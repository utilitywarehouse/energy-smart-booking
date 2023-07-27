package lowribeck

type auth struct {
	user     string
	password string
}

type GetCalendarAvailabilityRequest struct {
	RequestID       string `json:"RequestId,omitempty"`
	SendingSystem   string `json:"SendingSystem,omitempty"`
	ReceivingSystem string `json:"ReceivingSystem,omitempty"`
	CreatedDate     string `json:"CreatedDate,omitempty"`
	ReferenceID     string `json:"ReferenceId,omitempty"`
	PostCode        string `json:"PostCode,omitempty"`
	Mpan            string `json:"Mpan,omitempty"`
	Mprn            string `json:"Mprn,omitempty"`
}

type GetCalendarAvailabilityResponse struct {
	RequestID                  string              `json:"RequestId,omitempty"`
	SendingSystem              string              `json:"SendingSystem,omitempty"`
	ReceivingSystem            string              `json:"ReceivingSystem,omitempty"`
	CreatedDate                string              `json:"CreatedDate,omitempty"`
	Mpan                       string              `json:"Mpan,omitempty"`
	Mprn                       string              `json:"Mprn,omitempty"`
	ElecJobTypeCode            string              `json:"ElecJobTypeCode,omitempty"`
	GasJobTypeCode             string              `json:"GasJobTypeCode,omitempty"`
	CalendarAvailabilityResult []*AvailabilitySlot `json:"CalendarAvailabilityResult,omitempty"`
	ResponseMessage            string              `json:"ResponseMessage,omitempty"`
	ResponseCode               string              `json:"ResponseCode,omitempty"`
}

type AvailabilitySlot struct {
	AppointmentDate string `json:"AppointmentDate,omitempty"`
	AppointmentTime string `json:"AppointmentTime,omitempty"`
}

type CreateBookingRequest struct {
	RequestID       string `json:"RequestId,omitempty"`
	SendingSystem   string `json:"SendingSystem,omitempty"`
	ReceivingSystem string `json:"ReceivingSystem,omitempty"`
	CreatedDate     string `json:"CreatedDate,omitempty"`
	ReferenceID     string `json:"ReferenceId,omitempty"`
	AppointmentDate string `json:"AppointmentDate,omitempty"`
	AppointmentTime string `json:"AppointmentTime,omitempty"`
	PostCode        string `json:"PostCode,omitempty"`
	Mpan            string `json:"Mpan,omitempty"`
	Mprn            string `json:"Mprn,omitempty"`
}

type CreateBookingResponse struct {
	RequestID       string `json:"RequestId,omitempty"`
	ReferenceID     string `json:"ReferenceId,omitempty"`
	SendingSystem   string `json:"SendingSystem,omitempty"`
	ReceivingSystem string `json:"ReceivingSystem,omitempty"`
	CreatedDate     string `json:"CreatedDate,omitempty"`
	Mpan            string `json:"Mpan,omitempty"`
	Mprn            string `json:"Mprn,omitempty"`
	ResponseMessage string `json:"ResponseMessage,omitempty"`
	ResponseCode    string `json:"ResponseCode,omitempty"`
}

// {
// 	"RequestId": "1234234",
// 	"SendingSystem": "UTIL",
// 	"ReceivingSystem": "LB",
// 	"CreatedDate": "25/07/2020 12:47:41",
// 	 "ReferenceId": "UTIL/4568973",
// 				"AppointmentDate": "15/08/2023",
// 				"AppointmentTime": "10:00-14:00",
// 	 "PostCode": "LS8 4EX",
// 	"Mpan": "",
// 	"Mprn": ""
// 	}

// {
//     "ReferenceId": "UTIL/4568973",
//     "Mpan": "",
//     "Mprn": "",
//     "ResponseMessage": "Booking Confirmed",
//     "ResponseCode": "B01",
//     "RequestId": "1234234",
//     "SendingSystem": "LB",
//     "ReceivingSystem": "UTIL",
//     "CreatedDate": "26/07/2023 14:24:34"
// }
