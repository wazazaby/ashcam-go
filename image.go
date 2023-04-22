package ashcam

type InterestingCode string

const (
	VolcanicActivity   InterestingCode = "V"
	NoVolcanicActivity InterestingCode = "N"
	Unknown            InterestingCode = "U"
)

func (c InterestingCode) IsInteresting() bool {
	return c == VolcanicActivity
}

func (c InterestingCode) String() string {
	return string(c)
}

type SunInformations struct {
	Timezone string `json:"timezone"`

	CurrentTime          DateRFC1123Z `json:"time_in"`
	CurrentTimeTimestamp int          `json:"time_in_unixtime"`

	CivilTwilightSunrise DateRFC1123Z `json:"civil_twilight_sunrise"`
	CivilTwilightSunset  DateRFC1123Z `json:"civil_twilight_sunset"`

	CivilTwilightSunriseTimestamp int `json:"civil_twilight_sunrise_unixtime"`
	CivilTwilightSunsetTimestamp  int `json:"civil_twilight_sunset_unixtime"`
}

//

type Image struct {
	ID                int             `json:"imageId"`
	MD5               string          `json:"md5"`
	WebcamCode        string          `json:"webcamCode"`
	IsNewestForWebcam Bool            `json:"newestForWebcam"`
	Timestamp         int             `json:"imageTimestamp"`
	Date              DateRFC1123Z    `json:"imageDate"`
	InterestingCode   InterestingCode `json:"interestingCode"`
	IsNightTime       Bool            `json:"isNighttimeInd"`
	URL               string          `json:"imageUrl"`
	SunInformations   SunInformations `json:"suninfo"`
}

type Meta struct {
	ImageTotal          int    `json:"imageTotal"`
	FirstImageTimestamp int    `json:"firstImageTimestamp"`
	LastImageTimestamp  int    `json:"lastImageTimestamp"`
	APIURL              string `json:"apiUrl"`
	QuerySec            int    `json:"querySec"`
}

type ImageAPIResponse struct {
	Images []Image `json:"images"`
	Webcam Webcam  `json:"webcam"`
	Meta   Meta    `json:"meta"`
}
