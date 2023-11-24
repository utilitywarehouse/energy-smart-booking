package serialisers

import (
	"encoding/json"
	"fmt"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type PointOfSaleCustomerDetails struct{}

type orderSupply struct {
	MPXN       string `json:"mpxn"`
	TariffType uint32 `json:"tariff_type"`
}

type accountAddress struct {
	UPRN string `json:"uprn"`
	PAF  paf    `json:"paf"`
}

type paf struct {
	BuildingName            string `json:"building_name"`
	BuildingNumber          string `json:"building_number"`
	Department              string `json:"department"`
	DependentLocality       string `json:"dependent_locality"`
	DependentThoroughfare   string `json:"dependent_thoroughfare"`
	DoubleDependentLocality string `json:"double_dependent_locality"`
	Organisation            string `json:"organisation"`
	PostTown                string `json:"post_town"`
	Postcode                string `json:"postcode"`
	SubBuilding             string `json:"sub_building"`
	Thoroughfare            string `json:"thoroughfare"`
}

type accountDetails struct {
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
}

type pointOfSaleCustomerDetails struct {
	AccountNumber     string         `json:"account_number"`
	Details           accountDetails `json:"contact_details"`
	Address           accountAddress `json:"site_address"`
	ElecOrderSupplies orderSupply    `json:"electricity_order_supply"`
	GasOrderSupplies  orderSupply    `json:"gas_order_supply"`
}

func (s PointOfSaleCustomerDetails) Serialise(details models.PointOfSaleCustomerDetails) ([]byte, error) {

	marshalledDetails, err := json.Marshal(pointOfSaleCustomerDetails{
		AccountNumber: details.AccountNumber,
		Details: accountDetails{
			Title:     details.Details.Title,
			FirstName: details.Details.FirstName,
			LastName:  details.Details.LastName,
			Email:     details.Details.Email,
			Mobile:    details.Details.Mobile,
		},
		Address: accountAddress{
			UPRN: details.Address.UPRN,
			PAF: paf{
				BuildingName:            details.Address.PAF.BuildingName,
				BuildingNumber:          details.Address.PAF.BuildingNumber,
				Department:              details.Address.PAF.Department,
				DependentLocality:       details.Address.PAF.DependentLocality,
				DependentThoroughfare:   details.Address.PAF.DependentThoroughfare,
				DoubleDependentLocality: details.Address.PAF.DoubleDependentLocality,
				Organisation:            details.Address.PAF.Organisation,
				PostTown:                details.Address.PAF.PostTown,
				Postcode:                details.Address.PAF.Postcode,
				SubBuilding:             details.Address.PAF.SubBuilding,
				Thoroughfare:            details.Address.PAF.Thoroughfare,
			},
		},
		ElecOrderSupplies: orderSupply{
			MPXN:       details.ElecOrderSupplies.MPXN,
			TariffType: uint32(details.ElecOrderSupplies.TariffType.Number()),
		},
		GasOrderSupplies: orderSupply{
			MPXN:       details.GasOrderSupplies.MPXN,
			TariffType: uint32(details.GasOrderSupplies.TariffType.Number()),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal details, %w", err)
	}

	return marshalledDetails, nil
}

func (s PointOfSaleCustomerDetails) Deserialise(details []byte) (models.PointOfSaleCustomerDetails, error) {

	structuredDetails := pointOfSaleCustomerDetails{}

	err := json.Unmarshal(details, &structuredDetails)
	if err != nil {
		return models.PointOfSaleCustomerDetails{}, fmt.Errorf("failed to unmarshal details, %w", err)
	}

	return models.PointOfSaleCustomerDetails{
		AccountNumber: structuredDetails.AccountNumber,
		Details: models.AccountDetails{
			Title:     structuredDetails.Details.Title,
			FirstName: structuredDetails.Details.FirstName,
			LastName:  structuredDetails.Details.LastName,
			Email:     structuredDetails.Details.Email,
			Mobile:    structuredDetails.Details.Mobile,
		},
		Address: models.AccountAddress{
			UPRN: structuredDetails.Address.UPRN,
			PAF: models.PAF{
				BuildingName:            structuredDetails.Address.PAF.BuildingName,
				BuildingNumber:          structuredDetails.Address.PAF.BuildingNumber,
				Department:              structuredDetails.Address.PAF.Department,
				DependentLocality:       structuredDetails.Address.PAF.DependentLocality,
				DependentThoroughfare:   structuredDetails.Address.PAF.DependentThoroughfare,
				DoubleDependentLocality: structuredDetails.Address.PAF.DoubleDependentLocality,
				Organisation:            structuredDetails.Address.PAF.Organisation,
				PostTown:                structuredDetails.Address.PAF.PostTown,
				Postcode:                structuredDetails.Address.PAF.Postcode,
				SubBuilding:             structuredDetails.Address.PAF.SubBuilding,
				Thoroughfare:            structuredDetails.Address.PAF.Thoroughfare,
			},
		},
		ElecOrderSupplies: models.OrderSupply{
			MPXN:       structuredDetails.ElecOrderSupplies.MPXN,
			TariffType: bookingv1.TariffType(structuredDetails.ElecOrderSupplies.TariffType),
		},
		GasOrderSupplies: models.OrderSupply{
			MPXN:       structuredDetails.GasOrderSupplies.MPXN,
			TariffType: bookingv1.TariffType(structuredDetails.GasOrderSupplies.TariffType),
		},
	}, nil
}
