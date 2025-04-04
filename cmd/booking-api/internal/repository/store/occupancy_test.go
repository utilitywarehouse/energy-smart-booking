package store_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func Test_OccupancyStore_Insert(t *testing.T) {
	occupancyStore := store.NewOccupancy(pool)
	defer truncateDB(t)

	type inputParams struct {
		occupancy models.Occupancy
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should upsert a occupancy with occupancy-id-1",
			input: inputParams{
				occupancy: models.Occupancy{
					OccupancyID: "occupancy-id-1",
					SiteID:      "site-id-1",
					AccountID:   "account-id-1",
					CreatedAt:   time.Now(),
				},
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyStore.Begin()

			occupancyStore.Insert(tc.input.occupancy)

			err := occupancyStore.Commit(t.Context())

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_OccupancyStore_UpdateSiteID(t *testing.T) {
	err := populateDB(t.Context(), pool)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(pool)
	defer truncateDB(t)

	type inputParams struct {
		occupancyID string
		siteID      string
	}

	type testSetup struct {
		description string
		input       inputParams
		output      error
	}

	testCases := []testSetup{
		{
			description: "should update the site id of occupancy with occupancy-id",
			input: inputParams{
				occupancyID: "occupancy-id",
				siteID:      "new-site-id",
			},
			output: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			occupancyStore.Begin()

			occupancyStore.UpdateSiteID(tc.input.occupancyID, tc.input.siteID)

			err = occupancyStore.Commit(t.Context())

			if err != nil {
				if tc.output != nil {
					if diff := cmp.Diff(err.Error(), tc.output.Error()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}
		})

	}
}

func Test_OccupancyStore_GetOccupancyByID(t *testing.T) {
	err := populateDB(t.Context(), pool)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(pool)
	defer truncateDB(t)

	type inputParams struct {
		occupancyID string
	}

	type outputParams struct {
		occupancy *models.Occupancy
		err       error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the occupancy by occupancy-id",
			input: inputParams{
				occupancyID: "occupancy-id",
			},
			output: outputParams{
				occupancy: &models.Occupancy{
					OccupancyID: "occupancy-id",
					SiteID:      "site-id",
					AccountID:   "account-id",
				},
				err: nil,
			},
		},
		{
			description: "should fail to find a occupancy id because it does not exist",
			input: inputParams{
				occupancyID: "i do not exist",
			},
			output: outputParams{
				occupancy: nil,
				err:       store.ErrOccupancyNotFound,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			occupancy, err := occupancyStore.GetOccupancyByID(t.Context(), tc.input.occupancyID)

			if err != nil {
				if tc.output.err != nil {
					if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(occupancy, tc.output.occupancy); diff != "" {
				t.Fatal(diff)
			}
		})

	}
}

func Test_OccupancyStore_GetSiteExternalReferenceByAccountID(t *testing.T) {
	err := populateDB(t.Context(), pool)
	if err != nil {
		t.Fatal(err)
	}

	occupancyStore := store.NewOccupancy(pool)
	defer truncateDB(t)

	type inputParams struct {
		accountID string
	}

	type outputParams struct {
		site                 *models.Site
		occupancyEligibility *models.OccupancyEligibility
		err                  error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should get the site and the external reference by account-id",
			input: inputParams{
				accountID: "account-id-#1",
			},
			output: outputParams{
				site: &models.Site{
					SiteID:                  "site-id-a",
					Postcode:                "post-code-1",
					UPRN:                    "uprn",
					BuildingNameNumber:      "building-name-number",
					DependentThoroughfare:   "dependent-thoroughfare",
					Thoroughfare:            "thoroughfare",
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
				occupancyEligibility: &models.OccupancyEligibility{
					OccupancyID: "occupancy-id-#1",
					Reference:   "ref##1",
				},
				err: nil,
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {
			actualSite, actualOccupancyEligibility, err := occupancyStore.GetSiteExternalReferenceByAccountID(t.Context(), tc.input.accountID)

			if err != nil {
				if tc.output.err != nil {
					if diff := cmp.Diff(err, tc.output.err, cmpopts.EquateErrors()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			}

			if diff := cmp.Diff(actualSite, tc.output.site); diff != "" {
				t.Fatal(diff)
			}

			if diff := cmp.Diff(actualOccupancyEligibility, tc.output.occupancyEligibility); diff != "" {
				t.Fatal(diff)
			}
		})

	}
}
