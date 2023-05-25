package utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v3/utils"
	"github.com/shopspring/decimal"
)

type TestVal struct {
	IntVal       int64           `json:"intVal"`
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
	if err != nil {
		t.Fatal(err)
	}
	data["newField"] = "random value"
	err = utils.StrictJsonTransformer(data, toData)
	if err == nil {
		t.Fatal(fmt.Errorf("Json should throw an error"))
	}
	fmt.Println(err.Error())
}
