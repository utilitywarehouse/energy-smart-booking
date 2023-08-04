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

func (c *Client) DoRequest(ctx context.Context, req interface{}, endpoint string) ([]byte, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request: %w", err)
	}

	logrus.Debugf("request: [%s]", string(body))

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

	if resp.StatusCode != http.StatusOK {
		statusErr := fmt.Errorf("received status code [%d] (expected 200): %s", resp.StatusCode, bodyBytes)
		logrus.Error(statusErr)
		return nil, statusErr
	}

	logrus.Debugf("response: [%s]", string(bodyBytes))

	return bodyBytes, nil
}
