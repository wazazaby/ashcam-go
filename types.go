// This file contains the types shared by the API payloads.
package ashcam

import (
	"encoding/json"
	"time"
)

type DateRFC1123Z time.Time

func (d DateRFC1123Z) Time() time.Time {
	return time.Time(d)
}

func (d *DateRFC1123Z) UnmarshalJSON(b []byte) error {
	var date string
	if err := json.Unmarshal(b, &date); err != nil {
		return err
	}

	parsed, err := time.Parse(time.RFC1123Z, date)
	if err != nil {
		return err
	}

	*d = DateRFC1123Z(parsed)
	return nil
}

type YesNoUnknownState uint8

const (
	StateUnknown YesNoUnknownState = iota
	StateYes
	StateNo
)

const (
	stateUnknownLabel string = "?"
	stateYesLabel     string = "Y"
	stateNoLabel      string = "N"
)

func (i *YesNoUnknownState) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"` + stateYesLabel + `"`:
		*i = StateYes
	case `"` + stateNoLabel + `"`:
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
		return stateYesLabel
	case StateNo:
		return stateNoLabel
	default:
		return stateUnknownLabel
	}
}

// YesNo is a boolean sent to the API as the "Y" or "N" indicator it expects.
type YesNo bool

func (y YesNo) String() string {
	if y {
		return stateYesLabel
	}
	return stateNoLabel
}

func (y YesNo) MarshalJSON() ([]byte, error) {
	return []byte(`"` + y.String() + `"`), nil
}

var (
	_ json.Unmarshaler = (*YesNoUnknownState)(nil)
	_ json.Unmarshaler = (*DateRFC1123Z)(nil)
	_ json.Marshaler   = YesNoUnknownState(0)
	_ json.Marshaler   = YesNo(false)
)
