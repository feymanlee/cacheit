package cacheit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Message     string
	Value       int
	Time        time.Time
	IntSlice    []int
	StringSlice []string
	Map         map[string]interface{}
}

var testStructData = testStruct{
	Message:     "Hello, this is a mock message!",
	Value:       42,
	IntSlice:    []int{1, 2, 3, 4, 5},
	StringSlice: []string{"apple", "banana", "cherry"},
	Map: map[string]interface{}{
		"name":    "John Doe",
		"address": "123 Main St",
		"phone":   "555-555-5555",
		"email":   "john.doe@example.com",
	},
}

func TestJSONSerializer(t *testing.T) {
	t.Run("serialize and un serialize", func(t *testing.T) {
		serializer := &JSONSerializer{}

		data, err := serializer.Serialize(testStructData)
		assert.NoError(t, err, "Serialize should not return an error")
		var unSerializedData testStruct
		err = serializer.UnSerialize(data, &unSerializedData)
		assert.NoError(t, err, "UnSerialize should not return an error")
		assert.Equal(t, testStructData, unSerializedData, "Serialized and UnSerialized data should be equal")
	})
}
