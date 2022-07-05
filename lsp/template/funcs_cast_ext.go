package template

import (
	"fmt"
	"github.com/spf13/cast"
)

func toFloat64(v interface{}) float64 {
	return cast.ToFloat64(v)
}

func toInt(v interface{}) int {
	return cast.ToInt(v)
}

func toInt64(v interface{}) int64 {
	return cast.ToInt64(v)
}

func strval(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
