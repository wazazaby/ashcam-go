package ashcam

import (
	"fmt"
	"time"
)

// Meta is the envelope every API response carries. [ImagesMeta] and
// [WebcamsMeta] extend it.
type Meta struct {
	APIURL   string `json:"apiUrl"`
	QuerySec int    `json:"querySec"`
}

// DateRFC1123Z is a [time.Time] decoded from the RFC 1123Z layout the API
// serializes its dates in. The embedded time is directly usable, so
// image.Date.Unix() and image.Date.Before(t) both work.
type DateRFC1123Z struct {
	time.Time
}

func (d *DateRFC1123Z) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' {
		return fmt.Errorf("unable to parse %s as an RFC1123Z date", b)
	}

	parsed, err := time.Parse(time.RFC1123Z, string(b[1:len(b)-1]))
	if err != nil {
		return err
	}

	d.Time = parsed
	return nil
}

func (d DateRFC1123Z) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.Format(time.RFC1123Z) + `"`), nil
}

// SunInformations is the daylight state at a webcam's location, which the API
// leaves unset when it doesn't know the coordinates - [Webcam.Timezone] is the
// fallback. Each date is repeated as a Unix timestamp, matching the payload.
type SunInformations struct {
	CurrentTime                   DateRFC1123Z `json:"time_in"`
	CivilTwilightSunrise          DateRFC1123Z `json:"civil_twilight_sunrise"`
	CivilTwilightSunset           DateRFC1123Z `json:"civil_twilight_sunset"`
	Timezone                      string       `json:"timezone"`
	CurrentTimeTimestamp          int          `json:"time_in_unixtime"`
	CivilTwilightSunriseTimestamp int          `json:"civil_twilight_sunrise_unixtime"`
	CivilTwilightSunsetTimestamp  int          `json:"civil_twilight_sunset_unixtime"`
}

// YesNoUnknownState is a tri-state indicator the API returns as "Y", "N" or "?".
// [YesNo] is its request-side counterpart.
type YesNoUnknownState uint8

const (
	StateUnknown YesNoUnknownState = iota
	StateYes
	StateNo
)

func (i *YesNoUnknownState) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"Y"`:
		*i = StateYes
	case `"N"`:
		*i = StateNo
	default:
		*i = StateUnknown
	}
	return nil
}

func (i YesNoUnknownState) MarshalJSON() ([]byte, error) {
	return []byte(`"` + i.String() + `"`), nil
}

func (i YesNoUnknownState) String() string {
	switch i {
	case StateYes:
		return "Y"
	case StateNo:
		return "N"
	default:
		return "?"
	}
}

// YesNo is a boolean sent to the API as the "Y" or "N" indicator it expects. It
// is the request-side counterpart of [YesNoUnknownState], which also carries the
// unknown state the API can answer with.
type YesNo bool

func (y YesNo) String() string {
	if y {
		return "Y"
	}
	return "N"
}

func (y YesNo) MarshalJSON() ([]byte, error) {
	return []byte(`"` + y.String() + `"`), nil
}
