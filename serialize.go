package cacheit

import (
	"encoding/json"
)

type Serializer interface {
	Serialize(v any) ([]byte, error)
	UnSerialize(data []byte, v any) error
}

type JSONSerializer struct{}

//
// Serialize
//  @Description: Serialize data
//  @receiver d
//  @param v
//  @return []byte
//  @return error
//
func (d *JSONSerializer) Serialize(v any) ([]byte, error) {
	return json.Marshal(v)
}

//
// UnSerialize
//  @Description: UnSerialize data
//  @receiver d
//  @param data
//  @param v
//  @return error
//
func (d *JSONSerializer) UnSerialize(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
