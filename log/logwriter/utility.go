// Package logwriter contains the log handler implementations
package logwriter

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

const ParseErrorMsg = "******************ERROR DURING MARSHAL OF FULL MESSAGE*******************"

// ParseObject encodes the value of fullMessage into string
func ParseObject(fullMessage any, indent bool) string {
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
			msg = fmt.Sprintf("%v", v)
		} else {
			msg = string(blob)
		}
	}
	return msg
}

// ParseLogObject encodes the value of logObj into string
func ParseLogObject(logObj []any, indent bool) string {
	if len(logObj) == 1 {
		return ParseObject(logObj[0], indent)
	}
	msg := make([]string, 0, len(logObj))
	for _, val := range logObj {
		msg = append(msg, ParseObject(val, indent))
	}
	return ParseObject(msg, indent)
}

// GetLogObjectType returns the TypeOf objects in logObj as  a string delimited by comma(,)
func GetLogObjectType(logObj []any) string {
	if len(logObj) == 1 {
		return GetObjectType(logObj[0])
	}
	msg := make([]string, 0, len(logObj))
	for _, val := range logObj {
		msg = append(msg, GetObjectType(val))
	}
	return ParseObject(msg, false)
}

// GetObjectType returns the TypeOf objects in obj
func GetObjectType(obj any) string {
	msgType := "nil"
	if obj != nil {
		msgType = reflect.TypeOf(obj).String()
	}
	return msgType
}
