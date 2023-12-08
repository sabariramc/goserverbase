package logwriter

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const ParseErrorMsg = "******************ERROR DURING MARSHAL OF FULL MESSAGE*******************"

func ParseLogObject(fullMessage any, indent bool) string {
	if fullMessage == nil {
		return ""
	}
	var msg string
	switch v := fullMessage.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	case func() string:
		msg = v()
	case []byte:
		msg = string(v)
	default:
		var blob []byte
		var err error
		if indent {
			blob, err = json.MarshalIndent(v, "", "    ")
		} else {
			blob, err = json.Marshal(v)
		}
		if err != nil {
			msg = fmt.Sprintf("%v - %v", ParseErrorMsg, err)
		} else {
			msg = string(blob)
		}
	}
	return msg
}

func GetLogObjectType(object any) string {
	msgType := "nil"
	if object != nil {
		msgType = reflect.TypeOf(object).String()
	}
	return msgType
}
