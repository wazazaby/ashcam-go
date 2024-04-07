// This file contains type aliases
package ashcam

import (
	"encoding/json"
	"strconv"
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
	s, _ := strconv.Unquote(string(b))
	switch s {
	case stateYesLabel:
		*i = StateYes
	case stateNoLabel:
		*i = StateNo
	default:
		*i = StateUnknown
	}
	return nil
}

func (i YesNoUnknownState) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 3)
	buf = append(buf, '"')
	buf = append(buf, i.String()...)
	buf = append(buf, '"')
	return buf, nil
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

var (
	_ json.Unmarshaler = (*YesNoUnknownState)(nil)
	_ json.Unmarshaler = (*DateRFC1123Z)(nil)
)
