package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/protobuf/testing/protocmp"
)

// This test case tests for both upsert(only insert) and get actions.
func Test_PartialBookingStore_Insert_Get(t *testing.T) {
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

	partialBookingStore := store.NewPartialBooking(db)

	type inputParams struct {
		bookingID string
		event     *bookingv1.BookingCreatedEvent
	}

	type outputParams struct {
		event *models.PartialBooking
		err   error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should upsert a booking created event and retrieve it",
			input: inputParams{
				bookingID: "booking-id-1",
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId: "account-id-1",
					},
				},
			},
			output: outputParams{
				err: nil,
				event: &models.PartialBooking{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:   "booking-id-1",
						OccupancyId: "",
						Details: &bookingv1.Booking{
							AccountId: "account-id-1",
						},
					},
					BookingID:      "booking-id-1",
					Retries:        0,
					UpdatedAt:      nil,
					DeletedAt:      nil,
					CreatedAt:      time.Now(),
					DeletionReason: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := partialBookingStore.Get(ctx, "booking-id-1")

			if diff := cmp.Diff(tc.output.event, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_PartialBookingStore_Upsert_Get(t *testing.T) {
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

	partialBookingStore := store.NewPartialBooking(db)

	type inputParams struct {
		bookingID string
		event     *bookingv1.BookingCreatedEvent
		event2    *bookingv1.BookingCreatedEvent
	}

	type outputParams struct {
		event *models.PartialBooking
		err   error
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	testCases := []testSetup{
		{
			description: "should upsert a booking created event for booking-id-1 and get the upserted value",
			input: inputParams{
				bookingID: "booking-id-1",
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId:         "account-id-1",
						ExternalReference: "LBG1001",
					},
				},
				event2: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId:         "account-id-1",
						ExternalReference: "LBG1002",
					},
				},
			},
			output: outputParams{
				err: nil,
				event: &models.PartialBooking{
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:   "booking-id-1",
						OccupancyId: "",
						Details: &bookingv1.Booking{
							AccountId:         "account-id-1",
							ExternalReference: "LBG1002",
						},
					},
					BookingID:      "booking-id-1",
					Retries:        0,
					UpdatedAt:      nil,
					DeletedAt:      nil,
					CreatedAt:      time.Now(),
					DeletionReason: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			err = partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event2)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := partialBookingStore.Get(ctx, "booking-id-1")

			if diff := cmp.Diff(tc.output.event, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}
		})

	}
}

func Test_PartialBookingStore_Insert_Delete_Get(t *testing.T) {
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

	partialBookingStore := store.NewPartialBooking(db)

	type inputParams struct {
		bookingID      string
		deletionReason models.DeletionReason
		event          *bookingv1.BookingCreatedEvent
	}

	type outputParams struct {
		err            error
		partialBooking *models.PartialBooking
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	timeNow := time.Now()
	reasonCreated := models.DeletionReason_BookingCreated
	reasonExpired := models.DeletionReason_BookingExpired

	testCases := []testSetup{
		{
			description: "should upsert an booking created event, mark it as deleted and retrieve it",
			input: inputParams{
				bookingID:      "booking-id-1",
				deletionReason: reasonCreated,
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId: "account-id-1",
					},
				},
			},
			output: outputParams{
				err: nil,
				partialBooking: &models.PartialBooking{
					BookingID: "booking-id-1",
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:   "booking-id-1",
						OccupancyId: "",
						Details: &bookingv1.Booking{
							AccountId: "account-id-1",
						},
					},
					CreatedAt:      timeNow,
					UpdatedAt:      nil,
					DeletedAt:      &timeNow,
					Retries:        0,
					DeletionReason: &reasonCreated,
				},
			},
		},
		{
			description: "should upsert a booking created event, mark it as deleted due to expiration and retrieve it",
			input: inputParams{
				bookingID:      "booking-id-1",
				deletionReason: reasonExpired,
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId: "account-id-1",
					},
				},
			},
			output: outputParams{
				err: nil,
				partialBooking: &models.PartialBooking{
					BookingID: "booking-id-1",
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:   "booking-id-1",
						OccupancyId: "",
						Details: &bookingv1.Booking{
							AccountId: "account-id-1",
						},
					},
					CreatedAt:      timeNow,
					UpdatedAt:      nil,
					DeletedAt:      &timeNow,
					Retries:        0,
					DeletionReason: &reasonExpired,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			err = partialBookingStore.MarkAsDeleted(ctx, tc.input.bookingID, tc.input.deletionReason)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := partialBookingStore.Get(ctx, tc.input.bookingID)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			if diff := cmp.Diff(tc.output.partialBooking, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}

func Test_PartialBookingStore_Insert_UpdateRetries_Get(t *testing.T) {
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

	partialBookingStore := store.NewPartialBooking(db)

	type inputParams struct {
		bookingID string
		event     *bookingv1.BookingCreatedEvent
	}

	type outputParams struct {
		err            error
		partialBooking *models.PartialBooking
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	timeNow := time.Now()

	testCases := []testSetup{
		{
			description: "should upsert an booking created event, update the retries count and retrieve it",
			input: inputParams{
				bookingID: "booking-id-1",
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId: "account-id-1",
					},
				},
			},
			output: outputParams{
				err: nil,
				partialBooking: &models.PartialBooking{
					BookingID: "booking-id-1",
					Event: &bookingv1.BookingCreatedEvent{
						BookingId:   "booking-id-1",
						OccupancyId: "",
						Details: &bookingv1.Booking{
							AccountId: "account-id-1",
						},
					},
					CreatedAt:      timeNow,
					UpdatedAt:      &timeNow,
					DeletedAt:      nil,
					Retries:        1,
					DeletionReason: nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			err = partialBookingStore.UpdateRetries(ctx, tc.input.bookingID, 0)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := partialBookingStore.Get(ctx, tc.input.bookingID)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			if diff := cmp.Diff(tc.output.partialBooking, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}

// tests the GetPending
func Test_PartialBookingStore_GetPendingProcessing(t *testing.T) {
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

	partialBookingStore := store.NewPartialBooking(db)

	type inputParams struct {
		bookingID string
		accountID string
		event     *bookingv1.BookingCreatedEvent
	}

	type outputParams struct {
		err            error
		partialBooking []*models.PartialBooking
	}

	type testSetup struct {
		description string
		input       inputParams
		output      outputParams
	}

	timeNow := time.Now()

	testCases := []testSetup{
		{
			description: "should upsert an booking created event, update the retries count and retrieve it",
			input: inputParams{
				bookingID: "booking-id-1",
				event: &bookingv1.BookingCreatedEvent{
					BookingId:   "booking-id-1",
					OccupancyId: "",
					Details: &bookingv1.Booking{
						AccountId: "account-id-1",
					},
				},
			},
			output: outputParams{
				err: nil,
				partialBooking: []*models.PartialBooking{
					{
						BookingID: "booking-id-1",
						Event: &bookingv1.BookingCreatedEvent{
							BookingId:   "booking-id-1",
							OccupancyId: "",
							Details: &bookingv1.Booking{
								AccountId: "account-id-1",
							},
						},
						CreatedAt:      timeNow,
						UpdatedAt:      nil,
						DeletedAt:      nil,
						Retries:        0,
						DeletionReason: nil,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {

			err := partialBookingStore.Upsert(ctx, tc.input.bookingID, tc.input.event)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			actual, err := partialBookingStore.GetPending(ctx)
			if err != nil {
				t.Fatalf("should not have errored, %s", err)
			}

			if diff := cmp.Diff(tc.output.partialBooking, actual, cmpopts.IgnoreUnexported(), protocmp.Transform(), cmpopts.EquateApproxTime(time.Second)); diff != "" {
				t.Fatal(diff)
			}

		})

	}
}
