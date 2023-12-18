package evaluation

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/utilitywarehouse/energy-contracts/pkg/generated/platform"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

type mockWanStore struct{ store map[string]bool }

func (s *mockWanStore) GetWanCoverage(_ context.Context, postcode string) (bool, error) {
	wan, _ := s.store[postcode]
	return wan, nil
}

type mockAltHanStore struct{ store map[string]bool }

func (s *mockAltHanStore) GetAltHan(_ context.Context, mpxn string) (bool, error) {
	altHan, _ := s.store[mpxn]
	return altHan, nil
}

type mockEcoesAPI struct {
	technicalDetailResponses map[string]*models.ElectricityMeterTechnicalDetails
	relatedMPANResponses     map[string]*models.ElectricityMeterRelatedMPAN
}

func (e *mockEcoesAPI) GetMPANTechnicalDetails(_ context.Context, mpan string) (*models.ElectricityMeterTechnicalDetails, error) {
	details, ok := e.technicalDetailResponses[mpan]
	if !ok {
		return nil, fmt.Errorf("failed to get technical details by mpan: %s, no entry", mpan)
	}
	return details, nil
}

func (e *mockEcoesAPI) HasRelatedMPAN(_ context.Context, mpan string) (bool, error) {
	related, ok := e.relatedMPANResponses[mpan]
	if ok == false {
		return false, nil
	}
	return len(related.Relations) > 0, nil
}

type mockXoserveAPI struct {
	technicalDetailResponses map[string]*models.GasMeterTechnicalDetails
}

func (x *mockXoserveAPI) GetMPRNTechnicalDetails(_ context.Context, mprn string) (*models.GasMeterTechnicalDetails, error) {
	details, ok := x.technicalDetailResponses[mprn]
	if !ok {
		return nil, fmt.Errorf("failed to get technical details by mprn: %s, no entry", mprn)
	}
	return details, nil
}

type meterpointEligibilityMocks struct {
	wanStore                         map[string]bool
	altHanStore                      map[string]bool
	ecoesTechnicalDetailsResponses   map[string]*models.ElectricityMeterTechnicalDetails
	ecoesRelatedMPANResponses        map[string]*models.ElectricityMeterRelatedMPAN
	xoserveTechnicalDetailsResponses map[string]*models.GasMeterTechnicalDetails
}

type electricityMeterpointEligibilityTestCases struct {
	mpan     string
	postcode string

	description         string
	expectedEligibility bool
	expectedReason      MeterpointIneligibleReason
}

func mustTime(date string) time.Time {
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestGetElectricityMeterpointEligibility(t *testing.T) {
	electricityMeterpointTestMocks := meterpointEligibilityMocks{
		wanStore: map[string]bool{
			"post-code-1": true,
			"post-code-2": true,
			"post-code-4": true,
			"post-code-5": true,
			"post-code-6": true,
		},
		altHanStore: map[string]bool{
			"mpan-2": true,
		},
		ecoesTechnicalDetailsResponses: map[string]*models.ElectricityMeterTechnicalDetails{
			"mpan-1": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
				SettlementStandardConfiguration: "0123",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_HALF_HOURLY,
					InstalledAt: mustTime("2020-11-11"),
				}},
			},
			"mpan-2": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
				SettlementStandardConfiguration: "0123",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_HALF_HOURLY,
					InstalledAt: mustTime("2020-10-10"),
				}},
			},
			"mpan-3": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
				SettlementStandardConfiguration: "0123",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_HALF_HOURLY,
					InstalledAt: mustTime("2020-09-09"),
				}},
			},
			"mpan-4": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
				SettlementStandardConfiguration: "0123",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_HALF_HOURLY,
					InstalledAt: mustTime("2020-08-08"),
				}},
			},
			"mpan-5": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_02,
				SettlementStandardConfiguration: "0110",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_HALF_HOURLY,
					InstalledAt: mustTime("2020-07-07"),
				}},
			},
			"mpan-6": {
				ProfileClass:                    platform.ProfileClass_PROFILE_CLASS_01,
				SettlementStandardConfiguration: "0123",
				Meters: []models.ElectricityMeter{{
					MeterType:   platform.MeterTypeElec_METER_TYPE_ELEC_S2A,
					InstalledAt: mustTime("2020-06-06"),
				}},
			},
		},
		ecoesRelatedMPANResponses: map[string]*models.ElectricityMeterRelatedMPAN{
			"mpan-4": {
				Relations: []models.MPANRelation{
					{
						Primary:   "mpan-4",
						Secondary: "mpan-4-related",
					},
				},
			},
		},
		xoserveTechnicalDetailsResponses: map[string]*models.GasMeterTechnicalDetails{},
	}

	testCases := []electricityMeterpointEligibilityTestCases{
		{
			mpan:                "mpan-1",
			postcode:            "post-code-1",
			description:         "standard eligible test case",
			expectedEligibility: true,
		}, {
			mpan:                "mpan-2",
			postcode:            "post-code-2",
			description:         "ineligible because alt-HAN",
			expectedEligibility: false,
			expectedReason:      "Alt_HAN",
		},
		{
			mpan:                "mpan-3",
			postcode:            "post-code-3",
			description:         "ineligible because no WAN",
			expectedEligibility: false,
			expectedReason:      "not_WAN",
		},
		{
			mpan:                "mpan-4",
			postcode:            "post-code-4",
			description:         "ineligible because has related MPAN",
			expectedEligibility: false,
			expectedReason:      "related_meterpoints_present",
		},
		{
			mpan:                "mpan-5",
			postcode:            "post-code-5",
			description:         "ineligible because complex SSC",
			expectedEligibility: false,
			expectedReason:      "complex_SSC",
		},
		{
			mpan:                "mpan-6",
			postcode:            "post-code-6",
			description:         "ineligible because already smart meter",
			expectedEligibility: false,
			expectedReason:      "already_a_smart_meter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			evaluator := NewMeterpointEvaluator(
				&mockWanStore{store: electricityMeterpointTestMocks.wanStore},
				&mockAltHanStore{store: electricityMeterpointTestMocks.altHanStore},
				&mockEcoesAPI{
					technicalDetailResponses: electricityMeterpointTestMocks.ecoesTechnicalDetailsResponses,
					relatedMPANResponses:     electricityMeterpointTestMocks.ecoesRelatedMPANResponses,
				}, &mockXoserveAPI{
					technicalDetailResponses: electricityMeterpointTestMocks.xoserveTechnicalDetailsResponses,
				},
			)

			actualEligibility, actualReason, err := evaluator.GetElectricityMeterpointEligibility(context.Background(), tc.mpan, tc.postcode)
			if err != nil {
				t.Fatal(err)
			}
			if tc.expectedEligibility != actualEligibility {
				t.Fatalf("unexpected eligibility result for %s (expected %t, got %t)", tc.description, tc.expectedEligibility, actualEligibility)
			}
			if tc.expectedReason != actualReason {
				t.Fatalf("unexpected eligibility faiure reason for %s (expected %q, got %q)", tc.description, tc.expectedReason, actualReason)
			}
		})
	}
}
