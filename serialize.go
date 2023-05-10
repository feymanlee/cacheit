package cacheit

import (
	"encoding/json"
)

// Serializer serializer interface
type Serializer interface {
	Serialize(v any) ([]byte, error)
	UnSerialize(data []byte, v any) error
}

// JSONSerializer json serializer
type JSONSerializer struct{}

// Serialize Serialize data
func (d *JSONSerializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

// UnSerialize UnSerialize data
func (d *JSONSerializer) UnSerialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
