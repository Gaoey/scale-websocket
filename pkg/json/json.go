package json

import "encoding/json"

// mustJSON marshals the value to JSON or panics
func MustJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
