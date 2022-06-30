package template

import (
	"reflect"
	"strings"
)

func trunc(c int, s string) string {
	if c < 0 && len(s)+c > 0 {
		return s[len(s)+c:]
	}
	if c >= 0 && len(s) > c {
		return s[:c]
	}
	return s
}

func strslice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, strval(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, strval(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
	}
}

func join(sep string, v interface{}) string {
	return strings.Join(strslice(v), sep)
}
