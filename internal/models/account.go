package models

type Account struct {
	AccountID string
	Details   AccountDetails
	Address   AccountAddress
}

type AccountDetails struct {
	Title     string
	FirstName string
	LastName  string
	Email     string
	Mobile    string
}

func (m AccountDetails) Equals(b AccountDetails) bool {
	return m.Title == b.Title &&
		m.Email == b.Email &&
		m.FirstName == b.FirstName &&
		m.LastName == b.LastName &&
		m.Mobile == b.Mobile
}

func (m AccountDetails) Empty() bool {
	return m.Title == "" &&
		m.Email == "" &&
		m.FirstName == "" &&
		m.LastName == "" &&
		m.Mobile == ""
}

type AccountAddress struct {
	UPRN string
	PAF  PAF
}

type PAF struct {
	BuildingName            string
	BuildingNumber          string
	Department              string
	DependentLocality       string
	DependentThoroughfare   string
	DoubleDependentLocality string
	Organisation            string
	PostTown                string
	Postcode                string
	SubBuilding             string
	Thoroughfare            string
}
