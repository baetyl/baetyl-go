package json

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

func Marshal(v interface{}) ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, v)
}

func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.NewDecoder(reader)
}

func NewEncoder(writer io.Writer) *jsoniter.Encoder {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.NewEncoder(writer)
}
