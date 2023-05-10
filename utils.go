//
// Package cacheit
// @Author: feymanlee@gmail.com
// @Description:
// @File:  utils
// @Date: 2023/5/9 19:50
//

package cacheit

import (
	"fmt"

	"github.com/spf13/cast"
)

func isNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
		return true
	default:
		return false
	}
}

func toAnyE[T any](a any) (T, error) {
	var t T
	switch any(t).(type) {
	case bool:
		v, err := cast.ToBoolE(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case int:
		v, err := cast.ToIntE(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case int8:
		v, err := cast.ToInt8E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case int16:
		v, err := cast.ToInt16E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case int32:
		v, err := cast.ToInt32E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case int64:
		v, err := cast.ToInt64E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case uint:
		v, err := cast.ToUintE(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case uint8:
		v, err := cast.ToUint8E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case uint16:
		v, err := cast.ToUint16E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case uint32:
		v, err := cast.ToUint32E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case uint64:
		v, err := cast.ToUint64E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case float32:
		v, err := cast.ToFloat32E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case float64:
		v, err := cast.ToFloat64E(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	case string:
		v, err := cast.ToStringE(a)
		if err != nil {
			return t, err
		}
		t = any(v).(T)
	default:
		return t, fmt.Errorf("the type %T is not supported", t)
	}
	return t, nil
}
