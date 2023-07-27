package serializers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type SiteSerializer struct{}

type SiteAddress struct {
	Postcode                string `json:"postcode"`
	UPRN                    string `json:"uprn"`
	BuildingNameNumber      string `json:"building_name_number"`
	SubBuildingNameNumber   string `json:"sub_building_name_number"`
	DependentThoroughfare   string `json:"dependent_thoroughfare"`
	Thoroughfare            string `json:"thoroughfare"`
	DoubleDependentLocality string `json:"double_dependent_locality"`
	DependentLocality       string `json:"dependent_locality"`
	Locality                string `json:"locality"`
	County                  string `json:"county"`
	Town                    string `json:"town"`
	Department              string `json:"department"`
	Organisation            string `json:"organisation"`
	PoBox                   string `json:"po_box"`
	DeliveryPointSuffix     string `json:"delivery_point_suffix"`
}

func (s SiteSerializer) SerializeSiteAddress(ctx context.Context, siteAddress models.SiteAddress) ([]byte, error) {

	jsonSiteAddress := SiteAddress{
		Postcode:                siteAddress.Postcode,
		UPRN:                    siteAddress.UPRN,
		BuildingNameNumber:      siteAddress.BuildingNameNumber,
		SubBuildingNameNumber:   siteAddress.SubBuildingNameNumber,
		DependentThoroughfare:   siteAddress.DependentThoroughfare,
		Thoroughfare:            siteAddress.Thoroughfare,
		DoubleDependentLocality: siteAddress.DoubleDependentLocality,
		DependentLocality:       siteAddress.DependentLocality,
		Locality:                siteAddress.Locality,
		County:                  siteAddress.County,
		Town:                    siteAddress.Town,
		Department:              siteAddress.Department,
		Organisation:            siteAddress.Organisation,
		PoBox:                   siteAddress.PoBox,
		DeliveryPointSuffix:     siteAddress.DeliveryPointSuffix,
	}

	marshalledPAF, err := json.Marshal(jsonSiteAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %+v, %w", siteAddress, err)
	}

	return marshalledPAF, err
}

func (s SiteSerializer) UnserializeSiteAddress(ctx context.Context, blob []byte) (models.SiteAddress, error) {

	jsonSiteAddress := SiteAddress{}
	err := json.Unmarshal(blob, &jsonSiteAddress)
	if err != nil {
		return models.SiteAddress{}, fmt.Errorf("failed to unmarshal blob: %s, %w", string(blob), err)
	}

	return models.SiteAddress{
		Postcode:                jsonSiteAddress.Postcode,
		UPRN:                    jsonSiteAddress.UPRN,
		BuildingNameNumber:      jsonSiteAddress.BuildingNameNumber,
		SubBuildingNameNumber:   jsonSiteAddress.SubBuildingNameNumber,
		DependentThoroughfare:   jsonSiteAddress.DependentThoroughfare,
		Thoroughfare:            jsonSiteAddress.Thoroughfare,
		DoubleDependentLocality: jsonSiteAddress.DoubleDependentLocality,
		DependentLocality:       jsonSiteAddress.DependentLocality,
		Locality:                jsonSiteAddress.Locality,
		County:                  jsonSiteAddress.County,
		Town:                    jsonSiteAddress.Town,
		Department:              jsonSiteAddress.Department,
		Organisation:            jsonSiteAddress.Organisation,
		PoBox:                   jsonSiteAddress.PoBox,
		DeliveryPointSuffix:     jsonSiteAddress.DeliveryPointSuffix,
	}, nil
}
