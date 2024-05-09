// Package time defines custom wrapper for time.Time object with different formats for JSON encode/decode
package time

import (
	"time"
)

const defaultFormat = "2006-01-02T15:04:05.000Z07:00"

type CustomTime struct {
	time.Time
	format string
}

func Now(format string) *CustomTime {
	if format == "" {
		format = defaultFormat
	}
	return &CustomTime{
		format: format,
		Time:   time.Now(),
	}
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(t.format, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t CustomTime) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(t.format)
	return []byte(val), nil
}

// Date wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02" format
type Date struct {
	*CustomTime
}

func (t *Date) Now() *Date {
	t.CustomTime = Now("2006-01-02")
	return t
}

// DateTimeShort wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02T15:04:05Z07:00" format
type DateTimeShort struct {
	*CustomTime
}

func (t *DateTimeShort) Now() *DateTimeShort {
	t.CustomTime = Now("2006-01-02T15:04:05Z07:00")
	return t
}

// DateTime wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02T15:04:05.000Z07:00" format
type DateTime struct {
	*CustomTime
}

func (t *DateTime) Now() *DateTime {
	t.CustomTime = Now("2006-01-02T15:04:05.000Z07:00")
	return t
}
