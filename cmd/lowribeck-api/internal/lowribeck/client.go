package lowribeck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/utilitywarehouse/energy-smart-booking/cmd/lowribeck-api/internal/metrics"
	"github.com/utilitywarehouse/uwos-go/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotOKStatusCode = errors.New("error status code is not 200(OK)")
	ErrTimeout         = errors.New("timeout occurred")
)

const (
	contentHeader    = "Content-Type"
	contentJSON      = "application/json"
	availabilityURL  = "appointmentManagement/getCalendarAvailability"
	bookingURL       = "appointmentManagement/book"
	updateContactURL = "appointmentManagement/updateContact"
	healthCheckURL   = "health/get"
)

type Client struct {
	http    *http.Client
	auth    auth
	baseURL string
}

func New(c *http.Client, user, password, url string) *Client {
	return &Client{
		http: c,
		auth: auth{
			user:     user,
			password: password,
		},
		baseURL: url,
	}
}

func (c *Client) GetCalendarAvailability(ctx context.Context, req *GetCalendarAvailabilityRequest) (_ *GetCalendarAvailabilityResponse, err error) {
	ctx, span := tracing.Start(ctx, fmt.Sprintf("LowriBeck.%s", availabilityURL),
		trace.WithAttributes(attribute.String("postcode", req.PostCode)),
		trace.WithAttributes(attribute.String("lowribeck.reference", req.ReferenceID)),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(payload))))

	responseBody, err := c.doRequest(ctx, payload, availabilityURL)
	if err != nil {
		return nil, err
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(responseBody))))

	var ar GetCalendarAvailabilityResponse
	if err = json.Unmarshal(responseBody, &ar); err != nil {
		return nil, fmt.Errorf("unable to unmarshal get calendar availability body: %w", err)
	}

	return &ar, nil
}

func (c *Client) CreateBooking(ctx context.Context, req *CreateBookingRequest) (_ *CreateBookingResponse, err error) {
	ctx, span := tracing.Start(ctx, fmt.Sprintf("LowriBeck.%s", bookingURL),
		trace.WithAttributes(attribute.String("postcode", req.PostCode)),
		trace.WithAttributes(attribute.String("lowribeck.reference", req.ReferenceID)),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(payload))))

	responseBody, err := c.doRequest(ctx, payload, bookingURL)
	if err != nil {
		return nil, err
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(responseBody))))

	var br CreateBookingResponse
	if err = json.Unmarshal(responseBody, &br); err != nil {
		return nil, fmt.Errorf("unable to unmarshal create booking response body: %w", err)
	}

	return &br, nil
}

func (c *Client) GetCalendarAvailabilityPointOfSale(ctx context.Context, req *GetCalendarAvailabilityRequest) (_ *GetCalendarAvailabilityResponse, err error) {
	ctx, span := tracing.Start(ctx, fmt.Sprintf("LowriBeck.POS.%s", availabilityURL),
		trace.WithAttributes(attribute.String("postcode", req.PostCode)),
		trace.WithAttributes(attribute.String("mpan", req.Mpan)),
		trace.WithAttributes(attribute.String("mprn", req.Mprn)),
		trace.WithAttributes(attribute.String("elec.job", req.ElecJobTypeCode)),
		trace.WithAttributes(attribute.String("gas.job.", req.GasJobTypeCode)),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(payload))))

	responseBody, err := c.doRequest(ctx, payload, availabilityURL)
	if err != nil {
		return nil, err
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(responseBody))))

	var ar GetCalendarAvailabilityResponse
	if err = json.Unmarshal(responseBody, &ar); err != nil {
		return nil, fmt.Errorf("unable to unmarshal get calendar availability body: %w", err)
	}

	return &ar, nil
}

func (c *Client) CreateBookingPointOfSale(ctx context.Context, req *CreateBookingRequest) (_ *CreateBookingResponse, err error) {
	ctx, span := tracing.Start(ctx, fmt.Sprintf("LowriBeck.POS.%s", bookingURL),
		trace.WithAttributes(attribute.String("mpan", req.Mpan)),
		trace.WithAttributes(attribute.String("mprn", req.Mprn)),
		trace.WithAttributes(attribute.String("elec.job", req.ElecJobTypeCode)),
		trace.WithAttributes(attribute.String("gas.job.", req.GasJobTypeCode)),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(payload))))

	responseBody, err := c.doRequest(ctx, payload, bookingURL)
	if err != nil {
		return nil, err
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(responseBody))))

	var br CreateBookingResponse
	if err = json.Unmarshal(responseBody, &br); err != nil {
		return nil, fmt.Errorf("unable to unmarshal create booking response body: %w", err)
	}

	return &br, nil
}

func (c *Client) UpdateContactDetails(ctx context.Context, req *UpdateContactDetailsRequest) (_ *UpdateContactDetailsResponse, err error) {
	ctx, span := tracing.Start(ctx, fmt.Sprintf("LowriBeck.%s", updateContactURL),
		trace.WithAttributes(attribute.String("lowribeck.reference", req.ReferenceID)),
	)
	defer func() {
		tracing.RecordError(span, err)
		span.End()
	}()

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(payload))))

	responseBody, err := c.doRequest(ctx, payload, updateContactURL)
	if err != nil {
		return nil, err
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(responseBody))))

	var ucr UpdateContactDetailsResponse
	if err = json.Unmarshal(responseBody, &ucr); err != nil {
		return nil, fmt.Errorf("unable to unmarshal update contact details response body: %w", err)
	}

	return &ucr, nil
}

func (c *Client) doRequest(ctx context.Context, payload []byte, endpoint string) (_ []byte, err error) {

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}

	request.SetBasicAuth(c.auth.user, c.auth.password)
	request.Header.Add(contentHeader, contentJSON)

	resp, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send http request: %w", err)
	}
	defer resp.Body.Close()

	metrics.LBResponseCount.WithLabelValues(resp.Status, endpoint).Inc()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("received status code [%d] (expected 200): %s", resp.StatusCode, bodyBytes)
		slog.Error("status not ok", "error", statusErr)
		metrics.LBErrorsCount.WithLabelValues(metrics.LBStatus, endpoint).Inc()
		return nil, statusErr
	}

	return bodyBytes, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {

	requestURL := c.baseURL + healthCheckURL

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)
	if err != nil {
		return fmt.Errorf("unable to create new request: %w", err)
	}

	request.SetBasicAuth(c.auth.user, c.auth.password)
	request.Header.Add(contentHeader, contentJSON)

	resp, err := c.http.Do(request)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			slog.Error("healthcheck request timeout occurred")

			return ErrTimeout
		}
		return fmt.Errorf("unable to do healtcheck http request: %w", err)
	}
	defer resp.Body.Close()

	metrics.LBResponseCount.WithLabelValues(resp.Status, healthCheckURL).Inc()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		slog.Error("health check got an unauthorised (401) status code, check the username and password being used")

		return ErrNotOKStatusCode
	case http.StatusNotFound:
		slog.Error("health check failed got a not found(404) status code", "request_url", requestURL)

		return ErrNotOKStatusCode

	default:
		slog.Error("health check got status code", "status_code", resp.StatusCode)

		return ErrNotOKStatusCode
	}
}
