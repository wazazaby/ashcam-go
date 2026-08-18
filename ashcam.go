// Package ashcam is a client for the USGS ASHCAM API.
//
// See https://avo-volcview.wr.usgs.gov/ashcam-api/ for the API documentation.
package ashcam

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	// DefaultBaseURL keeps the full image archive for every webcam.
	DefaultBaseURL = "https://volcview.wr.usgs.gov/ashcam-api"

	// AVOBaseURL is the Alaska Volcano Observatory instance. It serves the same
	// webcam catalog from its own database, but only keeps a deep image archive
	// for the Alaska and Yellowstone webcams - as of August 2026 it holds 27
	// Kilauea images against 21k on the default host. Image IDs are not
	// comparable between the two instances, MD5 sums are.
	AVOBaseURL = "https://avo-volcview.wr.usgs.gov/ashcam-api"
)

var (
	ErrNotAuthorized = errors.New("not authorized, missing or invalid credentials")
	ErrNotFound      = errors.New("resource not found")
)

// APIError is returned when the API responds with a non 2xx status code.
// Unknown webcam codes and image identifiers are reported as a 500 by the API,
// not a 404.
type APIError struct {
	Method     string
	URL        string
	Body       string
	StatusCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s %s: request failed with status %d, body: %q", e.Method, e.URL, e.StatusCode, e.Body)
}

func (e *APIError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrNotAuthorized
	case http.StatusNotFound:
		return ErrNotFound
	}
	return nil
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type ClientOption func(*Client)

func WithHTTPClient(h HTTPClient) ClientOption {
	return func(c *Client) {
		c.httpClient = h
	}
}

// WithBaseURL picks the API instance to talk to - DefaultBaseURL, AVOBaseURL or
// a local mirror.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithCredentials sets the credentials sent to the endpoints requiring
// authentication - every write and admin endpoint.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username, c.password = username, password
	}
}

type Client struct {
	httpClient HTTPClient
	baseURL    string
	username   string
	password   string
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		httpClient: http.DefaultClient,
		baseURL:    DefaultBaseURL,
	}

	for _, applyOption := range options {
		applyOption(client)
	}

	return client
}

type requestBody struct {
	reader      io.Reader
	contentType string
}

// do sends a request to path, relative to the client base URL, and returns the
// response with its body still open unless an error is returned.
func (c *Client) do(ctx context.Context, method, path string, body *requestBody) (*http.Response, error) {
	url := c.baseURL + path

	var reader io.Reader
	if body != nil {
		reader = body.reader
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("unable to create %s %s request: %w", method, url, err)
	}

	if body != nil {
		req.Header.Set("Content-Type", body.contentType)
	}

	if c.username != "" || c.password != "" {
		req.Header.Set("username", c.username)
		req.Header.Set("password", c.password)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to do %s %s request: %w", method, url, err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		defer res.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		// Drain the rest of the body, else the connection can't be reused.
		io.Copy(io.Discard, res.Body)
		return nil, &APIError{
			Method:     method,
			URL:        url,
			Body:       string(bytes.TrimSpace(data)),
			StatusCode: res.StatusCode,
		}
	}

	return res, nil
}

func get[T any](ctx context.Context, c *Client, path string) (T, error) {
	var v T

	res, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return v, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("unable to decode %s response: %w", path, err)
	}

	// The decoder stops at the closing brace, so it never reaches EOF - the
	// connection can only be reused once the rest of the body is drained.
	io.Copy(io.Discard, res.Body)

	return v, nil
}

// raw returns the response body as-is. The API doesn't document the payload
// returned by its write and admin endpoints, so it's left to the caller to
// interpret.
func (c *Client) raw(ctx context.Context, method, path string, body *requestBody) ([]byte, error) {
	res, err := c.do(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s %s response: %w", method, path, err)
	}

	return data, nil
}

// rawJSON sends v as the JSON payload of a write endpoint.
func (c *Client) rawJSON(ctx context.Context, method, path string, v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal %s request body: %w", path, err)
	}
	body := &requestBody{reader: bytes.NewReader(data), contentType: "application/json"}
	return c.raw(ctx, method, path, body)
}
