package ashcam

import (
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
