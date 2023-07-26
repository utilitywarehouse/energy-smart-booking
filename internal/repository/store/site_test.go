package store_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"github.com/utilitywarehouse/energy-smart-booking/internal/repository/store"
)

func Test_SiteStore_Upsert(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	store := store.NewSite(db)

	type inputParams struct {
		site models.Site
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a site with site-id-1",
			input: inputParams{
				site: models.Site{
					SiteID:                  "site-id-1",
					Postcode:                "post-code-1",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number-1",
					DependentThoroughfare:   "dependent-thoroughfare-1",
					ThoroughFare:            "thoroughfare",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "po-box",
					DeliveryPointSuffix:     "delivery-point-suffix",
					SubBuildingNameNumber:   "sub-building-name-number",
				},
			},
			output: nil,
		},
		{
			description: "should upsert another site with site-id-2",
			input: inputParams{
				site: models.Site{
					SiteID:                  "site-id-2",
					Postcode:                "post-code-1",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number-1",
					DependentThoroughfare:   "dependent-thoroughfare-1",
					ThoroughFare:            "thoroughfare",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "po-box",
					DeliveryPointSuffix:     "delivery-point-suffix",
					SubBuildingNameNumber:   "sub-building-name-number",
				},
			},
			output: nil,
		},
		{
			description: "should upsert another site with the same previous site_id: site-id-1",
			input: inputParams{
				site: models.Site{
					SiteID:                  "site-id-1",
					Postcode:                "post-code-1",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number-1",
					DependentThoroughfare:   "dependent-thoroughfare-1",
					ThoroughFare:            "thoroughfare",
					DoubleDependentLocality: "ddl-1",
					DependentLocality:       "dl-1",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "po-box",
					DeliveryPointSuffix:     "delivery-point-suffix",
					SubBuildingNameNumber:   "sub-building-name-number",
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			store.Begin()

			store.Upsert(tc.input.site)

			err := store.Commit(ctx)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_SiteStore_Get(t *testing.T) {
	ctx := context.Background()

	testContainer, err := setupTestContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	if err != nil {
		t.Fatal(err)
	}

	db, err := store.Setup(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = populateDB(ctx, db)

	siteStore := store.NewSite(db)

	type outputParams struct {
		site *models.Site
		err  error
	}

	type testSetup struct {
		description string
		input       string
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get a site with id site-id",
			input:       "site-id-a",
			output: outputParams{
				site: &models.Site{
					SiteID:                  "site-id-a",
					Postcode:                "post-code-1",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number",
					DependentThoroughfare:   "dependent-thoroughfare",
					ThoroughFare:            "thoroughfare",
					DoubleDependentLocality: "double-dependent-locality",
					DependentLocality:       "dependent-locality",
					Locality:                "locality",
					County:                  "county",
					Town:                    "town",
					Department:              "department",
					Organisation:            "organisation",
					PoBox:                   "po-box",
					DeliveryPointSuffix:     "deliver-point-suffix",
					SubBuildingNameNumber:   "sub-building-name-number",
				},
				err: nil,
			},
		},
		{
			description: "should return a pgx err no rows because site id does not exist",
			input:       "i dont exist",
			output: outputParams{
				site: nil,
				err:  store.ErrSiteNotFound,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			site, err := siteStore.GetSiteBySiteID(ctx, tc.input)

			if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
				t.Fatal(err)
			}

			if diff := cmp.Diff(site, tc.output.site); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
