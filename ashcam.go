package ashcam

import (
	"net/http"
)

type ClientOption func(*Client)

func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = h
	}
}

type Client struct {
	httpClient *http.Client
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
