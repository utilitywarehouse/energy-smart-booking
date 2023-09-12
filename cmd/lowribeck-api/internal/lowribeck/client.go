package lowribeck

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/utilitywarehouse/uwos-go/v1/telemetry/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrNotOKStatusCode = errors.New("error status code is not 200(OK)")
	ErrTimeout         = errors.New("timeout occurred")
)

const (
	contentHeader   = "Content-Type"
	contentJSON     = "application/json"
	availabilityURL = "appointmentManagement/getCalendarAvailability"
	bookingURL      = "appointmentManagement/book"
	healthCheckURL  = "health/get"
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

type LBRequest interface {
	GetPostCode() string
	GetReference() string
}

func (c *Client) GetCalendarAvailability(ctx context.Context, req *GetCalendarAvailabilityRequest) (*GetCalendarAvailabilityResponse, error) {
	resp, err := c.DoRequest(ctx, req, availabilityURL)
	if err != nil {
		return nil, err
	}

	var ar GetCalendarAvailabilityResponse
	if err = json.Unmarshal(resp, &ar); err != nil {
		return nil, fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return &ar, nil
}

func (c *Client) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*CreateBookingResponse, error) {
	resp, err := c.DoRequest(ctx, req, bookingURL)
	if err != nil {
		return nil, err
	}

	var br CreateBookingResponse
	if err = json.Unmarshal(resp, &br); err != nil {
		return nil, fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return &br, nil
}

func (c *Client) DoRequest(ctx context.Context, req LBRequest, endpoint string) (_ []byte, err error) {
	ctx, span := tracing.Tracer().Start(ctx, fmt.Sprintf("LowriBeck.%s", endpoint),
		trace.WithAttributes(attribute.String("postcode", req.GetPostCode())),
		trace.WithAttributes(attribute.String("lowribeck.reference", req.GetReference())),
	)

	defer func() {
		tracing.RecordSpanError(span, err) // nolint: errcheck
		span.End()
	}()

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	span.AddEvent("request", trace.WithAttributes(attribute.String("req", string(body))))

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+endpoint,
		bytes.NewReader(body),
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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	span.AddEvent("response", trace.WithAttributes(attribute.String("resp", string(bodyBytes))))

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("received status code [%d] (expected 200): %s", resp.StatusCode, bodyBytes)
		logrus.Error(statusErr)
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
		if os.IsTimeout(err) {
			logrus.Error("healthcheck request timeout occurred")

			return ErrTimeout
		}
		return fmt.Errorf("unable to do healtcheck http request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		logrus.Error("health check got an unauthorized (401) status code, check the username and password being used") //nolint: misspell

		return ErrNotOKStatusCode
	case http.StatusNotFound:
		logrus.Errorf("health check failed got a not found(404) status code, the request URL is: %s", requestURL)

		return ErrNotOKStatusCode

	default:
		logrus.Errorf("health check got status code: %d", resp.StatusCode)

		return ErrNotOKStatusCode
	}
}
