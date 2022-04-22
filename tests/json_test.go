package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/utils"
	"github.com/shopspring/decimal"
)

type TestVal struct {
	IntVal       int64           `json:"intVal"`
	DeciVal      decimal.Decimal `json:"deciVal"`
	StrVal       string          `json:"strVal"`
	BoolVal      bool            `json:"boolVal"`
	TimeValUTC   time.Time       `json:"timeValUTC"`
	TimeValLocal time.Time       `json:"timeValLocal"`
}

func TestJsonDecoding(t *testing.T) {
	val, _ := decimal.NewFromString("123.1232")
	data := map[string]interface{}{
		"intVal":  10,
		"deciVal": val,
	}
	toData := &TestVal{}
	err := utils.JsonTransformer(data, toData)
	if err != nil {
		t.Fatal(err)
	}
	data["newField"] = "fadfa"
	err = utils.JsonTransformer(data, toData)
	if err == nil {
		t.Fatal(fmt.Errorf("Json should throw an error"))
	}
	fmt.Println(err.Error())
}
