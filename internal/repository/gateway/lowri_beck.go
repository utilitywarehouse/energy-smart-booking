package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	lowribeckv1 "github.com/utilitywarehouse/energy-contracts/pkg/generated/third_party/lowribeck/v1"
	"github.com/utilitywarehouse/energy-smart-booking/internal/models"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrBadParameters      = errors.New("bad parameters")
	ErrNotFound           = errors.New("not found")
	ErrInternal           = errors.New("internal error")
	ErrUnhandledErrorCode = errors.New("error code not handled")
)

type LowriBeckGateway struct {
	mai    MachineAuthInjector
	client LowriBeckClient
}

func NewLowriBeckGateway(mai MachineAuthInjector, client LowriBeckClient) LowriBeckGateway {
	return LowriBeckGateway{mai, client}
}

type AvailableSlotsResponse struct {
	BookingSlots []models.BookingSlot
}

type CreateBookingResponse struct {
	Success bool
}

func (g LowriBeckGateway) GetAvailableSlots(ctx context.Context, postcode, reference string) (AvailableSlotsResponse, error) {
	availableSlots, err := g.client.GetAvailableSlots(g.mai.ToCtx(ctx), &lowribeckv1.GetAvailableSlotsRequest{
		Postcode:  postcode,
		Reference: reference,
	})
	if err != nil {
		logrus.Errorf("failed to get available slots, %s, %s", ErrInternal, err)

		details := status.Convert(err).Details()

		for _, detail := range details {
			switch x := detail.(type) {
			case lowribeckv1.InvalidParameterResponse:
				logrus.Errorf("!found details!, %s", x.GetParameters().String())
			}
		}

		switch status.Convert(err).Code() {
		case codes.Internal:
			return AvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", ErrInternal)
		case codes.NotFound:
			return AvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", ErrNotFound)
		case codes.InvalidArgument:

			details := status.Convert(err).Details()

			for _, detail := range details {
				switch x := detail.(type) {
				case lowribeckv1.InvalidParameterResponse:
					logrus.Errorf("!found details!, %s", x.GetParameters().String())
				}
			}

			return AvailableSlotsResponse{}, fmt.Errorf("failed to get available slots, %w", ErrBadParameters)
		}
		return AvailableSlotsResponse{}, ErrUnhandledErrorCode
	}

	slots := []models.BookingSlot{}

	for _, elem := range availableSlots.GetSlots() {
		slots = append(slots, models.BookingSlot{
			Date:      time.Date(int(elem.Date.Year), time.Month(elem.Date.Month), int(elem.Date.Day), 0, 0, 0, 0, time.UTC),
			StartTime: int(elem.GetStartTime()),
			EndTime:   int(elem.GetEndTime()),
		})
	}

	return AvailableSlotsResponse{
		BookingSlots: slots,
	}, nil
}

func (g LowriBeckGateway) CreateBooking(ctx context.Context, postcode, reference string, slot models.BookingSlot, accountDetails models.AccountDetails, vulnerabilities []lowribeckv1.Vulnerability, other string) (CreateBookingResponse, error) {

	bookingResponse, err := g.client.CreateBooking(g.mai.ToCtx(ctx), &lowribeckv1.CreateBookingRequest{
		Postcode:  postcode,
		Reference: reference,
		Slot: &lowribeckv1.BookingSlot{
			Date: &date.Date{
				Year:  int32(slot.Date.Year()),
				Month: int32(slot.Date.Month()),
				Day:   int32(slot.Date.Day()),
			},
			StartTime: int32(slot.StartTime),
			EndTime:   int32(slot.EndTime),
		},
		VulnerabilityDetails: &lowribeckv1.VulnerabilityDetails{
			Vulnerabilities: vulnerabilities,
			Other:           other,
		},
		ContactDetails: &lowribeckv1.ContactDetails{
			Title:     accountDetails.Title,
			FirstName: accountDetails.FirstName,
			LastName:  accountDetails.LastName,
			Phone:     accountDetails.Mobile,
		},
	})
	if err != nil {
		details := status.Convert(err).Details()

		for _, detail := range details {
			switch x := detail.(type) {
			case lowribeckv1.InvalidParameterResponse:
				logrus.Errorf("!found details!, %s", x.GetParameters().String())
			}
		}

		switch status.Convert(err).Code() {
		case codes.Internal:
			return CreateBookingResponse{
				Success: false,
			}, fmt.Errorf("failed to call lowribeck create booking, %w", err)
		case codes.InvalidArgument:
			return CreateBookingResponse{
				Success: false,
			}, fmt.Errorf("failed to call lowribeck create booking, %w", err)
		default:
			return CreateBookingResponse{
				Success: false,
			}, fmt.Errorf("failed to call lowribeck create booking, %w", err)
		}
	}

	return CreateBookingResponse{
		Success: bookingResponse.Success,
	}, nil
}
