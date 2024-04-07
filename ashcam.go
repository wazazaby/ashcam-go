package ashcam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

type ClientOption func(*Client)

func WithHTTPClient(h HTTPClient) ClientOption {
	return func(c *Client) {
		c.httpClient = h
	}
}

type Client struct {
	httpClient HTTPClient
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		httpClient: http.DefaultClient,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

type ImageRequestTimeRange struct {
	Start, End time.Time
}

type ImageAPIRequestParameters struct {
	TimeRange   ImageRequestTimeRange
	WebcamCode  string
	DaysOld     int
	Limit       int
	NewestFirst bool
}

func (c Client) GetImages(p ImageAPIRequestParameters) (ImageAPIResponse, error) {
	return ImageAPIResponse{}, nil
}

func (c Client) GetWebcam(ctx context.Context, code string) (WebcamResponse, error) {
	var r WebcamResponse
	select {
	case <-ctx.Done():
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, ctx.Err())
	default:
	}

	url := concat(webcamEndpoint, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, err)
	}

	res, err := c.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, err)
	}

	if res.StatusCode != http.StatusOK {
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, ErrWebcamResourceNotFound)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, err)
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("unable to get webcam %q, err: %w", code, err)
	}

	return r, nil
}

func (c Client) GetWebcams(ctx context.Context) (WebcamsResponse, error) {
	var r WebcamsResponse
	select {
	case <-ctx.Done():
		return r, fmt.Errorf("unable to get all webcams, err: %w", ctx.Err())
	default:
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, webcamsEndpoint, http.NoBody)
	if err != nil {
		return r, fmt.Errorf("unable to get all webcams, err: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return r, fmt.Errorf("unable to get all webcams, err: %w", err)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return r, fmt.Errorf("unable to get all webcams, err: %w", err)
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("unable to get all webcams, err: %w", err)
	}

	return r, nil
}
