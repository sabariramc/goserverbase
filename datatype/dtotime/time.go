/*
Package dtotime defines custom wrappers for the time.Time object with different formats for API request and response.
These wrappers are not compatible with utils.JSONTransformer.
*/
package dtotime

import (
	"time"
)

// Date wraps time.Time and provides JSON marshaling and unmarshaling in the "2006-01-02" format.
type Date struct {
	time.Time
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in the Date.
func (t *Date) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02"`, string(b))
	if err != nil {
		return err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Now().Location())
	t.Time = date
	return
}

// MarshalJSON returns the JSON encoding of the Date.
func (t Date) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02"`)
	return []byte(val), nil
}

// DateTimeShort wraps time.Time and provides JSON marshaling and unmarshaling in the "2006-01-02T15:04:05Z07:00" format.
type DateTimeShort struct {
	time.Time
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in the DateTimeShort.
func (t *DateTimeShort) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05Z07:00"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

// MarshalJSON returns the JSON encoding of the DateTimeShort.
func (t DateTimeShort) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02T15:04:05Z07:00"`)
	return []byte(val), nil
}

// DateTime wraps time.Time and provides JSON marshaling and unmarshaling in the "2006-01-02T15:04:05.000Z07:00" format.
type DateTime struct {
	time.Time
}

// UnmarshalJSON parses the JSON-encoded data and stores the result in the DateTime.
func (t *DateTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05.000Z07:00"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

// MarshalJSON returns the JSON encoding of the DateTime.
func (t DateTime) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02T15:04:05.000Z07:00"`)
	return []byte(val), nil
}
