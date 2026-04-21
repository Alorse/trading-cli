package utils

import (
	"encoding/json"
	"io"
	"os"
)

// PrintJSON writes a JSON-encoded value to stdout with indentation.
func PrintJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteJSON writes a JSON-encoded value to a writer with indentation.
func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// ToJSONString returns a pretty-printed JSON string.
func ToJSONString(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
