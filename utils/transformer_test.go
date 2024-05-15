package utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sabariramc/goserverbase/v6/utils"
	"github.com/shopspring/decimal"
	"gotest.tools/assert"
)

func ExampleJSONTransformer() {
	a := map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
		"key4": "abc",
	}
	type sample struct {
		Key1 string
		B    int  `json:"key2"`
		C    bool `json:"key3"`
	}
	b := &sample{}
	err := utils.JSONTransformer(a, b)
	if err != nil {
		//error processing
	}
	fmt.Printf("%+v", b)
	// Output: &{Key1:value1 B:10 C:false}
}

func ExampleStrictJSONTransformer_success() {
	a := map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
	}

	type sample struct {
		Key1 string
		B    int  `json:"key2"`
		C    bool `json:"key3"`
	}
	b := &sample{}
	err := utils.StrictJSONTransformer(a, b)
	if err != nil {
		//err is nil
	}
	fmt.Printf("%+v", b)
	// Output: &{Key1:value1 B:10 C:false}
}

func ExampleStrictJSONTransformer_failure() {
	a := map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
		"key4": "fasf",
	}

	type sample struct {
		Key1 string
		B    int  `json:"key2"`
		C    bool `json:"key3"`
	}
	b := &sample{}
	err := utils.StrictJSONTransformer(a, b)
	if err != nil {
		fmt.Print(err)
	}
	// Output: StrictJSONTransformer: error decoding content: utils_test.sample.ReadObject: found unknown field: key4, error found in #10 byte of ...|lse,"key4":"fasf"}
	//|..., bigger context ...|{"key1":"value1","key2":10,"key3":false,"key4":"fasf"}
	//|...
}

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
	err := utils.StrictJSONTransformer(data, toData)
	assert.NilError(t, err)
	assert.Equal(t, 10, toData.IntVal)
	assert.DeepEqual(t, val, toData.DecimalVal)
	data["newField"] = "random value"
	err = utils.StrictJSONTransformer(data, toData)
	assert.Error(t, err, "StrictJsonTransformer: error decoding content: utils_test.TestVal.ReadObject: found unknown field: newField, error found in #10 byte of ...|\"newField\":\"random v|..., bigger context ...|{\"decimalVal\":\"123.1232\",\"intVal\":10,\"newField\":\"random value\"}\n|...")
}

func TestLeniantJSONTransform(t *testing.T) {
	a := map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
		"key4": "abc",
	}
	type sample struct {
		Key1 string
		B    int  `json:"key2"`
		C    bool `json:"key3"`
	}
	b := &sample{}
	err := utils.JSONTransformer(a, b)
	assert.NilError(t, err)
	assert.Equal(t, b.Key1, "value1")
	assert.Equal(t, b.B, 10)
	assert.Equal(t, b.C, false)
}

func TestStrictJSONTransform(t *testing.T) {
	a := map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
	}
	type sample struct {
		Key1 string
		B    int  `json:"key2"`
		C    bool `json:"key3"`
	}
	b := &sample{}
	err := utils.StrictJSONTransformer(a, b)
	assert.NilError(t, err)
	assert.Equal(t, b.Key1, "value1")
	assert.Equal(t, b.B, 10)
	assert.Equal(t, b.C, false)
	a = map[string]any{
		"key1": "value1",
		"key2": 10,
		"key3": false,
		"key4": "abc",
	}
	b = &sample{}
	err = utils.StrictJSONTransformer(a, b)
	assert.Error(t, err, "StrictJsonTransformer: error decoding content: utils_test.sample.ReadObject: found unknown field: key4, error found in #10 byte of ...|lse,\"key4\":\"abc\"}\n|..., bigger context ...|{\"key1\":\"value1\",\"key2\":10,\"key3\":false,\"key4\":\"abc\"}\n|...")
}
