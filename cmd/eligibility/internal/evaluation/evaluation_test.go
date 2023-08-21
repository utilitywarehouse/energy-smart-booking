package evaluation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-contracts/pkg/generated/smart"
	energy_domain "github.com/utilitywarehouse/energy-pkg/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/domain"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/eligibility/internal/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/test_common"
)

func TestPublishCampaignabilityIfChanged(t *testing.T) {
	ctx := context.Background()
	mockSync := test_common.MockSink{}
	evaluator := Evaluator{
		campaignabilitySync: &mockSync,
	}
	occupancy := &domain.Occupancy{
		ID: "occupancyID",
		Account: domain.Account{
			ID: "accountID",
		},
		Site:     nil,
		Services: nil,
		EvaluationResult: domain.OccupancyEvaluation{
			OccupancyID: "occupancyID",
		},
	}
	err := evaluator.publishCampaignabilityIfChanged(ctx, occupancy, domain.IneligibleReasons{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(mockSync.Msgs))
	assert.Equal(t, &smart.CampaignableOccupancyAddedEvent{OccupancyId: "occupancyID", AccountId: "accountID"}, mockSync.Msgs[0])
}

func TestRunFull(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	eMockSync := test_common.MockSink{}
	sMockSync := test_common.MockSink{}
	cMockSync := test_common.MockSink{}
	bMockSync := test_common.MockSink{}

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
		t.Run(tc.description, func(t *testing.T) {
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

type mockStore struct {
	occupancies         map[string]domain.Occupancy
	servicesByOccupancy map[string][]domain.Service
	meters              map[string]domain.Meter
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

func (s *mockStore) GetServicesWithBookingRef(_ context.Context, occupancyID string) ([]store.ServiceBookingRef, error) {
	services := s.servicesByOccupancy[occupancyID]
	refs := make([]store.ServiceBookingRef, len(services))

	for i, sv := range services {
		refs[i] = store.ServiceBookingRef{ServiceID: sv.ID, BookingRef: sv.BookingReference}
	}

	return refs, nil
}

func (s *mockStore) Get(_ context.Context, mpxn string) (domain.Meter, error) {
	if m, ok := s.meters[mpxn]; ok {
		return m, nil
	}
	return domain.Meter{}, store.ErrMeterNotFound
}
