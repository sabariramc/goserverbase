package dtotime_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v6/datatype/dtotime"
	"github.com/sabariramc/goserverbase/v6/utils"
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

func Example() {
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
	json.Unmarshal([]byte(input), dataT1)
	fmt.Println(dataT1)
	dataT2 := DBData{}
	utils.JSONTransformer(dataT1, &dataT2)
	fmt.Println(dataT2)
	//Output:
	//&{2014-06-12 00:00:00 +0530 IST 2006-03-04 15:04:05.112 +0530 IST}
	//{0 0001-01-01 00:00:00 +0000 UTC 0001-01-01 00:00:00 +0000 UTC}
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
