package dtotime_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/datatype/dtotime"
	"gotest.tools/assert"
)

type TestData struct {
	Dob       dtotime.Date     `json:"dob"`
	CreatedAt dtotime.DateTime `json:"createdAt"`
}

type DBData struct {
	ID        int64     `json:"id" bson:"id"`
	Dob       time.Time `json:"dob" bson:"dob"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

func TestDate(t *testing.T) {
	input := `{"dob":"2014-06-12","createdAt":"2006-03-04T15:04:05.112+05:30"}`
	data := &TestData{}
	json.Unmarshal([]byte(input), data)
	assert.Equal(t, data.Dob.Time, time.Date(2014, 6, 12, 0, 0, 0, 0, time.Local))
	assert.Equal(t, data.CreatedAt.Time, time.Date(2006, 3, 4, 15, 4, 5, 112000000, time.Local))
	out, err := json.Marshal(data)
	assert.NilError(t, err)
	assert.Equal(t, input, string(out))
}
