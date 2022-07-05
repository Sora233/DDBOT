package template

import (
	"fmt"
	"github.com/tidwall/gjson"
)

func toGJson(input interface{}) gjson.Result {
	switch e := input.(type) {
	case gjson.Result:
		return e
	case string:
		return gjson.Parse(e)
	case []byte:
		return gjson.ParseBytes(e)
	default:
		panic(fmt.Sprintf("invalid input type %T", input))
	}
}
