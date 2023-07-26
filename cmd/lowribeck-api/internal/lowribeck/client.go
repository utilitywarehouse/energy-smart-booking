package lowribeck

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	contentHeader   = "Content-Type"
	contentJSON     = "application/json"
	availabilityURL = "getCalendarAvailability"
	bookingURL      = "book"
)

type client struct {
	http    *http.Client
	auth    auth
	baseURL string
}

func New(c *http.Client, user, password, url string) *client {
	return &client{
		http: c,
		auth: auth{
			user:     user,
			password: password,
		},
		baseURL: url,
	}
}

func (c *client) GetCalendarAvailability(ctx context.Context, req *GetCalendarAvailabilityRequest) (*GetCalendarAvailabilityResponse, error) {
	resp, err := c.DoRequest(ctx, req, availabilityURL)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	var ar GetCalendarAvailabilityResponse
	if err = json.Unmarshal(respBody, &ar); err != nil {
		return nil, fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return &ar, nil
}

func (c *client) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*CreateBookingResponse, error) {
	resp, err := c.DoRequest(ctx, req, bookingURL)
	if err != nil {
		return nil, fmt.Errorf("unable to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read body: %w", err)
	}

	var br CreateBookingResponse
	if err = json.Unmarshal(respBody, &br); err != nil {
		return nil, fmt.Errorf("unable to unmarshal body: %w", err)
	}

	return &br, nil
}

func (c *client) DoRequest(ctx context.Context, req interface{}, url string) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	logrus.Debugf("request: [%s]", string(body))

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request")
	}

	request.SetBasicAuth(c.auth.user, c.auth.password)
	request.Header.Add(contentHeader, contentJSON)

	resp, err := c.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to send http request")
	}

	return resp, nil
}
