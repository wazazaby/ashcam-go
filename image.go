package ashcam

import (
	"encoding/json"
	"fmt"
)

const (
	imagesEndpoint string = "https://volcview.wr.usgs.gov/ashcam-api/imageApi/webcam"
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

type SunInformations struct {
	CurrentTime                   DateRFC1123Z `json:"time_in"`
	CivilTwilightSunrise          DateRFC1123Z `json:"civil_twilight_sunrise"`
	CivilTwilightSunset           DateRFC1123Z `json:"civil_twilight_sunset"`
	Timezone                      string       `json:"timezone"`
	CurrentTimeTimestamp          int          `json:"time_in_unixtime"`
	CivilTwilightSunriseTimestamp int          `json:"civil_twilight_sunrise_unixtime"`
	CivilTwilightSunsetTimestamp  int          `json:"civil_twilight_sunset_unixtime"`
}

//

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

type image Image

func (i *Image) UnmarshalJSON(b []byte) error {
	if string(b) == `[]` {
		*i = Image{}
		return nil
	}

	var img image
	if err := json.Unmarshal(b, &img); err != nil {
		return err
	}

	*i = Image(img)
	return nil
}

type Meta struct {
	APIURL              string `json:"apiUrl"`
	ImageTotal          int    `json:"imageTotal"`
	FirstImageTimestamp int    `json:"firstImageTimestamp"`
	LastImageTimestamp  int    `json:"lastImageTimestamp"`
	QuerySec            int    `json:"querySec"`
}

type ImageAPIResponse struct {
	Images []Image `json:"images"`
	Meta   Meta    `json:"meta"`
	Webcam Webcam  `json:"webcam"`
}

var (
	_ fmt.Stringer     = (*InterestingCode)(nil)
	_ json.Unmarshaler = (*InterestingCode)(nil)
)
