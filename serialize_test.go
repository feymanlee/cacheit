package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	Name string
	Age  int
}

func TestJSONSerializer(t *testing.T) {
	serializer := &JSONSerializer{}

	person := Person{
		Name: "John Doe",
		Age:  30,
	}
	t.Run("test JSON Serializer", func(t *testing.T) {
		// 测试序列化
		data, err := serializer.Serialize(person)
		assert.NoError(t, err, "failed to serialize data")

		// 测试反序列化
		var deserializedPerson Person
		err = serializer.UnSerialize(data, &deserializedPerson)
		assert.NoError(t, err, "failed to unserialize data")

		// 比较原始对象和反序列化后的对象
		assert.Equal(t, person, deserializedPerson, "the original object and deserialized object should be equal")
	})
}
