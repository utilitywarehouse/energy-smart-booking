package models

type PointOfSaleCustomerDetails struct {
	AccountNumber     string
	Details           AccountDetails
	Address           AccountAddress
	ElecOrderSupplies OrderSupply
	GasOrderSupplies  OrderSupply
}
