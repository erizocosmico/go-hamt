package hamt

import "reflect"

// equals compares the two values given and reports if they are equal.
// - If the types are not equal, the comparison returns false.
// - Is manually casted and compared if the types are of the following types:
//    string, byte, all ints, all floats, rune, bool, nil, ...
// - If the types are pointers, they are equals if they point to the same address.
// - Any other types are deeply compared using reflect.DeepEqual,
//   which means their comparisons are considerably slower.
func equals(b, a interface{}) bool {
	switch b := b.(type) {
	case string:
		a, ok := a.(string)
		return ok && a == b
	case byte:
		a, ok := a.(byte)
		return ok && a == b
	case int8:
		a, ok := a.(int8)
		return ok && a == b
	case uint16:
		a, ok := a.(uint16)
		return ok && a == b
	case int16:
		a, ok := a.(int16)
		return ok && a == b
	case uint32:
		a, ok := a.(uint32)
		return ok && a == b
	case int32:
		a, ok := a.(int32)
		return ok && a == b
	case uint64:
		a, ok := a.(uint64)
		return ok && a == b
	case int64:
		a, ok := a.(int64)
		return ok && a == b
	case uint:
		a, ok := a.(uint)
		return ok && a == b
	case int:
		a, ok := a.(int)
		return ok && a == b
	case uintptr:
		a, ok := a.(uintptr)
		return ok && a == b
	case float32:
		a, ok := a.(float32)
		return ok && a == b
	case float64:
		a, ok := a.(float64)
		return ok && a == b
	case bool:
		a, ok := a.(bool)
		return ok && a == b
	case []byte:
		a, ok := a.([]byte)
		return ok && string(a) == string(b)
	case nil:
		return a == isNil(b)
	default:
		v1 := reflect.ValueOf(a)
		v2 := reflect.ValueOf(b)

		if v1.Kind() != v2.Kind() {
			return false
		}

		switch v1.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Interface:
			return v1.Pointer() == v2.Pointer()
		}

		return reflect.DeepEqual(a, b)
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	} else {
		switch v := reflect.ValueOf(i); v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			return v.IsNil()
		}
	}
	return false
}
