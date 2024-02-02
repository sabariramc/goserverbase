package utils_test

import (
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v5/utils"
	"github.com/shopspring/decimal"
	"gotest.tools/assert"
)

type TestVal struct {
	IntVal       int             `json:"intVal"`
	DecimalVal   decimal.Decimal `json:"decimalVal"`
	StrVal       string          `json:"strVal"`
	BoolVal      bool            `json:"boolVal"`
	TimeValUTC   time.Time       `json:"timeValUTC"`
	TimeValLocal time.Time       `json:"timeValLocal"`
}

func TestJsonDecoding(t *testing.T) {
	val, _ := decimal.NewFromString("123.1232")
	data := map[string]interface{}{
		"intVal":     10,
		"decimalVal": val,
	}
	toData := &TestVal{}
	err := utils.StrictJsonTransformer(data, toData)
	assert.NilError(t, err)
	assert.Equal(t, 10, toData.IntVal)
	assert.DeepEqual(t, val, toData.DecimalVal)
	data["newField"] = "random value"
	err = utils.StrictJsonTransformer(data, toData)
	assert.Error(t, err, "StrictJsonTransformer: error decoding content: json: unknown field \"newField\"")
}
