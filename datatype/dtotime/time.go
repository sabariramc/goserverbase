/*
Package dtotime defines custom wrapper for time.Time object with different formats for API request and response
DO NOT USE IT FOR DB MODEL
NOT COMPATIBLE WITH utils.JSONTransformer

Example1:

	type TestData struct {
		Dob       dtotime.Date     `json:"dob"`
		CreatedAt dtotime.DateTime `json:"createdAt"`
	}

	type DBData struct {
		ID        int64     `json:"id" bson:"id"`
		Dob       time.Time `json:"dob" bson:"dob"`
		CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	}

	input := `{"dob":"2014-06-12","createdAt":"2006-03-04T15:04:05.112+05:30"}`
	dataT1 := &TestData{}
	json.Unmarshal([]byte(input), data)

	dataT2 := DBData{}
	utils.JSONTransformer(data, &insertData)
	fmt.Println(dataT2) //{0 0001-01-01 00:00:00 +0000 UTC 0001-01-01 00:00:00 +0000 UTC} since the dto overrides MarshalJSON function the format is not compatible with time.UnmarshalJSON
*/
package dtotime

import (
	"time"
)

// Date wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02" format
type Date struct {
	time.Time
}

func (t *Date) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02"`, string(b))
	date = time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second(), date.Nanosecond(), time.Now().Location())
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t Date) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02"`)
	return []byte(val), nil
}

// DateTimeShort wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02T15:04:05Z07:00" format
type DateTimeShort struct {
	time.Time
}

func (t *DateTimeShort) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05Z07:00"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t DateTimeShort) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02T15:04:05Z07:00"`)
	return []byte(val), nil
}

// DateTime wraps time.Time with the JSON format and encodes and decodes in the "2006-01-02T15:04:05.000Z07:00" format
type DateTime struct {
	time.Time
}

func (t *DateTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02T15:04:05.000Z07:00"`, string(b))
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t DateTime) MarshalJSON() ([]byte, error) {
	val := t.Time.Format(`"2006-01-02T15:04:05.000Z07:00"`)
	return []byte(val), nil
}
