package ashcam

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type Webcam struct {
	FirstImageDate        DateRFC1123Z      `json:"firstImageDate"`
	LastImageDate         DateRFC1123Z      `json:"lastImageDate"`
	CurrentImageURL       string            `json:"currentImageUrl"`
	Name                  string            `json:"webcamName"`
	Code                  string            `json:"webcamCode"`
	ClearImageURL         string            `json:"clearImageUrl"`
	Timezone              string            `json:"timezone"` // Timezone can be useful when sun informations are not set.
	VName                 string            `json:"vName"`
	CurrentThumbImageURL  string            `json:"currentThumbImageUrl"`
	CurrentMediumImageURL string            `json:"currentMediumImageUrl"`
	ExternalURL           string            `json:"externalUrl"`
	SunInformations       SunInformations   `json:"suninfo"`
	NewestImage           NewestImage       `json:"newestImage"`
	VNum                  int               `json:"vnum,string"`
	BearingDegrees        int               `json:"bearingDeg"`
	LastImageTimestamp    int               `json:"lastImageTimestamp"`
	FirstImageTimestamp   int               `json:"firstImageTimestamp"`
	ImageTotal            int               `json:"imageTotal"`
	Elevation             float64           `json:"elevationM"`
	Longitude             float64           `json:"longitude"`
	Latitude              float64           `json:"latitude"`
	HasImages             YesNoUnknownState `json:"hasImages"`
	IsFAA                 YesNoUnknownState `json:"faaInd"`
}

type WebcamsMeta struct {
	Meta
	Total int `json:"webcamTotal"`
}

type WebcamResponse struct {
	Meta   Meta   `json:"meta"`
	Webcam Webcam `json:"webcam"`
}

type WebcamsResponse struct {
	Webcams []Webcam    `json:"webcams"`
	Meta    WebcamsMeta `json:"meta"`
}

func (c *Client) GetWebcam(ctx context.Context, code string) (WebcamResponse, error) {
	return get[WebcamResponse](ctx, c, "/webcamApi/webcam/"+code)
}

func (c *Client) GetWebcams(ctx context.Context) (WebcamsResponse, error) {
	return get[WebcamsResponse](ctx, c, "/webcamApi/webcams")
}

// GeoArea is the latitude and longitude bounding box of a GeoJSON query.
type GeoArea struct {
	Lat1, Lat2   float64
	Long1, Long2 float64
}

func (a GeoArea) query() string {
	format := func(f float64) string {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return "lat1=" + format(a.Lat1) + "&lat2=" + format(a.Lat2) +
		"&long1=" + format(a.Long1) + "&long2=" + format(a.Long2)
}

type GeoJSONMetadata struct {
	Title     string `json:"title"`
	Params    string `json:"params"`
	Generated int    `json:"generated"`
	Count     int    `json:"count"`
}

type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type GeoJSONFeature struct {
	Type       string          `json:"type"`
	Properties Webcam          `json:"properties"`
	Geometry   GeoJSONGeometry `json:"geometry"`
	ID         int             `json:"id"`
}

type GeoJSONResponse struct {
	Type     string           `json:"type"`
	Metadata GeoJSONMetadata  `json:"metadata"`
	Features []GeoJSONFeature `json:"features"`
}

func (c *Client) GetWebcamsGeoJSON(ctx context.Context, area GeoArea) (GeoJSONResponse, error) {
	return get[GeoJSONResponse](ctx, c, "/webcamApi/geojson?"+area.query())
}

// WebcamInput is the payload used to create or update a webcam. Only Code and
// Name are required, but Latitude and Longitude are needed for the API to tell
// daytime images from nighttime ones. VNum and VName are the values assigned by
// the Smithsonian Institution to the volcano the webcam looks at.
type WebcamInput struct {
	Code        string  `json:"webcamCode"`
	Name        string  `json:"webcamName"`
	ExternalURL string  `json:"externalUrl"`
	VNum        string  `json:"vnum"`
	VName       string  `json:"vName"`
	Latitude    float64 `json:"latitude,string"`
	Longitude   float64 `json:"longitude,string"`
	Elevation   float64 `json:"elevationM,string"`
	BearingDeg  int     `json:"bearingDeg,string"`
	IsFAA       YesNo   `json:"faaInd"`
}

// CreateOrUpdateWebcam creates the webcam if its code is unknown, updates it
// otherwise. The code is part of the image paths and can't be changed once the
// webcam is created.
//
// Requires credentials.
func (c *Client) CreateOrUpdateWebcam(ctx context.Context, webcam WebcamInput) ([]byte, error) {
	return c.rawJSON(ctx, http.MethodPost, "/webcamApi/webcam", webcam)
}

// RefreshAllWebcams refreshes the hasImages and newestImage values of every webcam.
//
// Requires credentials.
func (c *Client) RefreshAllWebcams(ctx context.Context) ([]byte, error) {
	return c.raw(ctx, http.MethodPost, "/webcamApi/refreshAll", nil)
}

// AssignClearImage copies one of the webcam images to the clear image
// directory, enabling the clear image slider in the webcam viewer. The
// identifier is either the image ID or its MD5 sum.
//
// Requires credentials.
func (c *Client) AssignClearImage(ctx context.Context, webcamCode, imageIdentifier string) ([]byte, error) {
	path := fmt.Sprintf("/webcamApi/assignClearImage/%s/%s", webcamCode, imageIdentifier)
	return c.raw(ctx, http.MethodGet, path, nil)
}

// AssignVolcano sets the volcano number and name of a webcam.
//
// Requires credentials.
func (c *Client) AssignVolcano(ctx context.Context, webcamCode, vNum, vName string) ([]byte, error) {
	return c.rawJSON(ctx, http.MethodPost, "/webcamApi/assignVolcano", struct {
		Code  string `json:"webcamCode"`
		VNum  string `json:"vnum"`
		VName string `json:"vName"`
	}{webcamCode, vNum, vName})
}
