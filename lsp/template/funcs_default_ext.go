package template

import "reflect"

// empty returns true if the given value has the zero value for its type.
func empty(given interface{}) bool {
	g := reflect.ValueOf(given)
	if !g.IsValid() {
		return true
	}

	// Basically adapted from text/template.isTrue
	switch g.Kind() {
	default:
		return g.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return g.Len() == 0
	case reflect.Bool:
		return !g.Bool()
	case reflect.Complex64, reflect.Complex128:
		return g.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return g.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return g.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return g.Float() == 0
	case reflect.Struct:
		return false
	}
}

func nonEmpty(given interface{}) bool {
	return !empty(given)
}

// coalesce returns the first non-empty value.
func coalesce(v ...interface{}) interface{} {
	for _, val := range v {
		if !empty(val) {
			return val
		}
	}
	return nil
}

// all returns true if empty(x) is false for all values x in the list.
// If the list is empty, return true.
func all(v ...interface{}) bool {
	for _, val := range v {
		if empty(val) {
			return false
		}
	}
	return true
}

// any returns true if empty(x) is false for any x in the list.
// If the list is empty, return false.
func any(v ...interface{}) bool {
	for _, val := range v {
		if !empty(val) {
			return true
		}
	}
	return false
}

// ternary returns the first value if the last value is true, otherwise returns the second value.
func ternary(vt interface{}, vf interface{}, v bool) interface{} {
	if v {
		return vt
	}

	return vf
}
