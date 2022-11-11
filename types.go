// This file contains type aliases
package ashcam

import (
	"encoding/json"
	"fmt"
	"time"
)

type DateRFC1123Z time.Time

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

type Bool bool

func (i Bool) Bool() bool {
	return bool(i)
}

func (i *Bool) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"Y"`:
		*i = Bool(true)
	case `"N"`:
		*i = Bool(false)
	default:
		return fmt.Errorf("unsupported value %s for boolean indicator", string(b))
	}
	return nil
}

var (
	_ json.Unmarshaler = (*Bool)(nil)
	_ json.Unmarshaler = (*DateRFC1123Z)(nil)
)
