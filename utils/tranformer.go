package utils

import (
	"bytes"
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// LenientJSONTransformer copies fields from src(map/ struct object) to dest struct object
/*
Example:
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
	err := utils.LenientJSONTransformer(a, b)
	if err != nil {
		assert.NilError(t, err)
	}
	fmt.Printf("%+v", b) // Output: &{Key1:value1 B:10 C:false}
*/
func LenientJSONTransformer(src interface{}, dest interface{}) error {
	blob, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("LenientJsonTransformer: error encoding content: %w", err)
	}
	err = json.Unmarshal(blob, dest)
	if err != nil {
		return fmt.Errorf("LenientJsonTransformer: error decoding content: %w", err)
	}
	return nil
}

/*
StrictJSONTransformer copies fields from src(map/ struct object) to dest struct object throws error if there are keys in src that dosen't have slot in dest

Example 1:

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
	if err != nil { // err is nil
	    ...
	}
	fmt.Printf("%+v", b) // Output: &{Key1:value1 B:10 C:false}

Example 2:

	    a := map[string]any{
			"key1": "value1",
			"key2": 10,
			"key3": false,
	        "key4":"fasf",
		}

		type sample struct {
			Key1 string
			B    int  `json:"key2"`
			C    bool `json:"key3"`
		}
		b := &sample{}
		err := utils.StrictJSONTransformer(a, b)
		if err != nil { //err is not nil
	        fmt.Print(err) // Output: StrictJsonTransformer: error decoding content: utils_test.TestVal.ReadObject: found unknown field: newField, error found in #10 byte of ...|\"newField\":\"random v|..., bigger context ...|{\"decimalVal\":\"123.1232\",\"intVal\":10,\"newField\":\"random value\"}\n|...
	        ...
		}
*/
func StrictJSONTransformer(src interface{}, dest interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(src)
	if err != nil {
		return fmt.Errorf("StrictJsonTransformer: error encoding source: %w", err)
	}
	decoder := json.NewDecoder(&buf)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(dest)
	if err != nil {
		return fmt.Errorf("StrictJsonTransformer: error decoding content: %w", err)
	}
	return nil
}
