package cacheit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsNumeric(t *testing.T) {
	assert.True(t, isNumeric(42))
	assert.True(t, isNumeric(int32(-42)))
	assert.True(t, isNumeric(uint(42)))
	assert.True(t, isNumeric(float32(3.14)))
	assert.True(t, isNumeric(complex(3, 4)))

	assert.False(t, isNumeric("abc"))
	assert.False(t, isNumeric([]int{1, 2, 3}))
}

func Test_toAnyE(t *testing.T) {
	singleTestToAnyE[bool](t, "bool", false, false, false)
	singleTestToAnyE[int](t, "int", 42, 42, false)
	singleTestToAnyE[int8](t, "int8", 5, int8(5), false)
	singleTestToAnyE[int16](t, "int16", 6, int16(6), false)
	singleTestToAnyE[int32](t, "int32", 7, int32(7), false)
	singleTestToAnyE[int64](t, "int64", 8, int64(8), false)
	singleTestToAnyE[uint](t, "uint", 4, uint(4), false)
	singleTestToAnyE[uint8](t, "uint8", 5, uint8(5), false)
	singleTestToAnyE[uint16](t, "uint16", 6, uint16(6), false)
	singleTestToAnyE[uint32](t, "uint32", 7, uint32(7), false)
	singleTestToAnyE[uint64](t, "uint64", 8, uint64(8), false)
	singleTestToAnyE[float32](t, "float32", 3.12, float32(3.12), false)
	singleTestToAnyE[float64](t, "float64", 4.24, 4.24, false)
	singleTestToAnyE[string](t, "string", "abc", "abc", false)

	singleTestToAnyE[bool](t, "bool with err", time.Time{}, false, true)
	singleTestToAnyE[int](t, "int with err", time.Time{}, 42, true)
	singleTestToAnyE[int8](t, "int8 with err", time.Time{}, int8(5), true)
	singleTestToAnyE[int16](t, "int16 with err", time.Time{}, int16(6), true)
	singleTestToAnyE[int32](t, "int32 with err", time.Time{}, int32(7), true)
	singleTestToAnyE[int64](t, "int64 with err", time.Time{}, int64(8), true)
	singleTestToAnyE[uint](t, "uint with err", time.Time{}, uint(4), true)
	singleTestToAnyE[uint8](t, "uint8 with err", time.Time{}, uint8(5), true)
	singleTestToAnyE[uint16](t, "uint16 with err", time.Time{}, uint16(6), true)
	singleTestToAnyE[uint32](t, "uint32 with err", time.Time{}, uint32(7), true)
	singleTestToAnyE[uint64](t, "uint64 with err", time.Time{}, uint64(8), true)
	singleTestToAnyE[float32](t, "float32 with err", time.Time{}, float32(3.12), true)
	singleTestToAnyE[float64](t, "float64 with err", time.Time{}, 4.24, true)
	singleTestToAnyE[string](t, "string with err", testing.T{}, "abc", true)
	singleTestToAnyE[time.Time](t, "not supported err", "abc", time.Time{}, true)
}

func singleTestToAnyE[T any](t *testing.T, name string, arg any, expected T, wantErr bool) {
	t.Run(name, func(t *testing.T) {
		got, err := toAnyE[T](arg)
		if wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equalf(t, expected, got, "toAnyE(%v)", arg)
		}
	})
}
