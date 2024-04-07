package ashcam

import "errors"

const (
	webcamEndpoint  string = "https://volcview.wr.usgs.gov/ashcam-api/webcamApi/webcam/"
	webcamsEndpoint string = "https://volcview.wr.usgs.gov/ashcam-api/webcamApi/webcams"
)

var (
	ErrWebcamResourceNotFound = errors.New("webcam resource not found")
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
	NewestImage           Image             `json:"newestImage"`
	VNum                  int               `json:"vnum,string"`
	BearingDegrees        int               `json:"bearingDeg"`
	LastImageTimestamp    int               `json:"lastImageTimestamp"`
	FistImageTimestamp    int               `json:"firstImageTimestamp"`
	ImageTotal            int               `json:"imageTotal"`
	Elevation             float64           `json:"elevationM"`
	Longitude             float64           `json:"longitude"`
	Latitude              float64           `json:"latitude"`
	HasImages             YesNoUnknownState `json:"hasImages"`
}

type WebcamMeta struct {
	APIURL   string `json:"apiUrl"`
	QuerySec int    `json:"querySec"`
}

type WebcamsMeta struct {
	WebcamMeta
	Total int `json:"webcamTotal"`
}

type WebcamResponse struct {
	WebcamMeta WebcamMeta `json:"meta"`
	Webcam     Webcam     `json:"webcam"`
}

type WebcamsResponse struct {
	Webcams     []Webcam    `json:"webcams"`
	WebcamsMeta WebcamsMeta `json:"meta"`
}
