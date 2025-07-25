package evaluation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	smart "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart/v1"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/testcommon"
)

func TestRunFull(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	eMockSync := testcommon.MockSink{}
	sMockSync := testcommon.MockSink{}
	cMockSync := testcommon.MockSink{}
	bMockSync := testcommon.MockSink{}

	type testCase struct {
		description string
		occupancyID string
		evaluator   Evaluator
		checkOutput func()
	}

	var capacity float32 = 6

	testCases := []testCase{
		{
			description: "occupancy eligible for smart booking journey with booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(bMockSync.Msgs) == 1)
				assert.Equal(&smart.SmartBookingJourneyOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					Reference:   "booking-ref",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy eligible for smart booking journey but booking ref missing",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(bMockSync.Msgs) == 1)
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy eligible for smart booking journey with booking ref with no previous evaluation run",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: false,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.EligibleOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, eMockSync.Msgs[0])
				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
				assert.Equal(&smart.CampaignableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, cMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					Reference:   "booking-ref",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for smart booking journey with booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID:     "account-id",
							OptOut: true,
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonBookingOptOut},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						Capacity:   &capacity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				assert.True(len(bMockSync.Msgs) == 1)
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(_ *testing.T) {
			err := tc.evaluator.RunFull(ctx, tc.occupancyID)
			assert.NoError(err)
			tc.checkOutput()

			// cleanup
			eMockSync.Msgs = nil
			sMockSync.Msgs = nil
			cMockSync.Msgs = nil
			bMockSync.Msgs = nil
		})
	}
}

func TestRunSuppliability(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	eMockSync := testcommon.MockSink{}
	sMockSync := testcommon.MockSink{}
	cMockSync := testcommon.MockSink{}
	bMockSync := testcommon.MockSink{}

	type testCase struct {
		description string
		occupancyID string
		evaluator   Evaluator
		checkOutput func()
	}

	var capacity float32 = 6
	var largeCapacity float32 = 21

	testCases := []testCase{
		{
			description: "occupancy suppliable if gas service only",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonGasServiceOnly},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   &capacity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy suppliable if meter exists with no capacity defined",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonMissingMeterData},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   nil,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable if meter does not exist",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore:             &mockStore{meters: map[string]domain.Meter{}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_MISSING_METER_DATA},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy suppliable if meter has large capacity",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   &largeCapacity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_METER_LARGE_CAPACITY},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable if site does not exist",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   nil,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_MISSING_SITE_DATA},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable if site does not have wan coverage",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   nil,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable if no active services",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   nil,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable if meterpoint alt han",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeGas,
							BookingReference: "booking-ref",
							Meterpoint: &domain.Meterpoint{
								Mpxn:   "mpxn",
								AltHan: true,
							},
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeGas,
						Capacity:   nil,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_ALT_HAN},
				}, sMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable for smart booking journey with no previous suppliability evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, sMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable for smart booking journey with previous suppliability evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not suppliable for smart booking journey with different previous suppliability evaluation results",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SuppliableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, sMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming suppliable for smart booking journey with no booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming suppliable for smart booking journey with booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					Reference:   "booking-ref",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming suppliable for smart booking journey with booking ref but not all criteria evaluated",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only suppliable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(sMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.SuppliableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, sMockSync.Msgs[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(_ *testing.T) {
			err := tc.evaluator.RunSuppliability(ctx, tc.occupancyID)
			assert.NoError(err)
			tc.checkOutput()

			// cleanup
			eMockSync.Msgs = nil
			sMockSync.Msgs = nil
			cMockSync.Msgs = nil
			bMockSync.Msgs = nil
		})
	}
}

func TestRunCampaignability(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	eMockSync := testcommon.MockSink{}
	sMockSync := testcommon.MockSink{}
	cMockSync := testcommon.MockSink{}
	bMockSync := testcommon.MockSink{}

	type testCase struct {
		description string
		occupancyID string
		evaluator   Evaluator
		checkOutput func()
	}

	testCases := []testCase{
		{
			description: "occupancy not campaignable if account opt out",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID:     "account-id",
							OptOut: true,
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: false,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.CampaignableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_SMART_BOOKING_OPT_OUT},
				}, cMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not campaignable if no active services",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: false,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.CampaignableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE},
				}, cMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not campaignable for smart booking journey with no previous campaignability evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID:     "account-id",
							OptOut: true,
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: false,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.CampaignableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_SMART_BOOKING_OPT_OUT},
				}, cMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not campaignable for smart booking journey with previous campaignability evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID:     "account-id",
							OptOut: true,
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonBookingOptOut},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not campaignable for smart booking journey with different previous campaignability evaluation results",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID:     "account-id",
							OptOut: true,
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.CampaignableOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_SMART_BOOKING_OPT_OUT},
				}, cMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming campaignable for smart booking journey with no booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.CampaignableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, cMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming campaignable for smart booking journey with booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.CampaignableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, cMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					Reference:   "booking-ref",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming campaignable for smart booking journey with booking ref but not all criteria evaluated",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only campaignable + booking events should be published
				assert.True(len(eMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				assert.True(len(cMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.CampaignableOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, cMockSync.Msgs[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(_ *testing.T) {
			err := tc.evaluator.RunCampaignability(ctx, tc.occupancyID)
			assert.NoError(err)
			tc.checkOutput()

			// cleanup
			eMockSync.Msgs = nil
			sMockSync.Msgs = nil
			cMockSync.Msgs = nil
			bMockSync.Msgs = nil
		})
	}
}

func TestRunEligibility(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	eMockSync := testcommon.MockSink{}
	sMockSync := testcommon.MockSink{}
	cMockSync := testcommon.MockSink{}
	bMockSync := testcommon.MockSink{}

	type testCase struct {
		description string
		occupancyID string
		evaluator   Evaluator
		checkOutput func()
	}

	testCases := []testCase{
		{
			description: "occupancy not eligible for if site missing",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_MISSING_SITE_DATA},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if site does not have wan coverage",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if no active services",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NOT_ACTIVE},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if missing meterpoint data",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:               "service-id",
							Mpxn:             "mpxn",
							SupplyType:       energy_domain.SupplyTypeElectricity,
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_MISSING_METERPOINT_DATA},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if it has complex tariff",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_02,
								SSC:          "0003",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_COMPLEX_TARIFF},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if missing meter data",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore:             &mockStore{meters: map[string]domain.Meter{}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_MISSING_METER_DATA},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for if meter already smart",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2A.String(),
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_ALREADY_SMART},
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy eligible, meter is already smart but is in MSN exception list",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        exceptionMSN,
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  platform.MeterTypeElec_METER_TYPE_ELEC_S2A.String(),
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, eMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for smart booking journey with no previous eligibility evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     false,
							Eligibility:              nil,
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, eMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for smart booking journey with previous eligbility evaluation",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(cMockSync.Msgs) == 0)
				assert.True(len(sMockSync.Msgs) == 0)

				// evaluation result is same as previous evaluation
				assert.True(len(eMockSync.Msgs) == 0)

				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy not eligible for smart booking journey with different previous eligibility evaluation results",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: false,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							SuppliabilityEvaluated:   true,
							Suppliability:            domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.EligibleOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
					Reasons:     []smart.IneligibleReason{smart.IneligibleReason_INELIGIBLE_REASON_NO_WAN_COVERAGE},
				}, eMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming eligible for smart booking journey with no booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.EligibleOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, eMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyRemovedEvent{
					OccupancyId: "occupancy-id",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming eligible for smart booking journey with booking ref",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoWanCoverage},
							SuppliabilityEvaluated:   true,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 1)

				assert.Equal(&smart.EligibleOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, eMockSync.Msgs[0])
				assert.Equal(&smart.SmartBookingJourneyOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					Reference:   "booking-ref",
				}, bMockSync.Msgs[0])
			},
		},
		{
			description: "occupancy becoming eligible for smart booking journey with booking ref but not all criteria evaluated",
			occupancyID: "occupancy-id",
			evaluator: Evaluator{
				meterSerialNumberStore: &mockMeterSerialNumberStore{},
				occupancyStore: &mockStore{occupancies: map[string]domain.Occupancy{
					"occupancy-id": {
						ID: "occupancy-id",
						Account: domain.Account{
							ID: "account-id",
						},
						Site: &domain.Site{
							ID:          "site-id",
							Postcode:    "AP 24X",
							WanCoverage: true,
						},
						EvaluationResult: domain.OccupancyEvaluation{
							OccupancyID:              "occupancy-id",
							EligibilityEvaluated:     true,
							Eligibility:              domain.IneligibleReasons{domain.IneligibleReasonNoActiveService},
							SuppliabilityEvaluated:   false,
							Suppliability:            nil,
							CampaignabilityEvaluated: true,
							Campaignability:          nil,
						},
					},
				}},
				serviceStore: &mockStore{servicesByOccupancy: map[string][]domain.Service{
					"occupancy-id": {
						{
							ID:         "service-id",
							Mpxn:       "mpxn",
							SupplyType: energy_domain.SupplyTypeElectricity,
							Meterpoint: &domain.Meterpoint{
								Mpxn:         "mpxn",
								AltHan:       false,
								ProfileClass: platform.ProfileClass_PROFILE_CLASS_06,
								SSC:          "ssc",
							},
							BookingReference: "booking-ref",
						},
					},
				}},
				meterStore: &mockStore{meters: map[string]domain.Meter{
					"mpxn": {
						ID:         "meter-id",
						Mpxn:       "mpxn",
						MSN:        "msn",
						SupplyType: energy_domain.SupplyTypeElectricity,
						MeterType:  "some_type",
					},
				}},
				eligibilitySync:        &eMockSync,
				suppliabilitySync:      &sMockSync,
				campaignabilitySync:    &cMockSync,
				bookingEligibilitySync: &bMockSync,
			},
			checkOutput: func() {
				// only eligible + booking events should be published
				assert.True(len(sMockSync.Msgs) == 0)
				assert.True(len(cMockSync.Msgs) == 0)

				assert.True(len(eMockSync.Msgs) == 1)
				assert.True(len(bMockSync.Msgs) == 0)

				assert.Equal(&smart.EligibleOccupancyAddedEvent{
					OccupancyId: "occupancy-id",
					AccountId:   "account-id",
				}, eMockSync.Msgs[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(_ *testing.T) {
			err := tc.evaluator.RunEligibility(ctx, tc.occupancyID)
			assert.NoError(err)
			tc.checkOutput()

			// cleanup
			eMockSync.Msgs = nil
			sMockSync.Msgs = nil
			cMockSync.Msgs = nil
			bMockSync.Msgs = nil
		})
	}
}

type mockStore struct {
	occupancies         map[string]domain.Occupancy
	servicesByOccupancy map[string][]domain.Service
	meters              map[string]domain.Meter
}

type mockMeterSerialNumberStore struct{}

const exceptionMSN = "expcetion_msn"

func (m *mockMeterSerialNumberStore) FindMeterSerialNumber(s string) bool {
	return strings.EqualFold(s, exceptionMSN)
}

func (s *mockStore) LoadOccupancy(_ context.Context, id string) (domain.Occupancy, error) {
	if occ, ok := s.occupancies[id]; ok {
		return occ, nil
	}
	return domain.Occupancy{}, store.ErrOccupancyNotFound
}

func (s *mockStore) LoadLiveServicesByOccupancyID(_ context.Context, occupancyID string) ([]domain.Service, error) {
	return s.servicesByOccupancy[occupancyID], nil
}

func (s *mockStore) GetLiveServicesWithBookingRef(_ context.Context, occupancyID string) ([]store.ServiceBookingRef, error) {
	services := s.servicesByOccupancy[occupancyID]
	refs := make([]store.ServiceBookingRef, len(services))

	timestamp := time.Now()

	for i, sv := range services {
		if sv.BookingReference == "" {
			refs[i] = store.ServiceBookingRef{ServiceID: sv.ID, BookingRef: "removed", DeletedAt: &timestamp}
		} else {
			refs[i] = store.ServiceBookingRef{ServiceID: sv.ID, BookingRef: sv.BookingReference}
		}
	}

	return refs, nil
}

func (s *mockStore) Get(_ context.Context, mpxn string) (domain.Meter, error) {
	if m, ok := s.meters[mpxn]; ok {
		return m, nil
	}
	return domain.Meter{}, store.ErrMeterNotFound
}
