package models

type Site struct {
	SiteID                  string
	Postcode                string
	UPRN                    string
	BuildingNameNumber      string
	SubBuildingNameNumber   string
	DependentThoroughfare   string
	Thoroughfare            string
	DoubleDependentLocality string
	DependentLocality       string
	Locality                string
	County                  string
	Town                    string
	Department              string
	Organisation            string
	PoBox                   string
	DeliveryPointSuffix     string
}
