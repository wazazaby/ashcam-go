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

type InterestingCode uint8

const (
	UnknownVolcanicActivity InterestingCode = iota
	VolcanicActivity
	NoVolcanicActivity
)

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

// NewestImage is an Image tolerating the empty array the API serializes for a
// webcam without images. Only Webcam needs it, so Image itself keeps the
// default - and much faster - decoding.
type NewestImage Image

func (i *NewestImage) UnmarshalJSON(b []byte) error {
	if string(b) == `[]` {
		*i = NewestImage{}
		return nil
	}
	return json.Unmarshal(b, (*Image)(i))
}

type ImagesMeta struct {
	Meta
	ImageTotal          int `json:"imageTotal"`
	FirstImageTimestamp int `json:"firstImageTimestamp"`
	LastImageTimestamp  int `json:"lastImageTimestamp"`
}

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

type ImageRequestParameter func(*imageAPIRequestParameters)

func OldestImageFirst() ImageRequestParameter {
	return func(p *imageAPIRequestParameters) {
		p.oldestFirst = true
	}
}

// Limit sets the number of images to return, 0 returns all of them within the
// requested time range.
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

var (
	ErrDaysOldAndTimeRangeCantBeUsedTogether = errors.New("days old and time range parameters can't be used together")
	ErrDaysOldRequired                       = errors.New("days old parameter is required")
)

// GetImages returns the images of a webcam. Without DaysOld or TimeRange every
// image is returned, and the Limit and OldestImageFirst parameters are ignored
// by the API.
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

// GetInterestingImages returns the images displaying volcanic activity. A
// daysOld greater than 0 limits the results to the most recent images.
func (c *Client) GetInterestingImages(ctx context.Context, daysOld int) (ImagesResponse, error) {
	path := "/imageApi/interesting"
	if daysOld > 0 {
		path += "/" + strconv.Itoa(daysOld)
	}
	return get[ImagesResponse](ctx, c, path)
}

// GetUninterestingImages returns the images that don't display volcanic
// activity. As they're expected to be far more numerous than the interesting
// ones, daysOld is required.
func (c *Client) GetUninterestingImages(ctx context.Context, daysOld int) (ImagesResponse, error) {
	if daysOld <= 0 {
		return ImagesResponse{}, ErrDaysOldRequired
	}
	return get[ImagesResponse](ctx, c, "/imageApi/uninteresting/"+strconv.Itoa(daysOld))
}

// SetInterestingCode flags whether an image displays volcanic activity.
// Interesting images are never purged. The identifier is either the image ID or
// its MD5 sum - the MD5 is the one that identifies the same image across
// instances, see AVOBaseURL.
//
// Requires credentials.
func (c *Client) SetInterestingCode(ctx context.Context, imageIdentifier string, code InterestingCode) ([]byte, error) {
	path := fmt.Sprintf("/imageApi/interestingCode/%s/%s", imageIdentifier, code)
	return c.raw(ctx, http.MethodPut, path, nil)
}

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
// Requires credentials.
func (c *Client) UploadImage(ctx context.Context, upload ImageUpload) ([]byte, error) {
	body, err := upload.body()
	if err != nil {
		return nil, err
	}
	return c.raw(ctx, http.MethodPost, "/imageApi/uploadImage", body)
}
