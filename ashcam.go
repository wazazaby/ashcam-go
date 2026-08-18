package ashcam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrDaysOldAndTimeRangeCantBeUsedTogether = errors.New("days old and time range parameters can't be used together")
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
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

	for _, applyOption := range options {
		applyOption(client)
	}

	return client
}

type imageAPIRequestParameters struct {
	start, end  time.Time
	daysOld     int
	limit       int
	oldestFirst bool
}

type ImageRequestParameter func(*imageAPIRequestParameters)

func OldestImageFirst() ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.oldestFirst = true
	}
}

func Limit(n int) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.limit = n
	}
}

func DaysOld(n int) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.daysOld = n
	}
}

func TimeRange(start, end time.Time) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.start, p.end = start, end
	}
}

func (c *Client) GetImages(ctx context.Context, webcamCode string, parameters ...ImageRequestParameter) (ImageAPIResponse, error) {
	var r ImageAPIResponse

	p := imageAPIRequestParameters{}
	for _, applyParameter := range parameters {
		applyParameter(&p)
	}

	byDaysOld := p.daysOld > 0
	byTimeRange := !p.start.IsZero() && !p.end.IsZero()

	order := "newestFirst"
	if p.oldestFirst {
		order = "oldestFirst"
	}

	var url string
	switch {
	case byDaysOld && byTimeRange:
		return r, ErrDaysOldAndTimeRangeCantBeUsedTogether
	case byDaysOld:
		url = fmt.Sprintf("%s/%s/%d/%s/%d", imagesEndpoint, webcamCode, p.daysOld, order, p.limit)
	case byTimeRange:
		url = fmt.Sprintf("%s/%s/%d/%d/%s/%d", imagesEndpoint, webcamCode, p.start.Unix(), p.end.Unix(), order, p.limit)
	default:
		url = fmt.Sprintf("%s/%s", imagesEndpoint, webcamCode)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return r, fmt.Errorf("unable to create request for webcam %q, err: %w", webcamCode, err)
	}

	res, err := c.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return r, fmt.Errorf("unable to do request for webcam %q, err: %w", webcamCode, err)
	}

	if res.StatusCode != http.StatusOK {
		return r, fmt.Errorf("request failed with status %d for webcam %q, err: %w", res.StatusCode, webcamCode, ErrWebcamResourceNotFound)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return r, fmt.Errorf("unable to read response body for webcam %q, err: %w", webcamCode, err)
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf("unable unmarshal response data for webcam %q, err: %w", webcamCode, err)
	}

	return r, nil
}

func (c *Client) GetWebcam(ctx context.Context, code string) (WebcamResponse, error) {
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

func (c *Client) GetWebcams(ctx context.Context) (WebcamsResponse, error) {
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
