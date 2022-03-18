package utils

import (
	"encoding/json"
	"net/http"
)



func ParseBody(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}