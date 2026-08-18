package ashcam

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

// InterestingCode tells whether an image displays volcanic activity. Images
// flagged [VolcanicActivity] are never purged, see [Client.PurgeImages].
type InterestingCode uint8

const (
	UnknownVolcanicActivity InterestingCode = iota
	VolcanicActivity
	NoVolcanicActivity
)

// IsInteresting reports whether the code is [VolcanicActivity].
func (c InterestingCode) IsInteresting() bool {
	return c == VolcanicActivity
}

func (c InterestingCode) String() string {
	switch c {
	case VolcanicActivity:
		return "V"
	case NoVolcanicActivity:
		return "N"
	default:
		return "U"
	}
}

func (c *InterestingCode) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"V"`:
		*c = VolcanicActivity
	case `"N"`:
		*c = NoVolcanicActivity
	default:
		*c = UnknownVolcanicActivity
	}
	return nil
}

func (c InterestingCode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + c.String() + `"`), nil
}

// Image is one webcam image. MD5 identifies it across instances, ID does not,
// see [AVOBaseURL].
type Image struct {
	Date              DateRFC1123Z      `json:"imageDate"`
	MD5               string            `json:"md5"`
	WebcamCode        string            `json:"webcamCode"`
	URL               string            `json:"imageUrl"`
	SunInformations   SunInformations   `json:"suninfo"`
	ID                int               `json:"imageId"`
	Timestamp         int               `json:"imageTimestamp"`
	IsNewestForWebcam YesNoUnknownState `json:"newestForWebcam"`
	InterestingCode   InterestingCode   `json:"interestingCode"`
	IsNightTime       YesNoUnknownState `json:"isNighttimeInd"`
}

// NewestImage is an [Image] tolerating the empty array the API serializes for a
// webcam without images. Only [Webcam.NewestImage] needs it, so [Image] itself
// keeps the default - and much faster - decoding.
type NewestImage Image

func (i *NewestImage) UnmarshalJSON(b []byte) error {
	if string(b) == `[]` {
		*i = NewestImage{}
		return nil
	}
	return json.Unmarshal(b, (*Image)(i))
}

// ImagesMeta is the [Meta] of an [ImagesResponse]. The timestamps are those of
// the returned set, not of the webcam as a whole.
type ImagesMeta struct {
	Meta
	ImageTotal          int `json:"imageTotal"`
	FirstImageTimestamp int `json:"firstImageTimestamp"`
	LastImageTimestamp  int `json:"lastImageTimestamp"`
}

// ImagesResponse is what every image read endpoint returns. Webcam is only set
// by the ones scoped to a single webcam, so not by [Client.GetRecentImages],
// [Client.GetInterestingImages] or [Client.GetUninterestingImages].
type ImagesResponse struct {
	Images []Image    `json:"images"`
	Meta   ImagesMeta `json:"meta"`
	Webcam Webcam     `json:"webcam"`
}

type imageAPIRequestParameters struct {
	start, end  time.Time
	daysOld     int
	limit       int
	oldestFirst bool
}

// ImageRequestParameter narrows what [Client.GetImages] returns.
type ImageRequestParameter func(*imageAPIRequestParameters)

// OldestImageFirst returns the oldest images first instead of the newest.
//
// It can't be combined with a non-zero [Limit]: the API answers 500 to that
// combination - on both instances, for every webcam and both request forms, as
// of August 2026. Ask for the whole range and slice the result, or take the
// newest images instead.
func OldestImageFirst() ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.oldestFirst = true
	}
}

// Limit sets the number of images to return, 0 returns all of them within the
// requested time range. See [OldestImageFirst] for a pair the API rejects.
func Limit(n int) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.limit = n
	}
}

// DaysOld only returns the images of the last n days. It can't be combined with
// [TimeRange].
func DaysOld(n int) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.daysOld = n
	}
}

// TimeRange only returns the images taken between start and end. It can't be
// combined with [DaysOld].
func TimeRange(start, end time.Time) ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.start, p.end = start, end
	}
}

// The image request parameter errors, matched with [errors.Is].
var (
	ErrDaysOldAndTimeRangeCantBeUsedTogether = errors.New("days old and time range parameters can't be used together")
	ErrDaysOldRequired                       = errors.New("days old parameter is required")
)

// GetImages returns the images of a webcam. Without [DaysOld] or [TimeRange]
// every image is returned, and the API ignores [Limit] and [OldestImageFirst].
//
// It returns [ErrDaysOldAndTimeRangeCantBeUsedTogether] when both are given.
func (c *Client) GetImages(ctx context.Context, webcamCode string, parameters ...ImageRequestParameter) (ImagesResponse, error) {
	p := imageAPIRequestParameters{}
	for _, applyParameter := range parameters {
		applyParameter(&p)
	}

	byDaysOld := p.daysOld > 0
	byTimeRange := !p.start.IsZero() && !p.end.IsZero()
	if byDaysOld && byTimeRange {
		return ImagesResponse{}, ErrDaysOldAndTimeRangeCantBeUsedTogether
	}

	order := "newestFirst"
	if p.oldestFirst {
		order = "oldestFirst"
	}

	path := "/imageApi/webcam/" + webcamCode
	switch {
	case byDaysOld:
		path += fmt.Sprintf("/%d/%s/%d", p.daysOld, order, p.limit)
	case byTimeRange:
		path += fmt.Sprintf("/%d/%d/%s/%d", p.start.Unix(), p.end.Unix(), order, p.limit)
	}

	return get[ImagesResponse](ctx, c, path)
}

// GetRecentImages returns the limit most recent images, regardless of webcam.
func (c *Client) GetRecentImages(ctx context.Context, limit int) (ImagesResponse, error) {
	return get[ImagesResponse](ctx, c, "/imageApi/recent/"+strconv.Itoa(limit))
}

// GetInterestingImages returns the images flagged [VolcanicActivity]. A daysOld
// greater than 0 limits the results to the most recent images.
func (c *Client) GetInterestingImages(ctx context.Context, daysOld int) (ImagesResponse, error) {
	path := "/imageApi/interesting"
	if daysOld > 0 {
		path += "/" + strconv.Itoa(daysOld)
	}
	return get[ImagesResponse](ctx, c, path)
}

// GetUninterestingImages returns the images flagged [NoVolcanicActivity]. As
// they're expected to be far more numerous than the interesting ones, daysOld is
// required: the API has no route without it, and a daysOld below 1 returns
// [ErrDaysOldRequired].
func (c *Client) GetUninterestingImages(ctx context.Context, daysOld int) (ImagesResponse, error) {
	if daysOld <= 0 {
		return ImagesResponse{}, ErrDaysOldRequired
	}
	return get[ImagesResponse](ctx, c, "/imageApi/uninteresting/"+strconv.Itoa(daysOld))
}

// SetInterestingCode flags whether an image displays volcanic activity. Images
// flagged [VolcanicActivity] are never purged, see [Client.PurgeImages].
//
// The identifier is either [Image.ID] or [Image.MD5], the latter being the one
// that survives switching instance - see [AVOBaseURL] - and the one to use when
// the image is already loaded but its ID is unknown.
//
// Requires [WithCredentials]. The API doesn't document a response payload, so
// the body is returned as is.
func (c *Client) SetInterestingCode(ctx context.Context, imageIdentifier string, code InterestingCode) ([]byte, error) {
	path := fmt.Sprintf("/imageApi/interestingCode/%s/%s", imageIdentifier, code)
	return c.raw(ctx, http.MethodPut, path, nil)
}

// ImageUpload is the image [Client.UploadImage] adds to a webcam.
type ImageUpload struct {
	Image      io.Reader
	WebcamCode string
	// FileName is optional and defaults to image.jpg.
	FileName        string
	Timestamp       time.Time
	InterestingCode InterestingCode
	// NotNewest keeps the newest webcam image unchanged and defers the webcam
	// statistics to the slower periodic process. Intended for backfilling old
	// images - uploading them too quickly may deadlock the API database.
	NotNewest YesNo
}

func (u ImageUpload) body() (*requestBody, error) {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)
	w.WriteField("webcamCode", u.WebcamCode)
	w.WriteField("imageTimestamp", strconv.FormatInt(u.Timestamp.Unix(), 10))
	w.WriteField("interestingCode", u.InterestingCode.String())
	w.WriteField("notNewest", u.NotNewest.String())

	part, err := w.CreateFormFile("file", cmp.Or(u.FileName, "image.jpg"))
	if err != nil {
		return nil, fmt.Errorf("unable to create image upload part: %w", err)
	}

	if _, err := io.Copy(part, u.Image); err != nil {
		return nil, fmt.Errorf("unable to write image upload part: %w", err)
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("unable to close image upload writer: %w", err)
	}

	return &requestBody{reader: &buf, contentType: w.FormDataContentType()}, nil
}

// UploadImage adds a new image to a webcam.
//
// Requires [WithCredentials]. The API doesn't document a response payload, so
// the body is returned as is.
func (c *Client) UploadImage(ctx context.Context, upload ImageUpload) ([]byte, error) {
	body, err := upload.body()
	if err != nil {
		return nil, err
	}
	return c.raw(ctx, http.MethodPost, "/imageApi/uploadImage", body)
}
