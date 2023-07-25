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
