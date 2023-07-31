package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	bookingv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/smart_booking/booking/v1"
	"github.com/utilitywarehouse/energy-pkg/postgres"
	"github.com/utilitywarehouse/energy-smart-booking/cmd/booking-api/internal/repository/store"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustDateFromString(t *testing.T, value string) time.Time {
	t.Helper()
	date, err := models.DateFromString(value)
	must(t, err)
	return date
}

func storeInit(t *testing.T) (context.Context, *store.BookingStore) {
	t.Helper()
	ctx := context.Background()

	testContainer, err := postgres.SetupTestContainer(ctx)
	must(t, err)

	dsn, err := postgres.GetTestContainerDSN(testContainer)
	must(t, err)

	db, err := store.Setup(ctx, dsn)
	must(t, err)

	store := store.NewBooking(db)

	return ctx, store
}

type TestCaseBase[Input any, Expected any] struct {
	Description string
	I           Input
	E           Expected
}

func makeDummyBooking(
	bookingID, accountID, siteID string,
	bookingStatus bookingv1.BookingStatus,
	slot models.BookingSlot,
	vulnerabilities models.Vulnerabilities,
) models.Booking {
	return models.Booking{
		BookingID: bookingID,
		AccountID: accountID,
		Status:    bookingStatus,
		SiteID:    siteID,
		Contact: models.ContactDetails{
			Title:     "Mr.",
			FirstName: "Foo",
			LastName:  "Bar",
			Phone:     "5555555",
			Email:     "foobar@example.com",
		},
		Slot: slot,
		VulnerabilityDetails: models.VulnerabilityDetails{
			Vulnerabilities: vulnerabilities,
			Other:           "",
		},
	}
}

func makeBookingSlot(t *testing.T, datestr string, start, end int) models.BookingSlot {
	t.Helper()
	return models.BookingSlot{
		Date:      mustDateFromString(t, datestr),
		StartTime: start,
		EndTime:   end,
	}
}

func Test_BookingStore_Upsert(t *testing.T) {
	ctx, store := storeInit(t)

	type upsertTestCase TestCaseBase[models.Booking, error]

	testCases := []upsertTestCase{
		{
			Description: "basic upsert",
			I: makeDummyBooking(
				"booking-id-1",
				"account-id-1",
				"site-id-1",
				bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
				makeBookingSlot(t, "2023-09-16", 13, 15),
				models.Vulnerabilities{bookingv1.Vulnerability_VULNERABILITY_ILLNESS},
			),
			E: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			store.Begin()
			store.Upsert(tc.I)
			err := store.Commit(ctx)
			if !errors.Is(err, tc.E) {
				t.Fatalf("expected error: %s, actual returned error: %s", tc.E, err)
			}
		})
	}
}

func Test_BookingStore_UpsertAndQuery(t *testing.T) {
	ctx, store := storeInit(t)

	type upsertAndQueryInput struct {
		QueriedAccountID string
		Insertions       []models.Booking
	}
	type upsertAndQueryTestCase TestCaseBase[upsertAndQueryInput, []models.Booking]

	testCases := []upsertAndQueryTestCase{
		{
			Description: "basic upsert and get",
			I: upsertAndQueryInput{
				QueriedAccountID: "account-id-1",
				Insertions: []models.Booking{
					makeDummyBooking(
						"booking-id-1", "account-id-1", "site-id-1",
						bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						makeBookingSlot(t, "2023-05-01", 13, 15),
						models.Vulnerabilities{}),
				}},
			E: []models.Booking{
				makeDummyBooking(
					"booking-id-1", "account-id-1", "site-id-1",
					bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					makeBookingSlot(t, "2023-05-01", 13, 15),
					models.Vulnerabilities{}),
			},
		},
		{
			Description: "upsert and get multiple",
			I: upsertAndQueryInput{
				QueriedAccountID: "account-id-1",
				Insertions: []models.Booking{
					makeDummyBooking(
						"booking-id-1", "account-id-1", "site-id-1",
						bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
						makeBookingSlot(t, "2023-05-01", 13, 15),
						models.Vulnerabilities{}),
					makeDummyBooking(
						"booking-id-2", "account-id-1", "site-id-2",
						bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						makeBookingSlot(t, "2023-09-16", 13, 15),
						models.Vulnerabilities{}),
					makeDummyBooking(
						"booking-id-3", "account-id-2", "site-id-3",
						bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
						makeBookingSlot(t, "2023-08-07", 10, 12),
						models.Vulnerabilities{}),
				}},
			E: []models.Booking{
				makeDummyBooking(
					"booking-id-1", "account-id-1", "site-id-1",
					bookingv1.BookingStatus_BOOKING_STATUS_COMPLETED,
					makeBookingSlot(t, "2023-05-01", 13, 15),
					models.Vulnerabilities{}),
				makeDummyBooking(
					"booking-id-2", "account-id-1", "site-id-2",
					bookingv1.BookingStatus_BOOKING_STATUS_SCHEDULED,
					makeBookingSlot(t, "2023-09-16", 13, 15),
					models.Vulnerabilities{}),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			store.Begin()
			for _, b := range tc.I.Insertions {
				store.Upsert(b)
			}
			store.Commit(ctx)

			retrieved, err := store.GetBookingsByAccountID(ctx, tc.I.QueriedAccountID)
			must(t, err)

			if diff := cmp.Diff(tc.E, retrieved); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
