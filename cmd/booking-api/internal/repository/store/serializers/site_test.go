package serializers_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store/serializers"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_Serialize(t *testing.T) {

	ctx := context.Background()

	siteSerializer := serializers.SiteSerializer{}

	input := models.SiteAddress{
		UPRN:                    "uprn",
		BuildingNameNumber:      "building-name-number",
		SubBuildingNameNumber:   "sub-building-name-number",
		Locality:                "locality",
		County:                  "county",
		Town:                    "town",
		PoBox:                   "po-box",
		DeliveryPointSuffix:     "delivery-point-suffix",
		Department:              "department",
		DependentLocality:       "dependent-locality",
		DependentThoroughfare:   "dependent-thoroughfare",
		DoubleDependentLocality: "double-dependent-locality",
		Organisation:            "organisation",
		Postcode:                "postcode",
		Thoroughfare:            "thoroughfare",
	}

	expected := []byte(`{"postcode":"postcode","uprn":"uprn","building_name_number":"building-name-number","sub_building_name_number":"sub-building-name-number","dependent_thoroughfare":"dependent-thoroughfare","thoroughfare":"thoroughfare","double_dependent_locality":"double-dependent-locality","dependent_locality":"dependent-locality","locality":"locality","county":"county","town":"town","department":"department","organisation":"organisation","po_box":"po-box","delivery_point_suffix":"delivery-point-suffix"}`)

	actualBlob, err := siteSerializer.SerializeSiteAddress(ctx, input)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actualBlob); diff != "" {
		t.Fatal(diff)
	}

}

func Test_Unserialize(t *testing.T) {

	ctx := context.Background()

	siteSerializer := serializers.SiteSerializer{}

	input := []byte(`{"postcode":"postcode","uprn":"uprn","building_name_number":"building-name-number","sub_building_name_number":"sub-building-name-number","dependent_thoroughfare":"dependent-thoroughfare","thoroughfare":"thoroughfare","double_dependent_locality":"double-dependent-locality","dependent_locality":"dependent-locality","locality":"locality","county":"county","town":"town","department":"department","organisation":"organisation","po_box":"po-box","delivery_point_suffix":"delivery-point-suffix"}`)

	expected := models.SiteAddress{
		UPRN:                    "uprn",
		BuildingNameNumber:      "building-name-number",
		SubBuildingNameNumber:   "sub-building-name-number",
		Locality:                "locality",
		County:                  "county",
		Town:                    "town",
		PoBox:                   "po-box",
		DeliveryPointSuffix:     "delivery-point-suffix",
		Department:              "department",
		DependentLocality:       "dependent-locality",
		DependentThoroughfare:   "dependent-thoroughfare",
		DoubleDependentLocality: "double-dependent-locality",
		Organisation:            "organisation",
		Postcode:                "postcode",
		Thoroughfare:            "thoroughfare",
	}

	actual, err := siteSerializer.UnserializeSiteAddress(ctx, input)

	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatal(diff)
	}

}
