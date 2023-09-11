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

func (gca GetCalendarAvailabilityRequest) GetPostCode() string {
	return gca.PostCode
}

func (gca GetCalendarAvailabilityRequest) GetReference() string {
	return gca.ReferenceID
}

type GetCalendarAvailabilityResponse struct {
	RequestID                  string             `json:"RequestId,omitempty"`
	SendingSystem              string             `json:"SendingSystem,omitempty"`
	ReceivingSystem            string             `json:"ReceivingSystem,omitempty"`
	CreatedDate                string             `json:"CreatedDate,omitempty"`
	Mpan                       string             `json:"Mpan,omitempty"`
	Mprn                       string             `json:"Mprn,omitempty"`
	ElecJobTypeCode            string             `json:"ElecJobTypeCode,omitempty"`
	GasJobTypeCode             string             `json:"GasJobTypeCode,omitempty"`
	CalendarAvailabilityResult []AvailabilitySlot `json:"CalendarAvailabilityResult,omitempty"`
	ResponseMessage            string             `json:"ResponseMessage,omitempty"`
	ResponseCode               string             `json:"ResponseCode,omitempty"`
}

type AvailabilitySlot struct {
	AppointmentDate string `json:"AppointmentDate,omitempty"`
	AppointmentTime string `json:"AppointmentTime,omitempty"`
}

type CreateBookingRequest struct {
	RequestID            string `json:"RequestId,omitempty"`
	SendingSystem        string `json:"SendingSystem,omitempty"`
	ReceivingSystem      string `json:"ReceivingSystem,omitempty"`
	CreatedDate          string `json:"CreatedDate,omitempty"`
	AppointmentDate      string `json:"AppointmentDate,omitempty"`
	AppointmentTime      string `json:"AppointmentTime,omitempty"`
	ReferenceID          string `json:"ReferenceId,omitempty"`
	PostCode             string `json:"PostCode,omitempty"`
	Mpan                 string `json:"Mpan,omitempty"`
	Mprn                 string `json:"Mprn,omitempty"`
	ElecJobTypeCode      string `json:"ElecJobTypeCode,omitempty"`
	GasJobTypeCode       string `json:"GasJobTypeCode,omitempty"`
	Ssc                  string `json:"Ssc,omitempty"`
	SiteContactName      string `json:"SiteContactName,omitempty"`
	SiteContactNumber    string `json:"SiteContactNumber,omitempty"`
	SiteContactNumberAlt string `json:"SiteContactNumberAlt,omitempty"`
	AccessPassword       string `json:"AccessPassword,omitempty"`
	AdditionalInfo       string `json:"AdditionalInfo,omitempty"`
	Vulnerabilities      string `json:"Vulnerabilities,omitempty"`
	VulnerabilitiesOther string `json:"VulnerabilitiesOther,omitempty"`
	EsmeGUID             string `json:"EsmeGuid,omitempty"`
	GsmeGUID             string `json:"GsmeGuid,omitempty"`
	CommsHubGUID         string `json:"CommsHubGuid,omitempty"`
	IhdGUID              string `json:"IhdGuid,omitempty"`
	PpmidGUID            string `json:"PpmidGuid,omitempty"`
	CadGUID              string `json:"CadGuid,omitempty"`
	HcalcsGUID           string `json:"HcalcsGuid,omitempty"`
}

func (gba CreateBookingRequest) GetPostCode() string {
	return gba.PostCode
}

func (gba CreateBookingRequest) GetReference() string {
	return gba.ReferenceID
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
