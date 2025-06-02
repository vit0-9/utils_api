package models

import (
	"bytes"
	"encoding/json"
)

type SafeURLString string

func (s SafeURLString) MarshalJSON() ([]byte, error) {

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(string(s))
	if err != nil {
		return nil, err
	}

	encodedBytes := buf.Bytes()

	if len(encodedBytes) > 0 && encodedBytes[len(encodedBytes)-1] == '\n' {
		return encodedBytes[:len(encodedBytes)-1], nil
	}

	return encodedBytes, nil
}

func (s *SafeURLString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = SafeURLString(str)
	return nil
}
