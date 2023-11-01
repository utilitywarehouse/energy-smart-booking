package serializers

import (
	"encoding/json"
	"fmt"

	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type PointOfSaleCustomerDetails struct{}

type meterpoint struct {
	MPXN       string `json:"mpxn"`
	TariffType uint32 `json:"tariff_type"`
}

type accountAddress struct {
	UPRN string `json:"uprn"`
	PAF  pAF    `json:"paf"`
}

type pAF struct {
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
	AccountNumber string         `json:"account_number"`
	Details       accountDetails `json:"contact_details"`
	Address       accountAddress `json:"site_address"`
	Meterpoints   []meterpoint   `json:"meterpoints"`
}

func (s PointOfSaleCustomerDetails) Serialize(details models.PointOfSaleCustomerDetails) ([]byte, error) {

	meterpoints := []meterpoint{}

	for _, elem := range details.Meterpoints {
		meterpoints = append(meterpoints, meterpoint{
			MPXN:       elem.MPXN,
			TariffType: uint32(elem.TariffType.Number()),
		})
	}

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
			PAF: pAF{
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
		Meterpoints: meterpoints,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal details, %w", err)
	}

	return marshalledDetails, nil
}

func (s PointOfSaleCustomerDetails) Deserialize(details []byte) (models.PointOfSaleCustomerDetails, error) {

	structuredDetails := pointOfSaleCustomerDetails{}

	err := json.Unmarshal(details, &structuredDetails)
	if err != nil {
		return models.PointOfSaleCustomerDetails{}, fmt.Errorf("failed to unmarshal details, %w", err)
	}

	meterpoints := []models.Meterpoint{}

	for _, elem := range structuredDetails.Meterpoints {
		meterpoints = append(meterpoints, models.Meterpoint{
			MPXN:       elem.MPXN,
			TariffType: bookingv1.TariffType(elem.TariffType),
		})
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
		Meterpoints: meterpoints,
	}, nil
}
